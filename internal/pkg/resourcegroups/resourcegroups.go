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
package resourcegroups

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
)

// GetResourceGroupClient retrieves an Azure Resource Groups client.
//
// Takes in Azure credentials and a subscription ID. Returns a ResourceGroupsClient pointer and an error.

func GetResourceGroupClient(cred *azidentity.DefaultAzureCredential, subscriptionID string) (*armresources.ResourceGroupsClient, error) { //move to interface

	// Create a new Resource Groups client
	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	// Create a new Resource Groups client
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()
	if resourceGroupClient == nil {
		return nil, err
	}

	return resourceGroupClient, nil
}

// GetResourceGroup retrieves a resource group using the provided resource group name.
//
// ctx: The context within which the function is being executed.
// resourceGroupName: The name of the resource group to retrieve.
// Returns a pointer to armresources.ResourceGroup and an error.
func GetResourceGroup(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient, resourceGroupName string) (*armresources.ResourceGroup, error) {

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
func ListResourceGroup(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient) ([]*armresources.ResourceGroup, error) {

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
func CheckResourceGroupExists(ctx context.Context, resourceGroupClient *armresources.ResourceGroupsClient, resourceGroupName string) (bool, error) {

	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
	if err != nil {
		return false, err
	}
	return boolResp.Success, nil
}
