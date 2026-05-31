package knowledge

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestGraphAddNode(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{
		ID:    "file:test.go",
		Type:  NodeFile,
		Label: "test.go",
	})

	n, exists := g.GetNode("file:test.go")
	if !exists {
		t.Fatal("expected node to exist")
	}
	if n.Type != NodeFile {
		t.Fatalf("expected NodeFile, got %s", n.Type)
	}
	if n.Label != "test.go" {
		t.Fatalf("expected 'test.go', got '%s'", n.Label)
	}
}

func TestGraphAddEdge(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})

	g.AddEdge("a", "b", EdgeCalls, map[string]interface{}{"count": 5})

	out := g.GetOutbound("a")
	if len(out) != 1 {
		t.Fatalf("expected 1 outbound edge, got %d", len(out))
	}
	if out[0].To != "b" || out[0].Type != EdgeCalls {
		t.Fatalf("unexpected edge: %v", out[0])
	}

	in := g.GetInbound("b")
	if len(in) != 1 {
		t.Fatalf("expected 1 inbound edge, got %d", len(in))
	}

	// Test duplicate edge is ignored
	g.AddEdge("a", "b", EdgeCalls, nil)
	out = g.GetOutbound("a")
	if len(out) != 1 {
		t.Fatalf("expected still 1 outbound edge after duplicate, got %d", len(out))
	}
}

func TestGraphAddNodeIdempotent(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "x", Type: NodePackage, Label: "pkg1"})
	g.AddNode(Node{ID: "x", Type: NodePackage, Label: "pkg1-updated"})

	n, _ := g.GetNode("x")
	if n.Label != "pkg1-updated" {
		t.Fatalf("expected updated label, got %s", n.Label)
	}
}

func TestGraphRemoveNode(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddEdge("a", "b", EdgeCalls, nil)

	g.RemoveNode("a")

	if _, exists := g.GetNode("a"); exists {
		t.Fatal("expected node a to be removed")
	}

	// Check edges are cleaned up
	in := g.GetInbound("b")
	if len(in) != 0 {
		t.Fatalf("expected no inbound edges to b after removing a, got %d", len(in))
	}
}

func TestGraphFindNodesByType(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:a.go", Type: NodeFile, Label: "a.go"})
	g.AddNode(Node{ID: "file:b.go", Type: NodeFile, Label: "b.go"})
	g.AddNode(Node{ID: "pkg:main", Type: NodePackage, Label: "main"})

	files := g.FindNodesByType(NodeFile)
	if len(files) != 2 {
		t.Fatalf("expected 2 file nodes, got %d", len(files))
	}

	pkgs := g.FindNodesByType(NodePackage)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package node, got %d", len(pkgs))
	}
}

func TestGraphFindNodeByLabel(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:main.go", Type: NodeFile, Label: "main.go"})
	g.AddNode(Node{ID: "pkg:main", Type: NodePackage, Label: "main"})

	n, exists := g.FindNodeByLabel("main.go")
	if !exists || n.ID != "file:main.go" {
		t.Fatalf("expected to find node with label 'main.go'")
	}
}

func TestGraphStats(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodePackage, Label: "B"})
	g.AddEdge("a", "b", EdgeContains, nil)

	nodes, edges, types := g.Stats()
	if nodes != 2 {
		t.Fatalf("expected 2 nodes, got %d", nodes)
	}
	if edges != 1 {
		t.Fatalf("expected 1 edge, got %d", edges)
	}
	if types != 2 {
		t.Fatalf("expected 2 types, got %d", types)
	}
}

func TestGraphBlastRadius(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodeFile, Label: "C"})
	g.AddNode(Node{ID: "d", Type: NodePackage, Label: "D"})

	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("b", "c", EdgeCalls, nil)
	g.AddEdge("c", "d", EdgeContains, nil)

	layers, err := g.BlastRadius("a", 3, nil, DirectionOutbound)
	if err != nil {
		t.Fatalf("BlastRadius error: %v", err)
	}

	if len(layers) < 3 {
		t.Fatalf("expected at least 3 layers, got %d", len(layers))
	}
	// Layer 0: a
	if len(layers[0]) != 1 || layers[0][0].ID != "a" {
		t.Fatalf("expected layer 0 to contain 'a', got %v", layerIDs(layers[0]))
	}
	// Layer 1: b
	if len(layers[1]) != 1 || layers[1][0].ID != "b" {
		t.Fatalf("expected layer 1 to contain 'b'")
	}
	// Layer 2: c
	if len(layers[2]) != 1 || layers[2][0].ID != "c" {
		t.Fatalf("expected layer 2 to contain 'c'")
	}
}

