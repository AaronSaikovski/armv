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
	t.Parallel()

	const (
		sourceSubID  = "12345678-1234-1234-1234-123456789012"
		sourceRG     = "source-rg"
		targetRG     = "target-rg"
		targetRGID   = "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/target-rg"
		resourceID1  = "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/source-rg/providers/Microsoft.Storage/storageAccounts/test"
		resourceID2  = "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/source-rg/providers/Microsoft.Compute/virtualMachines/vm1"
	)
	rgID := targetRGID
	rid1 := resourceID1
	rid2 := resourceID2

	tests := []struct {
		name                  string
		targetResourceGroupId *string
		resourceIds           []*string
	}{
		{name: "complete resource move info", targetResourceGroupId: &rgID, resourceIds: []*string{&rid1, &rid2}},
		{name: "without target resource group ID", targetResourceGroupId: nil, resourceIds: []*string{&rid1}},
		{name: "empty resource IDs", targetResourceGroupId: &rgID, resourceIds: []*string{}},
		{name: "nil resource IDs", targetResourceGroupId: &rgID, resourceIds: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validation.NewAzureResourceMoveInfo(
				sourceSubID,
				sourceRG,
				targetRG,
				tt.targetResourceGroupId,
				tt.resourceIds,
				nil, // credentials (azcore.TokenCredential interface; nil is allowed here)
			)

			if got.SourceSubscriptionId != sourceSubID {
				t.Errorf("SourceSubscriptionId = %q, want %q", got.SourceSubscriptionId, sourceSubID)
			}
			if got.SourceResourceGroup != sourceRG {
				t.Errorf("SourceResourceGroup = %q, want %q", got.SourceResourceGroup, sourceRG)
			}
			if got.TargetResourceGroup != targetRG {
				t.Errorf("TargetResourceGroup = %q, want %q", got.TargetResourceGroup, targetRG)
			}

			switch {
			case tt.targetResourceGroupId == nil && got.TargetResourceGroupId != nil:
				t.Error("TargetResourceGroupId should be nil")
			case tt.targetResourceGroupId != nil && got.TargetResourceGroupId == nil:
				t.Error("TargetResourceGroupId should not be nil")
			case tt.targetResourceGroupId != nil && *got.TargetResourceGroupId != *tt.targetResourceGroupId:
				t.Errorf("TargetResourceGroupId = %q, want %q", *got.TargetResourceGroupId, *tt.targetResourceGroupId)
			}

			if len(got.ResourceIds) != len(tt.resourceIds) {
				t.Errorf("len(ResourceIds) = %d, want %d", len(got.ResourceIds), len(tt.resourceIds))
			}
		})
	}
}

func TestAzureResourceMoveInfoStruct(t *testing.T) {
	t.Parallel()

	info := validation.AzureResourceMoveInfo{
		SourceSubscriptionId: "test-sub-id",
		SourceResourceGroup:  "test-source-rg",
		TargetResourceGroup:  "test-target-rg",
	}

	if info.SourceSubscriptionId != "test-sub-id" {
		t.Errorf("SourceSubscriptionId = %q", info.SourceSubscriptionId)
	}
	if info.SourceResourceGroup != "test-source-rg" {
		t.Errorf("SourceResourceGroup = %q", info.SourceResourceGroup)
	}
	if info.TargetResourceGroup != "test-target-rg" {
		t.Errorf("TargetResourceGroup = %q", info.TargetResourceGroup)
	}
}
