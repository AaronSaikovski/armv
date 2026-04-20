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
	// DefaultOutputPath is the default directory for output files.
	DefaultOutputPath = "./output"

	// consoleTopFailures caps how many failing resource names appear in the
	// terminal summary banner; the full list is always in the Markdown report.
	consoleTopFailures = 3
)

// Config holds the application configuration resolved from CLI flags.
type Config struct {
	Version    string
	Args       utils.Args
	OutputPath string
}

// run executes the validation workflow end-to-end.
func run(ctx context.Context, cfg *Config) error {
	if !utils.CheckValidSubscriptionID(cfg.Args.SourceSubscriptionId) {
		return fmt.Errorf("invalid source subscription ID format: expected '00000000-0000-0000-0000-000000000000'")
	}
	if !utils.CheckValidSubscriptionID(cfg.Args.TargetSubscriptionId) {
		return fmt.Errorf("invalid target subscription ID format: expected '00000000-0000-0000-0000-000000000000'")
	}

	if cfg.Args.Debug {
		startTime := time.Now()
		defer func() {
			fmt.Printf("Elapsed time: %.2f seconds\n", time.Since(startTime).Seconds())
		}()
	}

	cred, err := auth.GetAzureDefaultCredential()
	if err != nil {
		return fmt.Errorf("failed to get Azure default credential: %w", err)
	}

	azureResourceMoveInfo := validation.NewAzureResourceMoveInfo(
		cfg.Args.SourceSubscriptionId,
		cfg.Args.SourceResourceGroup,
		cfg.Args.TargetResourceGroup,
		nil,
		nil,
		cred,
	)

	if err := checkLogin(ctx, &azureResourceMoveInfo); err != nil {
		return err
	}

	if err := getResourceGroupInfo(ctx, &azureResourceMoveInfo); err != nil {
		return err
	}

	resp, err := azureResourceMoveInfo.ValidateMove(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate resource move: %w", err)
	}

	reportCtx := poller.ReportContext{
		SourceSubscriptionID: cfg.Args.SourceSubscriptionId,
		SourceResourceGroup:  cfg.Args.SourceResourceGroup,
		TargetSubscriptionID: cfg.Args.TargetSubscriptionId,
		TargetResourceGroup:  cfg.Args.TargetResourceGroup,
		ResourceCount:        len(azureResourceMoveInfo.ResourceIds),
	}

	report, err := poller.PollApi(ctx, resp, cfg.OutputPath, reportCtx)
	if err != nil {
		return fmt.Errorf("failed to poll API: %w", err)
	}

	if report.Success {
		utils.OutputSuccess(report.StatusText)
	} else {
		utils.OutputFailSummary(len(report.Errors), poller.TopFailureNames(report, consoleTopFailures))
	}

	fmt.Println(aurora.Yellow(fmt.Sprintf("\n***  Output file written to: - %s ***", cfg.OutputPath)))
	return nil
}
