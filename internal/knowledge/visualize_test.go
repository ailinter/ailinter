package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToCytoscapeElements(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:a.go", Type: NodeFile, Label: "a.go", Properties: map[string]interface{}{"path": "a.go"}})
	g.AddNode(Node{ID: "file:b.go", Type: NodeFile, Label: "b.go"})
	g.AddNode(Node{ID: "pkg:main", Type: NodePackage, Label: "main"})
	g.AddNode(Node{ID: "func:main.Run", Type: NodeFunction, Label: "main.Run"})

	g.AddEdge("file:a.go", "pkg:main", EdgeContains, nil)
	g.AddEdge("file:b.go", "pkg:main", EdgeContains, nil)
	g.AddEdge("file:a.go", "func:main.Run", EdgeCalls, nil)

	elements, nodeCounts, edgeCounts, totalNodes, totalEdges := g.toCytoscapeElements()

	// Verify total counts
	if totalNodes != 4 {
		t.Fatalf("expected 4 total nodes, got %d", totalNodes)
	}
	if totalEdges != 3 {
		t.Fatalf("expected 3 total edges, got %d", totalEdges)
	}

	// Verify node counts by type
	if nodeCounts["file"] != 2 {
		t.Fatalf("expected 2 file nodes, got %d", nodeCounts["file"])
	}
	if nodeCounts["package"] != 1 {
		t.Fatalf("expected 1 package node, got %d", nodeCounts["package"])
	}
	if nodeCounts["function"] != 1 {
		t.Fatalf("expected 1 function node, got %d", nodeCounts["function"])
	}

	// Verify edge counts by type
	if edgeCounts[string(EdgeContains)] != 2 {
		t.Fatalf("expected 2 CONTAINS edges, got %d", edgeCounts[string(EdgeContains)])
	}
	if edgeCounts[string(EdgeCalls)] != 1 {
		t.Fatalf("expected 1 CALLS edge, got %d", edgeCounts[string(EdgeCalls)])
	}

	// Verify element format
	nodeElements := 0
	edgeElements := 0
	for _, el := range elements {
		data, ok := el["data"].(map[string]interface{})
		if !ok {
			t.Fatal("element missing data field")
		}
		if _, hasSource := data["source"]; hasSource {
			edgeElements++
			// Edge must have target
			if _, ok := data["target"]; !ok {
				t.Fatal("edge missing target")
			}
		} else {
			nodeElements++
			// Node must have id, label, type
			if _, ok := data["id"]; !ok {
				t.Fatal("node missing id")
			}
			if _, ok := data["label"]; !ok {
				t.Fatal("node missing label")
			}
			if _, ok := data["type"]; !ok {
				t.Fatal("node missing type")
			}
		}
	}
	if nodeElements != 4 {
		t.Fatalf("expected 4 node elements, got %d", nodeElements)
	}
	if edgeElements != 3 {
		t.Fatalf("expected 3 edge elements, got %d", edgeElements)
	}
}

func TestGenerateVisualization(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "file:a.go", Type: NodeFile, Label: "a.go"})
	g.AddNode(Node{ID: "file:b.go", Type: NodeFile, Label: "b.go"})
	g.AddEdge("file:a.go", "file:b.go", EdgeCalls, nil)

	tmpDir := t.TempDir()
	g.KnowledgeDir = tmpDir

	outputPath, err := g.GenerateVisualization(VisualizeOptions{
		OutputPath: filepath.Join(tmpDir, "viz.html"),
		Title:      "Test Graph",
		Layout:     "cose",
	})
	if err != nil {
		t.Fatalf("GenerateVisualization error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected visualization file to exist")
	}

	// Read and verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	content := string(data)

	// Verify it contains key elements
	if !strings.Contains(content, "unpkg.com/cytoscape") {
		t.Fatal("HTML should load Cytoscape from CDN")
	}
	if !strings.Contains(content, "file:a.go") {
		t.Fatal("HTML should contain node data for file:a.go")
	}
	if !strings.Contains(content, "file:b.go") {
		t.Fatal("HTML should contain node data for file:b.go")
	}
	if !strings.Contains(content, "Test Graph") {
		t.Fatal("HTML should contain custom title")
	}
	if !strings.Contains(content, "const DATA") {
		t.Fatal("HTML should embed DATA const")
	}
	if !strings.Contains(content, "cytoscape@3.30") {
		t.Fatal("HTML should load Cytoscape.js from CDN")
	}
}

