package refactoring

import (
	"embed"
	"fmt"
)

//go:embed patterns/*.md
var patternsFS embed.FS

// Pattern represents a refactoring strategy.
type Pattern struct {
	Name        string
	Description string
	Content     string // full markdown
}

// Patterns is a map from smell name to refactoring pattern.
var Patterns map[string]Pattern

func init() {
	Patterns = make(map[string]Pattern)
	entries, err := patternsFS.ReadDir("patterns")
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded patterns: %v", err))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := patternsFS.ReadFile("patterns/" + entry.Name())
		if err != nil {
			continue
		}
		name := entry.Name()
		name = name[:len(name)-3] // strip .md
		Patterns[name] = Pattern{
			Name:    name,
			Content: string(data),
		}
	}
}

// Lookup returns the refactoring pattern for a given smell, or nil.
func Lookup(smellName string) *Pattern {
	p, ok := Patterns[smellName]
	if !ok {
		return nil
	}
	return &p
}

// ListPatterns returns all available pattern names.
func ListPatterns() []string {
	names := make([]string, 0, len(Patterns))
	for name := range Patterns {
		names = append(names, name)
	}
	return names
}
