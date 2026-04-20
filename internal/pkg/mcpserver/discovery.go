package mcpserver

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Discovery tools let an LLM agent enumerate what the user can see before
// committing to a specific validate_move call. The typical flow is:
//   list_subscriptions → list_resource_groups → list_resources → validate_move

// --- list_subscriptions ---------------------------------------------------

// ListSubscriptionsInput carries only auth fields; there are no resource
// parameters because the operation is scoped to the credential's tenant.
type ListSubscriptionsInput struct {
	TenantID     string `json:"tenant_id,omitempty"     jsonschema:"optional service principal tenant UUID"`
	ClientID     string `json:"client_id,omitempty"     jsonschema:"optional service principal client UUID"`
	ClientSecret string `json:"client_secret,omitempty" jsonschema:"optional service principal client secret"`
	BearerToken  string `json:"bearer_token,omitempty"  jsonschema:"optional Azure AD bearer token for https://management.azure.com"`
}

type SubscriptionInfo struct {
	SubscriptionID string `json:"subscription_id"          jsonschema:"Azure subscription UUID — use this value as source_subscription_id / target_subscription_id in validate_move"`
	DisplayName    string `json:"display_name,omitempty"   jsonschema:"human-readable subscription name"`
	State          string `json:"state,omitempty"          jsonschema:"subscription state (Enabled, Disabled, Warned, etc.)"`
	ID             string `json:"id,omitempty"             jsonschema:"fully qualified ARM ID (/subscriptions/{sub})"`
}

type ListSubscriptionsOutput struct {
	Subscriptions []SubscriptionInfo `json:"subscriptions" jsonschema:"subscriptions visible to the supplied credential"`
	Count         int                `json:"count"         jsonschema:"number of subscriptions returned"`
}

func listSubscriptionsHandler(ctx context.Context, _ *mcp.CallToolRequest, in ListSubscriptionsInput) (*mcp.CallToolResult, ListSubscriptionsOutput, error) {
	cred, err := selectCredential(in.TenantID, in.ClientID, in.ClientSecret, in.BearerToken)
	if err != nil {
		return toolError(err), ListSubscriptionsOutput{}, nil
	}

	subs, err := auth.ListSubscriptions(ctx, cred)
	if err != nil {
		return toolError(fmt.Errorf("failed to list subscriptions: %w", err)), ListSubscriptionsOutput{}, nil
	}

	out := ListSubscriptionsOutput{
		Subscriptions: make([]SubscriptionInfo, 0, len(subs)),
		Count:         len(subs),
	}
	for _, s := range subs {
		info := SubscriptionInfo{}
		if s.SubscriptionID != nil {
			info.SubscriptionID = *s.SubscriptionID
		}
		if s.DisplayName != nil {
			info.DisplayName = *s.DisplayName
		}
		if s.State != nil {
			info.State = string(*s.State)
		}
		if s.ID != nil {
			info.ID = *s.ID
		}
		out.Subscriptions = append(out.Subscriptions, info)
	}
	return nil, out, nil
}

// --- list_resource_groups -------------------------------------------------

type ListResourceGroupsInput struct {
	SubscriptionID string `json:"subscription_id" jsonschema:"Azure subscription UUID to enumerate resource groups in (required)"`

	TenantID     string `json:"tenant_id,omitempty"     jsonschema:"optional service principal tenant UUID"`
	ClientID     string `json:"client_id,omitempty"     jsonschema:"optional service principal client UUID"`
	ClientSecret string `json:"client_secret,omitempty" jsonschema:"optional service principal client secret"`
	BearerToken  string `json:"bearer_token,omitempty"  jsonschema:"optional Azure AD bearer token for https://management.azure.com"`
}

type ResourceGroupInfo struct {
	Name     string `json:"name"               jsonschema:"resource group name — use as source_resource_group / target_resource_group in validate_move"`
	ID       string `json:"id"                 jsonschema:"fully qualified ARM ID (/subscriptions/{sub}/resourceGroups/{rg})"`
	Location string `json:"location,omitempty" jsonschema:"Azure region (e.g. australiaeast, eastus)"`
}

type ListResourceGroupsOutput struct {
	SubscriptionID string              `json:"subscription_id" jsonschema:"subscription that was enumerated"`
	ResourceGroups []ResourceGroupInfo `json:"resource_groups" jsonschema:"resource groups found in the subscription"`
	Count          int                 `json:"count"           jsonschema:"number of resource groups returned"`
}

