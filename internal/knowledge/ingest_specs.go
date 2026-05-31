package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// filePathPattern matches code file references in markdown: internal/something.go, cmd/ailinter/main.go, etc.
	filePathPattern = regexp.MustCompile(`\b((?:internal|cmd|pkg)/\S+\.(?:go|ts|tsx|js|py|rs|toml|yaml|yml|json))\b`)

	// pkgPattern matches Go package references like "analyzer package" or "config package"
	pkgPattern = regexp.MustCompile(`(?i)\b(\w+)\s+package\b`)

	// decisionPattern matches decision references: "DECISION:", "decision ID:", or "DEC-123"
	decisionPattern = regexp.MustCompile(`(?i)\b(DECISION|DEC|decision)[:\s]*\s*(\S+)?`)

	// headingPattern matches markdown headings
	headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
)

// IngestSpecs ingests all markdown spec files from the given directories.
func IngestSpecs(graph *Graph, dirs []string) error {
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			Logf("spec directory not found: %s", dir)
			continue
		}
		if !info.IsDir() {
			Logf("not a directory: %s", dir)
			continue
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			return ingestSpecFile(graph, path)
		})
		if err != nil {
			Logf("warning: walk specs %s: %v", dir, err)
		}
	}
	return nil
}

// ingestSpecFile parses a single markdown spec file and adds its nodes to the graph.
func ingestSpecFile(graph *Graph, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read spec file: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return nil
	}

	// Determine a relative display path
	relPath := path

	title := extractTitle(content, path)
	specNodeID := slug("spec", sanitizeID(title))

	// Check if this spec node already exists (idempotency)
	if _, exists := graph.GetNode(specNodeID); exists {
		return nil
	}

	graph.AddNode(Node{
		ID:    specNodeID,
		Type:  NodeSpec,
		Label: title,
		Properties: map[string]interface{}{
			"path": relPath,
			"type": "spec",
		},
	})

	// Extract subsections (## headings)
	subsectionCount := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		matches := headingPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		level := len(matches[1])
		if level != 2 { // only ## headings as subsections
			continue
		}
		subTitle := matches[2]
		subID := slug("spec", sanitizeID(title), sanitizeID(subTitle))
		if _, exists := graph.GetNode(subID); exists {
			continue
		}
		graph.AddNode(Node{
			ID:    subID,
			Type:  NodeSpec,
			Label: fmt.Sprintf("%s > %s", title, subTitle),
			Properties: map[string]interface{}{
				"path":       relPath,
				"type":       "subsection",
				"parent":     title,
				"subsection": subTitle,
			},
		})
		graph.AddEdge(specNodeID, subID, EdgeContains, nil)
		subsectionCount++
	}

	// Link to code files
	fileMatches := filePathPattern.FindAllStringSubmatch(content, -1)
	seenFiles := make(map[string]bool)
	for _, m := range fileMatches {
		filePath := m[1]
		if seenFiles[filePath] {
			continue
		}
		seenFiles[filePath] = true

		fileNodeID := slug("file", filePath)
		if _, exists := graph.GetNode(fileNodeID); exists {
			graph.AddEdge(specNodeID, fileNodeID, EdgeImplements, nil)
		}
	}

	// Link to packages
	pkgMatches := pkgPattern.FindAllStringSubmatch(content, -1)
	seenPkgs := make(map[string]bool)
	for _, m := range pkgMatches {
		pkgName := strings.ToLower(m[1])
		if seenPkgs[pkgName] {
			continue
		}
		seenPkgs[pkgName] = true

		pkgNodeID := slug("pkg", pkgName)
		if _, exists := graph.GetNode(pkgNodeID); exists {
			graph.AddEdge(specNodeID, pkgNodeID, EdgeImplements, nil)
		}
	}

	// Link to decisions
	decMatches := decisionPattern.FindAllStringSubmatch(content, -1)
	seenDecs := make(map[string]bool)
	for _, m := range decMatches {
		decID := strings.TrimSpace(m[2])
		if decID == "" {
			continue
		}
		if seenDecs[decID] {
			continue
		}
		seenDecs[decID] = true

		// Search for decision nodes with matching label or content
		for _, node := range graph.FindNodesByType(NodeDecision) {
			if strings.Contains(node.Label, decID) {
				graph.AddEdge(specNodeID, node.ID, EdgeReferences, nil)
			}
		}
	}

	return nil
}

// extractTitle gets the title from a markdown file (# heading) or uses filename.
func extractTitle(content, path string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	// Fall back to filename without extension
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// sanitizeID replaces characters not suitable for node IDs.
func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, "(", "-")
	s = strings.ReplaceAll(s, ")", "-")
	s = strings.ReplaceAll(s, "[", "-")
	s = strings.ReplaceAll(s, "]", "-")
	s = strings.ReplaceAll(s, "{", "-")
	s = strings.ReplaceAll(s, "}", "-")
	s = strings.ReplaceAll(s, "'", "-")
	s = strings.ReplaceAll(s, "\"", "-")
	s = strings.ReplaceAll(s, ",", "-")

	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	return strings.Trim(s, "-")
}
