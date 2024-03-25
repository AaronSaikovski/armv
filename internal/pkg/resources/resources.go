package resources

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	resourcesClient *armresources.Client
)

// GetResourcesClient returns a new instance of the armresources.Client for the given Azure credential and subscription ID.
//
// Parameters:
// - cred: The Azure credential used to authenticate the client.
// - subscriptionID: The ID of the subscription to create the client for.
//
// Returns:
// - *armresources.Client: The created client instance.
// - error: An error if the client creation fails.
func GetResourcesClient(cred *azidentity.DefaultAzureCredential, subscriptionID string) (*armresources.Client, error) {
	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	resourcesClient = resourcesClientFactory.NewClient()

	if resourcesClient == nil {
		return nil, err
	}

	return resourcesClient, nil
}

// GetResources retrieves a list of resources in a specific resource group.
//
// ctx: the context for the request
// resourceGroupName: the name of the resource group to retrieve resources from
// []*armresources.GenericResourceExpanded: a list of expanded generic resources
// error: an error if the operation fails
func GetResources(ctx context.Context, resourcesClient *armresources.Client, resourceGroupName string) ([]*armresources.GenericResourceExpanded, error) {

	resourcePager := resourcesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	resourceItems := make([]*armresources.GenericResourceExpanded, 0)

	for resourcePager.More() {

		pageResp, err := resourcePager.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		resourceItems = append(resourceItems, pageResp.ResourceListResult.Value...)

	}

	return resourceItems, nil
}

// GetResourceIds generates resource IDs for the given resource group.
//
// ctx: the context object.
// resourceGroupName: the name of the resource group.
// []string, error: returns a slice of resource IDs and an error if any.
func GetResourceIds(ctx context.Context, resourcesClient *armresources.Client, resourceGroupName string) ([]string, error) {

	resourceIds := make([]string, 0)
	resourcesList, err := GetResources(ctx, resourcesClient, resourceGroupName)
	if err != nil {
		return nil, err
	}

	for _, val := range resourcesList {
		resourceIds = append(resourceIds, *val.ID)
	}

	return resourceIds, nil

}
