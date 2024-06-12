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

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/logrusorgru/aurora"
)

var (
	args utils.Args
)

// run - main run method
func Run(ctx context.Context, versionString string) error {

	startTime := time.Now()
	defer func() {
		fmt.Printf("Elapsed time: %.2f seconds\n", time.Since(startTime).Seconds())
	}()

	//set the version build info
	args.SetVersion(versionString)

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	/* ********************************************************************** */

	//populate the AzureResourceInfo struct
	azureResourceMoveInfo := validation.AzureResourceMoveInfo{
		SourceSubscriptionId: args.SourceSubscriptionId,
		SourceResourceGroup:  args.SourceResourceGroup,
		TargetResourceGroup:  args.TargetResourceGroup,
	}

	/* ********************************************************************** */

	// Get default cred
	cred, err := func() (*azidentity.DefaultAzureCredential, error) {
		return azidentity.NewDefaultAzureCredential(nil)
	}()

	if err != nil {
		return fmt.Errorf("failed to get Azure default credential: %w", err)
	}

	/* ********************************************************************** */
	// check we are logged into the Azure source subscription
	login, err := auth.CheckLogin(ctx, cred, args.SourceSubscriptionId)

	if err != nil {
		return fmt.Errorf("getlogin error: %w", err)
	}

	if !login {
		return fmt.Errorf("you are not logged into the azure subscription '%s', please login and retry operation", azureResourceMoveInfo.SourceSubscriptionId)
	}
	fmt.Println(aurora.Sprintf(aurora.Yellow("Logged into Subscription Id: %s\n"), azureResourceMoveInfo.SourceSubscriptionId))

	/* ********************************************************************** */

	//Get the resource group client
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(cred, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return fmt.Errorf("failed to get resource group client: %w", err)
	}

	/* ********************************************************************** */

	// check source and destination resource groups exists
	srcRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, azureResourceMoveInfo.SourceResourceGroup)
	if err != nil {
		return err
	}
	if !srcRsgExists {
		return fmt.Errorf("source resource group '%s' does not exist", args.SourceResourceGroup)
	}

	/* ********************************************************************** */

	// check destination and destination resource groups exists
	dstRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return err
	}
	if !dstRsgExists {
		return fmt.Errorf("destination resource group '%s' does not exist", args.TargetResourceGroup)
	}

	/* ********************************************************************** */

	// Get resource client
	resourcesClient, err := resources.GetResourcesClient(cred, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	// Get all resource IDs from source resource group
	azureResourceMoveInfo.ResourceIds, err = resources.GetResourceIds(ctx, resourcesClient, azureResourceMoveInfo.SourceResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get resource IDs: %w", err)
	}
	/* ********************************************************************** */

	// get the target resource group ID
	azureResourceMoveInfo.TargetResourceGroupId, err = resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get target resource group ID: %w", err)
	}

	/* ********************************************************************** */

	//Validate resources - return runtime poller
	resp, err := azureResourceMoveInfo.ValidateMove(ctx, cred)
	if err != nil {
		return fmt.Errorf("failed to validate resource move: %w", err)
	}

	/* ********************************************************************** */

	// Poll the API and show a status...this is a blocking call
	pollErr := poller.PollApi(ctx, resp)
	if pollErr != nil {
		return fmt.Errorf("failed to poll API: %w", pollErr)
	}

	/* ********************************************************************** */

	return nil
}
