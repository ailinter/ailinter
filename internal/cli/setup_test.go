package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/git"
)

// ──────────────────────────────────────────────
// Pure function tests (no IO, no scanner)
// ──────────────────────────────────────────────

func TestAllAgentNames(t *testing.T) {
	names := allAgentNames()
	expected := []string{"opencode", "claude", "cursor", "copilot"}
	if len(names) != len(expected) {
		t.Fatalf("got %d names, want %d: %v", len(names), len(expected), names)
	}
	for i, n := range expected {
		if names[i] != n {
			t.Errorf("names[%d] = %q, want %q", i, names[i], n)
		}
	}
}

func TestParseAgentInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []agentKind
	}{
		{"single_name", "opencode", []agentKind{agentOpenCode}},
		{"alias_oc", "oc", []agentKind{agentOpenCode}},
		{"alias_open", "open", []agentKind{agentOpenCode}},
		{"alias_cc", "cc", []agentKind{agentClaude}},
		{"alias_cur", "cur", []agentKind{agentCursor}},
		{"alias_gh", "gh", []agentKind{agentCopilot}},
		{"alias_cp", "cp", []agentKind{agentCopilot}},
		{"alias_claude_code", "claude-code", []agentKind{agentClaude}},
		{"alias_github", "github", []agentKind{agentCopilot}},
		{"multiple_comma", "opencode,cursor", []agentKind{agentOpenCode, agentCursor}},
		{"dedup_same_agent", "opencode,oc,open", []agentKind{agentOpenCode}},
		{"with_spaces", "  opencode , claude  ", []agentKind{agentOpenCode, agentClaude}},
		{"all_aliases", "open,cc,cur,gh", []agentKind{agentOpenCode, agentClaude, agentCursor, agentCopilot}},
		{"empty_input", "", nil},
		{"invalid_only", "frobulator", nil},
		{"mixed_valid_invalid", "opencode,frobulator", []agentKind{agentOpenCode}},
		{"case_insensitive", "OpenCode,CLAUDE", []agentKind{agentOpenCode, agentClaude}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAgentInput(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("parseAgentInput(%q) = %v (len %d), want %v (len %d)",
					tc.input, got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("parseAgentInput(%q)[%d] = %q, want %q",
						tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestRelPath(t *testing.T) {
	dir := t.TempDir()

	t.Run("file inside directory", func(t *testing.T) {
		path := filepath.Join(dir, "file.txt")
		result := relPath(dir, path)
		if result != "file.txt" {
			t.Logf("relPath(%q, %q) = %q", dir, path, result)
		}
	})

	t.Run("file outside directory returns relative path", func(t *testing.T) {
		// filepath.Rel can return "../other.txt" for parent paths.
		outside := filepath.Join(dir, "..", "other.txt")
		result := relPath(dir, outside)
		// It should NOT return the original path (which means Rel succeeded)
		if result == outside {
			t.Errorf("expected filepath.Rel to succeed, got original path %q", result)
		}
	})

	t.Run("identical paths", func(t *testing.T) {
		result := relPath(dir, dir)
		if result != "." {
			t.Logf("relPath(%q, %q) = %q", dir, dir, result)
		}
	})
}

func TestGetAgentDefs(t *testing.T) {
	cwd := t.TempDir()

	t.Run("opencode has 3 files", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentOpenCode)
		if len(defs) != 3 {
			t.Fatalf("opencode should have 3 defs, got %d", len(defs))
		}
		if !strings.Contains(defs[0].path, "opencode.json") {
			t.Errorf("expected opencode.json, got %s", defs[0].path)
		}
		if !strings.Contains(defs[0].content, "mcp") {
			t.Error("opencode config should contain MCP settings")
		}
		if !strings.Contains(defs[1].path, "ailinter.md") {
			t.Errorf("expected ailinter.md, got %s", defs[1].path)
		}
		if !strings.Contains(defs[2].path, "SKILL.md") {
			t.Errorf("expected SKILL.md, got %s", defs[2].path)
		}
	})

	t.Run("claude has 2 files", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentClaude)
		if len(defs) != 2 {
			t.Fatalf("claude should have 2 defs, got %d", len(defs))
		}
		if !strings.Contains(defs[0].path, "settings.json") {
			t.Errorf("expected settings.json, got %s", defs[0].path)
		}
		if !strings.Contains(defs[1].path, "CLAUDE.md") {
			t.Errorf("expected CLAUDE.md, got %s", defs[1].path)
		}
		if !strings.Contains(defs[1].content, "Code Quality") {
			t.Error("Claude instructions should mention Code Quality")
		}
	})

	t.Run("cursor has 2 files", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentCursor)
		if len(defs) != 2 {
			t.Fatalf("cursor should have 2 defs, got %d", len(defs))
		}
		if !strings.Contains(defs[0].path, "mcp.json") {
			t.Errorf("expected mcp.json, got %s", defs[0].path)
		}
		if !strings.Contains(defs[1].path, "ailinter.mdc") {
			t.Errorf("expected ailinter.mdc, got %s", defs[1].path)
		}
	})

	t.Run("copilot has 1 file", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentCopilot)
		if len(defs) != 1 {
			t.Fatalf("copilot should have 1 def, got %d", len(defs))
		}
		if !strings.Contains(defs[0].path, "copilot-instructions.md") {
			t.Errorf("expected copilot-instructions.md, got %s", defs[0].path)
		}
	})

	t.Run("unknown agent returns nil", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentKind("unknown"))
		if defs != nil {
			t.Errorf("expected nil for unknown agent, got %d defs", len(defs))
		}
	})
}

