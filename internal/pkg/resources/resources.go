package resources

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	resourcesClient *armresources.Client
)

// GetResources retrieves a list of resources in a specific resource group.
//
// ctx: the context for the request
// resourceGroupName: the name of the resource group to retrieve resources from
// []*armresources.GenericResourceExpanded: a list of expanded generic resources
// error: an error if the operation fails
func GetResources(ctx context.Context, resourceGroupName string) ([]*armresources.GenericResourceExpanded, error) {

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
func GetResourceIds(ctx context.Context, resourceGroupName string) ([]string, error) {

	resourceIds := make([]string, 0)
	resourcesList, err := GetResources(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	for _, val := range resourcesList { // loop in resourcesList
		resourceIds = append(resourceIds, *val.ID)
	}

	return resourceIds, nil

}
