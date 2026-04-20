/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package mcpserver exposes ARMV as a Model Context Protocol server so LLM agents
// (Claude Desktop, Claude Code, etc.) can invoke resource-move validation as a tool.
//
// The server speaks MCP over stdio; the host process must not write anything
// else to stdout or the JSON-RPC framing will break.
package mcpserver

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/validator"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverName is advertised to MCP clients. serverVersion is injected from main at startup.
const serverName = "armv"

// ValidateMoveInput is the MCP tool input contract. Service principal fields are
// optional; when all three are supplied they take precedence over the ambient
// DefaultAzureCredential chain (env vars / managed identity / az CLI).
type ValidateMoveInput struct {
	SourceSubscriptionID string `json:"source_subscription_id" jsonschema:"source Azure subscription UUID (required)"`
	SourceResourceGroup  string `json:"source_resource_group"  jsonschema:"source resource group name (required)"`
	TargetSubscriptionID string `json:"target_subscription_id" jsonschema:"target Azure subscription UUID (required)"`
	TargetResourceGroup  string `json:"target_resource_group"  jsonschema:"target resource group name (required)"`

	TenantID     string `json:"tenant_id,omitempty"     jsonschema:"optional service principal tenant UUID; supply with client_id and client_secret to bypass DefaultAzureCredential"`
	ClientID     string `json:"client_id,omitempty"     jsonschema:"optional service principal client (application) UUID"`
	ClientSecret string `json:"client_secret,omitempty" jsonschema:"optional service principal client secret"`

	BearerToken string `json:"bearer_token,omitempty" jsonschema:"optional Azure AD bearer token for https://management.azure.com (obtain via 'az account get-access-token' or similar); when set, takes precedence over all other auth fields and no credentials are stored on the server"`
}

// ValidateMoveOutput is the structured result returned to the MCP client.
type ValidateMoveOutput struct {
	Success               bool     `json:"success"                           jsonschema:"true when the Azure validate-move API returned 204 No Content"`
	ResourceIDs           []string `json:"resource_ids"                      jsonschema:"fully qualified IDs of every resource enumerated in the source resource group"`
	TargetResourceGroupID string   `json:"target_resource_group_id"          jsonschema:"fully qualified ID of the target resource group"`
	HTTPStatusCode        int      `json:"http_status_code"                  jsonschema:"HTTP status code of the final validate-move response (204 = ok, 409 = conflict)"`
	HTTPStatus            string   `json:"http_status"                       jsonschema:"HTTP status string of the final validate-move response"`
	Diagnostics           string   `json:"diagnostics,omitempty"             jsonschema:"raw response body, typically the 409 error payload when validation fails"`
}

// Run starts the MCP server on stdio and blocks until ctx is cancelled or the
// client disconnects. version is surfaced to MCP clients via the Implementation struct.
func Run(ctx context.Context, version string) error {
	server := newServer(version)
	return server.Run(ctx, &mcp.StdioTransport{})
}

// newServer constructs the MCP server with all tools registered. Split from Run
// so tests can exercise the handlers via in-memory transports.
func newServer(version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: version}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_move",
		Description: "Validate whether all resources in an Azure source resource group can be moved to a target resource group (optionally in a different subscription) without performing the move. Wraps the Azure 'validate move resources' API.",
	}, validateMoveHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_subscriptions",
		Description: "List every Azure subscription the supplied credential can see. Use this as the first step in a discovery flow before calling validate_move, so you can offer the user a picklist instead of asking them to recall subscription UUIDs.",
	}, listSubscriptionsHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resource_groups",
		Description: "List every resource group in a given subscription. Typically called after list_subscriptions and before validate_move, once the user has picked a subscription.",
	}, listResourceGroupsHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resources",
		Description: "List every Azure resource in a given resource group (name, type, location, ARM ID). Useful for inspecting what's in an RG before running validate_move, or for pinpointing which resource type is likely to block a move.",
	}, listResourcesHandler)

	return server
}