func TestGetAgentDefs_TemplateContent(t *testing.T) {
	cwd := t.TempDir()

	t.Run("opencode MCP config is valid JSON", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentOpenCode)
		content := defs[0].content // opencode.json
		if !strings.Contains(content, `"command"`) {
			t.Error("opencode MCP config should contain 'command' key")
		}
		if !strings.Contains(content, "ailinter") {
			t.Error("opencode MCP config should reference ailinter")
		}
	})

	t.Run("claude MCP config is valid JSON", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentClaude)
		content := defs[0].content // .claude/settings.json
		if !strings.Contains(content, `"mcpServers"`) {
			t.Error("claude MCP config should contain mcpServers key")
		}
	})

	t.Run("cursor MCP config is valid JSON", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentCursor)
		content := defs[0].content // .cursor/mcp.json
		if !strings.Contains(content, `"mcpServers"`) {
			t.Error("cursor MCP config should contain mcpServers key")
		}
	})

	t.Run("opencode skill contains required MCP tools", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentOpenCode)
		skillContent := defs[2].content                                                                                                                       // SKILL.md
		requiredTools := []string{"analyze_code", "scan_for_secrets", "get_refactoring_strategy", "assess_file", "list_hotspots", "set_config", "get_config"} // gitleaks:allow
		for _, tool := range requiredTools {
			if !strings.Contains(skillContent, tool) {
				t.Errorf("skill content should mention '%s'", tool)
			}
		}
	})

	t.Run("claude instructions contain quality score reference", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentClaude)
		content := defs[1].content // CLAUDE.md
		requiredSections := []string{"Go Ahead", "Proceed with Care", "Stop & Refactor", "Quality Score Reference", "Available MCP Tools"}
		for _, section := range requiredSections {
			if !strings.Contains(content, section) {
				t.Errorf("Claude instructions should contain '%s' section", section)
			}
		}
	})

	t.Run("cursor rules mention available smells", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentCursor)
		content := defs[1].content // ailinter.mdc
		requiredSmells := []string{"deep_nesting", "brain_method", "bumpy_road", "complex_conditional", "god_class", "long_parameter_list", "primitive_obsession", "duplicated_code"}
		for _, smell := range requiredSmells {
			if !strings.Contains(content, smell) {
				t.Errorf("cursor rules should mention '%s' smell", smell)
			}
		}
	})

	t.Run("copilot instructions contain all MCP tools", func(t *testing.T) {
		defs := getAgentDefs(cwd, agentCopilot)
		content := defs[0].content // copilot-instructions.md
		if !strings.Contains(content, "analyze_code") {
			t.Error("copilot instructions should mention analyze_code")
		}
		if !strings.Contains(content, "scan_for_secrets") {
			t.Error("copilot instructions should mention scan_for_secrets")
		}
		if !strings.Contains(content, "get_refactoring_strategy") {
			t.Error("copilot instructions should mention get_refactoring_strategy")
		}
		if !strings.Contains(content, "assess_file") {
			t.Error("copilot instructions should mention assess_file")
		}
		if !strings.Contains(content, "list_hotspots") {
			t.Error("copilot instructions should mention list_hotspots")
		}
	})
}

