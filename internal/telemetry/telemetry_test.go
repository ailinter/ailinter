package telemetry

import (
	"context"
	"os"
	"testing"
)

func TestResolveEnabled_NoEnv(t *testing.T) {
	os.Unsetenv("AILINTER_NO_TELEMETRY")
	if !resolveEnabled() {
		t.Error("expected enabled by default")
	}
}

func TestResolveEnabled_Var_1(t *testing.T) {
	os.Setenv("AILINTER_NO_TELEMETRY", "1")
	defer os.Unsetenv("AILINTER_NO_TELEMETRY")

	if resolveEnabled() {
		t.Error("expected disabled with AILINTER_NO_TELEMETRY=1")
	}
}

func TestResolveEnabled_Var_True(t *testing.T) {
	os.Setenv("AILINTER_NO_TELEMETRY", "true")
	defer os.Unsetenv("AILINTER_NO_TELEMETRY")

	if resolveEnabled() {
		t.Error("expected disabled with AILINTER_NO_TELEMETRY=true")
	}
}

func TestResolveEnabled_Var_0(t *testing.T) {
	os.Setenv("AILINTER_NO_TELEMETRY", "0")
	defer os.Unsetenv("AILINTER_NO_TELEMETRY")

	if !resolveEnabled() {
		t.Error("expected enabled with AILINTER_NO_TELEMETRY=0")
	}
}

func TestLoadOrCreateInstallID(t *testing.T) {
	id, _ := loadOrCreateInstallID()
	if id == "" {
		t.Skip("install ID generation might fail without config dir")
		return
	}
	if len(id) != 32 {
		t.Errorf("expected 32-hex install ID, got %d chars", len(id))
	}
}

func TestInit(t *testing.T) {
	Version = "test"
	defer func() { Version = "" }()

	Init(context.Background())
	defer Shutdown(context.Background())
}

func TestNoopRecordings(t *testing.T) {
	RecordInstallation()
	RecordCLIInvocation("check")
	RecordMCPToolCall("analyze_code")
	RecordFileAnalyzed("go", ".go")
	RecordQualityScore("go", 100)
	RecordSmellsDetected("deep_nesting", "go", 1)
	RecordSecretsDetected("aws", "critical", ".env", 1)
	RecordDuration("check_file", "go", 1.5)
	RecordError("test_error")
}
