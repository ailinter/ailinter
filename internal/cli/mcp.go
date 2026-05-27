package cli

import (
	"fmt"
	"os"

	"github.com/ailinter/ailinter/internal/mcp"
	"github.com/spf13/cobra"
)

// MCPCommand returns the `mcp` subcommand.
func MCPCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start ailinter as an MCP server (stdio)",
		Long: `Start the ailinter Model Context Protocol server on stdio.

This allows AI assistants (Cursor, Claude, Copilot) to call ailinter tools directly.
Add to your MCP configuration:

  {
    "mcpServers": {
      "ailinter": {
        "command": "ailinter",
        "args": ["mcp"]
      }
    }
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "ailinter MCP server starting on stdio...")
			return mcp.Serve(version)
		},
	}
}
