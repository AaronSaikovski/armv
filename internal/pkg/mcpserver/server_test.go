/*
MIT License

# Copyright (c) 2024 Aaron Saikovski
*/

package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSelectCredential(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name     string
		in       ValidateMoveInput
		wantType string // "default", "sp", "bearer", or "" when an error is expected
		wantErr  string // substring the error must contain
	}{
		{
			name:     "no auth fields -> DefaultAzureCredential (picks up az login)",
			in:       ValidateMoveInput{},
			wantType: "default",
		},
		{
			name: "all three SP fields -> ClientSecretCredential",
			in: ValidateMoveInput{
				TenantID:     validUUID,
				ClientID:     validUUID,
				ClientSecret: "secret-value",
			},
			wantType: "sp",
		},
		{
			name:     "bearer_token only -> StaticTokenCredential",
			in:       ValidateMoveInput{BearerToken: "eyJhbGciOi.fake.token"},
			wantType: "bearer",
		},
		{
			name: "bearer_token + any SP field -> error",
			in: ValidateMoveInput{
				BearerToken: "eyJhbGciOi.fake.token",
				TenantID:    validUUID,
			},
			wantErr: "bearer_token cannot be combined",
		},
		{
			name:    "only tenant_id set -> error",
			in:      ValidateMoveInput{TenantID: validUUID},
			wantErr: "require all three",
		},
		{
			name:    "tenant_id + client_id, missing secret -> error",
			in:      ValidateMoveInput{TenantID: validUUID, ClientID: validUUID},
			wantErr: "require all three",
		},
		{
			name: "invalid tenant_id -> error",
			in: ValidateMoveInput{
				TenantID:     "not-a-uuid",
				ClientID:     validUUID,
				ClientSecret: "secret",
			},
			wantErr: "invalid tenant_id",
		},
		{
			name: "invalid client_id -> error",
			in: ValidateMoveInput{
				TenantID:     validUUID,
				ClientID:     "not-a-uuid",
				ClientSecret: "secret",
			},
			wantErr: "invalid client_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := selectCredential(tt.in.TenantID, tt.in.ClientID, tt.in.ClientSecret, tt.in.BearerToken)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cred == nil {
				t.Fatal("expected a credential, got nil")
			}

			switch tt.wantType {
			case "sp":
				if _, ok := cred.(*azidentity.ClientSecretCredential); !ok {
					t.Fatalf("expected *azidentity.ClientSecretCredential, got %T", cred)
				}
			case "default":
				if _, ok := cred.(*azidentity.DefaultAzureCredential); !ok {
					t.Fatalf("expected *azidentity.DefaultAzureCredential, got %T", cred)
				}
			case "bearer":
				if _, ok := cred.(*auth.StaticTokenCredential); !ok {
					t.Fatalf("expected *auth.StaticTokenCredential, got %T", cred)
				}
			}
		})
	}
}

func TestValidateInputs(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name    string
		in      ValidateMoveInput
		wantErr string // substring; empty = no error
	}{
		{
			name: "all valid",
			in: ValidateMoveInput{
				SourceSubscriptionID: validUUID,
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: validUUID,
				TargetResourceGroup:  "rg-tgt",
			},
		},
		{
			name: "bad source subscription",
			in: ValidateMoveInput{
				SourceSubscriptionID: "not-a-uuid",
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: validUUID,
				TargetResourceGroup:  "rg-tgt",
			},
			wantErr: "invalid source_subscription_id",
		},
		{
			name: "bad target subscription",
			in: ValidateMoveInput{
				SourceSubscriptionID: validUUID,
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: "nope",
				TargetResourceGroup:  "rg-tgt",
			},
			wantErr: "invalid target_subscription_id",
		},
		{
			name: "missing source RG",
			in: ValidateMoveInput{
				SourceSubscriptionID: validUUID,
				TargetSubscriptionID: validUUID,
				TargetResourceGroup:  "rg-tgt",
			},
			wantErr: "source_resource_group is required",
		},
		{
			name: "missing target RG",
			in: ValidateMoveInput{
				SourceSubscriptionID: validUUID,
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: validUUID,
			},
			wantErr: "target_resource_group is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputs(tt.in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestNewServerRegistersValidateMove(t *testing.T) {
	// Smoke test: constructing the server must not panic, and it should at least
	// return a non-nil *mcp.Server. AddTool panics on schema errors, so reaching
	// here proves the ValidateMoveInput/Output schemas are inference-compatible.
	s := newServer("test-version")
	if s == nil {
		t.Fatal("newServer returned nil")
	}
}

// TestProgressNotifier_NoToken verifies that clients which don't opt into
// progress notifications (no progressToken in _meta) get a nil callback, which
// is how validator.Validate knows to skip all NotifyProgress calls.
func TestProgressNotifier_NoToken(t *testing.T) {
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{},
	}
	if got := progressNotifier(context.Background(), req); got != nil {
		t.Fatalf("expected nil ProgressFn when no progress token was supplied, got %T", got)
	}
}

// TestProgressNotifier_WithToken verifies that when the client supplies a
// progress token, we return a non-nil callback. We can't actually invoke the
// callback here because it calls req.Session.NotifyProgress, and Session is
// not wired up in this unit-test-scope request; the full round-trip is
// covered by the MCP SDK's own tests.
func TestProgressNotifier_WithToken(t *testing.T) {
	params := &mcp.CallToolParamsRaw{}
	params.SetProgressToken("token-xyz")

	req := &mcp.CallToolRequest{
		Params: params,
	}
	if got := progressNotifier(context.Background(), req); got == nil {
		t.Fatal("expected non-nil ProgressFn when progress token was supplied, got nil")
	}
}