func TestGraphBlastRadiusInbound(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodeFunction, Label: "C"})

	g.AddEdge("c", "b", EdgeCalls, nil)
	g.AddEdge("b", "a", EdgeCalls, nil)

	layers, err := g.BlastRadius("a", 2, nil, DirectionInbound)
	if err != nil {
		t.Fatalf("BlastRadius error: %v", err)
	}

	// Layer 0: a
	// Layer 1: b (inbound edge from b→a means a is reached from b)
	if len(layers) < 2 {
		t.Fatalf("expected at least 2 layers, got %d", len(layers))
	}
	found := false
	for _, n := range layers[1] {
		if n.ID == "b" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected layer 1 to contain 'b', got %v", layerIDs(layers[1]))
	}
}

func TestGraphBlastRadiusFiltered(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodeFile, Label: "C"})

	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("a", "c", EdgeContains, nil)

	// Only follow CALLS edges
	layers, err := g.BlastRadius("a", 2, []EdgeType{EdgeCalls}, DirectionOutbound)
	if err != nil {
		t.Fatalf("BlastRadius error: %v", err)
	}

	if len(layers) < 2 {
		t.Fatalf("expected at least 2 layers, got %d", len(layers))
	}
	if len(layers[1]) != 1 || layers[1][0].ID != "b" {
		t.Fatalf("expected only 'b' in layer 1 (filtered by CALLS), got %v", layerIDs(layers[1]))
	}
}

func TestGraphFindPath(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFunction, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodePackage, Label: "C"})
	g.AddNode(Node{ID: "d", Type: NodeFile, Label: "D"})

	g.AddEdge("a", "b", EdgeContains, nil)
	g.AddEdge("b", "c", EdgeCalls, nil)
	g.AddEdge("c", "d", EdgeContains, nil)

	paths, err := g.FindPath("a", "d", 10)
	if err != nil {
		t.Fatalf("FindPath error: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("expected at least one path")
	}
	// Should find a→b→c→d (3 hops)
	if len(paths[0]) != 4 {
		t.Fatalf("expected path of length 4, got %d: %v", len(paths[0]), pathIDs(paths[0]))
	}
}

func TestGraphFindPathNoPath(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "z", Type: NodeFile, Label: "Z"})

	paths, err := g.FindPath("a", "z", 10)
	if err != nil {
		t.Fatalf("FindPath error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got %d", len(paths))
	}
}

func TestGraphFindPathSameNode(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})

	paths, err := g.FindPath("a", "a", 10)
	if err != nil {
		t.Fatalf("FindPath error: %v", err)
	}
	if len(paths) != 1 || len(paths[0]) != 1 {
		t.Fatalf("expected single-node path for self-loop")
	}
}

func TestGraphFindPathNotFound(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})

	_, err := g.FindPath("a", "nonexistent", 10)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}

	_, err = g.FindPath("nonexistent", "a", 10)
	if err == nil {
		t.Fatal("expected error for nonexistent start")
	}
}

func TestGraphBlastRadiusNotFound(t *testing.T) {
	g := NewGraph()
	_, err := g.BlastRadius("nonexistent", 3, nil, DirectionOutbound)
	if err == nil {
		t.Fatal("expected error for nonexistent start node")
	}
}

func TestGraphTraverse(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodeFile, Label: "C"})
	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("b", "c", EdgeCalls, nil)

	var visited []string
	g.Traverse("a", 3, DirectionOutbound, nil, func(node *Node, edges []Edge, hop int) bool {
		visited = append(visited, node.ID)
		return true
	})

	if len(visited) != 3 {
		t.Fatalf("expected 3 visited nodes, got %d: %v", len(visited), visited)
	}
	if visited[0] != "a" || visited[1] != "b" || visited[2] != "c" {
		t.Fatalf("unexpected traversal order: %v", visited)
	}
}

