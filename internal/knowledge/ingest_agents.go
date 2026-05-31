package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// opencodeConfig represents the structure of opencode.json for agent parsing.
type opencodeConfig struct {
	Agent map[string]opencodeAgent `json:"agent"`
}

// opencodeAgent represents a single agent configuration in opencode.json.
type opencodeAgent struct {
	Description string `json:"description"`
	Mode        string `json:"mode,omitempty"`
	Prompt      string `json:"prompt,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ownershipRule maps an agent name to the packages/modules it owns.
type ownershipRule struct {
	AgentName string
	Modules   []string // package prefixes this agent owns
}

// defaultOwnershipRules defines known ownership relationships for AILINTER agents.
var defaultOwnershipRules = []ownershipRule{
	{AgentName: "quality-guardian", Modules: []string{"internal/analyzer", "internal/metalinter"}},
	{AgentName: "software-engineer", Modules: []string{"internal/cli", "internal/mcp", "internal/knowledge", "internal/parser"}},
	{AgentName: "devops-engineer", Modules: []string{"internal/telemetry"}},
	{AgentName: "qa-engineer", Modules: []string{"internal/analyzer/test", "internal/mcp/test"}},
	{AgentName: "security-engineer", Modules: []string{"internal/secrets", "internal/vulnerability"}},
	{AgentName: "product-manager", Modules: []string{"spec", "docs", "roadmap"}},
	{AgentName: "marketing-director", Modules: []string{"marketing", "blog", "community"}},
	{AgentName: "research-analyst", Modules: []string{"research", "competitive", "analysis"}},
	{AgentName: "developer-advocate", Modules: []string{"docs", "examples", "integrations"}},
	{AgentName: "ci-guardian", Modules: []string{".github", "ci", "pipeline"}},
}

// IngestAgentOwnership reads agent configuration and creates ownership edges in the graph.
func IngestAgentOwnership(graph *Graph, opencodeJSONPath string) error {
	Logf("ingesting agent ownership from %s", opencodeJSONPath)

	data, err := os.ReadFile(opencodeJSONPath)
	if err != nil {
		Logf("warning: could not read opencode.json at %s: %v", opencodeJSONPath, err)
		Logf("using default ownership rules instead")
		applyDefaultOwnership(graph)
		return nil
	}

	var cfg opencodeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		Logf("warning: could not parse opencode.json: %v, using defaults", err)
		applyDefaultOwnership(graph)
		return nil
	}

	for agentName, agentCfg := range cfg.Agent {
		agentNodeID := slug("agent", agentName)
		if _, exists := graph.GetNode(agentNodeID); exists {
			continue
		}

		desc := truncate(agentCfg.Description, 100)
		graph.AddNode(Node{
			ID:    agentNodeID,
			Type:  NodeAgent,
			Label: agentName,
			Properties: map[string]interface{}{
				"description": desc,
				"mode":        agentCfg.Mode,
				"color":       agentCfg.Color,
			},
		})

		// Infer ownership from description keywords
		ownedPrefixes := inferOwnershipFromDescription(agentName, agentCfg.Description)

		// Also check default rules
		for _, rule := range defaultOwnershipRules {
			if rule.AgentName == agentName {
				ownedPrefixes = append(ownedPrefixes, rule.Modules...)
			}
		}

		// Create OWNS edges
		for _, prefix := range ownedPrefixes {
			prefix = strings.TrimPrefix(prefix, "./")
			prefix = strings.TrimPrefix(prefix, "/")

			// Check for matching nodes
			for _, node := range graph.Nodes {
				label := strings.ToLower(node.Label)

				// Match against file paths, package names, or node labels
				if strings.Contains(label, strings.ToLower(prefix)) || strings.HasPrefix(label, strings.ToLower(prefix)) {
					graph.AddEdge(agentNodeID, node.ID, EdgeOwns, map[string]interface{}{
						"rule": prefix,
					})
				}
			}
		}
	}

	return nil
}

// applyDefaultOwnership applies default ownership rules without parsing opencode.json.
func applyDefaultOwnership(graph *Graph) {
	for _, rule := range defaultOwnershipRules {
		agentNodeID := slug("agent", rule.AgentName)
		if _, exists := graph.GetNode(agentNodeID); exists {
			continue
		}

		graph.AddNode(Node{
			ID:    agentNodeID,
			Type:  NodeAgent,
			Label: rule.AgentName,
			Properties: map[string]interface{}{
				"description": fmt.Sprintf("Owns %s", strings.Join(rule.Modules, ", ")),
				"mode":        "subagent",
			},
		})

		for _, module := range rule.Modules {
			for _, node := range graph.Nodes {
				if strings.Contains(strings.ToLower(node.Label), strings.ToLower(module)) {
					graph.AddEdge(agentNodeID, node.ID, EdgeOwns, nil)
				}
			}
		}
	}
}

// inferOwnershipFromDescription tries to determine owned code areas from an agent's description.
func inferOwnershipFromDescription(agentName, description string) []string {
	var modules []string
	desc := strings.ToLower(description)

	keywords := map[string][]string{
		"code quality":   {"internal/analyzer", "internal/metalinter"},
		"refactoring":    {"internal/refactoring"},
		"infrastructure": {"internal/telemetry", "deploy", "ci"},
		"security":       {"internal/secrets", "internal/vulnerability"},
		"test":           {"_test.go", "testdata"},
		"cli":            {"internal/cli"},
		"mcp":            {"internal/mcp"},
		"parser":         {"internal/parser"},
		"documentation":  {"docs", "readme", "spec"},
		"marketing":      {"marketing", "landing", "blog"},
		"research":       {"research", "competitive"},
		"ci":             {".github", "ci"},
	}

	for keyword, prefixes := range keywords {
		if strings.Contains(desc, keyword) {
			modules = append(modules, prefixes...)
		}
	}

	return modules
}
