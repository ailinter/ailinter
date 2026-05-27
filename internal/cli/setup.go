package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
)

type thresholdProfile struct {
	name   string
	config string
}

var profiles = []thresholdProfile{
	{"default", defaultConfig},
	{"strict", strictConfig},
	{"relaxed", relaxedConfig},
}

type setupResult struct {
	created []string
	skipped []string
}

func isInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
}

func runInteractiveSetup(cwd string, skipAgents bool, withVSCode bool) error {
	scanner := bufio.NewScanner(os.Stdin)
	result := &setupResult{}

	fmt.Println()
	fmt.Println("  ailinter Setup")
	fmt.Println("  ──────────────")
	fmt.Println("  Set up Code Quality + Security for your AI tools.")
	fmt.Println()

	profileName := selectProfile(scanner)
	selectedAgents := selectAgents(scanner)
	setupHook := askConfirm(scanner, "Set up git pre-commit hook? (y/N)")

	fmt.Println()
	fmt.Println("Creating files...")
	fmt.Println()

	writeConfig(cwd, profileName, result)
	writeAgentsMD(cwd, skipAgents, result)

	for _, agent := range selectedAgents {
		writeAgentFiles(cwd, agent, result)
	}

	if withVSCode {
		writeVSCodeFiles(cwd, result)
	}

	if setupHook {
		writeGitHook(cwd, result)
	}

	printResult(result)
	fmt.Println()
	fmt.Println("To enable git hooks: git config core.hooksPath .githooks")
	fmt.Println()
	fmt.Println("ailinter initialized! Run 'ailinter check .' to analyze your codebase.")
	fmt.Println()
	return nil
}

func selectProfile(scanner *bufio.Scanner) string {
	fmt.Println("Select threshold profile:")
	fmt.Println("  1. Default (balanced)")
	fmt.Println("  2. Strict (catches more)")
	fmt.Println("  3. Relaxed (fewer warnings)")

	for {
		fmt.Print("Choice [1]: ")
		if !scanner.Scan() {
			return "default"
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" || input == "1" {
			return "default"
		}
		if input == "2" {
			return "strict"
		}
		if input == "3" {
			return "relaxed"
		}
		fmt.Println("Please enter 1, 2, or 3.")
	}
}

type agentKind string

const (
	agentOpenCode agentKind = "opencode"
	agentClaude   agentKind = "claude"
	agentCursor   agentKind = "cursor"
	agentCopilot  agentKind = "copilot"
)

var allAgents = []agentKind{agentOpenCode, agentClaude, agentCursor, agentCopilot}

func allAgentNames() []string {
	return []string{"opencode", "claude", "cursor", "copilot"}
}

var agentAliases = map[string]agentKind{
	"opencode":    agentOpenCode,
	"open":        agentOpenCode,
	"oc":          agentOpenCode,
	"claude":      agentClaude,
	"claude-code": agentClaude,
	"cc":          agentClaude,
	"cursor":      agentCursor,
	"cur":         agentCursor,
	"copilot":     agentCopilot,
	"github":      agentCopilot,
	"gh":          agentCopilot,
	"cp":          agentCopilot,
}

func parseAgentInput(input string) []agentKind {
	seen := make(map[agentKind]bool)
	var result []agentKind
	for _, name := range strings.Split(input, ",") {
		name = strings.TrimSpace(strings.ToLower(name))
		if canonical, ok := agentAliases[name]; ok && !seen[canonical] {
			seen[canonical] = true
			result = append(result, canonical)
		}
	}
	return result
}

func selectAgents(scanner *bufio.Scanner) []agentKind {
	fmt.Println()
	fmt.Println("Select AI agents to configure (comma-separated):")
	fmt.Println("  opencode  — OpenCode subagent + skill + MCP config")
	fmt.Println("  claude    — Claude Code CLAUDE.md + MCP config")
	fmt.Println("  cursor    — Cursor rules + MCP config")
	fmt.Println("  copilot   — GitHub Copilot instructions")
	fmt.Println("  all       — All of the above")
	fmt.Print("Agents (all): ")

	if !scanner.Scan() {
		return allAgents
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" || strings.ToLower(input) == "all" {
		return allAgents
	}

	selected := parseAgentInput(input)
	if len(selected) == 0 {
		fmt.Println("No valid agents selected — using 'all'.")
		return allAgents
	}
	return selected
}

func askConfirm(scanner *bufio.Scanner, prompt string) bool {
	fmt.Println()
	fmt.Print(prompt + " ")
	if !scanner.Scan() {
		return false
	}
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "y" || input == "yes"
}

func writeConfig(cwd, profileName string, result *setupResult) {
	configPath := filepath.Join(cwd, ".ailinter.toml")
	if _, err := os.Stat(configPath); err == nil {
		result.skipped = append(result.skipped, ".ailinter.toml")
		return
	}
	content := defaultConfig
	for _, p := range profiles {
		if p.name == profileName {
			content = p.config
			break
		}
	}
	os.WriteFile(configPath, []byte(content), 0644)
	result.created = append(result.created, ".ailinter.toml")
}

