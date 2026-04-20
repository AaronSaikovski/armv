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
