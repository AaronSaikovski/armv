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

// Package app provides the main application logic for Azure Resource Movability Validator.
// It orchestrates the validation workflow including authentication, resource group checks,
// and resource move validation operations.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AaronSaikovski/armv/cmd/armv/poller"
	"github.com/AaronSaikovski/armv/internal/pkg/auth"
	"github.com/AaronSaikovski/armv/internal/pkg/validation"
	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/logrusorgru/aurora"
)

const (
	// DefaultOutputPath is the default directory for output files
	DefaultOutputPath = "./output"
)

// Config holds the application configuration
type Config struct {
	Args       utils.Args
	OutputPath string
}

// run - main run method (package-private, called by cobra command)
func run(ctx context.Context, cfg *Config) error {

	// Validate subscription IDs
	if !utils.CheckValidSubscriptionID(cfg.Args.SourceSubscriptionId) {
		return fmt.Errorf("invalid Source Subscription ID format: should be '0000-0000-0000-000000000000'")
	}
	if !utils.CheckValidSubscriptionID(cfg.Args.TargetSubscriptionId) {
		return fmt.Errorf("invalid Target Subscription ID format: should be '0000-0000-0000-000000000000'")
	}

	//Debug
	if cfg.Args.Debug {
		startTime := time.Now()
		defer func() {
			fmt.Printf("Elapsed time: %.2f seconds\n", time.Since(startTime).Seconds())
		}()
	}

	/* ********************************************************************** */

	//populate the AzureResourceInfo struct
	azureResourceMoveInfo := validation.NewAzureResourceMoveInfo(
		cfg.Args.SourceSubscriptionId,
		cfg.Args.SourceResourceGroup,
		cfg.Args.TargetResourceGroup,
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
	if err := checkLogin(ctx, &azureResourceMoveInfo); err != nil {
		return err
	}

	/* ********************************************************************** */

	//Get the resource group info
	if err := getResourceGroupInfo(ctx, &azureResourceMoveInfo); err != nil {
		return err
	}

	/* ********************************************************************** */

	//Validate resources - return runtime poller
	resp, err := azureResourceMoveInfo.ValidateMove(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate resource move: %w", err)
	}

	/* ********************************************************************** */

	// Poll the API and show a status.
	if err := poller.PollApi(ctx, resp, cfg.OutputPath); err != nil {
		return fmt.Errorf("failed to poll API: %w", err)
	}

	/* ********************************************************************** */

	fmt.Println(aurora.Yellow(fmt.Sprintf("\n\n***  Output file written to: - %s ***", cfg.OutputPath)))

	/* ********************************************************************** */

	return nil
}
