package app

import (
	"context"

	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the root cobra command for the armv CLI.
// Note: MCP server subcommand has been disabled.
func NewRootCommand(version string) *cobra.Command {
	var (
		sourceSubscriptionId string
		sourceResourceGroup  string
		targetSubscriptionId string
		targetResourceGroup  string
		debug                bool
		outputPath           string
	)

	rootCmd := &cobra.Command{
		Use:   "armv",
		Short: "Azure Resource Movability Validator",
		Long:  utils.AppDescription,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			cfg := &Config{
				Version: version,
				Args: utils.Args{
					SourceSubscriptionId: sourceSubscriptionId,
					SourceResourceGroup:  sourceResourceGroup,
					TargetSubscriptionId: targetSubscriptionId,
					TargetResourceGroup:  targetResourceGroup,
					Debug:                debug,
					OutputPath:           outputPath,
				},
				OutputPath: outputPath,
			}

			return run(ctx, cfg)
		},
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.Flags().StringVar(&sourceSubscriptionId, "source-subscription-id", "", "Source Subscription Id (required)")
	rootCmd.Flags().StringVar(&sourceResourceGroup, "source-resource-group", "", "Source Resource Group (required)")
	rootCmd.Flags().StringVar(&targetSubscriptionId, "target-subscription-id", "", "Target Subscription Id (required)")
	rootCmd.Flags().StringVar(&targetResourceGroup, "target-resource-group", "", "Target Resource Group (required)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode with timing information")
	rootCmd.Flags().StringVar(&outputPath, "output-path", DefaultOutputPath, "Output path to write results")

	// Required flags apply only to the root invocation.
	// Note: MCP server subcommand has been disabled.
	for _, flagName := range []string{
		"source-subscription-id",
		"source-resource-group",
		"target-subscription-id",
		"target-resource-group",
	} {
		cobra.CheckErr(rootCmd.MarkFlagRequired(flagName))
	}

	// MCP subcommand disabled: rootCmd.AddCommand(newMCPCommand(version))

	return rootCmd
}
