/*
MIT License

# Copyright (c) 2024 Aaron Saikovski
*/

package mcpserver

import (
	"strings"
	"testing"
)

func TestValidateListResourcesInput(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name    string
		in      ListResourcesInput
		wantErr string
	}{
		{
			name: "valid",
			in: ListResourcesInput{
				SubscriptionID: validUUID,
				ResourceGroup:  "rg-app",
			},
		},
		{
			name: "invalid subscription UUID",
			in: ListResourcesInput{
				SubscriptionID: "not-a-uuid",
				ResourceGroup:  "rg-app",
			},
			wantErr: "invalid subscription_id",
		},
		{
			name: "missing resource group",
			in: ListResourcesInput{
				SubscriptionID: validUUID,
			},
			wantErr: "resource_group is required",
		},
		{
			name: "empty input",
			in:   ListResourcesInput{},
			// Subscription check runs first, so that's the error we expect.
			wantErr: "invalid subscription_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListResourcesInput(tt.in)
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

func TestValidateListResourceGroupsInput(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name    string
		in      ListResourceGroupsInput
		wantErr string
	}{
		{
			name: "valid",
			in:   ListResourceGroupsInput{SubscriptionID: validUUID},
		},
		{
			name:    "invalid subscription UUID",
			in:      ListResourceGroupsInput{SubscriptionID: "not-a-uuid"},
			wantErr: "invalid subscription_id",
		},
		{
			name:    "empty subscription_id",
			in:      ListResourceGroupsInput{},
			wantErr: "invalid subscription_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListResourceGroupsInput(tt.in)
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

// TestDiscoveryInputsShareAuthFields is a compile-time + registration smoke test:
// it confirms that every discovery input struct passes schema inference through
// mcp.AddTool without panicking, which requires the embedded auth fields to be
// well-formed (valid json/jsonschema tags, no conflicting field names, etc.).
// newServer() exercises the same path that Run() would at startup.
func TestDiscoveryInputsShareAuthFields(t *testing.T) {
	s := newServer("test-version")
	if s == nil {
		t.Fatal("newServer returned nil — discovery tool registration likely panicked")
	}
}
