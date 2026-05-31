package knowledge

import (
	"fmt"
	"sort"
)

// direction constants for traversal
const (
	DirectionOutbound = "outbound"
	DirectionInbound  = "inbound"
	DirectionBoth     = "both"
)

// TraverseOptions controls graph traversal behavior.
type TraverseOptions struct {
	MaxHops   int
	Direction string
	EdgeTypes []EdgeType
	Visitor   func(node *Node, edges []Edge, hop int) bool
}

// BlastRadius returns all nodes reachable from startNodeID within maxHops,
// organized by hop distance (layers). Layers[0] is the start node.
// If edgeTypes is non-empty, only traverses edges of those types.
// direction: "outbound", "inbound", or "both".
func (g *Graph) BlastRadius(startNodeID string, maxHops int, edgeTypes []EdgeType, direction string) ([][]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.Nodes[startNodeID]; !exists {
		return nil, fmt.Errorf("node not found: %s", startNodeID)
	}
	if maxHops < 0 {
		maxHops = 0
	}

	layers := make([][]*Node, 0, maxHops+1)
	visited := map[string]int{startNodeID: 0}
	layers = append(layers, []*Node{g.Nodes[startNodeID]})

	allowedTypes := makeSet(edgeTypes)

	for hop := 1; hop <= maxHops; hop++ {
		prevLayer := layers[hop-1]
		currentLayer := collectNeighbors(g, prevLayer, hop, direction, allowedTypes, visited)
		if len(currentLayer) == 0 {
			break
		}
		layers = append(layers, currentLayer)
	}
	return layers, nil
}

// collectNeighbors gathers all unvisited neighbors from a layer of nodes at the given hop.
func collectNeighbors(g *Graph, nodes []*Node, hop int, direction string, allowedTypes map[string]struct{}, visited map[string]int) []*Node {
	var currentLayer []*Node
	for _, node := range nodes {
		edges := getEdges(g, node.ID, direction)
		for _, e := range edges {
			if !isAllowedEdge(e, allowedTypes) {
				continue
			}
			neighborID := neighborID(e, node.ID, direction)
			if _, seen := visited[neighborID]; !seen {
				visited[neighborID] = hop
				if n, ok := g.Nodes[neighborID]; ok {
					currentLayer = append(currentLayer, n)
				}
			}
		}
	}
	return currentLayer
}

// getEdges returns the edges for a node filtered by direction.
func getEdges(g *Graph, nodeID, direction string) []Edge {
	switch direction {
	case DirectionInbound:
		return g.EdgesIn[nodeID]
	case DirectionBoth:
		return append(append([]Edge{}, g.EdgesOut[nodeID]...), g.EdgesIn[nodeID]...)
	default:
		return g.EdgesOut[nodeID]
	}
}

// isAllowedEdge returns true if the edge type is in the allowed set (or if allowed is empty).
func isAllowedEdge(e Edge, allowedTypes map[string]struct{}) bool {
	if len(allowedTypes) == 0 {
		return true
	}
	_, ok := allowedTypes[string(e.Type)]
	return ok
}

// neighborID returns the neighbor ID for the given edge and traversal direction.
func neighborID(e Edge, nodeID, direction string) string {
	if direction == DirectionInbound {
		return e.From
	}
	if direction == DirectionBoth {
		if e.From == nodeID {
			return e.To
		}
		return e.From
	}
	return e.To
}

// FindPath finds the shortest path(s) from startNodeID to targetNodeID using BFS.
// Returns all shortest paths (there may be multiple at the same length).
func (g *Graph) FindPath(startNodeID, targetNodeID string, maxHops int) ([][]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if err := validateNodes(g, startNodeID, targetNodeID); err != nil {
		return nil, err
	}
	if startNodeID == targetNodeID {
		return [][]*Node{{g.Nodes[startNodeID]}}, nil
	}
	if maxHops <= 0 {
		maxHops = 10
	}

	return bfsShortestPaths(g, startNodeID, targetNodeID, maxHops), nil
}

