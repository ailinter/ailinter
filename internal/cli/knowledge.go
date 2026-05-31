package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ailinter/ailinter/internal/knowledge"
	"github.com/spf13/cobra"
)

// KnowledgeCommand returns the parent command for knowledge graph operations.
func KnowledgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Internal knowledge graph operations for AILINTER agent team",
		Long: `The knowledge graph stores semantic relationships between code files,
packages, functions, specs, decisions, bugs, commits, and agents.
Use these commands to inspect and visualize the graph.

The graph is persisted in $WORKSPACE/.ailinter-knowledge/snapshot.json
and is rebuilt automatically when source files change.
`,
	}
	cmd.AddCommand(KnowledgeVisualizeCommand())
	return cmd
}

// KnowledgeVisualizeCommand returns the "knowledge visualize" subcommand.
func KnowledgeVisualizeCommand() *cobra.Command {
	var (
		outputPath string
		layout     string
		serverPort int
		useServer  bool
		repoRoot   string
	)

	cmd := &cobra.Command{
		Use:   "visualize",
		Short: "Generate interactive HTML visualization of the knowledge graph",
		Long: `Generate a self-contained HTML file with an interactive Cytoscape.js
graph visualization of the AILINTER knowledge graph.

The visualization includes:
  - Dark theme with 8 node types (agent, bug, commit, file, function, package, spec, test)
  - 7+ edge types with distinct colors and styles
  - Search, filter, and click-to-inspect
  - Fcose layout (compound spring embedder)

Use --server to start an ephemeral HTTP server instead of a static file.
The server auto-shuts down after 10 minutes of inactivity.

Examples:
  ailinter knowledge visualize
  ailinter knowledge visualize --layout cose
  ailinter knowledge visualize --server --port 8742
  ailinter knowledge visualize -o ~/Desktop/kgraph.html
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine workspace root
			if repoRoot == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine working directory: %w", err)
				}
				repoRoot = cwd
			}

			// Determine knowledge dir
			workspaceRoot := repoRoot
			if filepath.Base(workspaceRoot) == "ailinter" {
				workspaceRoot = filepath.Join(repoRoot, "..")
			}
			knowledgeDir := filepath.Join(workspaceRoot, ".ailinter-knowledge")

			// Ensure knowledge dir exists
			if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
				return fmt.Errorf("ensure knowledge dir: %w", err)
			}

			// Try loading graph from snapshot, or build fresh
			graph, loaded, err := tryLoadKnowledgeGraph(knowledgeDir)
			if err != nil {
				return fmt.Errorf("load knowledge graph: %w", err)
			}
			if !loaded {
				fmt.Fprintf(os.Stderr, "No graph snapshot found at %s\n", filepath.Join(knowledgeDir, "snapshot.json"))
				fmt.Fprintf(os.Stderr, "Run the MCP server with --knowledge to build the graph first,\n")
				fmt.Fprintf(os.Stderr, "or place a snapshot.json in %s\n", knowledgeDir)
				return fmt.Errorf("knowledge graph not built yet")
			}

			graph.KnowledgeDir = knowledgeDir

			if useServer {
				url, cancel, err := graph.StartVisualizationServer(serverPort)
				if err != nil {
					return fmt.Errorf("start server: %w", err)
				}
				nodes, edges, _ := graph.Stats()
				fmt.Fprintf(os.Stderr, "Knowledge graph visualization server started\n")
				fmt.Fprintf(os.Stderr, "  URL:  %s\n", url)
				fmt.Fprintf(os.Stderr, "  Nodes: %d, Edges: %d\n", nodes, edges)
				fmt.Fprintf(os.Stderr, "  Server auto-shuts down after 10 minutes of inactivity\n")
				fmt.Println(url) // Print URL to stdout for scripts
				// Wait forever — the server will close on inactivity or cancel
				<-make(chan struct{})
				cancel()
				return nil
			}

			// Generate static HTML
			opts := knowledge.VisualizeOptions{
				Layout:      layout,
				OpenBrowser: true,
			}
			if outputPath != "" {
				opts.OutputPath = outputPath
			}

			resultPath, err := graph.GenerateVisualization(opts)
			if err != nil {
				return fmt.Errorf("generate visualization: %w", err)
			}

			nodes, edges, _ := graph.Stats()
			fmt.Fprintf(os.Stderr, "Knowledge graph visualization generated:\n")
			fmt.Fprintf(os.Stderr, "  File:  %s\n", resultPath)
			fmt.Fprintf(os.Stderr, "  Nodes: %d, Edges: %d\n", nodes, edges)
			fmt.Println(resultPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output HTML file path (default: .ailinter-knowledge/visualize.html)")
	cmd.Flags().StringVar(&layout, "layout", "cose", "Layout algorithm: cose, breadthfirst, concentric, circle, grid")
	cmd.Flags().IntVar(&serverPort, "port", 8742, "Port for HTTP server mode")
	cmd.Flags().BoolVar(&useServer, "server", false, "Start HTTP server instead of writing file")
	cmd.Flags().StringVar(&repoRoot, "repo-root", "", "Repository root path (default: current directory)")

	return cmd
}

// tryLoadKnowledgeGraph attempts to load the graph from a JSON snapshot.
func tryLoadKnowledgeGraph(knowledgeDir string) (*knowledge.Graph, bool, error) {
	snapPath := filepath.Join(knowledgeDir, "snapshot.json")
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		return nil, false, nil
	}

	graph, err := knowledge.LoadGraphJSON(snapPath)
	if err != nil {
		return nil, false, fmt.Errorf("load snapshot: %w", err)
	}

	return graph, true, nil
}
