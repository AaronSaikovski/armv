package validator

import (
	"context"
	"strings"
	"testing"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
)

// TestValidateEarlyReturnsDoNotFireProgress locks in that the input-validation
// and nil-credential guards run before any progress callbacks. This matters
// because the MCP server wires the callback to NotifyProgress; firing one for
// a call that never actually started work would confuse the client UI.
func TestValidateEarlyReturnsDoNotFireProgress(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"

	// Build a real (but harmless) credential so we can test the UUID guards in
	// isolation from the "credential is required" guard. DefaultAzureCredential
	// construction doesn't actually contact Azure until a call is made.
	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		t.Fatalf("failed to construct DefaultAzureCredential: %v", err)
	}

	tests := []struct {
		name    string
		in      Input
		cred    any // azcore.TokenCredential, or nil to test the nil-cred path
		wantErr string
	}{
		{
			name: "invalid source subscription UUID",
			in: Input{
				SourceSubscriptionID: "not-a-uuid",
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: validUUID,
				TargetResourceGroup:  "rg-tgt",
			},
			cred:    cred,
			wantErr: "invalid source subscription",
		},
		{
			name: "invalid target subscription UUID",
			in: Input{
				SourceSubscriptionID: validUUID,
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: "also-bad",
				TargetResourceGroup:  "rg-tgt",
			},
			cred:    cred,
			wantErr: "invalid target subscription",
		},
		{
			name: "nil credential",
			in: Input{
				SourceSubscriptionID: validUUID,
				SourceResourceGroup:  "rg-src",
				TargetSubscriptionID: validUUID,
				TargetResourceGroup:  "rg-tgt",
			},
			cred:    nil,
			wantErr: "credential is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var progressCalls []string
			onProgress := func(msg string) {
				progressCalls = append(progressCalls, msg)
			}

			var result *Result
			var callErr error
			if tt.cred == nil {
				result, callErr = Validate(context.Background(), tt.in, nil, onProgress)
			} else {
				cred, _ := auth.GetAzureDefaultCredential()
				result, callErr = Validate(context.Background(), tt.in, cred, onProgress)
			}

			if callErr == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(callErr.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", callErr.Error(), tt.wantErr)
			}
			if result != nil {
				t.Fatalf("expected nil result on error, got %+v", result)
			}
			if len(progressCalls) != 0 {
				t.Fatalf("expected 0 progress callbacks before validation passed, got %d: %v", len(progressCalls), progressCalls)
			}
		})
	}
}

// TestValidateAcceptsNilProgress verifies that passing a nil ProgressFn does not
// panic — the internal notify() helper must treat nil as a no-op. This is the
// path taken when an MCP client did not send a progressToken.
func TestValidateAcceptsNilProgress(t *testing.T) {
	// Trigger the earliest guard (invalid UUID) so we return before any Azure
	// call, but still exercise the nil-progress path through the notify wrapper.
	in := Input{
		SourceSubscriptionID: "not-a-uuid",
		SourceResourceGroup:  "rg-src",
		TargetSubscriptionID: "11111111-1111-1111-1111-111111111111",
		TargetResourceGroup:  "rg-tgt",
	}

	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		t.Fatalf("failed to construct DefaultAzureCredential: %v", err)
	}

	// If notify() panics on nil, this fails.
	_, callErr := Validate(context.Background(), in, cred, nil)
	if callErr == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}
