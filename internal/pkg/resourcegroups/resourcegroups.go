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
