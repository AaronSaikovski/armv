

package app

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/resourcegroups"
	"github.com/AaronSaikovski/armv/internal/pkg/resources"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
)

// getResourceGroupInfo populates the source/target resource group details and
// the full resource-ID list on the supplied AzureResourceMoveInfo.
func getResourceGroupInfo(ctx context.Context, azureResourceMoveInfo *validation.AzureResourceMoveInfo) error {
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(azureResourceMoveInfo.Credentials, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return fmt.Errorf("failed to get resource group client: %w", err)
	}

	srcRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, azureResourceMoveInfo.SourceResourceGroup)
	if err != nil {
		return fmt.Errorf("checking source resource group %q: %w", azureResourceMoveInfo.SourceResourceGroup, err)
	}
	if !srcRsgExists {
		return fmt.Errorf("source resource group %q does not exist", azureResourceMoveInfo.SourceResourceGroup)
	}

	dstRsgExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return fmt.Errorf("checking target resource group %q: %w", azureResourceMoveInfo.TargetResourceGroup, err)
	}
	if !dstRsgExists {
		return fmt.Errorf("destination resource group %q does not exist", azureResourceMoveInfo.TargetResourceGroup)
	}

	resourcesClient, err := resources.GetResourcesClient(azureResourceMoveInfo.Credentials, azureResourceMoveInfo.SourceSubscriptionId)
	if err != nil {
		return fmt.Errorf("failed to get resources client: %w", err)
	}

	azureResourceMoveInfo.ResourceIds, err = resources.GetResourceIds(ctx, resourcesClient, azureResourceMoveInfo.SourceResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get resource IDs: %w", err)
	}

	if len(azureResourceMoveInfo.ResourceIds) == 0 {
		return fmt.Errorf("no resources found in source resource group %q", azureResourceMoveInfo.SourceResourceGroup)
	}

	azureResourceMoveInfo.TargetResourceGroupId, err = resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, azureResourceMoveInfo.TargetResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get target resource group ID: %w", err)
	}

	return nil
}
