// Package resources provides functions for managing Azure resources within resource groups,
// including listing resources and extracting resource IDs with optimized memory allocation.
package resources

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// GetResourcesClient returns an armresources.Client for the given subscription.
func GetResourcesClient(cred azcore.TokenCredential, subscriptionID string) (*armresources.Client, error) {
	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("resources: new client factory: %w", err)
	}
	return resourcesClientFactory.NewClient(), nil
}

// GetResources retrieves a list of resources in a specific resource group.
func GetResources(ctx context.Context, resourcesClient *armresources.Client, resourceGroupName string) ([]*armresources.GenericResourceExpanded, error) {
	resourcePager := resourcesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	resourceItems := make([]*armresources.GenericResourceExpanded, 0, 32)
	for resourcePager.More() {
		pageResp, err := resourcePager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("resources: list page for %q: %w", resourceGroupName, err)
		}
		resourceItems = append(resourceItems, pageResp.ResourceListResult.Value...)
	}

	return resourceItems, nil
}

// GetResourceIds returns the Azure resource IDs for every resource in the group.
func GetResourceIds(ctx context.Context, resourcesClient *armresources.Client, resourceGroupName string) ([]*string, error) {
	resourcesList, err := GetResources(ctx, resourcesClient, resourceGroupName)
	if err != nil {
		return nil, err
	}

	resourceIds := make([]*string, len(resourcesList))
	for i, val := range resourcesList {
		resourceIds[i] = val.ID
	}

	return resourceIds, nil
}
