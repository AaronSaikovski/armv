package auth

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// TestStaticTokenCredential_GetToken verifies the credential returns the exact
// token string it was constructed with and sets a future expiry, so the Azure
// SDK's bearer-token policy doesn't immediately try to refresh. If this ever
// broke, the bearer_token path in the MCP server would silently fall over on
// real Azure calls, with only a 401 to show for it.
func TestStaticTokenCredential_GetToken(t *testing.T) {
	const wantToken = "eyJhbGciOiJSUzI1NiJ9.fake.payload"

	cred := NewStaticTokenCredential(wantToken)
	if cred == nil {
		t.Fatal("NewStaticTokenCredential returned nil")
	}

	before := time.Now()
	got, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{})
	if err != nil {
		t.Fatalf("unexpected error from GetToken: %v", err)
	}

	if got.Token != wantToken {
		t.Errorf("Token = %q, want %q", got.Token, wantToken)
	}

	// Must expire in the future so the bearer policy treats it as fresh.
	if !got.ExpiresOn.After(before) {
		t.Errorf("ExpiresOn %v is not after construction time %v", got.ExpiresOn, before)
	}

	// Should be roughly an hour out (matches a freshly minted Azure management
	// API token). Give generous slack so this isn't timing-fragile.
	oneHour := before.Add(1 * time.Hour)
	if got.ExpiresOn.Before(oneHour.Add(-5 * time.Minute)) {
		t.Errorf("ExpiresOn %v is too close to now; expected ~1h from construction (near %v)", got.ExpiresOn, oneHour)
	}
	if got.ExpiresOn.After(oneHour.Add(5 * time.Minute)) {
		t.Errorf("ExpiresOn %v is too far in the future; expected ~1h from construction (near %v)", got.ExpiresOn, oneHour)
	}
}

// TestStaticTokenCredential_GetTokenIgnoresScopes confirms that the scope options
// are ignored — a static token is tied to whichever resource was used when it
// was minted, and we don't try to reinterpret it.
func TestStaticTokenCredential_GetTokenIgnoresScopes(t *testing.T) {
	cred := NewStaticTokenCredential("token-abc")

	got1, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got2, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://graph.microsoft.com/.default"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got1.Token != got2.Token {
		t.Errorf("token differs between scopes; StaticTokenCredential should ignore scopes: got1=%q got2=%q", got1.Token, got2.Token)
	}
}
