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

// Package auth provides authentication and authorization utilities for Azure services.
// It handles credential management, subscription validation, and client creation.
package auth

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

// CheckLogin verifies the credential can reach the given subscription by
// issuing a subscription Get request against it.
func CheckLogin(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) (bool, error) {
	client, err := SubscriptionClientCred(cred)
	if err != nil {
		return false, fmt.Errorf("auth: creating subscription client: %w", err)
	}

	if _, err := client.Get(ctx, subscriptionID, nil); err != nil {
		return false, fmt.Errorf("auth: subscription %q get: %w", subscriptionID, err)
	}

	return true, nil
}

// GetAzureDefaultCredential returns a DefaultAzureCredential, which walks the
// standard Azure credential chain (env vars, managed identity, workload
// identity, az CLI, etc.).
func GetAzureDefaultCredential() (*azidentity.DefaultAzureCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("auth: default credential: %w", err)
	}
	return cred, nil
}

// NewClientSecretCredential builds a service principal credential from
// explicit tenant/client/secret values. Used by the MCP server to accept
// per-call credentials rather than relying on ambient `az login`.
func NewClientSecretCredential(tenantID, clientID, clientSecret string) (*azidentity.ClientSecretCredential, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: client secret credential: %w", err)
	}
	return cred, nil
}

// NewResourceClient creates an armresources.Client for the given subscription.
func NewResourceClient(subscriptionID string, cred azcore.TokenCredential) (*armresources.Client, error) {
	clientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: new resources client factory: %w", err)
	}
	return clientFactory.NewClient(), nil
}

// SubscriptionClientCred creates an armsubscription.SubscriptionsClient.
func SubscriptionClientCred(cred azcore.TokenCredential) (*armsubscription.SubscriptionsClient, error) {
	client, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: new subscriptions client: %w", err)
	}
	return client, nil
}

// ListSubscriptions returns every subscription the given credential can enumerate.
// Used by the MCP discovery tool so an LLM can offer the user a picklist before
// asking for a specific ID. Pagination is handled internally.
func ListSubscriptions(ctx context.Context, cred azcore.TokenCredential) ([]*armsubscription.Subscription, error) {
	client, err := SubscriptionClientCred(cred)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(nil)
	subs := make([]*armsubscription.Subscription, 0, 8)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("auth: list subscriptions page: %w", err)
		}
		subs = append(subs, page.Value...)
	}
	return subs, nil
}
