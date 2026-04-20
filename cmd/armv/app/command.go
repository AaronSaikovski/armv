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

	"github.com/AaronSaikovski/armv/pkg/utils"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the root cobra command for the armv CLI, including
// the `mcp` subcommand that runs ARMV as a Model Context Protocol server.
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

	// Required flags apply only to the root invocation; `armv mcp serve`
	// doesn't need them because the MCP client supplies parameters per call.
	for _, flagName := range []string{
		"source-subscription-id",
		"source-resource-group",
		"target-subscription-id",
		"target-resource-group",
	} {
		cobra.CheckErr(rootCmd.MarkFlagRequired(flagName))
	}

	rootCmd.AddCommand(newMCPCommand(version))

	return rootCmd
}
