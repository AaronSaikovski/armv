package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// DoLogin performs the login process for a given subscription ID.
//
// Parameter:
// subscriptionID: a string representing the subscription ID.
//
// Return type:
// bool
func DoLogin(subscriptionID string) bool {

	cred, err := GetAzureDefaultCredential()
	if err != nil {
		return false
	}

	client, err := SubscriptionClientCred(cred)
	if err != nil {
		return false
	}

	clientErr := GetSubscriptionClient(client, subscriptionID)
	return clientErr == nil

}

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

// SubscriptionClientCred creates a new SubscriptionsClient using the provided Azure SDK DefaultAzureCredential.
//
// Takes a pointer to a DefaultAzureCredential as a parameter. Returns a pointer to a SubscriptionsClient and an error.
func SubscriptionClientCred(cred *azidentity.DefaultAzureCredential) (*armsubscription.SubscriptionsClient, error) {
	// Azure SDK Resource Management clients accept the credential as a parameter.
	// The client will authenticate with the credential as necessary.
	client, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetSubscriptionClient retrieves a subscription client.
//
// Takes a pointer to a SubscriptionsClient and a subscriptionID string.
// Returns an error.
func GetSubscriptionClient(client *armsubscription.SubscriptionsClient, subscriptionID string) error {
	_, err := client.Get(context.TODO(), subscriptionID, nil)

	if err != nil {
		return err
	}

	return nil

}

// GetAzCachedAccessToken retrieves an Azure cached access token.
//
// ctx - the context in which the function is being called.
// // *exported.AzCachedAccessToken - returns a pointer to an Azure cached access token.
// func GetAzCachedAccessToken(ctx context.Context) *exported.AzCachedAccessToken {

// 	cred, err := GetAzureDefaultCredential()
// 	if err != nil {
// 		return nil
// 	}
// 	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{})

// 	if err != nil {
// 		return nil
// 	}

// 	return &token

// }