// ──────────────────────────────────────────────
// Scanner-based interactive function tests
// ──────────────────────────────────────────────

func TestSelectProfile(t *testing.T) {
	t.Run("default on empty input", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("\n"))
		profile := selectProfile(scanner)
		if profile != "default" {
			t.Errorf("expected 'default', got %q", profile)
		}
	})

	t.Run("choice 1 = default", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("1\n"))
		profile := selectProfile(scanner)
		if profile != "default" {
			t.Errorf("expected 'default', got %q", profile)
		}
	})

	t.Run("choice 2 = strict", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("2\n"))
		profile := selectProfile(scanner)
		if profile != "strict" {
			t.Errorf("expected 'strict', got %q", profile)
		}
	})

	t.Run("choice 3 = relaxed", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("3\n"))
		profile := selectProfile(scanner)
		if profile != "relaxed" {
			t.Errorf("expected 'relaxed', got %q", profile)
		}
	})

	t.Run("invalid input retries then succeeds", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("4\n2\n"))
		profile := selectProfile(scanner)
		if profile != "strict" {
			t.Errorf("expected 'strict' after retry, got %q", profile)
		}
	})

	t.Run("scanner EOF returns default", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader(""))
		profile := selectProfile(scanner)
		if profile != "default" {
			t.Errorf("expected 'default' on EOF, got %q", profile)
		}
	})
}

func TestSelectAgents(t *testing.T) {
	t.Run("default to all on empty input", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("\n"))
		agents := selectAgents(scanner)
		if len(agents) != 4 {
			t.Errorf("expected 4 agents, got %d: %v", len(agents), agents)
		}
	})

	t.Run("explicit 'all'", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("all\n"))
		agents := selectAgents(scanner)
		if len(agents) != 4 {
			t.Errorf("expected 4 agents for 'all', got %d: %v", len(agents), agents)
		}
	})

	t.Run("single agent selection", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("opencode\n"))
		agents := selectAgents(scanner)
		if len(agents) != 1 || agents[0] != agentOpenCode {
			t.Errorf("expected [opencode], got %v", agents)
		}
	})

	t.Run("multiple agents selection", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("opencode,cursor\n"))
		agents := selectAgents(scanner)
		if len(agents) != 2 {
			t.Errorf("expected 2 agents, got %d: %v", len(agents), agents)
		}
	})

	t.Run("invalid input defaults to all", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("frobulator\n"))
		agents := selectAgents(scanner)
		if len(agents) != 4 {
			t.Errorf("expected 4 agents on invalid, got %d: %v", len(agents), agents)
		}
	})

	t.Run("scanner EOF returns all", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader(""))
		agents := selectAgents(scanner)
		if len(agents) != 4 {
			t.Errorf("expected 4 agents on EOF, got %d: %v", len(agents), agents)
		}
	})
}

func TestAskConfirm(t *testing.T) {
	t.Run("'y' returns true", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("y\n"))
		if !askConfirm(scanner, "Test?") {
			t.Error("expected true for 'y'")
		}
	})

	t.Run("'yes' returns true", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("yes\n"))
		if !askConfirm(scanner, "Test?") {
			t.Error("expected true for 'yes'")
		}
	})

	t.Run("'n' returns false", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("n\n"))
		if askConfirm(scanner, "Test?") {
			t.Error("expected false for 'n'")
		}
	})

	t.Run("empty input returns false", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader("\n"))
		if askConfirm(scanner, "Test?") {
			t.Error("expected false for empty")
		}
	})

	t.Run("scanner EOF returns false", func(t *testing.T) {
		scanner := bufio.NewScanner(strings.NewReader(""))
		if askConfirm(scanner, "Test?") {
			t.Error("expected false on EOF")
		}
	})
}

// ──────────────────────────────────────────────
// File system-based function tests
// ──────────────────────────────────────────────

