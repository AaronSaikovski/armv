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
