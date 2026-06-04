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

This allows AI assistants to call ailinter tools directly.
The MCP client name is auto-detected from the initialization handshake
and reported in telemetry for usage analytics.

Supported clients (auto-detected): cursor, claude, cline, copilot, windsurf,
continue, cody, goose, and more.

Add to your MCP configuration:

  {
    "mcpServers": {
      "ailinter": {
        "command": "ailinter",
        "args": ["mcp"]
      }
    }
  }

To override auto-detection, set the AILINTER_MCP_CLIENT environment variable:

  {
    "mcpServers": {
      "ailinter": {
        "command": "ailinter",
        "args": ["mcp"],
        "env": {
          "AILINTER_MCP_CLIENT": "cursor"
        }
      }
    }
  }

Valid env var values: cursor, claude, cline, copilot, windsurf`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "ailinter MCP server starting on stdio...")
			return mcp.Serve(version)
		},
	}
}
