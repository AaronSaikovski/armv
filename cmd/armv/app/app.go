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
	"time"

	"github.com/AaronSaikovski/armv/cmd/armv/constants"
	"github.com/AaronSaikovski/armv/cmd/armv/poller"
	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/logrusorgru/aurora"
)

var (
	args utils.Args
)

// run - main run method
func Run(versionString string) error {

	startTime := time.Now()

	//set the version build info
	args.SetVersion(versionString)

	// check params
	if err := checkParams(); err != nil {
		return err
	}

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
	if !auth.GetLogin(ctx, args.SourceSubscriptionId) {
		return fmt.Errorf("you are not logged into the azure subscription '%s', please login and retry operation", args.SourceSubscriptionId)
	}
	fmt.Println(aurora.Sprintf(aurora.Yellow("Logged into Subscription Id: %s\n"), args.SourceSubscriptionId))

	/* ********************************************************************** */

	//Get the resource group client
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(cred, args.SourceSubscriptionId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	//check source and destination resource groups exists
	srcRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, args.SourceResourceGroup)
	if err != nil {
		return err
	}
	if !srcRsgExists {
		return fmt.Errorf("source resource group '%s' does not exist", args.SourceResourceGroup)
	}

	/* ********************************************************************** */

	//check destination and destination resource groups exists
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
	//fmt.Printf("Resource Ids: %s\n", resourceIds)

	/* ********************************************************************** */

	// get the target resource group ID
	targetResourceGroupId, err := resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, args.TargetResourceGroup)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	//Validate resources
	resp, err := validation.ValidateMove(ctx, args.SourceSubscriptionId, args.SourceResourceGroup, resourceIds, targetResourceGroupId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	// Poll the API and show a status
	pollResp, err := poller.PollApi(ctx, resp)
	if err != nil {
		return err
	}

	//204 == validation successful - no content
	//409 - with error validation failed
	if pollResp.RespStatusCode == constants.API_RESOURCE_MOVE_OK {
		utils.OutputSuccess(pollResp.RespStatus)
	} else {

		resp, err := utils.PrettyJsonString(string(pollResp.RespBody))
		if err != nil {
			return err
		}
		utils.OutputFail(args.SourceResourceGroup, resp)
	}

	/* ********************************************************************** */

	fmt.Printf("Elapsed time: %.2f seconds\n", time.Since(startTime).Seconds())
	return nil
}
