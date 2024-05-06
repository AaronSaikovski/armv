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

	//restoreColorMode := colorable.EnableColorsStdout(nil)
	//defer restoreColorMode()

	/* ********************************************************************** */

	// Create a context with cancellation capability
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* ********************************************************************** */

	// Get default cred
	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		return err
	}

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

	//not nice but it works
	//convert a slice of strings ([]string) to a slice of string pointers ([]*string)
	var resourcePointers []*string
	for _, id := range resourceIds {
		resourcePointers = append(resourcePointers, &id)
	}

	/* ********************************************************************** */

	// get the target resource group ID
	targetResourceGroupId, err := resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, args.TargetResourceGroup)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	// get the move params
	moveParams := validation.MoveInfoParams(resourcePointers, targetResourceGroupId)

	//Validate resources
	resp, err := validation.ValidateMoveResources(ctx, args.SourceSubscriptionId, args.SourceResourceGroup, moveParams)
	if err != nil {
		return err
	}

	var bodyText string

	// Pooling loop
	for {
		fmt.Println("Polling....")
		w, err := resp.Poll(ctx)
		if err != nil {
			return err
		}

		if resp.Done() {
			bodyText = w.Status
			break
		}

	}

	//204 == validation successful - no content
	//409 - with error validation failed
	fmt.Println(bodyText)

	/* ********************************************************************** */

	fmt.Println("Done!")

	return nil
}
