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
package app

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/mattn/go-colorable"
)

var (
	args utils.Args
)

// run - main run method
func Run() error {

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	restoreColorMode := colorable.EnableColorsStdout(nil)
	defer restoreColorMode()

	// Get default cred
	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		return err
	}

	// Create a context with cancellation capability
	//ctx := context.Background()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// resourcesClientFactory, err = armresources.NewClientFactory(args.SourceSubscriptionId, cred, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	// populate the params struct
	// inputParams = types.Params{
	// 	SourceSubscriptionId: args.SourceSubscriptionId,
	// 	SourceResourceGroup:  args.SourceResourceGroup,
	// 	TargetSubscriptionId: args.TargetSubscriptionId,
	// 	TargetResourceGroup:  args.TargetResourceGroup,
	// }

	//Print the args
	// fmt.Printf("Source Subscription Id: %s\n", args.SourceSubscriptionId)
	// fmt.Printf("Source Resource Group: %s\n", args.SourceResourceGroup)
	// fmt.Printf("Target Subscription Id: %s\n", args.TargetSubscriptionId)
	// fmt.Printf("Target Resource Group: %s\n", args.TargetResourceGroup)

	/* ********************************************************************** */
	// check we are logged into the Azure source subscription
	isLoggedIn := auth.GetLogin(args.SourceSubscriptionId)
	if !isLoggedIn {
		return fmt.Errorf("you are not logged into the azure subscription '%s', please login and retry operation", args.SourceSubscriptionId)
	} else {
		fmt.Printf("Logged into Subscription Id: %s\n", args.SourceSubscriptionId)
	}

	/* ********************************************************************** */

	//Get the resource group client
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(cred, args.SourceSubscriptionId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	//check source and destination resource groups exist
	srcRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, args.SourceResourceGroup)
	if err != nil {
		return err
	}
	if !srcRsgExists {
		return fmt.Errorf("source resource group '%s' does not exist", args.SourceResourceGroup)
	}

	/* ********************************************************************** */

	//check destination and destination resource groups exist
	dstRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, args.TargetResourceGroup)
	if err != nil {
		return err
	}
	if !dstRsgExists {
		return fmt.Errorf("destination resource group '%s' does not exist", args.TargetResourceGroup)
	}

	/* ********************************************************************** */

	// Get resource client
	resourcesClient, err := resources.GetResourcesClient(cred, args.SourceSubscriptionId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	// Get all resource IDs from source resource group
	resourceIds, err := resources.GetResourceIds(ctx, resourcesClient, args.SourceResourceGroup)
	if err != nil {
		return err
	}
	fmt.Printf("Resource Ids: %s\n", resourceIds)

	/* ********************************************************************** */

	// Get our bearer token because we're already signed into Azure
	// token, err := auth.GetAzureAccessToken(ctx)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("token: %s\n", token)

	/* ********************************************************************** */

	// Create a channel to receive results
	// resultChan := make(chan string)

	// // Create a WaitGroup to wait for all goroutines to finish
	// var wg sync.WaitGroup

	// // Increment the WaitGroup counter
	// wg.Add(1)

	// // Call the API in a goroutine
	// go api.CallValidationApi(args.SourceSubscriptionId, args.SourceResourceGroup, strings.Join(resourceIds, ""), ctx, &wg, resultChan)

	// // Wait for all goroutines to finish
	// wg.Wait()

	// // Close the result channel to signal completion
	// close(resultChan)

	// resp, err := api.CallValidationApi(args.SourceSubscriptionId, args.SourceResourceGroup, strings.Join(resourceIds, ""), ctx)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("response body: %s\n", resp)

	/* ********************************************************************** */

	//not nice but it works
	//convert a slice of strings ([]string) to a slice of string pointers ([]*string)
	var resourcePointers []*string
	for _, id := range resourceIds {
		resourcePointers = append(resourcePointers, &id)
	}

	// get the target resource group ID
	targetResourceGroupId, err := resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, args.TargetResourceGroup)
	if err != nil {
		return err
	}

	// get the move params
	moveParams := validation.MoveInfoParams(resourcePointers, targetResourceGroupId)

	resp, err := validation.ValidateMoveResources(ctx, args.SourceSubscriptionId, args.SourceResourceGroup, moveParams)
	if err != nil {
		return err
	}

	var bodyText string

	for {
		fmt.Println("Polling....") //add a status bar?
		w, err := resp.Poll(ctx)
		if err != nil {
			return err
		}

		//fmt.Printf("status: %s\n", w.Status)

		if resp.Done() {
			bodyText = w.Status
			break
		}

	}

	//204 == validation successful - no content
	//409 - with error validation failed
	fmt.Println(bodyText)

	// r, err := resp.Result(ctx)

	// fmt.Println(resp.Result(ctx))
	// fmt.Println(w)

	// w, err := resp.PollUntilDone(context.Background(), nil)

	// if err != nil {
	// 	// Handle error...
	// }

	// fmt.Println(resp.Result(ctx))
	// fmt.Println(w)

	//doesnt work!!
	//resp.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: 1 * time.Second})
	// w, err = resp.PollUntilDone(context.Background(), nil)
	// if err != nil {
	// 	// Handle error...
	// }

	// if resp.Done() {

	// 	fmt.Printf("validate response: %s\n", resp.Result)
	// }

	//poller := resp.Poller

	/* ********************************************************************** */

	fmt.Println("Done!")

	return nil
}
