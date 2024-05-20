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
// resourcesClient: the client for interacting with Azure resources.
// resourceGroupName: the name of the resource group.
// []*string, error: returns a slice of pointers to resource IDs and an error if any.
func GetResourceIds(ctx context.Context, resourcesClient *armresources.Client, resourceGroupName string) ([]*string, error) {
	resourceIds := make([]*string, 0)
	resourcesList, err := GetResources(ctx, resourcesClient, resourceGroupName)
	if err != nil {
		return nil, err
	}

	for _, val := range resourcesList {
		// Copying pointer to the ID string
		id := *val.ID
		resourceIds = append(resourceIds, &id)
	}

	return resourceIds, nil
}
