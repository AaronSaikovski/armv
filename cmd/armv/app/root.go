

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

	if err := poller.PollApi(ctx, resp, cfg.OutputPath); err != nil {
		return fmt.Errorf("failed to poll API: %w", err)
	}

	fmt.Println(aurora.Yellow(fmt.Sprintf("\n\n***  Output file written to: - %s ***", cfg.OutputPath)))

	return nil
}
