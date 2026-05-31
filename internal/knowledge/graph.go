package knowledge

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// NodeType identifies the kind of knowledge graph node.
type NodeType string

// EdgeType identifies the semantic relationship between two nodes.
type EdgeType string

// Node type constants.
const (
	NodeFile     NodeType = "file"
	NodePackage  NodeType = "package"
	NodeFunction NodeType = "function"
	NodeSpec     NodeType = "spec"
	NodeDecision NodeType = "decision"
	NodeCommit   NodeType = "commit"
	NodeBug      NodeType = "bug"
	NodeTest     NodeType = "test"
	NodeModule   NodeType = "module"
	NodeAgent    NodeType = "agent"
)

// Edge type constants.
const (
	EdgeImports    EdgeType = "IMPORTS"
	EdgeCalls      EdgeType = "CALLS"
	EdgeContains   EdgeType = "CONTAINS"
	EdgeImplements EdgeType = "IMPLEMENTS"
	EdgeReferences EdgeType = "REFERENCES"
	EdgeCausedBy   EdgeType = "CAUSED_BY"
	EdgeTests      EdgeType = "TESTS"
	EdgeDependsOn  EdgeType = "DEPENDS_ON"
	EdgeAffects    EdgeType = "AFFECTS"
	EdgeOwns       EdgeType = "OWNS"
	EdgeChanged    EdgeType = "CHANGED"
)