func validateMoveHandler(ctx context.Context, req *mcp.CallToolRequest, in ValidateMoveInput) (*mcp.CallToolResult, ValidateMoveOutput, error) {
	cred, err := selectCredential(in.TenantID, in.ClientID, in.ClientSecret, in.BearerToken)
	if err != nil {
		return toolError(err), ValidateMoveOutput{}, nil
	}

	if err := validateInputs(in); err != nil {
		return toolError(err), ValidateMoveOutput{}, nil
	}

	result, err := validator.Validate(ctx, validator.Input{
		SourceSubscriptionID: in.SourceSubscriptionID,
		SourceResourceGroup:  in.SourceResourceGroup,
		TargetSubscriptionID: in.TargetSubscriptionID,
		TargetResourceGroup:  in.TargetResourceGroup,
	}, cred, progressNotifier(ctx, req))
	if err != nil {
		return toolError(err), ValidateMoveOutput{}, nil
	}

	out := ValidateMoveOutput{
		Success:               result.Success,
		ResourceIDs:           result.ResourceIDs,
		TargetResourceGroupID: result.TargetResourceGroupID,
		HTTPStatusCode:        result.HTTPStatusCode,
		HTTPStatus:            result.HTTPStatus,
	}
	if !result.Success && len(result.ResponseBody) > 0 {
		out.Diagnostics = string(result.ResponseBody)
	}
	return nil, out, nil
}

// selectCredential resolves the credential to use for this call, in priority order:
//  1. bearer_token (client-supplied access token; nothing cached on the server)
//  2. tenant_id + client_id + client_secret (service principal)
//  3. DefaultAzureCredential (az login, managed identity, env vars, etc.)
//
// Mixing a bearer token with SP fields is rejected as ambiguous; partial SP input
// is rejected to surface configuration mistakes instead of silently falling back.
// Shared across all tools so every handler accepts the same auth fields.
func selectCredential(tenantID, clientID, clientSecret, bearerToken string) (azcore.TokenCredential, error) {
	spCount := 0
	if tenantID != "" {
		spCount++
	}
	if clientID != "" {
		spCount++
	}
	if clientSecret != "" {
		spCount++
	}

	if bearerToken != "" {
		if spCount > 0 {
			return nil, fmt.Errorf("bearer_token cannot be combined with tenant_id/client_id/client_secret; pick one auth method")
		}
		return auth.NewStaticTokenCredential(bearerToken), nil
	}

	switch spCount {
	case 0:
		return auth.GetAzureDefaultCredential()
	case 3:
		if !utils.CheckValidTenantID(tenantID) {
			return nil, fmt.Errorf("invalid tenant_id %q: must be a UUID", tenantID)
		}
		if !utils.CheckValidTenantID(clientID) {
			return nil, fmt.Errorf("invalid client_id %q: must be a UUID", clientID)
		}
		return auth.NewClientSecretCredential(tenantID, clientID, clientSecret)
	default:
		return nil, fmt.Errorf("service principal credentials require all three of tenant_id, client_id, and client_secret; got %d of 3", spCount)
	}
}

func validateInputs(in ValidateMoveInput) error {
	if !utils.CheckValidSubscriptionID(in.SourceSubscriptionID) {
		return fmt.Errorf("invalid source_subscription_id %q: must be a UUID", in.SourceSubscriptionID)
	}
	if !utils.CheckValidSubscriptionID(in.TargetSubscriptionID) {
		return fmt.Errorf("invalid target_subscription_id %q: must be a UUID", in.TargetSubscriptionID)
	}
	if in.SourceResourceGroup == "" {
		return fmt.Errorf("source_resource_group is required")
	}
	if in.TargetResourceGroup == "" {
		return fmt.Errorf("target_resource_group is required")
	}
	return nil
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

// progressNotifier returns a validator.ProgressFn that forwards each phase update
// to the MCP client as a standard progress notification. If the client did not
// include a ProgressToken in the initial tool call — meaning it has opted out
// of receiving progress — this returns nil, so validator.Validate becomes a no-op
// on the progress path.
//
// Each call increments a monotonic counter for Progress, per the spec's
// requirement that it "should increase every time progress is made". Total is
// omitted because validate-move has no reliable upper bound (Azure can take
// seconds to 30 minutes depending on resource count and type).
func progressNotifier(ctx context.Context, req *mcp.CallToolRequest) validator.ProgressFn {
	token := req.Params.GetProgressToken()
	if token == nil {
		return nil
	}

	var counter atomic.Int64
	return func(message string) {
		progress := float64(counter.Add(1))
		_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
			ProgressToken: token,
			Message:       message,
			Progress:      progress,
		})
	}
}
