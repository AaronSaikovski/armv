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

package test

import (
	"testing"

	"github.com/AaronSaikovski/armv/internal/pkg/validation"
)

func TestNewAzureResourceMoveInfo(t *testing.T) {
	sourceSubID := "12345678-1234-1234-1234-123456789012"
	sourceRG := "source-rg"
	targetRG := "target-rg"
	targetRGID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/target-rg"
	resourceID1 := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/source-rg/providers/Microsoft.Storage/storageAccounts/test"
	resourceID2 := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/source-rg/providers/Microsoft.Compute/virtualMachines/vm1"

	tests := []struct {
		name                  string
		sourceSubscriptionId  string
		sourceResourceGroup   string
		targetResourceGroup   string
		targetResourceGroupId *string
		resourceIds           []*string
	}{
		{
			name:                  "complete resource move info",
			sourceSubscriptionId:  sourceSubID,
			sourceResourceGroup:   sourceRG,
			targetResourceGroup:   targetRG,
			targetResourceGroupId: &targetRGID,
			resourceIds:           []*string{&resourceID1, &resourceID2},
		},
		{
			name:                  "without target resource group ID",
			sourceSubscriptionId:  sourceSubID,
			sourceResourceGroup:   sourceRG,
			targetResourceGroup:   targetRG,
			targetResourceGroupId: nil,
			resourceIds:           []*string{&resourceID1},
		},
		{
			name:                  "empty resource IDs",
			sourceSubscriptionId:  sourceSubID,
			sourceResourceGroup:   sourceRG,
			targetResourceGroup:   targetRG,
			targetResourceGroupId: &targetRGID,
			resourceIds:           []*string{},
		},
		{
			name:                  "nil resource IDs",
			sourceSubscriptionId:  sourceSubID,
			sourceResourceGroup:   sourceRG,
			targetResourceGroup:   targetRG,
			targetResourceGroupId: &targetRGID,
			resourceIds:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.NewAzureResourceMoveInfo(
				tt.sourceSubscriptionId,
				tt.sourceResourceGroup,
				tt.targetResourceGroup,
				tt.targetResourceGroupId,
				tt.resourceIds,
				nil, // credentials
			)

			if got.SourceSubscriptionId != tt.sourceSubscriptionId {
				t.Errorf("SourceSubscriptionId = %v, want %v", got.SourceSubscriptionId, tt.sourceSubscriptionId)
			}
			if got.SourceResourceGroup != tt.sourceResourceGroup {
				t.Errorf("SourceResourceGroup = %v, want %v", got.SourceResourceGroup, tt.sourceResourceGroup)
			}
			if got.TargetResourceGroup != tt.targetResourceGroup {
				t.Errorf("TargetResourceGroup = %v, want %v", got.TargetResourceGroup, tt.targetResourceGroup)
			}

			// Check pointer equality for target resource group ID
			if tt.targetResourceGroupId == nil {
				if got.TargetResourceGroupId != nil {
					t.Errorf("TargetResourceGroupId should be nil")
				}
			} else {
				if got.TargetResourceGroupId == nil {
					t.Errorf("TargetResourceGroupId should not be nil")
				} else if *got.TargetResourceGroupId != *tt.targetResourceGroupId {
					t.Errorf("TargetResourceGroupId = %v, want %v", *got.TargetResourceGroupId, *tt.targetResourceGroupId)
				}
			}

			// Check resource IDs length
			if len(got.ResourceIds) != len(tt.resourceIds) {
				t.Errorf("ResourceIds length = %v, want %v", len(got.ResourceIds), len(tt.resourceIds))
			}
		})
	}
}

func TestAzureResourceMoveInfoStruct(t *testing.T) {
	// Test that the struct can be created and fields are accessible
	sourceSubID := "test-sub-id"
	sourceRG := "test-source-rg"
	targetRG := "test-target-rg"

	info := validation.AzureResourceMoveInfo{
		SourceSubscriptionId: sourceSubID,
		SourceResourceGroup:  sourceRG,
		TargetResourceGroup:  targetRG,
	}

	if info.SourceSubscriptionId != sourceSubID {
		t.Errorf("SourceSubscriptionId = %v, want %v", info.SourceSubscriptionId, sourceSubID)
	}
	if info.SourceResourceGroup != sourceRG {
		t.Errorf("SourceResourceGroup = %v, want %v", info.SourceResourceGroup, sourceRG)
	}
	if info.TargetResourceGroup != targetRG {
		t.Errorf("TargetResourceGroup = %v, want %v", info.TargetResourceGroup, targetRG)
	}
}
