package auth

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// AuthtoAzure returns a DefaultAzureCredential and an error.
//
// No parameters.
// Return type (*azidentity.DefaultAzureCredential, error).
func AuthtoAzure() (*azidentity.DefaultAzureCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return cred, nil
}