// validateNodes checks that both start and target nodes exist in the graph.
func validateNodes(g *Graph, startNodeID, targetNodeID string) error {
	if _, exists := g.Nodes[startNodeID]; !exists {
		return fmt.Errorf("start node not found: %s", startNodeID)
	}
	if _, exists := g.Nodes[targetNodeID]; !exists {
		return fmt.Errorf("target node not found: %s", targetNodeID)
	}
	return nil
}

// pathEntry tracks a node and its path during BFS.
type pathEntry struct {
	node *Node
	path []*Node
}

// bfsShortestPaths performs BFS to find all shortest paths to the target.
func bfsShortestPaths(g *Graph, startNodeID, targetNodeID string, maxHops int) [][]*Node {
	queue := []pathEntry{{node: g.Nodes[startNodeID], path: []*Node{g.Nodes[startNodeID]}}}
	visited := map[string]int{startNodeID: 0}
	var foundPaths [][]*Node
	foundAtHop := -1

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]
		currentHop := len(entry.path) - 1

		if (foundAtHop != -1 && currentHop >= foundAtHop) || currentHop >= maxHops {
			continue
		}

		for _, e := range g.EdgesOut[entry.node.ID] {
			if e.To == targetNodeID {
				path := append(append([]*Node{}, entry.path...), g.Nodes[e.To])
				foundPaths = append(foundPaths, path)
				if foundAtHop == -1 {
					foundAtHop = currentHop + 1
				}
				continue
			}
			if _, seen := visited[e.To]; seen {
				continue
			}
			visited[e.To] = currentHop + 1
			queue = append(queue, pathEntry{
				node: g.Nodes[e.To],
				path: append(append([]*Node{}, entry.path...), g.Nodes[e.To]),
			})
		}
	}
	return foundPaths
}

// Traverse walks the graph from startNodeID in BFS order, calling visitor
// for each node at the given hop. If visitor returns false, traversal stops.
func (g *Graph) Traverse(startNodeID string, maxHops int, direction string, edgeTypes []EdgeType, visitor func(node *Node, edges []Edge, hop int) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.Nodes[startNodeID]; !exists {
		return
	}

	opts := TraverseOptions{
		MaxHops:   maxHops,
		Direction: direction,
		EdgeTypes: edgeTypes,
		Visitor:   visitor,
	}
	traverseBFS(g, startNodeID, opts)
}

// traverseBFS performs BFS traversal of the graph starting from startNodeID.
func traverseBFS(g *Graph, startNodeID string, opts TraverseOptions) {
	allowedTypes := makeSet(opts.EdgeTypes)
	visited := map[string]int{startNodeID: 0}
	queue := []bfsEntry{{node: g.Nodes[startNodeID], hop: 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if !opts.Visitor(curr.node, nil, curr.hop) {
			return
		}
		if curr.hop >= opts.MaxHops {
			continue
		}

		for _, e := range getEdges(g, curr.node.ID, opts.Direction) {
			if !isAllowedEdge(e, allowedTypes) {
				continue
			}
			nid := neighborID(e, curr.node.ID, opts.Direction)
			if _, seen := visited[nid]; !seen {
				visited[nid] = curr.hop + 1
				if n, ok := g.Nodes[nid]; ok {
					queue = append(queue, bfsEntry{node: n, hop: curr.hop + 1})
				}
			}
		}
	}
}

// bfsEntry tracks a node and its hop during BFS traversal.
type bfsEntry struct {
	node *Node
	hop  int
}

// makeSet converts a slice to a set map for O(1) lookups.
func makeSet(s []EdgeType) map[string]struct{} {
	if len(s) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(s))
	for _, item := range s {
		set[string(item)] = struct{}{}
	}
	return set
}

// SortNodesByLabel sorts a slice of nodes by their label.
func SortNodesByLabel(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Label < nodes[j].Label
	})
}