func TestGraphTraverseStop(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddEdge("a", "b", EdgeCalls, nil)

	var visited []string
	g.Traverse("a", 3, DirectionOutbound, nil, func(node *Node, edges []Edge, hop int) bool {
		visited = append(visited, node.ID)
		return false // stop after first node
	})

	if len(visited) != 1 {
		t.Fatalf("expected only 1 visited node, got %d", len(visited))
	}
}

func TestGraphSerializationRoundTrip(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:a.go", Type: NodeFile, Label: "a.go", Properties: map[string]interface{}{"path": "a.go"}})
	g.AddNode(Node{ID: "pkg:main", Type: NodePackage, Label: "main"})
	g.AddEdge("file:a.go", "pkg:main", EdgeContains, nil)
	g.LastBuilt = time.Now()
	g.SourceFiles["a.go"] = time.Now()

	// Create temp file
	tmpDir := t.TempDir()
	snapPath := filepath.Join(tmpDir, "snapshot.json")

	if err := g.ExportJSON(snapPath); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	loaded, err := LoadGraphJSON(snapPath)
	if err != nil {
		t.Fatalf("LoadGraphJSON error: %v", err)
	}

	// Verify nodes
	if len(loaded.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(loaded.Nodes))
	}
	n, exists := loaded.GetNode("file:a.go")
	if !exists || n.Label != "a.go" {
		t.Fatalf("expected node 'file:a.go' with label 'a.go'")
	}

	// Verify edges
	out := loaded.GetOutbound("file:a.go")
	if len(out) != 1 || out[0].To != "pkg:main" {
		t.Fatalf("expected 1 edge from file:a.go to pkg:main")
	}

	// Verify source files
	if len(loaded.SourceFiles) != 1 {
		t.Fatalf("expected 1 source file, got %d", len(loaded.SourceFiles))
	}
}

func TestGraphSerializationRoundTripFullGraph(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:a.go", Type: NodeFile, Label: "a.go"})
	g.AddNode(Node{ID: "file:b.go", Type: NodeFile, Label: "b.go"})
	g.AddNode(Node{ID: "file:c_test.go", Type: NodeTest, Label: "c_test.go"})
	g.AddNode(Node{ID: "pkg:main", Type: NodePackage, Label: "main"})
	g.AddNode(Node{ID: "func:main.Run", Type: NodeFunction, Label: "main.Run", Properties: map[string]interface{}{
		"line":    10,
		"package": "main",
	}})
	g.AddNode(Node{ID: "spec:readme", Type: NodeSpec, Label: "README"})

	g.AddEdge("file:a.go", "pkg:main", EdgeContains, nil)
	g.AddEdge("file:b.go", "pkg:main", EdgeContains, nil)
	g.AddEdge("pkg:main", "func:main.Run", EdgeContains, nil)
	g.AddEdge("file:c_test.go", "pkg:main", EdgeTests, nil)
	g.AddEdge("func:main.Run", "func:main.Run", EdgeCalls, map[string]interface{}{"self": true})
	g.AddEdge("spec:readme", "file:a.go", EdgeImplements, nil)

	tmpDir := t.TempDir()
	snapPath := filepath.Join(tmpDir, "snapshot.json")
	if err := g.ExportJSON(snapPath); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	loaded, err := LoadGraphJSON(snapPath)
	if err != nil {
		t.Fatalf("LoadGraphJSON error: %v", err)
	}

	if len(loaded.Nodes) != 6 {
		t.Fatalf("expected 6 nodes, got %d", len(loaded.Nodes))
	}
	if len(loaded.FindNodesByType(NodeFile)) != 2 {
		t.Fatalf("expected 2 file nodes")
	}
	if len(loaded.FindNodesByType(NodeTest)) != 1 {
		t.Fatalf("expected 1 test node")
	}

	// Verify edge counts (1 + 1 + 1 + 1 + 1 + 1 = 6)
	_, edgeCount, _ := loaded.Stats()
	if edgeCount != 6 {
		t.Fatalf("expected 6 edges, got %d", edgeCount)
	}
}

