package app

import (
	"context"

	"github.com/AaronSaikovski/armv/internal/pkg/mcpserver"
	"github.com/spf13/cobra"
)

// newMCPCommand returns the `armv mcp` parent command and its `serve` subcommand.
// `armv mcp serve` runs ARMV as a Model Context Protocol server over stdio,
// exposing a single `validate_move` tool for LLM clients like Claude Desktop.
func newMCPCommand(version string) *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server commands",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run ARMV as an MCP server over stdio",
		Long: `Run ARMV as a Model Context Protocol server on stdin/stdout.

Intended to be launched by an MCP client (Claude Desktop, Claude Code, VS Code, etc.).
The server exposes a 'validate_move' tool that accepts the same four inputs as the CLI
plus optional service principal credentials (tenant_id, client_id, client_secret).
When SP credentials are omitted the server falls back to DefaultAzureCredential.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return mcpserver.Run(ctx, version)
		},
		SilenceUsage: true, // stdio transport: Cobra usage text must not hit stdout on error
	}

	mcpCmd.AddCommand(serveCmd)
	return mcpCmd
}