func TestWriteConfig(t *testing.T) {
	t.Run("default profile creates config", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeConfig(dir, "default", result)
		data, err := os.ReadFile(filepath.Join(dir, ".ailinter.toml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "deep_nesting") {
			t.Error("config should contain rules")
		}
		if len(result.created) != 1 || result.created[0] != ".ailinter.toml" {
			t.Errorf("expected .ailinter.toml in created, got %v", result.created)
		}
	})

	t.Run("strict profile", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeConfig(dir, "strict", result)
		data, _ := os.ReadFile(filepath.Join(dir, ".ailinter.toml"))
		if !strings.Contains(string(data), "strict thresholds") {
			t.Error("strict config should mention strict thresholds")
		}
	})

	t.Run("relaxed profile", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeConfig(dir, "relaxed", result)
		data, _ := os.ReadFile(filepath.Join(dir, ".ailinter.toml"))
		if !strings.Contains(string(data), "relaxed thresholds") {
			t.Error("relaxed config should mention relaxed thresholds")
		}
	})

	t.Run("skipped if config already exists", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte("existing"), 0644)
		result := &setupResult{}
		writeConfig(dir, "default", result)
		if len(result.skipped) != 1 || result.skipped[0] != ".ailinter.toml" {
			t.Errorf("expected .ailinter.toml in skipped, got %v", result.skipped)
		}
		if len(result.created) != 0 {
			t.Errorf("expected no created files, got %v", result.created)
		}
		// Verify existing content was not overwritten
		data, _ := os.ReadFile(filepath.Join(dir, ".ailinter.toml"))
		if string(data) != "existing" {
			t.Errorf("existing file was overwritten: got %q", string(data))
		}
	})
}

func TestWriteAgentsMD(t *testing.T) {
	t.Run("skip when flag is true", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentsMD(dir, true, result)
		if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
			t.Error("AGENTS.md should not be created when skip=true")
		}
	})

	t.Run("creates AGENTS.md with content", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentsMD(dir, false, result)
		data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "Code Quality") {
			t.Error("AGENTS.md should mention Code Quality")
		}
		if len(result.created) != 1 || result.created[0] != "AGENTS.md" {
			t.Errorf("expected AGENTS.md in created, got %v", result.created)
		}
	})

	t.Run("skipped if AGENTS.md already exists", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("existing"), 0644)
		result := &setupResult{}
		writeAgentsMD(dir, false, result)
		if len(result.skipped) != 1 || result.skipped[0] != "AGENTS.md" {
			t.Errorf("expected AGENTS.md in skipped, got %v", result.skipped)
		}
	})
}

func TestWriteAgentFiles(t *testing.T) {
	t.Run("creates opencode agent files", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentFiles(dir, agentOpenCode, result)
		// Should create 3 files
		if len(result.created) != 3 {
			t.Fatalf("expected 3 created files, got %d: %v", len(result.created), result.created)
		}
		// Verify files exist on disk
		for _, path := range result.created {
			fullPath := filepath.Join(dir, path)
			if _, err := os.Stat(fullPath); err != nil {
				t.Errorf("created file not found on disk: %s (%v)", fullPath, err)
			}
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		dir := t.TempDir()
		// Create one file ahead of time
		opencodePath := filepath.Join(dir, "opencode.json")
		os.WriteFile(opencodePath, []byte("existing"), 0644)

		result := &setupResult{}
		writeAgentFiles(dir, agentOpenCode, result)
		// One file should be skipped, two created
		if len(result.skipped) != 1 {
			t.Errorf("expected 1 skipped, got %d: %v", len(result.skipped), result.skipped)
		}
		if len(result.created) != 2 {
			t.Errorf("expected 2 created, got %d: %v", len(result.created), result.created)
		}
	})

	t.Run("creates claude files", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentFiles(dir, agentClaude, result)
		if len(result.created) != 2 {
			t.Fatalf("expected 2 created files, got %d", len(result.created))
		}
	})

	t.Run("creates cursor files", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentFiles(dir, agentCursor, result)
		if len(result.created) != 2 {
			t.Fatalf("expected 2 created files, got %d", len(result.created))
		}
	})

	t.Run("creates copilot file", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentFiles(dir, agentCopilot, result)
		if len(result.created) != 1 {
			t.Fatalf("expected 1 created file, got %d", len(result.created))
		}
	})
}

