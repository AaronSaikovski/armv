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

// Package resourcegroups provides functions for interacting with Azure Resource Groups,
// including client creation, resource group ID lookup, and existence checks.
package resourcegroups

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// GetResourceGroupClient creates a new ResourceGroupsClient for the given credential and subscription.
func GetResourceGroupClient(cred azcore.TokenCredential, subscriptionID string) (*armresources.ResourceGroupsClient, error) {
	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("resourcegroups: new client factory: %w", err)
	}
	return resourcesClientFactory.NewResourceGroupsClient(), nil
}

// GetResourceGroupId retrieves the Azure resource ID of a resource group by name.
func GetResourceGroupId(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient, resourceGroupName string) (*string, error) {
	resourceGroupResp, err := resourceGroupClient.Get(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, fmt.Errorf("resourcegroups: get %q: %w", resourceGroupName, err)
	}
	return resourceGroupResp.ResourceGroup.ID, nil
}

// CheckResourceGroupExists returns whether a resource group exists.
func CheckResourceGroupExists(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient, resourceGroupName string) (bool, error) {
	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, fmt.Errorf("resourcegroups: check existence of %q: %w", resourceGroupName, err)
	}
	return boolResp.Success, nil
}

// ListResourceGroup enumerates every resource group the credential can see in
// the subscription associated with the given client. Used by the MCP discovery
// tool; pagination is handled internally.
func ListResourceGroup(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient) ([]*armresources.ResourceGroup, error) {
	pager := resourceGroupClient.NewListPager(nil)

	groups := make([]*armresources.ResourceGroup, 0, 16)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("resourcegroups: list page: %w", err)
		}
		groups = append(groups, page.ResourceGroupListResult.Value...)
	}
	return groups, nil
}
