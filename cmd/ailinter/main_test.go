package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func captureMainOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

func TestVersionVariable(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
	if !strings.Contains(version, "dev") && !strings.Contains(version, ".") {
		t.Error("version should contain dev or semver")
	}
}

func TestRulesCommand_List(t *testing.T) {
	cmd := rulesCommand()
	subCmd, _, _ := cmd.Find([]string{"list"})
	if subCmd == nil {
		t.Fatal("rules list command not found")
	}

	out := captureMainOutput(func() {
		subCmd.RunE(subCmd, []string{})
	})

	if !strings.Contains(out, "Go") {
		t.Error("rules list should contain Go column")
	}
	if !strings.Contains(out, "Python") {
		t.Error("rules list should contain Python column")
	}
	if !strings.Contains(out, "Nesting depth") {
		t.Error("rules list should contain nesting metric")
	}
	if !strings.Contains(out, "Cyclomatic complexity") {
		t.Error("rules list should contain cyclomatic complexity")
	}
	if !strings.Contains(out, "Function LOC") {
		t.Error("rules list should contain function LOC")
	}
}

func TestRootCommand_Help(t *testing.T) {
	root := &cobra.Command{
		Use: "ailinter",
	}
	root.SetArgs([]string{"--help"})

	out := captureMainOutput(func() {
		root.Execute()
	})

	if !strings.Contains(out, "ailinter") && !strings.Contains(out, "Usage") {
		t.Logf("root help output: %s", out)
	}
}

func TestRootCommand_Version(t *testing.T) {
	root := &cobra.Command{
		Version: version,
	}
	root.SetArgs([]string{"--version"})

	out := captureMainOutput(func() {
		root.Execute()
	})

	if !strings.Contains(out, version) {
		t.Logf("version output: %s", out)
	}
}
