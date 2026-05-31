package cli

import (
	"fmt"
	"os"

	"github.com/ailinter/ailinter/internal/mcp"
	"github.com/ailinter/ailinter/internal/telemetry"
	"github.com/spf13/cobra"
)

// MCPCommand returns the `mcp` subcommand.
func MCPCommand(version string) *cobra.Command {
	var enableKnowledge bool

	cmd := &cobra.Command{
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
  }

Use --knowledge to enable the internal knowledge graph for agent team use.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if c := os.Getenv("AILINTER_MCP_CLIENT"); c != "" {
				telemetry.SetMCPClient(c)
			}
			fmt.Fprintln(os.Stderr, "ailinter MCP server starting on stdio...")
			if enableKnowledge {
				fmt.Fprintln(os.Stderr, "knowledge graph enabled (internal agent tools)")
			}
			return mcp.Serve(version, enableKnowledge)
		},
	}

	cmd.Flags().BoolVar(&enableKnowledge, "knowledge", false,
		"Enable the internal knowledge graph for AILINTER agent team (MCP tools: knowledge_blast_radius, etc.)")

	return cmd
}
