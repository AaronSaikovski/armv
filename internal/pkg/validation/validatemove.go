package validation

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// ValidateMove starts a long-running Azure operation that validates whether
// the configured resources can be moved from the source to the target group.
// The returned poller must be driven to completion by the caller.
func (azureResourceMoveInfo *AzureResourceMoveInfo) ValidateMove(ctx context.Context) (*runtime.Poller[armresources.ClientValidateMoveResourcesResponse], error) {
	moveInfo := armresources.MoveInfo{
		Resources:           azureResourceMoveInfo.ResourceIds,
		TargetResourceGroup: azureResourceMoveInfo.TargetResourceGroupId,
	}

	client, err := auth.NewResourceClient(azureResourceMoveInfo.SourceSubscriptionId, azureResourceMoveInfo.Credentials)
	if err != nil {
		return nil, err
	}

	poller, err := client.BeginValidateMoveResources(ctx, azureResourceMoveInfo.SourceResourceGroup, moveInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("validation: begin validate move: %w", err)
	}
	return poller, nil
}
