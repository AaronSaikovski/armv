package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestIntegration_ToolsListAdvertisesAllTools wires the server to a client via
// in-memory transports and checks that tools/list returns every tool we registered.
// This is the only test that touches the MCP protocol end-to-end; it catches
// schema inference panics in AddTool and accidental tool removal from newServer.
func TestIntegration_ToolsListAdvertisesAllTools(t *testing.T) {
	ctx := t.Context()
	cs := connectInMemory(t, ctx)

	want := map[string]bool{
		"validate_move":        false,
		"list_subscriptions":   false,
		"list_resource_groups": false,
		"list_resources":       false,
	}

	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("error iterating tools: %v", err)
		}
		if _, expected := want[tool.Name]; !expected {
			t.Errorf("unexpected tool advertised: %q", tool.Name)
			continue
		}
		want[tool.Name] = true

		if tool.Description == "" {
			t.Errorf("tool %q has empty description — LLMs rely on it for selection", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil InputSchema", tool.Name)
		}
	}

	for name, seen := range want {
		if !seen {
			t.Errorf("tool %q was not advertised by tools/list", name)
		}
	}
}

// TestIntegration_ValidateMoveRejectsBadInput confirms the SDK's schema validation
// catches malformed input before our handler runs. "abc" is not a UUID, so this
// must come back as a tool-result with IsError=true, not a transport-level error.
// (If it reached our handler, validateInputs would catch it and also return
// IsError=true — either way the client sees a clear failure.)
func TestIntegration_ValidateMoveRejectsBadInput(t *testing.T) {
	ctx := t.Context()
	cs := connectInMemory(t, ctx)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "validate_move",
		Arguments: map[string]any{
			"source_subscription_id": "abc",
			"source_resource_group":  "rg-src",
			"target_subscription_id": "11111111-1111-1111-1111-111111111111",
			"target_resource_group":  "rg-tgt",
		},
	})
	if err != nil {
		t.Fatalf("CallTool returned transport-level error, wanted an IsError result: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true for malformed UUID, got IsError=false with content %+v", res.Content)
	}
}

// TestIntegration_ListResourcesRejectsMissingResourceGroup confirms the handler's
// post-schema validator (validateListResourcesInput) fires when the SDK can't
// catch the problem — resource_group is a plain string, so the schema accepts
// an empty value; our helper must reject it.
func TestIntegration_ListResourcesRejectsMissingResourceGroup(t *testing.T) {
	ctx := t.Context()
	cs := connectInMemory(t, ctx)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_resources",
		Arguments: map[string]any{
			"subscription_id": "11111111-1111-1111-1111-111111111111",
			"resource_group":  "",
		},
	})
	if err != nil {
		t.Fatalf("CallTool returned transport-level error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError=true for empty resource_group")
	}

	// Sanity-check that the error text names the failing field so the LLM can
	// act on it rather than guessing.
	var found bool
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok && strings.Contains(tc.Text, "resource_group") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error content does not mention %q; LLM won't know which field to fix. Got: %+v", "resource_group", res.Content)
	}
}

// connectInMemory stands up newServer() on one end of an in-memory transport
// pair and returns a connected ClientSession.
func connectInMemory(t *testing.T, ctx context.Context) *mcp.ClientSession {
	t.Helper()

	server := newServer("test-version")
	t1, t2 := mcp.NewInMemoryTransports()

	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server.Connect failed: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client.Connect failed: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	return cs
}