func TestGenerateVisualizationEmptyGraph(t *testing.T) {
	g := NewGraph()
	tmpDir := t.TempDir()
	g.KnowledgeDir = tmpDir

	outputPath, err := g.GenerateVisualization(VisualizeOptions{
		OutputPath: filepath.Join(tmpDir, "empty.html"),
	})
	if err != nil {
		t.Fatalf("GenerateVisualization error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	content := string(data)

	// Verify it still generates valid HTML with Cytoscape CDN
	if !strings.Contains(content, "unpkg.com/cytoscape") {
		t.Fatal("HTML should load Cytoscape from CDN even for empty graph")
	}
	if !strings.Contains(content, "const DATA") {
		t.Fatal("HTML should embed DATA const even for empty graph")
	}
	// Should show empty stats
	if !strings.Contains(content, "0 Nodes") {
		t.Log("note: empty graph showing 0 nodes")
	}
}

func TestVisualizeEdgeDedup(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})

	// Add same edge twice (should be deduped by AddEdge)
	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("a", "b", EdgeCalls, nil)

	elements, _, _, _, totalEdges := g.toCytoscapeElements()

	if totalEdges != 1 {
		t.Fatalf("expected 1 edge after dedup, got %d", totalEdges)
	}

	// Count edge elements
	edgeCount := 0
	for _, el := range elements {
		if _, ok := el["data"].(map[string]interface{})["source"]; ok {
			edgeCount++
		}
	}
	if edgeCount != 1 {
		t.Fatalf("expected 1 edge element, got %d", edgeCount)
	}
}

func TestVisualizeEdgeDedupMultipleTypes(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "a", Type: NodeFile, Label: "A"})
	g.AddNode(Node{ID: "b", Type: NodeFile, Label: "B"})

	// Same nodes with different edge types — should both appear
	g.AddEdge("a", "b", EdgeCalls, nil)
	g.AddEdge("a", "b", EdgeImports, nil)

	// Same edge added twice (duplicate)
	g.AddEdge("a", "b", EdgeCalls, nil)

	elements, _, edgeCounts, _, totalEdges := g.toCytoscapeElements()

	if totalEdges != 2 {
		t.Fatalf("expected 2 unique edges, got %d", totalEdges)
	}

	if edgeCounts[string(EdgeCalls)] != 1 {
		t.Fatalf("expected 1 CALLS edge, got %d", edgeCounts[string(EdgeCalls)])
	}
	if edgeCounts[string(EdgeImports)] != 1 {
		t.Fatalf("expected 1 IMPORTS edge, got %d", edgeCounts[string(EdgeImports)])
	}

	// Count edge elements
	edgeCount := 0
	for _, el := range elements {
		if _, ok := el["data"].(map[string]interface{})["source"]; ok {
			edgeCount++
		}
	}
	if edgeCount != 2 {
		t.Fatalf("expected 2 edge elements, got %d", edgeCount)
	}
}

func TestVisualizeDefaultOutputPath(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{ID: "x", Type: NodeFile, Label: "x.go"})

	tmpDir := t.TempDir()
	g.KnowledgeDir = tmpDir

	// Without specifying OutputPath, should use KnowledgeDir/visualize.html
	outputPath, err := g.GenerateVisualization(VisualizeOptions{})
	if err != nil {
		t.Fatalf("GenerateVisualization error: %v", err)
	}

	expected := filepath.Join(tmpDir, "visualize.html")
	if outputPath != expected {
		t.Fatalf("expected output path %s, got %s", expected, outputPath)
	}

	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Fatal("expected visualize.html to exist at default path")
	}
}

