package auth

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// StaticTokenCredential wraps a pre-fetched Azure AD bearer token so the SDK can
// use it without performing any login flow of its own. The caller is responsible
// for acquiring the token (e.g. `az account get-access-token` or an upstream
// OAuth exchange) and for supplying a fresh one when the previous token expires.
//
// Because the token is opaque to the server, there is no refresh path: if the
// Azure API returns 401, the error propagates back to the MCP client so the
// client can fetch a new token and retry.
type StaticTokenCredential struct {
	token     string
	expiresOn time.Time
}

// NewStaticTokenCredential wraps a bearer token for use with azcore clients.
// ExpiresOn is set one hour from now, which matches a freshly minted Azure
// Management API token; if the supplied token is already near expiry the Azure
// call will fail with 401 and the error will surface to the caller.
func NewStaticTokenCredential(token string) *StaticTokenCredential {
	return &StaticTokenCredential{
		token:     token,
		expiresOn: time.Now().Add(1 * time.Hour),
	}
}

// GetToken satisfies azcore.TokenCredential. It ignores the requested scopes
// since a static token is tied to whatever resource the caller originally
// requested when minting it (for this server, that must be https://management.azure.com).
func (c *StaticTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     c.token,
		ExpiresOn: c.expiresOn,
	}, nil
}