func writeAgentsMD(cwd string, skip bool, result *setupResult) {
	if skip {
		return
	}
	path := filepath.Join(cwd, "AGENTS.md")
	if _, err := os.Stat(path); err == nil {
		result.skipped = append(result.skipped, "AGENTS.md")
		return
	}
	os.WriteFile(path, []byte(defaultAgentsMD), 0644)
	result.created = append(result.created, "AGENTS.md")
}

func relPath(cwd, path string) string {
	r, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return r
}

func writeAgentFiles(cwd string, agent agentKind, result *setupResult) {
	defs := getAgentDefs(cwd, agent)
	for _, def := range defs {
		dir := filepath.Dir(def.path)
		if dir != cwd {
			os.MkdirAll(dir, 0755)
		}
		if _, err := os.Stat(def.path); err == nil {
			result.skipped = append(result.skipped, relPath(cwd, def.path))
			continue
		}
		os.WriteFile(def.path, []byte(def.content), 0644)
		result.created = append(result.created, relPath(cwd, def.path))
	}
}

type fileDef struct {
	path    string
	content string
}

func getAgentDefs(cwd string, agent agentKind) []fileDef {
	switch agent {
	case agentOpenCode:
		return []fileDef{
			{filepath.Join(cwd, "opencode.json"), opencodeMCPConfig},
			{filepath.Join(cwd, ".opencode", "agent", "ailinter.md"), opencodeAgentConfig},
			{filepath.Join(cwd, ".opencode", "skills", "ailinter", "SKILL.md"), opencodeSkill},
		}
	case agentClaude:
		return []fileDef{
			{filepath.Join(cwd, ".claude", "settings.json"), claudeMCPConfig},
			{filepath.Join(cwd, "CLAUDE.md"), claudeInstructions},
		}
	case agentCursor:
		return []fileDef{
			{filepath.Join(cwd, ".cursor", "mcp.json"), cursorMCPConfig},
			{filepath.Join(cwd, ".cursor", "rules", "ailinter.mdc"), cursorRules},
		}
	case agentCopilot:
		return []fileDef{
			{filepath.Join(cwd, ".github", "copilot-instructions.md"), copilotInstructions},
		}
	}
	return nil
}

func writeVSCodeFiles(cwd string, result *setupResult) {
	vsDir := filepath.Join(cwd, ".vscode")
	os.MkdirAll(vsDir, 0755)

	tasksPath := filepath.Join(vsDir, "tasks.json")
	if _, err := os.Stat(tasksPath); err == nil {
		result.skipped = append(result.skipped, relPath(cwd, tasksPath))
	} else {
		os.WriteFile(tasksPath, []byte(defaultVSCodeTasks), 0644)
		result.created = append(result.created, relPath(cwd, tasksPath))
	}

	settingsPath := filepath.Join(vsDir, "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		result.skipped = append(result.skipped, relPath(cwd, settingsPath))
	} else {
		os.WriteFile(settingsPath, []byte(vscodeSettings), 0644)
		result.created = append(result.created, relPath(cwd, settingsPath))
	}

	extPath := filepath.Join(vsDir, "extensions.json")
	if _, err := os.Stat(extPath); err == nil {
		result.skipped = append(result.skipped, relPath(cwd, extPath))
	} else {
		os.WriteFile(extPath, []byte(vscodeExtensions), 0644)
		result.created = append(result.created, relPath(cwd, extPath))
	}
}

func writeGitHook(cwd string, result *setupResult) {
	githooksDir := filepath.Join(cwd, ".githooks")
	os.MkdirAll(githooksDir, 0755)

	hookPath := filepath.Join(githooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		result.skipped = append(result.skipped, relPath(cwd, hookPath))
		return
	}
	os.WriteFile(hookPath, []byte(gitPreCommitHook), 0755)
	result.created = append(result.created, relPath(cwd, hookPath))

	updateGitignoreForHooks(cwd)
}

func updateGitignoreForHooks(cwd string) {
	gitignorePath := filepath.Join(cwd, ".gitignore")
	if data, err := os.ReadFile(gitignorePath); err == nil {
		if strings.Contains(string(data), ".githooks") {
			return
		}
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString("\n# ailinter hooks\n.githooks/\n")
			f.Close()
		}
	} else {
		os.WriteFile(gitignorePath, []byte("# ailinter hooks\n.githooks/\n"), 0644)
	}
}

func printResult(result *setupResult) {
	for _, f := range result.created {
		fmt.Printf("  created  %s\n", f)
	}
	for _, f := range result.skipped {
		fmt.Printf("  skipped  %s (already exists)\n", f)
	}
}
