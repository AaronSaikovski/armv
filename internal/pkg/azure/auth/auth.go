package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/internal/exported"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// GetAzureDefaultCredential retrieves the default Azure credential.
//
// No parameters.
// Returns a pointer to azidentity.DefaultAzureCredential and an error.

func GetAzureDefaultCredential() (*azidentity.DefaultAzureCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

// GetAzCachedAccessToken retrieves an Azure cached access token.
//
// ctx - the context in which the function is being called.
// *exported.AzCachedAccessToken - returns a pointer to an Azure cached access token.
func GetAzCachedAccessToken(ctx context.Context) *exported.AzCachedAccessToken {

	cred, err := GetAzureDefaultCredential()
	if err != nil {
		return nil
	}
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{})

	if err != nil {
		return nil
	}

	return &token

}
