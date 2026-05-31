package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool names for the knowledge graph MCP tools.
const (
	ToolBlastRadius   = "knowledge_blast_radius"
	ToolTraceDecision = "knowledge_trace_decision"
	ToolSpecToCode    = "knowledge_spec_to_code"
	ToolCausalChain   = "knowledge_causal_chain"
	ToolGraphQuery    = "knowledge_graph_query"
	ToolReindex       = "knowledge_reindex"
	ToolStats         = "knowledge_stats"
	ToolVisualize     = "knowledge_visualize"
)

// RegisterKnowledgeTools registers all knowledge graph tools on an MCP server.
func RegisterKnowledgeTools(s *server.MCPServer, g *Graph, repoRoot string) {
	// Tool 1: knowledge_blast_radius
	s.AddTool(mcp.NewTool(
		ToolBlastRadius,
		mcp.WithDescription("Find all files, packages, specs, and tests affected by a change to a given node. Returns nodes organized by hop distance from the start node."),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("The starting node ID (e.g., 'file:internal/mcp/server.go', 'pkg:analyzer', 'func:main.Run')"),
		),
		mcp.WithNumber("max_hops",
			mcp.Description("Maximum traversal depth (default: 3)"),
		),
		mcp.WithString("direction",
			mcp.Description("Traversal direction: 'outbound', 'inbound', or 'both' (default: 'outbound')"),
		),
		mcp.WithString("edge_types",
			mcp.Description("Comma-separated edge types to filter by (e.g., 'CALLS,IMPLEMENTS'). Empty means all."),
		),
	), newBlastRadiusHandler(g))

	// Tool 2: knowledge_trace_decision
	s.AddTool(mcp.NewTool(
		ToolTraceDecision,
		mcp.WithDescription("Trace a decision through the knowledge graph. Find how a decision affects specs, files, and functions."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query: file path, package name, decision ID, or free text"),
		),
	), newTraceDecisionHandler(g))

	// Tool 3: knowledge_spec_to_code
	s.AddTool(mcp.NewTool(
		ToolSpecToCode,
		mcp.WithDescription("Map a specification section to the code files and functions that implement it."),
		mcp.WithString("spec_query",
			mcp.Required(),
			mcp.Description("Spec title or path to search for"),
		),
		mcp.WithString("detail_level",
			mcp.Description("'summary' or 'full' (default: 'summary')"),
		),
	), newSpecToCodeHandler(g))

	// Tool 4: knowledge_causal_chain
	s.AddTool(mcp.NewTool(
		ToolCausalChain,
		mcp.WithDescription("Show the causal chain for a package: recent commits, bugs, and incidents affecting it."),
		mcp.WithString("package_name",
			mcp.Required(),
			mcp.Description("Package name to investigate (e.g., 'analyzer', 'mcp', 'config')"),
		),
		mcp.WithNumber("time_range_days",
			mcp.Description("Number of days of history to include (default: 90)"),
		),
	), newCausalChainHandler(g))

	// Tool 5: knowledge_graph_query
	s.AddTool(mcp.NewTool(
		ToolGraphQuery,
		mcp.WithDescription("Query the knowledge graph with flexible traversal parameters. Returns all reachable nodes with their paths."),
		mcp.WithString("start_node_id",
			mcp.Required(),
			mcp.Description("The starting node ID"),
		),
		mcp.WithString("edge_types",
			mcp.Description("Comma-separated edge types to follow (e.g., 'CALLS,CONTAINS'). Empty means all."),
		),
		mcp.WithString("direction",
			mcp.Description("'outbound', 'inbound', or 'both' (default: 'outbound')"),
		),
		mcp.WithNumber("max_hops",
			mcp.Description("Maximum traversal depth (default: 5)"),
		),
	), newGraphQueryHandler(g))

	// Tool 6 (bonus): knowledge_reindex
	s.AddTool(mcp.NewTool(
		ToolReindex,
		mcp.WithDescription("Trigger a full rebuild of the knowledge graph from all sources."),
		mcp.WithString("repo_root",
			mcp.Required(),
			mcp.Description("Absolute path to the repository root"),
		),
	), newReindexHandler(g, repoRoot))

	// Tool 7 (bonus): knowledge_stats
	s.AddTool(mcp.NewTool(
		ToolStats,
		mcp.WithDescription("Return knowledge graph statistics: node/edge counts, last build time, source file freshness."),
	), newStatsHandler(g))

	// Tool 8: knowledge_visualize
	s.AddTool(mcp.NewTool(
		ToolVisualize,
		mcp.WithDescription("Generate an interactive HTML visualization of the knowledge graph. Starts an ephemeral HTTP server and returns a URL to open in a browser. The server auto-shuts down after 10 minutes."),
		mcp.WithString("layout",
			mcp.Description("Layout algorithm: 'cose', 'breadthfirst', 'concentric' (default: 'cose')"),
		),
	), newVisualizeHandler(g))
}

