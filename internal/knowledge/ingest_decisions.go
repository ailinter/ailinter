package knowledge

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DecisionEntry represents a decision stored in the RAG memory system.
type DecisionEntry struct {
	ID       string    `json:"id"`
	Content  string    `json:"content"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
}

var (
	// fileRefPattern matches file path references in decision text
	fileRefPattern = regexp.MustCompile(`\b((?:internal|cmd|pkg)/\S+\.(?:go|ts|tsx|js|py|rs|toml|yaml|yml|json))\b`)

	// pkgRefPattern matches Go package references
	pkgRefPattern = regexp.MustCompile(`\b(\w+)\.(\w+)\s*(?:package|module|function)\b`)
)

// IngestDecisions ingests decision entries from the RAG memory system into the graph.
func IngestDecisions(graph *Graph, entries []DecisionEntry) error {
	for _, entry := range entries {
		if entry.ID == "" || entry.Content == "" {
			continue
		}

		summary := truncate(entry.Content, 80)
		decNodeID := slug("decision", sanitizeID(entry.ID))

		// Skip if already ingested (idempotent)
		if _, exists := graph.GetNode(decNodeID); exists {
			continue
		}

		graph.AddNode(Node{
			ID:    decNodeID,
			Type:  NodeDecision,
			Label: summary,
			Properties: map[string]interface{}{
				"category": entry.Category,
				"date":     entry.Date.Format(time.RFC3339),
				"content":  entry.Content,
			},
		})

		// Reference code files mentioned in the decision
		fileMatches := fileRefPattern.FindAllStringSubmatch(entry.Content, -1)
		seenFiles := make(map[string]bool)
		for _, m := range fileMatches {
			filePath := m[1]
			if seenFiles[filePath] {
				continue
			}
			seenFiles[filePath] = true

			fileNodeID := slug("file", filePath)
			if _, exists := graph.GetNode(fileNodeID); exists {
				graph.AddEdge(decNodeID, fileNodeID, EdgeReferences, nil)
			}
		}

		// Reference packages mentioned in the decision
		pkgMatches := pkgRefPattern.FindAllStringSubmatch(entry.Content, -1)
		seenPkgs := make(map[string]bool)
		for _, m := range pkgMatches {
			pkgName := strings.ToLower(m[1])
			if seenPkgs[pkgName] {
				continue
			}
			seenPkgs[pkgName] = true

			pkgNodeID := slug("pkg", pkgName)
			if _, exists := graph.GetNode(pkgNodeID); exists {
				graph.AddEdge(decNodeID, pkgNodeID, EdgeReferences, nil)
			}
		}

		// Cross-reference with spec nodes: search for spec titles in decision text
		for _, spec := range graph.FindNodesByType(NodeSpec) {
			if spec.Properties == nil {
				continue
			}
			specTitle := mustGetString(spec.Properties, "subsection", "")
			if specTitle == "" {
				specTitle = spec.Label
			}
			if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(specTitle)) {
				graph.AddEdge(decNodeID, spec.ID, EdgeReferences, nil)
			}
		}
	}

	return nil
}

// truncate truncates a string to the given maximum length.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// decodeDecisionID provides a unique decision ID based on index and category.
func decodeDecisionID(index int, category string) string {
	return fmt.Sprintf("dec-%s-%04d", category, index)
}
