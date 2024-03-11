package resources

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/communication/armcommunication/v2"
)

var (
	resourcesClientFactory *armresources.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	resourcesClient     *armresources.Client
)
var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	virtualNetworkName = "sample-virtual-network"
)

// import (
// 	"fmt"

// 	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
// 	"context"
// 	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
// 	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
// 	"fmt"
// )

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"os"

// 	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
// 	"github.com/Azure/go-autorest/autorest/azure"
// 	"github.com/Azure/go-autorest/autorest/to"
// )

// var (
// 	subscriptionID     string
// 	location           = "westus"
// 	resourceGroupName  = "sample-resource-group"
// 	virtualNetworkName = "sample-virtual-network"
// )

var (
// resourcesClientFactory *armresources.ClientFactory
// resourceGroupClient *armresources.ResourceGroupsClient
// resourcesClient     *armresources.Client

//groupsClient resources.GroupsClient

// resourcesClient resources.Client
)

// var resourceProviderNamespace = "Microsoft.Network"
// var resourceType = "virtualNetworks"
// var apiVersion = "2021-02-01"

// func CheckResourceExists(ctx context.Context) (bool, error) {

// 	boolResp, err := resourcesClient.CheckExistence(
// 		ctx,
// 		resourceGroupName,
// 		resourceProviderNamespace,
// 		"/",
// 		resourceType,
// 		virtualNetworkName,
// 		apiVersion,
// 		nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	return boolResp.Success, nil
// }

// func GetResource(ctx context.Context) (*armresources.GenericResource, error) {

// 	resp, err := resourcesClient.Get(
// 		ctx,
// 		resourceGroupName,
// 		resourceProviderNamespace,
// 		"/",
// 		resourceType,
// 		virtualNetworkName,
// 		apiVersion,
// 		nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.GenericResource, nil
// }

// listResources lists all resources inside a resource group and prints them.
func ListResources(resourceGroupName string) {
	// fmt.Println("List resources inside the resource group")
	// resourcesList, err := groupsClient.ListResources(resourceGroupName, "", "", nil)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// if resourcesList.Value != nil && len(*resourcesList.Value) > 0 {

	// }

	// onErrorFail(err, "ListResources failed")
	// if resourcesList.Value != nil && len(*resourcesList.Value) > 0 {
	// 	fmt.Printf("Resources in '%s' resource group\n", groupName)
	// 	for _, resource := range *resourcesList.Value {
	// 		tags := "\n"
	// 		if resource.Tags == nil || len(*resource.Tags) <= 0 {
	// 			tags += "\t\t\tNo tags yet\n"
	// 		} else {
	// 			for k, v := range *resource.Tags {
	// 				tags += fmt.Sprintf("\t\t\t%s = %s\n", k, *v)
	// 			}
	// 		}
	// 		fmt.Printf("\tResource '%s'\n", *resource.Name)
	// 		elements := map[string]interface{}{
	// 			"ID":       *resource.ID,
	// 			"Location": *resource.Location,
	// 			"Type":     *resource.Type,
	// 			"Tags":     tags,
	// 		}
	// 		for k, v := range elements {
	// 			fmt.Printf("\t\t%s: %s\n", k, v)
	// 		}
	// 	}
	// } else {
	// 	fmt.Printf("There aren't any resources inside '%s' resource group\n", groupName)
	// }
}

// func onErrorFail(err error, message string) {
// 	if err != nil {
// 		fmt.Printf("%s: %s\n", message, err)
// 		//groupsClient.Delete(groupName, nil)
// 		//os.Exit(1)
// 	}
// }

// func cleanup(ctx context.Context) error {

// 	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = pollerResp.PollUntilDone(ctx, nil)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

// 	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
// 		ctx,
// 		resourceGroupName,
// 		armresources.ResourceGroup{
// 			Location: to.Ptr(location),
// 		},
// 		nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &resourceGroupResp.ResourceGroup, nil
// }

func ExampleServicesClient_NewListByResourceGroupPager() {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}
	ctx := context.Background()
	clientFactory, err := armcommunication.NewClientFactory("<subscription-id>", cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := clientFactory.NewServicesClient().NewListByResourceGroupPager("MyResourceGroup", nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			// You could use page here. We use blank identifier for just demo purposes.
			_ = v
		}
		// If the HTTP response code is 200 as defined in example definition, your page structure would look as follows. Please pay attention that all the values in the output are fake values for just demo purposes.
		// page.ServiceResourceList = armcommunication.ServiceResourceList{
		// 	Value: []*armcommunication.ServiceResource{
		// 		{
		// 			Name: to.Ptr("MyCommunicationResource"),
		// 			Type: to.Ptr("Microsoft.Communication/CommunicationServices"),
		// 			ID: to.Ptr("/subscriptions/11112222-3333-4444-5555-666677778888/resourceGroups/MyResourceGroup/providers/Microsoft.Communication/CommunicationServices/MyCommunicationResource"),
		// 			Location: to.Ptr("Global"),
		// 			Properties: &armcommunication.ServiceProperties{
		// 				DataLocation: to.Ptr("United States"),
		// 				HostName: to.Ptr("mycommunicationservice.comms.azure.net"),
		// 				ProvisioningState: to.Ptr(armcommunication.CommunicationServicesProvisioningStateSucceeded),
		// 				Version: to.Ptr("0.2.0"),
		// 			},
		// 	}},
		// }
	}
}

//func (client *Client) NewListByResourceGroupPager(resourceGroupName string, options *ClientListByResourceGroupOptions) *runtime.Pager[ClientListByResourceGroupResponse]	//func (client Client) ListByResourceGroup(ctx context.Context, resourceGroupName string, filter string, expand string, top *int32) (result ListResultPage, err error)

// func checkExistResource(ctx context.Context) (bool, error) {

// 	boolResp, err := resourcesClient.CheckExistence(
// 		ctx,
// 		resourceGroupName,
// 		resourceProviderNamespace,
// 		"/",
// 		resourceType,
// 		virtualNetworkName,
// 		apiVersion,
// 		nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	return boolResp.Success, nil
// }

// func GetResource(ctx context.Context) (*armresources.GenericResource, error) {

// 	resp, err := resourcesClient.Get(
// 		ctx,
// 		resourceGroupName,
// 		resourceProviderNamespace,
// 		"/",
// 		resourceType,
// 		virtualNetworkName,
// 		apiVersion,
// 		nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.GenericResource, nil
// }

// func GetResource(ctx context.Context) (*armresources.GenericResource, error) {

// 	resp, err := resourcesClient.Get(
// 		ctx,
// 		resourceGroupName,
// 		resourceProviderNamespace,
// 		"/",
// 		resourceType,
// 		virtualNetworkName,
// 		apiVersion,
// 		nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.GenericResource, nil
// }
