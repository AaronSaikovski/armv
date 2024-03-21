package resourcegroups

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
)

// GetResourceGroup retrieves a resource group using the provided resource group name.
//
// ctx: The context within which the function is being executed.
// resourceGroupName: The name of the resource group to retrieve.
// Returns a pointer to armresources.ResourceGroup and an error.
func GetResourceGroup(ctx context.Context, resourceGroupName string) (*armresources.ResourceGroup, error) {

	resourceGroupResp, err := resourceGroupClient.Get(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

// ListResourceGroup fetches a list of resource groups.
//
// ctx - the context within which the function is executed.
// []*armresources.ResourceGroup, error - returns a slice of resource groups and an error if any.
func ListResourceGroup(ctx context.Context) ([]*armresources.ResourceGroup, error) {

	resultPager := resourceGroupClient.NewListPager(nil)

	resourceGroups := make([]*armresources.ResourceGroup, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		resourceGroups = append(resourceGroups, pageResp.ResourceGroupListResult.Value...)
	}
	return resourceGroups, nil
}

// CheckResourceGroupExists checks if a resource group exists.
//
// ctx: the context for the request.
// resourceGroupName: the name of the resource group to check.
// (bool, error): returns a boolean indicating if the resource group exists and an error if any.
func CheckResourceGroupExists(ctx context.Context, resourceGroupName string) (bool, error) {

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}
