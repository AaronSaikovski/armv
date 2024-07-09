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
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/AaronSaikovski/armv/pkg/utils"
)

var (
	args utils.Args
)

// run - main run method
func Run(ctx context.Context, versionString string) error {

	//set the version build info
	args.SetVersion(versionString)

	// check params
	if err := checkParams(); err != nil {
		return err
	}

	//Debug
	if args.Debug {
		startTime := time.Now()
		defer func() {
			fmt.Printf("Elapsed time: %.2f seconds\n", time.Since(startTime).Seconds())
		}()
	}

	/* ********************************************************************** */

	//populate the AzureResourceInfo struct
	azureResourceMoveInfo := validation.NewAzureResourceMoveInfo(
		args.SourceSubscriptionId,
		args.SourceResourceGroup,
		args.TargetResourceGroup,
		nil,
		nil,
		nil)

	/* ********************************************************************** */

	// Get default credentials
	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		return fmt.Errorf("failed to get Azure default credential: %w", err)
	}

	//assign credential
	azureResourceMoveInfo.Credentials = cred

	/* ********************************************************************** */
	// check we are logged into the Azure source subscription
	checkLogin(ctx, &azureResourceMoveInfo)

	/* ********************************************************************** */

	//Get the resource group info
	rsgErr := getResourceGroupInfo(ctx, &azureResourceMoveInfo)
	if rsgErr != nil {
		return rsgErr
	}

	/* ********************************************************************** */

	//Validate resources - return runtime poller
	resp, err := azureResourceMoveInfo.ValidateMove(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate resource move: %w", err)
	}

	/* ********************************************************************** */

	// Poll the API and show a status.
	pollErr := poller.PollApi(ctx, resp)
	if pollErr != nil {
		return fmt.Errorf("failed to poll API: %w", pollErr)
	}

	/* ********************************************************************** */

	ctx.Done()

	return nil
}