// Node represents a single node in the knowledge graph.
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Edge represents a directed edge between two graph nodes.
type Edge struct {
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       EdgeType               `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Graph is an in-memory directed knowledge graph. Thread-safe via sync.RWMutex.
type Graph struct {
	mu           sync.RWMutex
	Nodes        map[string]*Node                 `json:"-"`
	EdgesOut     map[string][]Edge                `json:"-"`
	EdgesIn      map[string][]Edge                `json:"-"`
	LastBuilt    time.Time                        `json:"last_built"`
	SourceFiles  map[string]time.Time             `json:"source_files"`
	byType       map[NodeType]map[string]struct{} // node IDs by type for fast lookup
	KnowledgeDir string                           // workspace-relative directory for snapshots and cache
}

// NewGraph creates an empty knowledge graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes:       make(map[string]*Node),
		EdgesOut:    make(map[string][]Edge),
		EdgesIn:     make(map[string][]Edge),
		SourceFiles: make(map[string]time.Time),
		byType:      make(map[NodeType]map[string]struct{}),
	}
}

// AddNode adds a node to the graph. If a node with the same ID exists, it is
// replaced (idempotent for re-ingestion).
func (g *Graph) AddNode(n Node) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If node already exists, clean up its entries first
	if _, exists := g.Nodes[n.ID]; exists {
		oldType := g.Nodes[n.ID].Type
		if oldType != n.Type {
			delete(g.byType[oldType], n.ID)
			if len(g.byType[oldType]) == 0 {
				delete(g.byType, oldType)
			}
		}
	}

	g.Nodes[n.ID] = &n

	if g.byType[n.Type] == nil {
		g.byType[n.Type] = make(map[string]struct{})
	}
	g.byType[n.Type][n.ID] = struct{}{}
}

// RemoveNode removes a node and all its edges from the graph.
func (g *Graph) RemoveNode(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.Nodes[id]
	if !exists {
		return
	}

	// Remove from type index
	delete(g.byType[node.Type], id)
	if len(g.byType[node.Type]) == 0 {
		delete(g.byType, node.Type)
	}

	// Remove all edges involving this node
	for _, out := range g.EdgesOut[id] {
		g.EdgesIn[out.To] = filterEdges(g.EdgesIn[out.To], func(e Edge) bool {
			return e.From == id
		})
	}
	for _, in := range g.EdgesIn[id] {
		g.EdgesOut[in.From] = filterEdges(g.EdgesOut[in.From], func(e Edge) bool {
			return e.To == id
		})
	}

	delete(g.Nodes, id)
	delete(g.EdgesOut, id)
	delete(g.EdgesIn, id)
}

// GetNode returns a node by its ID.
func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.Nodes[id]
	return n, ok
}

// AddEdge creates a directed edge from→to. Duplicate edges are silently ignored.
func (g *Graph) AddEdge(from, to string, edgeType EdgeType, props map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check for existing edge to avoid duplicates
	for _, e := range g.EdgesOut[from] {
		if e.From == from && e.To == to && e.Type == edgeType {
			return // already exists
		}
	}

	edge := Edge{
		From:       from,
		To:         to,
		Type:       edgeType,
		Properties: props,
	}

	g.EdgesOut[from] = append(g.EdgesOut[from], edge)
	g.EdgesIn[to] = append(g.EdgesIn[to], edge)
}

// RemoveEdges removes all edges between two given nodes.
func (g *Graph) RemoveEdges(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.EdgesOut[from] = filterEdges(g.EdgesOut[from], func(e Edge) bool {
		return e.From == from && e.To == to // return true to remove
	})
	g.EdgesIn[to] = filterEdges(g.EdgesIn[to], func(e Edge) bool {
		return e.From == from && e.To == to // return true to remove
	})
}

// RemoveEdgesForNode removes all edges (outbound and inbound) for a given node.
func (g *Graph) RemoveEdgesForNode(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Remove outbound edges of id from the inbound lists of their targets
	for _, out := range g.EdgesOut[id] {
		g.EdgesIn[out.To] = filterEdges(g.EdgesIn[out.To], func(e Edge) bool {
			return e.From == id // return true to remove
		})
	}

	// Remove inbound edges of id from the outbound lists of their sources
	for _, in := range g.EdgesIn[id] {
		g.EdgesOut[in.From] = filterEdges(g.EdgesOut[in.From], func(e Edge) bool {
			return e.To == id // return true to remove
		})
	}

	delete(g.EdgesOut, id)
	delete(g.EdgesIn, id)
}

// GetOutbound returns all outbound edges from a node.
func (g *Graph) GetOutbound(id string) []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	edges := g.EdgesOut[id]
	result := make([]Edge, len(edges))
	copy(result, edges)
	return result
}

// GetInbound returns all inbound edges to a node.
func (g *Graph) GetInbound(id string) []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	edges := g.EdgesIn[id]
	result := make([]Edge, len(edges))
	copy(result, edges)
	return result
}

// FindNodesByType returns all nodes of a given type.
func (g *Graph) FindNodesByType(nt NodeType) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ids, ok := g.byType[nt]
	if !ok {
		return nil
	}
	nodes := make([]*Node, 0, len(ids))
	for id := range ids {
		if n, exists := g.Nodes[id]; exists {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// FindNodeByLabel finds the first node whose Label matches the given string.
func (g *Graph) FindNodeByLabel(label string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, n := range g.Nodes {
		if n.Label == label {
			return n, true
		}
	}
	return nil, false
}

// FindNodeByIDPrefix finds a node whose ID starts with the given prefix.
func (g *Graph) FindNodeByIDPrefix(prefix string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for id, n := range g.Nodes {
		if strings.HasPrefix(id, prefix) {
			return n, true
		}
	}
	return nil, false
}

// SearchNodesByLabel returns all nodes whose label contains the query string (case-insensitive).
func (g *Graph) SearchNodesByLabel(query string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*Node
	for _, n := range g.Nodes {
		if strings.Contains(strings.ToLower(n.Label), query) {
			results = append(results, n)
		}
	}
	return results
}

// SearchNodesByTypeAndLabel returns nodes of a specific type whose label contains the query.
func (g *Graph) SearchNodesByTypeAndLabel(nt NodeType, query string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	query = strings.ToLower(query)
	ids, ok := g.byType[nt]
	if !ok {
		return nil
	}
	var results []*Node
	for id := range ids {
		if n, exists := g.Nodes[id]; exists && strings.Contains(strings.ToLower(n.Label), query) {
			results = append(results, n)
		}
	}
	return results
}

// Stats returns the number of nodes, edges, and node types in the graph.
func (g *Graph) Stats() (nodes int, edges int, types int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes = len(g.Nodes)
	for _, out := range g.EdgesOut {
		edges += len(out)
	}
	types = len(g.byType)
	return
}

// Logf logs a message with the [knowledge] prefix.
func Logf(format string, v ...interface{}) {
	log.Printf("[knowledge] "+format, v...)
}

// slug builds a node ID from a prefix and optional parts, joining with ":".
func slug(prefix string, parts ...string) string {
	return prefix + ":" + strings.Join(parts, ".")
}

// MustGetString extracts a string from a map or returns the default.
func mustGetString(m map[string]interface{}, key, def string) string {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	return s
}

// mustMarshalJSON is a safe JSON marshal helper for internal use.
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(data)
}

// isJSON checks if a string looks like JSON.
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

func filterEdges(edges []Edge, fn func(Edge) bool) []Edge {
	result := make([]Edge, 0, len(edges))
	for _, e := range edges {
		if !fn(e) {
			result = append(result, e)
		}
	}
	return result
}

// NodeType returns the NodeType from a string.
func ParseNodeType(s string) NodeType {
	return NodeType(s)
}
