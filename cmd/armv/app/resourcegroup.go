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

	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
)

// GetResourceGroupInfo retrieves information about the source and destination resource groups.
//
// It takes the following parameters:
// - ctx: the context.Context object for controlling the execution flow.
// - azureResourceMoveInfo: a pointer to the validation.AzureResourceMoveInfo struct containing the necessary information for the move operation.
//
// It returns an error if any of the following steps fail:
// - Retrieving the resource group client.
// - Checking if the source resource group exists.
// - Checking if the destination resource group exists.
// - Retrieving the resource client.
// - Retrieving the resource IDs from the source resource group.
// - Retrieving the target resource group ID.
//
// The function populates the azureResourceMoveInfo struct with the retrieved information.
func getResourceGroupInfo(ctx context.Context, azureResourceMoveInfo *validation.AzureResourceMoveInfo) error {
	/* ********************************************************************** */

	//Get the resource group client
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(azureResourceMoveInfo.Credentials, azureResourceMoveInfo.SourceSubscriptionId)
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
		return fmt.Errorf("source resource group '%s' does not exist", azureResourceMoveInfo.SourceResourceGroup)
	}

	/* ********************************************************************** */

	// check destination and destination resource groups exists
	dstRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return err
	}
	if !dstRsgExists {
		return fmt.Errorf("destination resource group '%s' does not exist", azureResourceMoveInfo.TargetResourceGroup)
	}

	/* ********************************************************************** */

	// Get resource client
	resourcesClient, err := resources.GetResourcesClient(azureResourceMoveInfo.Credentials, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return err
	}

	/* ********************************************************************** */

	// Get all resource IDs from source resource group
	azureResourceMoveInfo.ResourceIds, err = resources.GetResourceIds(ctx, resourcesClient, azureResourceMoveInfo.SourceResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get resource IDs: %w", err)
	}

	// Validate that we have resources to move
	if len(azureResourceMoveInfo.ResourceIds) == 0 {
		return fmt.Errorf("no resources found in source resource group '%s'", azureResourceMoveInfo.SourceResourceGroup)
	}
	/* ********************************************************************** */

	// get the target resource group ID
	azureResourceMoveInfo.TargetResourceGroupId, err = resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get target resource group ID: %w", err)
	}

	return nil
}
