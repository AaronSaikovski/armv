// Package validation provides Azure resource move validation functionality.
// It handles the validation of whether resources can be moved between resource groups.
package validation

import "github.com/Azure/azure-sdk-for-go/sdk/azcore"

// AzureResourceMoveInfo carries the state required to validate a move across
// subscriptions and resource groups.
type AzureResourceMoveInfo struct {
	SourceSubscriptionId  string
	SourceResourceGroup   string
	TargetResourceGroup   string
	TargetResourceGroupId *string
	ResourceIds           []*string
	Credentials           azcore.TokenCredential
}

// NewAzureResourceMoveInfo constructs an AzureResourceMoveInfo with the supplied fields.
func NewAzureResourceMoveInfo(
	sourceSubscriptionId string,
	sourceResourceGroup string,
	targetResourceGroup string,
	targetResourceGroupId *string,
	resourceIds []*string,
	credentials azcore.TokenCredential,
) AzureResourceMoveInfo {
	return AzureResourceMoveInfo{
		SourceSubscriptionId:  sourceSubscriptionId,
		SourceResourceGroup:   sourceResourceGroup,
		TargetResourceGroup:   targetResourceGroup,
		TargetResourceGroupId: targetResourceGroupId,
		ResourceIds:           resourceIds,
		Credentials:           credentials,
	}
}
