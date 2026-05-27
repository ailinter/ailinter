package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckCommand_Exists(t *testing.T) {
	cmd := CheckCommand()
	if cmd == nil {
		t.Fatal("CheckCommand returned nil")
	}
}

func TestMCPCommand_Exists(t *testing.T) {
	cmd := MCPCommand()
	if cmd == nil {
		t.Fatal("MCPCommand returned nil")
	}
}

func TestInitCommand_Exists(t *testing.T) {
	cmd := InitCommand()
	if cmd == nil {
		t.Fatal("InitCommand returned nil")
	}
}

func TestInitCommand_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)

	cmd := InitCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".ailinter.toml")); err != nil {
		t.Error(".ailinter.toml not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Error("AGENTS.md not created")
	}
}

func TestInitCommand_NoAgents(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)

	cmd := InitCommand()
	cmd.SetArgs([]string{"--no-agents"})
	cmd.Execute()

	if _, err := os.Stat(filepath.Join(dir, ".ailinter.toml")); err != nil {
		t.Error(".ailinter.toml not created")
	}
}

func TestCheckCommand_FileNotFound(t *testing.T) {
	cmd := CheckCommand()
	cmd.SetArgs([]string{"/nonexistent/path.go"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestInitCommand_Idempotent(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	cmd := InitCommand()
	cmd.SetArgs([]string{})
	cmd.Execute() // first init
	cmd = InitCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute() // second init — should not error
	if err != nil {
		t.Errorf("second init should not error: %v", err)
	}
}