func listResourceGroupsHandler(ctx context.Context, _ *mcp.CallToolRequest, in ListResourceGroupsInput) (*mcp.CallToolResult, ListResourceGroupsOutput, error) {
	if err := validateListResourceGroupsInput(in); err != nil {
		return toolError(err), ListResourceGroupsOutput{}, nil
	}

	cred, err := selectCredential(in.TenantID, in.ClientID, in.ClientSecret, in.BearerToken)
	if err != nil {
		return toolError(err), ListResourceGroupsOutput{}, nil
	}

	client, err := resourcegroups.GetResourceGroupClient(cred, in.SubscriptionID)
	if err != nil {
		return toolError(fmt.Errorf("failed to build resource group client: %w", err)), ListResourceGroupsOutput{}, nil
	}

	rgs, err := resourcegroups.ListResourceGroup(ctx, client)
	if err != nil {
		return toolError(fmt.Errorf("failed to list resource groups: %w", err)), ListResourceGroupsOutput{}, nil
	}

	out := ListResourceGroupsOutput{
		SubscriptionID: in.SubscriptionID,
		ResourceGroups: make([]ResourceGroupInfo, 0, len(rgs)),
		Count:          len(rgs),
	}
	for _, rg := range rgs {
		info := ResourceGroupInfo{}
		if rg.Name != nil {
			info.Name = *rg.Name
		}
		if rg.ID != nil {
			info.ID = *rg.ID
		}
		if rg.Location != nil {
			info.Location = *rg.Location
		}
		out.ResourceGroups = append(out.ResourceGroups, info)
	}
	return nil, out, nil
}

// --- list_resources -------------------------------------------------------

type ListResourcesInput struct {
	SubscriptionID string `json:"subscription_id" jsonschema:"Azure subscription UUID containing the resource group (required)"`
	ResourceGroup  string `json:"resource_group"  jsonschema:"name of the resource group to enumerate (required)"`

	TenantID     string `json:"tenant_id,omitempty"     jsonschema:"optional service principal tenant UUID"`
	ClientID     string `json:"client_id,omitempty"     jsonschema:"optional service principal client UUID"`
	ClientSecret string `json:"client_secret,omitempty" jsonschema:"optional service principal client secret"`
	BearerToken  string `json:"bearer_token,omitempty"  jsonschema:"optional Azure AD bearer token for https://management.azure.com"`
}

type ResourceInfo struct {
	Name     string `json:"name"               jsonschema:"resource name"`
	Type     string `json:"type,omitempty"     jsonschema:"ARM resource type (e.g. Microsoft.Storage/storageAccounts)"`
	ID       string `json:"id"                 jsonschema:"fully qualified ARM resource ID — valid input for validate_move"`
	Location string `json:"location,omitempty" jsonschema:"Azure region"`
}

type ListResourcesOutput struct {
	SubscriptionID string         `json:"subscription_id" jsonschema:"subscription that was queried"`
	ResourceGroup  string         `json:"resource_group"  jsonschema:"resource group that was enumerated"`
	Resources      []ResourceInfo `json:"resources"       jsonschema:"resources in the resource group"`
	Count          int            `json:"count"           jsonschema:"number of resources returned"`
}

func listResourcesHandler(ctx context.Context, _ *mcp.CallToolRequest, in ListResourcesInput) (*mcp.CallToolResult, ListResourcesOutput, error) {
	if err := validateListResourcesInput(in); err != nil {
		return toolError(err), ListResourcesOutput{}, nil
	}

	cred, err := selectCredential(in.TenantID, in.ClientID, in.ClientSecret, in.BearerToken)
	if err != nil {
		return toolError(err), ListResourcesOutput{}, nil
	}

	client, err := resources.GetResourcesClient(cred, in.SubscriptionID)
	if err != nil {
		return toolError(fmt.Errorf("failed to build resources client: %w", err)), ListResourcesOutput{}, nil
	}

	items, err := resources.GetResources(ctx, client, in.ResourceGroup)
	if err != nil {
		return toolError(fmt.Errorf("failed to list resources: %w", err)), ListResourcesOutput{}, nil
	}

	out := ListResourcesOutput{
		SubscriptionID: in.SubscriptionID,
		ResourceGroup:  in.ResourceGroup,
		Resources:      make([]ResourceInfo, 0, len(items)),
		Count:          len(items),
	}
	for _, r := range items {
		info := ResourceInfo{}
		if r.Name != nil {
			info.Name = *r.Name
		}
		if r.Type != nil {
			info.Type = *r.Type
		}
		if r.ID != nil {
			info.ID = *r.ID
		}
		if r.Location != nil {
			info.Location = *r.Location
		}
		out.Resources = append(out.Resources, info)
	}
	return nil, out, nil
}

func validateListResourcesInput(in ListResourcesInput) error {
	if !utils.CheckValidSubscriptionID(in.SubscriptionID) {
		return fmt.Errorf("invalid subscription_id %q: must be a UUID", in.SubscriptionID)
	}
	if in.ResourceGroup == "" {
		return fmt.Errorf("resource_group is required")
	}
	return nil
}

func validateListResourceGroupsInput(in ListResourceGroupsInput) error {
	if !utils.CheckValidSubscriptionID(in.SubscriptionID) {
		return fmt.Errorf("invalid subscription_id %q: must be a UUID", in.SubscriptionID)
	}
	return nil
}