func TestWriteVSCodeFiles(t *testing.T) {
	t.Run("creates all three VS Code files", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeVSCodeFiles(dir, result)
		if len(result.created) != 3 {
			t.Fatalf("expected 3 created files, got %d: %v", len(result.created), result.created)
		}
		// Verify each file
		expectedFiles := []string{".vscode/tasks.json", ".vscode/settings.json", ".vscode/extensions.json"}
		for _, f := range expectedFiles {
			fullPath := filepath.Join(dir, f)
			if _, err := os.Stat(fullPath); err != nil {
				t.Errorf("file not found: %s (%v)", fullPath, err)
			}
		}
		// Verify content
		tasksData, _ := os.ReadFile(filepath.Join(dir, ".vscode/tasks.json"))
		if !strings.Contains(string(tasksData), "ailinter") {
			t.Error("tasks.json should reference ailinter")
		}
		settingsData, _ := os.ReadFile(filepath.Join(dir, ".vscode/settings.json"))
		if !strings.Contains(string(settingsData), "ailinter.enable") {
			t.Error("settings.json should have ailinter.enable")
		}
		extData, _ := os.ReadFile(filepath.Join(dir, ".vscode/extensions.json"))
		if !strings.Contains(string(extData), "ailinter") {
			t.Error("extensions.json should recommend ailinter")
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		dir := t.TempDir()
		// Create one existing file
		vsDir := filepath.Join(dir, ".vscode")
		os.MkdirAll(vsDir, 0755)
		os.WriteFile(filepath.Join(vsDir, "tasks.json"), []byte("existing"), 0644)

		result := &setupResult{}
		writeVSCodeFiles(dir, result)
		// tasks.json should be skipped, others created
		if len(result.skipped) != 1 {
			t.Errorf("expected 1 skipped (tasks.json), got %d: %v", len(result.skipped), result.skipped)
		}
		if len(result.created) != 2 {
			t.Errorf("expected 2 created, got %d: %v", len(result.created), result.created)
		}
	})
}

func TestWriteGitHook(t *testing.T) {
	t.Run("creates pre-commit hook", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeGitHook(dir, result)
		if len(result.created) != 1 {
			t.Fatalf("expected 1 created, got %d: %v", len(result.created), result.created)
		}
		hookPath := filepath.Join(dir, ".githooks", "pre-commit")
		if _, err := os.Stat(hookPath); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(hookPath)
		if !strings.Contains(string(data), "ailinter") {
			t.Error("hook should contain ailinter")
		}
		// Verify hook is executable
		info, _ := os.Stat(hookPath)
		if info.Mode().Perm()&0111 == 0 {
			t.Error("hook should be executable")
		}
	})

	t.Run("skips if hook already exists", func(t *testing.T) {
		dir := t.TempDir()
		// Pre-create hook
		hookDir := filepath.Join(dir, ".githooks")
		os.MkdirAll(hookDir, 0755)
		os.WriteFile(filepath.Join(hookDir, "pre-commit"), []byte("#!/bin/sh\nexisting"), 0755)

		result := &setupResult{}
		writeGitHook(dir, result)
		if len(result.skipped) != 1 {
			t.Errorf("expected 1 skipped, got %d: %v", len(result.skipped), result.skipped)
		}
		if len(result.created) != 0 {
			t.Errorf("expected 0 created, got %d: %v", len(result.created), result.created)
		}
	})
}

func TestUpdateGitignoreForHooks(t *testing.T) {
	t.Run("appends to existing .gitignore without hooks entry", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\nnode_modules/\n"), 0644)
		updateGitignoreForHooks(dir)
		data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if !strings.Contains(string(data), ".githooks/") {
			t.Error(".gitignore should contain .githooks/")
		}
		if !strings.Contains(string(data), "ailinter hooks") {
			t.Error(".gitignore should mention ailinter hooks")
		}
	})

	t.Run("does not duplicate if .githooks already in .gitignore", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".githooks/\n"), 0644)
		updateGitignoreForHooks(dir)
		data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
		// Count occurrences
		count := strings.Count(string(data), ".githooks/")
		if count != 1 {
			t.Errorf("expected 1 occurrence of .githooks/, got %d: %s", count, string(data))
		}
	})

	t.Run("creates .gitignore if none exists", func(t *testing.T) {
		dir := t.TempDir()
		updateGitignoreForHooks(dir)
		data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if !strings.Contains(string(data), ".githooks/") {
			t.Error("new .gitignore should contain .githooks/")
		}
	})
}

// ──────────────────────────────────────────────
// init.go function tests
// ──────────────────────────────────────────────

