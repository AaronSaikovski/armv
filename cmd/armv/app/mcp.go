//go:build ignore

package app

import (
	"context"
	"fmt"

	"github.com/AaronSaikovski/armv/internal/pkg/mcpserver"
	"github.com/spf13/cobra"
)

// newMCPCommand returns the `armv mcp` parent command and its `serve` subcommand.
// `armv mcp serve` runs ARMV as a Model Context Protocol server over stdio,
// exposing validation and discovery tools to MCP clients.
//
// Intended to be launched by an MCP client (Claude Desktop, Claude Code, VS Code, etc.).
//
// DISABLED: MCP server capability has been removed.
func newMCPCommand(version string) *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
		Long:  `MCP server commands (disabled)`,
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run ARMV as an MCP server over stdio",
		Long: `Run ARMV as an MCP server over stdio.

Intended to be launched by an MCP client (Claude Desktop, Claude Code, VS Code, etc.).

DISABLED: MCP server capability has been removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "MCP server capability has been disabled")
			return nil
		},
	}

	mcpCmd.AddCommand(serveCmd)
	return mcpCmd
}
