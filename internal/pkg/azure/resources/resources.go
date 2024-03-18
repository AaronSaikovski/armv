package resources

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	resourceGroupClient    *armresources.ResourceGroupsClient
	resourcesClient        *armresources.Client
	resourcesClientFactory *armresources.ClientFactory
)

// GetResources retrieves all resources for a given resource group.
//
// ctx: context.Context - The context for the request.
// resourceGroupName: string - The name of the resource group.
// *runtime.Pager - A pager for iterating over the list of resources.
func GetResources(ctx context.Context, resourceGroupName string) *runtime.Pager {

	resp := resourcesClient.NewListByResourceGroupPager(resourceGroupName, nil)
	return resp

}

// GetResourceIds description of the Go function.
//
// ctx context.Context, resourceGroupName string.
// []string, error.
func GetResourceIds(ctx context.Context, resourceGroupName string) ([]string, error) {

	var resourceIds []string
	resources := GetResources(ctx, resourceGroupName)

	for resources.More() {
		resource, err := resources.Next()
		if err != nil {
			return nil, err
		}
		resourceIds = append(resourceIds, *resource.Value.ID)
	}
	return resourceIds, nil

}
