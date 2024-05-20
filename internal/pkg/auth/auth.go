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

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// GetLogin checks if the user is logged into the Azure subscription with the given subscription ID.
//
// Parameters:
// - ctx: The context.Context object for controlling the request lifetime.
// - subscriptionID: The ID of the Azure subscription to check.
//
// Returns:
// - bool: True if the user is logged into the subscription, false otherwise.
func GetLogin(ctx context.Context, subscriptionID string) bool {
	cred, err := GetAzureDefaultCredential()
	if err != nil {
		return false
	}

	client, err := SubscriptionClientCred(cred)
	if err != nil {
		return false
	}

	if err := GetSubscriptionClient(ctx, client, subscriptionID); err != nil {
		return false
	}

	return true
}

// GetAzureDefaultCredential returns a new instance of the azidentity.DefaultAzureCredential.
//
// It takes no parameters.
// It returns a pointer to azidentity.DefaultAzureCredential and an error.
func GetAzureDefaultCredential() (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)

}

// NewResourceClient creates a new instance of the armresources.Client for the given Azure credential and subscription ID.
//
// Parameters:
// - subscriptionID: The ID of the subscription to create the client for.
// - cred: The Azure credential used to authenticate the client.
//
// Returns:
// - *armresources.Client: The created client instance.
// - error: An error if the client creation fails.
func NewResourceClient(subscriptionID string, cred *azidentity.DefaultAzureCredential) (*armresources.Client, error) {
	clientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	return clientFactory.NewClient(), nil
}

// SubscriptionClientCred creates a new instance of the armsubscription.SubscriptionsClient for the given azidentity.DefaultAzureCredential.
//
// Parameters:
// - cred: The azidentity.DefaultAzureCredential used to authenticate the client.
//
// Returns:
// - *armsubscription.SubscriptionsClient: The created client instance.
// - error: An error if the client creation fails.
func SubscriptionClientCred(cred *azidentity.DefaultAzureCredential) (*armsubscription.SubscriptionsClient, error) {
	return armsubscription.NewSubscriptionsClient(cred, nil)
}

// GetSubscriptionClient retrieves the subscription client for the given subscription ID.
//
// Parameters:
// - ctx: The context.Context object for controlling the request lifetime.
// - client: The armsubscription.SubscriptionsClient used to make the request.
// - subscriptionID: The ID of the subscription to retrieve the client for.
//
// Returns:
// - error: An error if the request fails, nil otherwise.
func GetSubscriptionClient(ctx context.Context, client *armsubscription.SubscriptionsClient, subscriptionID string) error {
	_, err := client.Get(ctx, subscriptionID, nil)
	return err
}
