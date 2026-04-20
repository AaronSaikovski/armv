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

// Package validator provides the library-friendly entry point for ARMV's core
// validation workflow, without any CLI presentation concerns (no stdout writes,
// no progress bar, no file output). Used by both the CLI adapter and the MCP
// server so they share one implementation of the Azure orchestration.
package validator

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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// Input collects the four parameters every validate-move invocation needs.
type Input struct {
	SourceSubscriptionID string
	SourceResourceGroup  string
	TargetSubscriptionID string
	TargetResourceGroup  string
}

// Result is the outcome of a validation, suitable for programmatic rendering.
type Result struct {
	SourceSubscriptionID  string
	SourceResourceGroup   string
	TargetSubscriptionID  string
	TargetResourceGroup   string
	TargetResourceGroupID string
	ResourceIDs           []string
	HTTPStatusCode        int
	HTTPStatus            string
	ResponseBody          []byte
	Success               bool
}

// ProgressFn is an optional hook the caller supplies to receive human-readable
// phase updates as validation runs. Used by the MCP server to forward progress
// notifications to the client. Pass nil to disable.
type ProgressFn func(message string)

// Validate runs the full resource-move validation flow without touching stdout,
// rendering a progress bar, or writing files. The caller supplies a credential
// so either DefaultAzureCredential or an explicit service principal can be used.
// If onProgress is non-nil, it is called with a short message at each phase
// (credential check, resource group checks, Azure API start, each poll tick).
func Validate(ctx context.Context, in Input, cred azcore.TokenCredential, onProgress ProgressFn) (*Result, error) {
	notify := func(msg string) {
		if onProgress != nil {
			onProgress(msg)
		}
	}

	if !utils.CheckValidSubscriptionID(in.SourceSubscriptionID) {
		return nil, fmt.Errorf("invalid source subscription ID %q: must be a UUID", in.SourceSubscriptionID)
	}
	if !utils.CheckValidSubscriptionID(in.TargetSubscriptionID) {
		return nil, fmt.Errorf("invalid target subscription ID %q: must be a UUID", in.TargetSubscriptionID)
	}
	if cred == nil {
		return nil, fmt.Errorf("credential is required")
	}

	notify("Verifying Azure credentials")
	ok, err := auth.CheckLogin(ctx, cred, in.SourceSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("login error: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("credential is not authorised for subscription %q", in.SourceSubscriptionID)
	}

	info := validation.NewAzureResourceMoveInfo(
		in.SourceSubscriptionID,
		in.SourceResourceGroup,
		in.TargetResourceGroup,
		nil,
		nil,
		cred,
	)

	notify("Enumerating resource groups and resources")
	if err := populateResourceInfo(ctx, &info); err != nil {
		return nil, err
	}

	notify(fmt.Sprintf("Starting Azure validate-move for %d resource(s)", len(info.ResourceIds)))
	respPoller, err := info.ValidateMove(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start validate move: %w", err)
	}

	respData, err := poller.PollAndCollect(ctx, respPoller, func(elapsed time.Duration) {
		notify(fmt.Sprintf("Polling Azure validate-move (elapsed %ds)", int(elapsed.Seconds())))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to poll validate-move API: %w", err)
	}

	notify(fmt.Sprintf("Validation complete (HTTP %d)", respData.RespStatusCode))

	resourceIDs := make([]string, 0, len(info.ResourceIds))
	for _, id := range info.ResourceIds {
		if id != nil {
			resourceIDs = append(resourceIDs, *id)
		}
	}
	targetRGID := ""
	if info.TargetResourceGroupId != nil {
		targetRGID = *info.TargetResourceGroupId
	}

	return &Result{
		SourceSubscriptionID:  in.SourceSubscriptionID,
		SourceResourceGroup:   in.SourceResourceGroup,
		TargetSubscriptionID:  in.TargetSubscriptionID,
		TargetResourceGroup:   in.TargetResourceGroup,
		TargetResourceGroupID: targetRGID,
		ResourceIDs:           resourceIDs,
		HTTPStatusCode:        respData.RespStatusCode,
		HTTPStatus:            respData.RespStatus,
		ResponseBody:          respData.RespBody,
		Success:               poller.ResourceMoveOK(respData.RespStatusCode),
	}, nil
}

func populateResourceInfo(ctx context.Context, info *validation.AzureResourceMoveInfo) error {
	resourceGroupClient, err := resourcegroups.GetResourceGroupClient(info.Credentials, info.SourceSubscriptionId)
	if err != nil {
		return fmt.Errorf("failed to get resource group client: %w", err)
	}

	srcExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, info.SourceResourceGroup)
	if err != nil {
		return err
	}
	if !srcExists {
		return fmt.Errorf("source resource group %q does not exist", info.SourceResourceGroup)
	}

	dstExists, err := resourcegroups.CheckResourceGroupExists(ctx, resourceGroupClient, info.TargetResourceGroup)
	if err != nil {
		return err
	}
	if !dstExists {
		return fmt.Errorf("target resource group %q does not exist", info.TargetResourceGroup)
	}

	resourcesClient, err := resources.GetResourcesClient(info.Credentials, info.SourceSubscriptionId)
	if err != nil {
		return err
	}

	info.ResourceIds, err = resources.GetResourceIds(ctx, resourcesClient, info.SourceResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get resource IDs: %w", err)
	}
	if len(info.ResourceIds) == 0 {
		return fmt.Errorf("no resources found in source resource group %q", info.SourceResourceGroup)
	}

	info.TargetResourceGroupId, err = resourcegroups.GetResourceGroupId(ctx, resourceGroupClient, info.TargetResourceGroup)
	if err != nil {
		return fmt.Errorf("failed to get target resource group ID: %w", err)
	}

	return nil
}
