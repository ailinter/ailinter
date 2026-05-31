package knowledge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// InitKnowledgeServer initializes the knowledge graph and registers all MCP tools
// on the given MCP server. Returns the graph and a cancel function for the file watcher.
//
// The knowledge graph is an internal operational tool for the 13 AILINTER agents.
// It is gated behind the --knowledge flag or the "knowledge" entry in enabled_tools.
func InitKnowledgeServer(mcpServer *server.MCPServer, repoRoot string, enableKnowledge bool) (*Graph, context.CancelFunc, error) {
	if !enableKnowledge {
		Logf("knowledge graph disabled (set enabled_tools=* or --knowledge)")
		return nil, nil, nil
	}

	Logf("initializing knowledge graph...")

	// 0. Determine workspace root for persistence
	workspaceRoot := repoRoot
	if filepath.Base(workspaceRoot) == "ailinter" {
		workspaceRoot = filepath.Join(repoRoot, "..")
	}
	knowledgeDir := filepath.Join(workspaceRoot, ".ailinter-knowledge")

	// 1. Ensure storage directory exists
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("ensure knowledge dir: %w", err)
	}

	// 2. Try loading from snapshot (pass knowledgeDir for path resolution)
	graph, loaded, err := tryLoadSnapshot(knowledgeDir)
	if err != nil {
		Logf("warning: could not load snapshot: %v, will rebuild", err)
	}

	if !loaded {
		// 3. Full rebuild
		graph = NewGraph()
		graph.KnowledgeDir = knowledgeDir
		if err := buildFullGraph(graph, repoRoot); err != nil {
			return nil, nil, fmt.Errorf("build knowledge graph: %w", err)
		}
	} else {
		// Ensure loaded graph also has the knowledge dir set
		graph.KnowledgeDir = knowledgeDir
	}

	// 4. Start file watcher for incremental updates
	ctx, cancel := context.WithCancel(context.Background())
	fw, err := NewFileWatcher(graph, repoRoot)
	if err != nil {
		Logf("warning: could not start file watcher: %v", err)
		cancel()
		return graph, nil, nil
	}

	go func() {
		if err := fw.Start(ctx); err != nil {
			Logf("file watcher error: %v", err)
		}
	}()

	// 5. Register MCP tools
	RegisterKnowledgeTools(mcpServer, graph, repoRoot)

	// Count the registered nodes
	nodes, edges, _ := graph.Stats()
	Logf("knowledge graph ready: %d nodes, %d edges", nodes, edges)

	return graph, func() {
		cancel()
		fw.Stop()
		saveSnapshotOnShutdown(graph)
	}, nil
}

// tryLoadSnapshot attempts to load the graph from a JSON snapshot in the given directory.
// Returns (graph, true, nil) on success, or (nil, false, nil) if no snapshot exists.
func tryLoadSnapshot(knowledgeDir string) (*Graph, bool, error) {
	snapPath := filepath.Join(knowledgeDir, "snapshot.json")
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		return nil, false, nil // no snapshot yet
	}

	graph, err := LoadGraphJSON(snapPath)
	if err != nil {
		return nil, false, fmt.Errorf("load snapshot: %w", err)
	}

	// Check if snapshot is fresh enough (not older than source files)
	if !graph.NeedsRebuild() {
		Logf("loaded knowledge graph from snapshot (%d nodes)", len(graph.Nodes))
		return graph, true, nil
	}

	Logf("snapshot is stale, will rebuild")
	return nil, false, nil
}

// buildFullGraph performs a complete rebuild of the knowledge graph from all sources.
// The graph must already have its KnowledgeDir set.
func buildFullGraph(graph *Graph, repoRoot string) error {
	start := time.Now()

	// Ingest Go codebase
	if err := IngestGoCodebase(graph, repoRoot); err != nil {
		return fmt.Errorf("ingest go codebase: %w", err)
	}

	// Ingest markdown specs from satellite repos
	specDirs := []string{
		filepath.Join(repoRoot, "..", "research"),
		filepath.Join(repoRoot, "..", "ops"),
	}
	if err := IngestSpecs(graph, specDirs); err != nil {
		Logf("warning: spec ingestion partial: %v", err)
	}

	// Ingest decisions from persisted decision cache
	memories := loadDecisionCache()
	if len(memories) > 0 {
		if err := IngestDecisions(graph, memories); err != nil {
			Logf("warning: decision ingestion partial: %v", err)
		}
	}

	// Ingest git history
	if err := IngestGitHistory(graph, repoRoot); err != nil {
		Logf("warning: git history ingestion partial: %v", err)
	}

	// Ingest agent ownership from opencode.json
	opencodePath := filepath.Join(repoRoot, "..", "opencode.json")
	if err := IngestAgentOwnership(graph, opencodePath); err != nil {
		Logf("warning: agent ownership ingestion partial: %v", err)
	}

	graph.LastBuilt = time.Now()

	// Persist snapshot
	if err := graph.ExportJSON(graph.snapshotPath()); err != nil {
		Logf("warning: could not persist snapshot: %v", err)
	}

	elapsed := time.Since(start)
	Logf("knowledge graph built in %v", elapsed)

	return nil
}

// saveSnapshotOnShutdown persists the graph snapshot on graceful shutdown.
func saveSnapshotOnShutdown(g *Graph) {
	Logf("shutting down: persisting knowledge graph snapshot...")

	g.mu.RLock()
	g.LastBuilt = time.Now()
	g.mu.RUnlock()

	if err := g.ExportJSON(g.snapshotPath()); err != nil {
		Logf("warning: could not save snapshot on shutdown: %v", err)
	} else {
		nodes, edges, _ := g.Stats()
		Logf("snapshot saved: %d nodes, %d edges", nodes, edges)
	}
}

// IsKnowledgeEnabled checks if the knowledge graph should be enabled based on
// the enabled_tools config or the --knowledge flag.
func IsKnowledgeEnabled(enabledTools []string) bool {
	if len(enabledTools) == 0 {
		return false
	}
	for _, t := range enabledTools {
		if t == "*" || t == "knowledge" || t == "knowledge_graph" {
			return true
		}
	}
	return false
}

// RegisterKnowledgeToolsOnServer is a convenience wrapper that creates a standalone
// MCP server with only the knowledge graph tools registered. Used for testing.
func RegisterKnowledgeToolsOnServer(g *Graph, repoRoot string, name, version string) *server.MCPServer {
	s := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(true),
	)

	RegisterKnowledgeTools(s, g, repoRoot)

	return s
}
