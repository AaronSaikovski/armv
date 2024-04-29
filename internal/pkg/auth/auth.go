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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// GetLogin Gets the login creds for a given user context
//
// Parameter:
// subscriptionID: a string representing the subscription ID.
//
// Return type:
// bool
func GetLogin(subscriptionID string) bool {

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

// Get the bearer token as already signed into Azure
func GetAzCachedAccessToken(ctx context.Context) (string, error) {

	cred, err := GetAzureDefaultCredential() //azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", err
	}

	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return "", err
	}

	return token.Token, nil

}