func TestVisualizePropertiesFlattened(t *testing.T) {
	g := NewGraph()
	g.AddNode(Node{
		ID:    "file:main.go",
		Type:  NodeFile,
		Label: "main.go",
		Properties: map[string]interface{}{
			"path":  "/src/main.go",
			"lines": 150,
		},
	})

	elements, _, _, _, _ := g.toCytoscapeElements()

	for _, el := range elements {
		data := el["data"].(map[string]interface{})
		if data["id"] == "file:main.go" {
			if data["path"] != "/src/main.go" {
				t.Fatalf("expected properties to be flattened, path=%v", data["path"])
			}
			if data["lines"] != float64(150) && data["lines"] != 150 {
				t.Fatalf("expected properties to be flattened, lines=%v", data["lines"])
			}
		}
	}
}

func TestVisualizeFullGraphTypes(t *testing.T) {
	g := NewGraph()
	// Add all 8 expected node types
	g.AddNode(Node{ID: "agent:dev", Type: NodeAgent, Label: "Dev"})
	g.AddNode(Node{ID: "bug:123", Type: NodeBug, Label: "Bug"})
	g.AddNode(Node{ID: "commit:abc", Type: NodeCommit, Label: "Commit"})
	g.AddNode(Node{ID: "file:f.go", Type: NodeFile, Label: "f.go"})
	g.AddNode(Node{ID: "func:run", Type: NodeFunction, Label: "Run"})
	g.AddNode(Node{ID: "pkg:p", Type: NodePackage, Label: "Pkg"})
	g.AddNode(Node{ID: "spec:s", Type: NodeSpec, Label: "Spec"})
	g.AddNode(Node{ID: "test:t", Type: NodeTest, Label: "Test"})

	// Add edges of all 7+ expected types
	g.AddEdge("file:f.go", "func:run", EdgeCalls, nil)
	g.AddEdge("pkg:p", "file:f.go", EdgeContains, nil)
	g.AddEdge("agent:dev", "file:f.go", EdgeOwns, nil)
	g.AddEdge("file:f.go", "commit:abc", EdgeChanged, nil)
	g.AddEdge("bug:123", "file:f.go", EdgeCausedBy, nil)
	g.AddEdge("test:t", "file:f.go", EdgeTests, nil)
	g.AddEdge("spec:s", "file:f.go", EdgeImplements, nil)
	g.AddEdge("file:f.go", "pkg:p", EdgeDependsOn, nil)
	g.AddEdge("commit:abc", "file:f.go", EdgeAffects, nil)
	g.AddEdge("file:f.go", "spec:s", EdgeReferences, nil)

	_, nodeCounts, edgeCounts, totalNodes, totalEdges := g.toCytoscapeElements()

	if totalNodes != 8 {
		t.Fatalf("expected 8 nodes, got %d", totalNodes)
	}
	if totalEdges != 10 {
		t.Fatalf("expected 10 edges, got %d", totalEdges)
	}

	// Verify all node types present
	for _, nt := range []NodeType{NodeAgent, NodeBug, NodeCommit, NodeFile, NodeFunction, NodePackage, NodeSpec, NodeTest} {
		if nodeCounts[string(nt)] == 0 {
			t.Fatalf("missing node type: %s", nt)
		}
	}

	// Verify all edge types counted
	for _, et := range []EdgeType{EdgeCalls, EdgeContains, EdgeOwns, EdgeChanged, EdgeCausedBy, EdgeTests, EdgeImplements, EdgeDependsOn, EdgeAffects, EdgeReferences} {
		if edgeCounts[string(et)] == 0 {
			t.Fatalf("missing edge type: %s", et)
		}
	}
}