func TestGraphSearchByLabel(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:analyzer.go", Type: NodeFile, Label: "internal/analyzer/analyzer.go"})
	g.AddNode(Node{ID: "file:server.go", Type: NodeFile, Label: "internal/mcp/server.go"})
	g.AddNode(Node{ID: "file:config.go", Type: NodeFile, Label: "internal/config/config.go"})

	results := g.SearchNodesByLabel("analyzer")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'analyzer', got %d", len(results))
	}

	results = g.SearchNodesByTypeAndLabel(NodeFile, "internal")
	if len(results) != 3 {
		t.Fatalf("expected 3 file results for 'internal', got %d", len(results))
	}
}

func TestGraphConcurrency(t *testing.T) {
	g := NewGraph()
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			id := fmt.Sprintf("node:%d", n)
			g.AddNode(Node{ID: id, Type: NodeFile, Label: id})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if len(g.Nodes) != 10 {
		t.Fatalf("expected 10 nodes after concurrent writes, got %d", len(g.Nodes))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(n int) {
			id := fmt.Sprintf("node:%d", n)
			g.GetNode(id)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		prefix string
		parts  []string
		want   string
	}{
		{"file", []string{"a.go"}, "file:a.go"},
		{"pkg", []string{"main"}, "pkg:main"},
		{"func", []string{"main", "Run"}, "func:main.Run"},
		{"func", []string{"pkg", "Type", "Method"}, "func:pkg.Type.Method"},
		{"spec", []string{"readme"}, "spec:readme"},
		{"commit", []string{"abc123def456"}, "commit:abc123def456"},
	}

	for _, tt := range tests {
		got := slug(tt.prefix, tt.parts...)
		if got != tt.want {
			t.Errorf("slug(%q, %v) = %q, want %q", tt.prefix, tt.parts, got, tt.want)
		}
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"My Feature/Enhancement", "my-feature-enhancement"},
		{"foo.bar:baz", "foo-bar-baz"},
		{"(parentheses) [brackets] {braces}", "parentheses-brackets-braces"},
		{"Special'Chars\"Here", "special-chars-here"},
		{"  Trim  Me  ", "trim-me"},
		{"already-kebab", "already-kebab"},
	}

	for _, tt := range tests {
		got := sanitizeID(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc..."},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestIsBugCommit(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"fix: resolve nil pointer", true},
		{"bug: incorrect timeout value", true},
		{"regression: test suite fails", true},
		{"feat: add new feature", false},
		{"chore: update dependencies", false},
		{"security: fix XSS vulnerability", true},
		{"docs: update readme", false},
		{"hotfix: critical production crash", true},
	}

	for _, tt := range tests {
		got := isBugCommit(tt.msg)
		if got != tt.want {
			t.Errorf("isBugCommit(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

func TestGraphRemoveEdgesForNode(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})
	g.AddNode(Node{ID: "c", Type: NodeFile, Label: "C"})
	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("b", "c", EdgeCalls, nil)
	g.AddEdge("a", "c", EdgeContains, nil)

	g.RemoveEdgesForNode("b")

	// a's edges should no longer include b
	for _, e := range g.GetOutbound("a") {
		if e.To == "b" {
			t.Fatal("expected no edge from a to b after removing b's edges")
		}
	}

	// c should have no inbound from b
	for _, e := range g.GetInbound("c") {
		if e.From == "b" {
			t.Fatal("expected no inbound edge from b to c after removing b's edges")
		}
	}
}

func TestGraphNodeFindByIDPrefix(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:internal/analyzer/analyzer.go", Type: NodeFile, Label: "analyzer.go"})
	g.AddNode(Node{ID: "file:internal/mcp/server.go", Type: NodeFile, Label: "server.go"})

	n, exists := g.FindNodeByIDPrefix("file:internal/analyzer")
	if !exists || n.ID != "file:internal/analyzer/analyzer.go" {
		t.Fatalf("expected to find node by ID prefix")
	}
}

// --- Helpers ---

func layerIDs(nodes []*Node) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	return ids
}

func pathIDs(nodes []*Node) []string {
	return layerIDs(nodes)
}