func TestParseAgentName(t *testing.T) {
	t.Run("recognizes canonical names", func(t *testing.T) {
		canonical, ok := parseAgentName("opencode")
		if !ok || canonical != agentOpenCode {
			t.Errorf("expected (opencode, true), got (%q, %v)", canonical, ok)
		}
	})

	t.Run("recognizes aliases", func(t *testing.T) {
		canonical, ok := parseAgentName("oc")
		if !ok || canonical != agentOpenCode {
			t.Errorf("expected (opencode, true), got (%q, %v)", canonical, ok)
		}
	})

	t.Run("rejects unknown names", func(t *testing.T) {
		_, ok := parseAgentName("frobulator")
		if ok {
			t.Error("expected false for unknown agent")
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		_, ok := parseAgentName("")
		if ok {
			t.Error("expected false for empty string")
		}
	})
}

func TestWriteAgentSetups(t *testing.T) {
	t.Run("'all' writes files for all agents", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentSetups(dir, "all", result)
		// opencode (3) + claude (2) + cursor (2) + copilot (1) = 8
		if len(result.created) != 8 {
			t.Fatalf("expected 8 created files, got %d: %v", len(result.created), result.created)
		}
	})

	t.Run("known agent name writes its files", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentSetups(dir, "opencode", result)
		if len(result.created) != 3 {
			t.Fatalf("expected 3 created files for opencode, got %d: %v", len(result.created), result.created)
		}
	})

	t.Run("agent alias works", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentSetups(dir, "oc", result)
		if len(result.created) != 3 {
			t.Fatalf("expected 3 created files for alias 'oc', got %d: %v", len(result.created), result.created)
		}
	})

	t.Run("unknown agent writes nothing", func(t *testing.T) {
		dir := t.TempDir()
		result := &setupResult{}
		writeAgentSetups(dir, "frobulator", result)
		if len(result.created) != 0 {
			t.Errorf("expected 0 created files for unknown agent, got %d: %v", len(result.created), result.created)
		}
	})
}

// ──────────────────────────────────────────────
// check.go function tests
// ──────────────────────────────────────────────

func TestSmellInRanges(t *testing.T) {
	tests := []struct {
		name   string
		smell  analyzer.Smell
		ranges []struct{ Start, End int }
		want   bool
	}{
		{
			name:   "smell start within range",
			smell:  analyzer.Smell{LineStart: 5, LineEnd: 10},
			ranges: []struct{ Start, End int }{{Start: 3, End: 7}},
			want:   true,
		},
		{
			name:   "smell end within range",
			smell:  analyzer.Smell{LineStart: 5, LineEnd: 10},
			ranges: []struct{ Start, End int }{{Start: 8, End: 12}},
			want:   true,
		},
		{
			name:   "smell completely before range",
			smell:  analyzer.Smell{LineStart: 1, LineEnd: 3},
			ranges: []struct{ Start, End int }{{Start: 5, End: 10}},
			want:   false,
		},
		{
			name:   "smell completely after range",
			smell:  analyzer.Smell{LineStart: 15, LineEnd: 20},
			ranges: []struct{ Start, End int }{{Start: 5, End: 10}},
			want:   false,
		},
		{
			name:   "smell encompasses range",
			smell:  analyzer.Smell{LineStart: 1, LineEnd: 20},
			ranges: []struct{ Start, End int }{{Start: 5, End: 10}},
			want:   true,
		},
		{
			name:   "exact match",
			smell:  analyzer.Smell{LineStart: 5, LineEnd: 10},
			ranges: []struct{ Start, End int }{{Start: 5, End: 10}},
			want:   true,
		},
		{
			name:   "multiple ranges, matches second",
			smell:  analyzer.Smell{LineStart: 15, LineEnd: 20},
			ranges: []struct{ Start, End int }{{Start: 1, End: 5}, {Start: 12, End: 18}},
			want:   true,
		},
		{
			name:   "single line smell within range",
			smell:  analyzer.Smell{LineStart: 7, LineEnd: 7},
			ranges: []struct{ Start, End int }{{Start: 5, End: 10}},
			want:   true,
		},
		{
			name:   "empty ranges",
			smell:  analyzer.Smell{LineStart: 5, LineEnd: 10},
			ranges: []struct{ Start, End int }{},
			want:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ranges := make([]git.LineRange, len(tc.ranges))
			for i, r := range tc.ranges {
				ranges[i] = git.LineRange{Start: r.Start, End: r.End}
			}
			got := smellInRanges(tc.smell, ranges)
			if got != tc.want {
				t.Errorf("smellInRanges(%+v, %v) = %v, want %v", tc.smell, ranges, got, tc.want)
			}
		})
	}
}
