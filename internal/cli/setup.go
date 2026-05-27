package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
)

type agentSetup struct {
	name     string
	dir      string
	files    map[string]string
	postHook func(cwd string) string
}

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

func selectAgents(scanner *bufio.Scanner) []string {
	fmt.Println()
	fmt.Println("Select AI agents to configure (comma-separated):")
	fmt.Println("  opencode  — OpenCode subagent + skill + MCP config")
	fmt.Println("  claude    — Claude Code CLAUDE.md + MCP config")
	fmt.Println("  cursor    — Cursor rules + MCP config")
	fmt.Println("  copilot   — GitHub Copilot instructions")
	fmt.Println("  all       — All of the above")
	fmt.Print("Agents (all): ")

	if !scanner.Scan() {
		return allAgentNames()
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" || strings.ToLower(input) == "all" {
		return allAgentNames()
	}

	selected := make(map[string]bool)
	for _, name := range strings.Split(input, ",") {
		name = strings.TrimSpace(strings.ToLower(name))
		switch name {
		case "opencode", "open", "oc":
			selected["opencode"] = true
		case "claude", "claude-code", "cc":
			selected["claude"] = true
		case "cursor", "cur":
			selected["cursor"] = true
		case "copilot", "github", "gh", "cp":
			selected["copilot"] = true
		}
	}

	if len(selected) == 0 {
		fmt.Println("No valid agents selected — using 'all'.")
		return allAgentNames()
	}

	var result []string
	for _, name := range []string{"opencode", "claude", "cursor", "copilot"} {
		if selected[name] {
			result = append(result, name)
		}
	}
	return result
}

func allAgentNames() []string {
	return []string{"opencode", "claude", "cursor", "copilot"}
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

func writeAgentFiles(cwd, agent string, result *setupResult) {
	defs := getAgentDefs(cwd, agent)
	for _, def := range defs {
		dir := filepath.Dir(def.path)
		if dir != cwd {
			os.MkdirAll(dir, 0755)
		}
		if _, err := os.Stat(def.path); err == nil {
			result.skipped = append(result.skipped, rel(cwd, def.path))
			continue
		}
		os.WriteFile(def.path, []byte(def.content), 0644)
		result.created = append(result.created, rel(cwd, def.path))
	}
}

type fileDef struct {
	path    string
	content string
}

func getAgentDefs(cwd, agent string) []fileDef {
	var defs []fileDef
	switch agent {
	case "opencode":
		defs = []fileDef{
			{filepath.Join(cwd, "opencode.json"), opencodeMCPConfig},
			{filepath.Join(cwd, ".opencode", "agent", "ailinter.md"), opencodeAgentConfig},
			{filepath.Join(cwd, ".opencode", "skills", "ailinter", "SKILL.md"), opencodeSkill},
		}
	case "claude":
		defs = []fileDef{
			{filepath.Join(cwd, ".claude", "settings.json"), claudeMCPConfig},
			{filepath.Join(cwd, "CLAUDE.md"), claudeInstructions},
		}
	case "cursor":
		defs = []fileDef{
			{filepath.Join(cwd, ".cursor", "mcp.json"), cursorMCPConfig},
			{filepath.Join(cwd, ".cursor", "rules", "ailinter.mdc"), cursorRules},
		}
	case "copilot":
		defs = []fileDef{
			{filepath.Join(cwd, ".github", "copilot-instructions.md"), copilotInstructions},
		}
	}
	return defs
}

func writeVSCodeFiles(cwd string, result *setupResult) {
	vsDir := filepath.Join(cwd, ".vscode")
	os.MkdirAll(vsDir, 0755)

	tasksPath := filepath.Join(vsDir, "tasks.json")
	if _, err := os.Stat(tasksPath); err == nil {
		result.skipped = append(result.skipped, rel(cwd, tasksPath))
	} else {
		os.WriteFile(tasksPath, []byte(defaultVSCodeTasks), 0644)
		result.created = append(result.created, rel(cwd, tasksPath))
	}

	settingsPath := filepath.Join(vsDir, "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		result.skipped = append(result.skipped, rel(cwd, settingsPath))
	} else {
		os.WriteFile(settingsPath, []byte(vscodeSettings), 0644)
		result.created = append(result.created, rel(cwd, settingsPath))
	}

	extPath := filepath.Join(vsDir, "extensions.json")
	if _, err := os.Stat(extPath); err == nil {
		result.skipped = append(result.skipped, rel(cwd, extPath))
	} else {
		os.WriteFile(extPath, []byte(vscodeExtensions), 0644)
		result.created = append(result.created, rel(cwd, extPath))
	}
}

func writeGitHook(cwd string, result *setupResult) {
	githooksDir := filepath.Join(cwd, ".githooks")
	os.MkdirAll(githooksDir, 0755)

	hookPath := filepath.Join(githooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		result.skipped = append(result.skipped, rel(cwd, hookPath))
		return
	}
	os.WriteFile(hookPath, []byte(gitPreCommitHook), 0755)
	result.created = append(result.created, rel(cwd, hookPath))

	gitignorePath := filepath.Join(cwd, ".gitignore")
	hasHookIgnore := false
	if data, err := os.ReadFile(gitignorePath); err == nil {
		hasHookIgnore = strings.Contains(string(data), ".githooks")
	}

	if !hasHookIgnore {
		if _, err := os.Stat(gitignorePath); err == nil {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString("\n# ailinter hooks\n.githooks/\n")
				f.Close()
			}
		} else {
			os.WriteFile(gitignorePath, []byte("# ailinter hooks\n.githooks/\n"), 0644)
		}
	}
}

func rel(cwd, path string) string {
	r, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return r
}

func printResult(result *setupResult) {
	for _, f := range result.created {
		fmt.Printf("  created  %s\n", f)
	}
	for _, f := range result.skipped {
		fmt.Printf("  skipped  %s (already exists)\n", f)
	}
}