// --- Handler factories ---

func newBlastRadiusHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		nodeID, _ := args["node_id"].(string)
		if nodeID == "" {
			return mcp.NewToolResultError("node_id is required"), nil
		}

		maxHops := 3
		if v, ok := args["max_hops"].(float64); ok && v > 0 {
			maxHops = int(v)
		}

		direction := "outbound"
		if d, ok := args["direction"].(string); ok && d != "" {
			direction = d
		}

		var edgeTypes []EdgeType
		if et, ok := args["edge_types"].(string); ok && et != "" {
			for _, part := range strings.Split(et, ",") {
				edgeTypes = append(edgeTypes, EdgeType(strings.TrimSpace(part)))
			}
		}

		layers, err := g.BlastRadius(nodeID, maxHops, edgeTypes, direction)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("blast radius error: %v", err)), nil
		}

		result := blastRadiusResult{
			StartNodeID: nodeID,
			MaxHops:     maxHops,
			Direction:   direction,
			TotalNodes:  countLayers(layers),
			Layers:      make([]layerResult, len(layers)),
		}

		for i, layer := range layers {
			layerNodes := make([]nodeSummary, len(layer))
			for j, n := range layer {
				layerNodes[j] = nodeSummary{
					ID:    n.ID,
					Type:  string(n.Type),
					Label: n.Label,
				}
			}
			result.Layers[i] = layerResult{
				Hop:   i,
				Nodes: layerNodes,
			}
		}

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}

func newTraceDecisionHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		// Search for decision nodes matching the query
		decNodes := g.SearchNodesByTypeAndLabel(NodeDecision, query)
		if len(decNodes) == 0 {
			// Try searching all nodes
			allNodes := g.SearchNodesByLabel(query)
			for _, n := range allNodes {
				if n.Type == NodeDecision {
					decNodes = append(decNodes, n)
				}
			}
		}

		if len(decNodes) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("no decisions found matching: %s", query)), nil
		}

		result := traceDecisionResult{
			Query:     query,
			Decisions: make([]decisionTrace, 0, len(decNodes)),
		}

		for _, dec := range decNodes {
			trace := decisionTrace{
				DecisionID:    dec.ID,
				DecisionLabel: dec.Label,
				Category:      mustGetString(dec.Properties, "category", ""),
				Date:          mustGetString(dec.Properties, "date", ""),
				Specs:         make([]nodeSummary, 0),
				Files:         make([]nodeSummary, 0),
				Functions:     make([]nodeSummary, 0),
			}

			for _, edge := range g.GetOutbound(dec.ID) {
				if edge.Type == EdgeReferences {
					if n, ok := g.GetNode(edge.To); ok {
						switch n.Type {
						case NodeSpec:
							trace.Specs = append(trace.Specs, nodeSummary{ID: n.ID, Type: string(n.Type), Label: n.Label})
						case NodeFile:
							trace.Files = append(trace.Files, nodeSummary{ID: n.ID, Type: string(n.Type), Label: n.Label})
						case NodeFunction:
							trace.Functions = append(trace.Functions, nodeSummary{ID: n.ID, Type: string(n.Type), Label: n.Label})
						}
					}
				}
			}

			result.Decisions = append(result.Decisions, trace)
		}

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}

func newSpecToCodeHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		specQuery, _ := args["spec_query"].(string)
		if specQuery == "" {
			return mcp.NewToolResultError("spec_query is required"), nil
		}

		specNodes := g.SearchNodesByTypeAndLabel(NodeSpec, specQuery)
		if len(specNodes) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("no specs found matching: %s", specQuery)), nil
		}

		result := specToCodeResult{
			Query: specQuery,
			Specs: make([]specCodeMapping, 0, len(specNodes)),
		}

		for _, spec := range specNodes {
			mapping := specCodeMapping{
				SpecTitle: spec.Label,
				SpecPath:  mustGetString(spec.Properties, "path", ""),
				Files:     make([]nodeSummary, 0),
				Functions: make([]nodeSummary, 0),
			}

			for _, edge := range g.GetOutbound(spec.ID) {
				if edge.Type == EdgeImplements || edge.Type == EdgeReferences {
					if n, ok := g.GetNode(edge.To); ok {
						switch n.Type {
						case NodeFile:
							mapping.Files = append(mapping.Files, nodeSummary{
								ID:    n.ID,
								Type:  string(n.Type),
								Label: n.Label,
							})
						case NodeFunction:
							fnLine := int(mustGetFloat(n.Properties, "line", 0))
							mapping.Functions = append(mapping.Functions, nodeSummary{
								ID:    n.ID,
								Type:  string(n.Type),
								Label: n.Label,
								Line:  fnLine,
							})
						}
					}
				}
			}

			result.Specs = append(result.Specs, mapping)
		}

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}

func newCausalChainHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		pkgName, _ := args["package_name"].(string)
		if pkgName == "" {
			return mcp.NewToolResultError("package_name is required"), nil
		}

		timeRangeDays := 90
		if v, ok := args["time_range_days"].(float64); ok && v > 0 {
			timeRangeDays = int(v)
		}

		cutoff := time.Now().AddDate(0, 0, -timeRangeDays)

		// Find the package node
		pkgNodeID := slug("pkg", pkgName)
		if _, exists := g.GetNode(pkgNodeID); !exists {
			return mcp.NewToolResultError(fmt.Sprintf("package not found: %s", pkgName)), nil
		}

		result := causalChainResult{
			PackageName: pkgName,
			TimeRange:   fmt.Sprintf("%d days", timeRangeDays),
			Commits:     make([]commitSummary, 0),
			Bugs:        make([]bugSummary, 0),
		}

		// Find files in this package
		fileIDs := make(map[string]bool)
		for _, edge := range g.GetOutbound(pkgNodeID) {
			if edge.Type == EdgeContains {
				if n, ok := g.GetNode(edge.To); ok {
					if n.Type == NodeFile || n.Type == NodeFunction {
						fileIDs[n.ID] = true
						if n.Type == NodeFunction {
							filePath := mustGetString(n.Properties, "file", "")
							if filePath != "" {
								fileIDs[slug("file", filePath)] = true
							}
						}
					}
				}
			}
		}

		// Trace commits that changed these files
		seenCommits := make(map[string]bool)
		for fileID := range fileIDs {
			for _, edge := range g.GetInbound(fileID) {
				if edge.Type != EdgeChanged {
					continue
				}
				commitNode, ok := g.GetNode(edge.From)
				if !ok || commitNode.Type != NodeCommit {
					continue
				}

				tsStr := mustGetString(commitNode.Properties, "timestamp", "")
				if tsStr != "" {
					ts, err := time.Parse(time.RFC3339, tsStr)
					if err == nil && ts.Before(cutoff) {
						continue
					}
				}

				if seenCommits[commitNode.ID] {
					continue
				}
				seenCommits[commitNode.ID] = true

				result.Commits = append(result.Commits, commitSummary{
					Hash:      mustGetString(commitNode.Properties, "hash", ""),
					Message:   commitNode.Label,
					Timestamp: mustGetString(commitNode.Properties, "timestamp", ""),
				})

				// Find bugs from this commit (inbound CAUSED_BY edges)
				for _, inEdge := range g.GetInbound(commitNode.ID) {
					if inEdge.Type == EdgeCausedBy {
						bugNode, ok := g.GetNode(inEdge.From)
						if ok && bugNode.Type == NodeBug {
							result.Bugs = append(result.Bugs, bugSummary{
								Message:  bugNode.Label,
								Severity: mustGetString(bugNode.Properties, "severity", "unknown"),
							})
						}
					}
				}
			}
		}

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}

func newGraphQueryHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		startNodeID, _ := args["start_node_id"].(string)
		if startNodeID == "" {
			return mcp.NewToolResultError("start_node_id is required"), nil
		}

		maxHops := 5
		if v, ok := args["max_hops"].(float64); ok && v > 0 {
			maxHops = int(v)
		}

		direction := "outbound"
		if d, ok := args["direction"].(string); ok && d != "" {
			direction = d
		}

		var edgeTypes []EdgeType
		if et, ok := args["edge_types"].(string); ok && et != "" {
			for _, part := range strings.Split(et, ",") {
				edgeTypes = append(edgeTypes, EdgeType(strings.TrimSpace(part)))
			}
		}

		layers, err := g.BlastRadius(startNodeID, maxHops, edgeTypes, direction)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query error: %v", err)), nil
		}

		result := graphQueryResult{
			StartNodeID: startNodeID,
			MaxHops:     maxHops,
			Direction:   direction,
			EdgeTypes:   edgeTypesToString(edgeTypes),
			TotalNodes:  countLayers(layers),
			Paths:       make([]pathResult, 0),
		}

		visited := map[string]bool{startNodeID: true}
		for hopNum := 1; hopNum < len(layers); hopNum++ {
			for _, node := range layers[hopNum] {
				if visited[node.ID] {
					continue
				}
				visited[node.ID] = true
				result.Paths = append(result.Paths, pathResult{
					NodeID:    node.ID,
					NodeType:  string(node.Type),
					NodeLabel: node.Label,
					Hop:       hopNum,
				})
			}
		}

		return mcp.NewToolResultStructuredOnly(result), nil
	}
}

func newReindexHandler(g *Graph, repoRoot string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		root, _ := args["repo_root"].(string)
		if root == "" {
			root = repoRoot
		}

		Logf("starting full knowledge graph reindex from %s", root)

		// Rebuild from scratch
		newGraph := NewGraph()
		newGraph.KnowledgeDir = g.KnowledgeDir

		if err := IngestGoCodebase(newGraph, root); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("go ingestion failed: %v", err)), nil
		}

		specDirs := []string{
			filepath.Join(root, "..", "research"),
			filepath.Join(root, "..", "ops"),
		}
		if err := IngestSpecs(newGraph, specDirs); err != nil {
			Logf("warning: spec ingestion: %v", err)
		}

		memories := loadDecisionCache()
		if len(memories) > 0 {
			if err := IngestDecisions(newGraph, memories); err != nil {
				Logf("warning: decision ingestion: %v", err)
			}
		}

		if err := IngestGitHistory(newGraph, root); err != nil {
			Logf("warning: git ingestion: %v", err)
		}

		opencodePath := filepath.Join(root, "..", "opencode.json")
		if err := IngestAgentOwnership(newGraph, opencodePath); err != nil {
			Logf("warning: agent ownership ingestion: %v", err)
		}

		newGraph.LastBuilt = time.Now()

		if err := newGraph.ExportJSON(newGraph.snapshotPath()); err != nil {
			Logf("warning: snapshot persist: %v", err)
		}

		// Replace the graph
		g.mu.Lock()
		g.Nodes = newGraph.Nodes
		g.EdgesOut = newGraph.EdgesOut
		g.EdgesIn = newGraph.EdgesIn
		g.LastBuilt = newGraph.LastBuilt
		g.SourceFiles = newGraph.SourceFiles
		g.byType = newGraph.byType
		g.mu.Unlock()

		nodes, edges, types := newGraph.Stats()

		return mcp.NewToolResultStructuredOnly(map[string]interface{}{
			"status":     "ok",
			"nodes":      nodes,
			"edges":      edges,
			"types":      types,
			"last_built": newGraph.LastBuilt.Format(time.RFC3339),
		}), nil
	}
}

func newStatsHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodes, edges, types := g.Stats()

		typeCounts := make(map[string]int)
		g.mu.RLock()
		for _, node := range g.Nodes {
			typeCounts[string(node.Type)]++
		}
		g.mu.RUnlock()

		return mcp.NewToolResultStructuredOnly(map[string]interface{}{
			"total_nodes":     nodes,
			"total_edges":     edges,
			"node_type_count": types,
			"node_counts":     typeCounts,
			"last_built":      g.LastBuilt.Format(time.RFC3339),
			"source_files":    len(g.SourceFiles),
		}), nil
	}
}

// --- Result types ---

type blastRadiusResult struct {
	StartNodeID string        `json:"start_node_id"`
	MaxHops     int           `json:"max_hops"`
	Direction   string        `json:"direction"`
	TotalNodes  int           `json:"total_nodes"`
	Layers      []layerResult `json:"layers"`
}

type layerResult struct {
	Hop   int           `json:"hop"`
	Nodes []nodeSummary `json:"nodes"`
}

type nodeSummary struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Line  int    `json:"line,omitempty"`
}

type traceDecisionResult struct {
	Query     string          `json:"query"`
	Decisions []decisionTrace `json:"decisions"`
}

type decisionTrace struct {
	DecisionID    string        `json:"decision_id"`
	DecisionLabel string        `json:"decision_label"`
	Category      string        `json:"category"`
	Date          string        `json:"date"`
	Specs         []nodeSummary `json:"specs"`
	Files         []nodeSummary `json:"files"`
	Functions     []nodeSummary `json:"functions"`
}

type specToCodeResult struct {
	Query string            `json:"query"`
	Specs []specCodeMapping `json:"specs"`
}

type specCodeMapping struct {
	SpecTitle string        `json:"spec_title"`
	SpecPath  string        `json:"spec_path"`
	Files     []nodeSummary `json:"files"`
	Functions []nodeSummary `json:"functions"`
}

type causalChainResult struct {
	PackageName string          `json:"package_name"`
	TimeRange   string          `json:"time_range"`
	Commits     []commitSummary `json:"commits"`
	Bugs        []bugSummary    `json:"bugs"`
}

type commitSummary struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type bugSummary struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type graphQueryResult struct {
	StartNodeID string       `json:"start_node_id"`
	MaxHops     int          `json:"max_hops"`
	Direction   string       `json:"direction"`
	EdgeTypes   string       `json:"edge_types"`
	TotalNodes  int          `json:"total_nodes"`
	Paths       []pathResult `json:"paths"`
}

type pathResult struct {
	NodeID    string `json:"node_id"`
	NodeType  string `json:"node_type"`
	NodeLabel string `json:"node_label"`
	Hop       int    `json:"hop"`
}

func newVisualizeHandler(g *Graph) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		layout := "cose"
		if v, ok := args["layout"].(string); ok && v != "" {
			layout = v
		}

		// Generate HTML
		if _, err := g.GenerateVisualization(VisualizeOptions{
			Layout:      layout,
			OpenBrowser: false,
		}); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("generate visualization: %v", err)), nil
		}

		// Start ephemeral HTTP server on port 8742 (or find free port)
		url, cancel, err := g.StartVisualizationServer(8742)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("start server: %v", err)), nil
		}
		// Ensure cancellation on context done
		go func() {
			<-ctx.Done()
			cancel()
		}()

		outputPath := g.KnowledgeDir + "/visualize.html"
		nodes, edges, _ := g.Stats()

		return mcp.NewToolResultStructuredOnly(map[string]interface{}{
			"url":   url,
			"path":  outputPath,
			"nodes": nodes,
			"edges": edges,
		}), nil
	}
}

// --- Helpers ---

func countLayers(layers [][]*Node) int {
	count := 0
	for _, l := range layers {
		count += len(l)
	}
	return count
}

func edgeTypesToString(ets []EdgeType) string {
	if len(ets) == 0 {
		return "all"
	}
	parts := make([]string, len(ets))
	for i, et := range ets {
		parts[i] = string(et)
	}
	return strings.Join(parts, ",")
}

func mustGetFloat(m map[string]interface{}, key string, def float64) float64 {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	}
	return def
}

// loadDecisionCache reads persisted decision entries from the cache file.
func loadDecisionCache() []DecisionEntry {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	cacheFile := filepath.Join(home, ".ailinter", "knowledge", "decisions.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}
	var entries []DecisionEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}
