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
package validation

import (
	"context"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// moveInfoParams returns an armresources.MoveInfo struct with the given resourceIds and targetResourceGroup.
//
// Parameters:
// - resourceIds: A slice of pointers to strings representing the resource IDs.
// - targetResourceGroup: A pointer to a string representing the target resource group.
//
// Return:
// - armresources.MoveInfo: The armresources.MoveInfo struct with the given resourceIds and targetResourceGroup.
func moveInfoParams(resourceIds []*string, targetResourceGroup *string) armresources.MoveInfo {

	// return moveInfoParams
	return armresources.MoveInfo{
		Resources:           resourceIds,
		TargetResourceGroup: targetResourceGroup,
	}
}

// This operation checks whether the specified resources can be moved to the target. The resources
// to be moved must be in the same source resource group in the source subscription being used. The target
// resource group may be in a different subscription. If validation succeeds, it returns HTTP response code 204 (no content).
// If validation fails, it returns HTTP response code 409 (Conflict) with an
// error message. Retrieve the URL in the Location header value to check the result of the long-running operation.
// If the operation fails it returns an *azcore.ResponseError type.
func validateMoveResources(ctx context.Context, sourceSubscriptionID string, sourceResourceGroupName string, moveInfoParams armresources.MoveInfo) (*runtime.Poller[armresources.ClientValidateMoveResourcesResponse], error) {

	// Authorisation
	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		return nil, err
	}

	// Create a client
	client, err := auth.NewResourceClient(sourceSubscriptionID, cred)
	if err != nil {
		return nil, err
	}

	// Validate move resources
	return client.BeginValidateMoveResources(ctx, sourceResourceGroupName, moveInfoParams, nil)

}

// ValidateMove validates a move operation of specified resources to a target resource group.
//
// Parameters:
// - ctx: the context for the operation.
// - sourceSubscriptionID: the ID of the source subscription.
// - sourceResourceGroupName: the name of the source resource group.
// - resourceIds: pointers to strings representing the resource IDs.
// - targetResourceGroup: pointer to a string representing the target resource group.
// Return:
// - *runtime.Poller[armresources.ClientValidateMoveResourcesResponse]: a poller for the move operation.
// - error: an error if the operation encounters any issues.
func ValidateMove(ctx context.Context, sourceSubscriptionID string, sourceResourceGroupName string, resourceIds []*string, targetResourceGroup *string) (*runtime.Poller[armresources.ClientValidateMoveResourcesResponse], error) {

	//get move params struct
	moveParams := moveInfoParams(resourceIds, targetResourceGroup)

	//validate moce
	resp, err := validateMoveResources(ctx, sourceSubscriptionID, sourceResourceGroupName, moveParams)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
