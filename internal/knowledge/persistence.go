package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SnapshotNode is a JSON-serializable representation of a graph Node.
type SnapshotNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// SnapshotEdge is a JSON-serializable representation of a graph Edge.
type SnapshotEdge struct {
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// graphSnapshot is the top-level JSON structure for persistence.
type graphSnapshot struct {
	Nodes       []SnapshotNode       `json:"nodes"`
	Edges       []SnapshotEdge       `json:"edges"`
	LastBuilt   time.Time            `json:"last_built"`
	SourceFiles map[string]time.Time `json:"source_files"`
}

// ExportJSON serializes the graph to a JSON file.
func (g *Graph) ExportJSON(path string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	snap := g.buildSnapshot()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	return nil
}

// buildSnapshot creates a graphSnapshot from the graph's current state.
// Caller must hold g.mu read lock.
func (g *Graph) buildSnapshot() graphSnapshot {
	snap := graphSnapshot{
		LastBuilt:   g.LastBuilt,
		SourceFiles: g.SourceFiles,
	}
	for _, id := range sortedNodeIDs(g) {
		n := g.Nodes[id]
		snap.Nodes = append(snap.Nodes, snapshotNode(n))
		snap.Edges = append(snap.Edges, snapshotEdges(g, id)...)
	}
	return snap
}

// snapshotNode converts a graph Node to a serializable SnapshotNode.
func snapshotNode(n *Node) SnapshotNode {
	return SnapshotNode{
		ID:         n.ID,
		Type:       string(n.Type),
		Label:      n.Label,
		Properties: copyProps(n.Properties),
	}
}

// snapshotEdges converts all outbound edges of a node to serializable SnapshotEdges,
// deduplicating by (from, to, type).
func snapshotEdges(g *Graph, nodeID string) []SnapshotEdge {
	var edges []SnapshotEdge
	seen := make(map[string]bool)
	for _, e := range g.EdgesOut[nodeID] {
		key := edgeKey(e)
		if seen[key] {
			continue
		}
		seen[key] = true
		edges = append(edges, SnapshotEdge{
			From:       e.From,
			To:         e.To,
			Type:       string(e.Type),
			Properties: copyProps(e.Properties),
		})
	}
	return edges
}

// sortedNodeIDs returns all node IDs sorted for deterministic output.
func sortedNodeIDs(g *Graph) []string {
	nodeIDs := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	return nodeIDs
}

// edgeKey returns a deterministic key for deduplicating edges.
func edgeKey(e Edge) string {
	return e.From + "→" + e.To + "→" + string(e.Type)
}

// LoadGraphJSON loads a graph from a JSON snapshot file.
func LoadGraphJSON(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}

	var snap graphSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}

	g := NewGraph()
	g.LastBuilt = snap.LastBuilt
	g.SourceFiles = snap.SourceFiles

	restoreNodes(g, snap.Nodes)
	restoreEdges(g, snap.Edges)

	return g, nil
}

// restoreNodes adds nodes from a snapshot into the graph.
func restoreNodes(g *Graph, nodes []SnapshotNode) {
	for _, sn := range nodes {
		g.Nodes[sn.ID] = &Node{
			ID:         sn.ID,
			Type:       NodeType(sn.Type),
			Label:      sn.Label,
			Properties: sn.Properties,
		}
		nt := NodeType(sn.Type)
		if g.byType[nt] == nil {
			g.byType[nt] = make(map[string]struct{})
		}
		g.byType[nt][sn.ID] = struct{}{}
	}
}

// restoreEdges adds edges from a snapshot into the graph.
func restoreEdges(g *Graph, edges []SnapshotEdge) {
	for _, se := range edges {
		edge := Edge{
			From:       se.From,
			To:         se.To,
			Type:       EdgeType(se.Type),
			Properties: se.Properties,
		}
		g.EdgesOut[se.From] = append(g.EdgesOut[se.From], edge)
		g.EdgesIn[se.To] = append(g.EdgesIn[se.To], edge)
	}
}

// NeedsRebuild returns true if any source file has been modified since the
// graph was last built, or if no source files are tracked.
func (g *Graph) NeedsRebuild() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.SourceFiles) == 0 {
		return true
	}

	for path, mtime := range g.SourceFiles {
		info, err := os.Stat(path)
		if err != nil {
			return true // file disappeared
		}
		// Truncate to second precision: many filesystems don't preserve
		// nanosecond timestamps, causing false-positive staleness.
		if info.ModTime().Truncate(time.Second).After(mtime.Truncate(time.Second)) {
			return true
		}
	}
	return false
}

// snapshotPath returns the path to the serialized graph snapshot inside the workspace.
func (g *Graph) snapshotPath() string {
	return filepath.Join(g.KnowledgeDir, "snapshot.json")
}

// gitCachePath returns the path to the git history cache inside the workspace.
func (g *Graph) gitCachePath() string {
	return filepath.Join(g.KnowledgeDir, "git-cache.json")
}

// ensureKnowledgeDir creates the workspace knowledge directory.
func (g *Graph) ensureKnowledgeDir() error {
	if g.KnowledgeDir == "" {
		return fmt.Errorf("knowledge dir not set on graph")
	}
	if err := os.MkdirAll(g.KnowledgeDir, 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", g.KnowledgeDir, err)
	}
	return nil
}

func copyProps(props map[string]interface{}) map[string]interface{} {
	if props == nil {
		return nil
	}
	c := make(map[string]interface{}, len(props))
	for k, v := range props {
		c[k] = v
	}
	return c
}
