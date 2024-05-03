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
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// type MoveInfoParams struct {
// 	// The IDs of the resources.
// 	Resources []*string

// 	// The target resource group.
// 	TargetResourceGroup *string
// }

// GetMoveParams generates the MoveInfo object with the specified resource IDs and target resource group.
//
// Parameters:
// - resourceIds: a slice of pointers to strings representing the IDs of the resources.
// - targetResourceGroup: a pointer to a string representing the target resource group.
//
// Returns:
// - MoveInfo: the MoveInfo object with the specified resource IDs and target resource group.
func MoveInfoParams(resourceIds []*string, targetResourceGroup *string) armresources.MoveInfo {

	moveInfoParams := armresources.MoveInfo{
		Resources:           resourceIds,
		TargetResourceGroup: targetResourceGroup,
	}

	return moveInfoParams
}

// This operation checks whether the specified resources can be moved to the target. The resources
// to be moved must be in the same source resource group in the source subscription being used. The target
// resource group may be in a different subscription. If validation succeeds, it returns HTTP response code 204 (no content).
// If validation fails, it returns HTTP response code 409 (Conflict) with an
// error message. Retrieve the URL in the Location header value to check the result of the long-running operation.
// If the operation fails it returns an *azcore.ResponseError type.
func ValidateMoveResources(ctx context.Context, sourceSubscriptionID string, sourceResourceGroupName string, moveInfoParams armresources.MoveInfo) (armresources.ClientValidateMoveResourcesResponse, error) {

	// Authorisation
	cred, _ := auth.GetAzureDefaultCredential()
	// if err != nil {
	// 	return nil, err
	// }

	// Create a client
	client, _ := auth.NewResourceClient(sourceSubscriptionID, cred)
	// if err != nil {
	// 	return nil, err
	// }

	// Call Validate move api
	validateMove, _ := client.BeginValidateMoveResources(ctx, sourceResourceGroupName, moveInfoParams, nil)
	// if err != nil {
	// 	return nil, err
	// }

	//ref: https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-core-concepts
	// https://github.com/Azure/azure-sdk-for-go/blob/sdk/azcore/v1.11.1/sdk/azcore/runtime/poller.go#L316
	for {
		resp, err := validateMove.Poll(ctx)

		if err != nil {
			//return resp, err
			fmt.Println(resp)
		}

		if validateMove.Done() {
			break
		}

		// Do other work while waiting.
	}

	//azcore.ResponseError

	return validateMove.Result(ctx)

}
