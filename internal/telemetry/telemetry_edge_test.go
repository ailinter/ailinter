package telemetry

import (
	"testing"
)

func TestRecordingsWithEnabledButNil(t *testing.T) {
	IsEnabled = true
	defer func() {
		IsEnabled = false
	}()

	recordings := []func(){
		RecordInstallation,
		func() { RecordCLIInvocation("check") },
		func() { RecordMCPToolCall("analyze_code") },
		func() { RecordFileAnalyzed("go", ".go") },
		func() { RecordQualityScore("go", 100) },
		func() { RecordSmellsDetected("deep_nesting", "go", 1) },
		func() { RecordSecretsDetected("aws", "critical", ".env", 1) },
		func() { RecordDuration("check_file", "go", 1.5) },
		func() { RecordError("test_error") },
	}

	for i, fn := range recordings {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("recording %d panicked with IsEnabled=true and nil instruments: %v", i, r)
				}
			}()
			fn()
		}()
	}
}

func TestRecordingsWithDisabled(t *testing.T) {
	IsEnabled = false

	recordings := []func(){
		RecordInstallation,
		func() { RecordCLIInvocation("check") },
		func() { RecordMCPToolCall("analyze_code") },
		func() { RecordFileAnalyzed("go", ".go") },
		func() { RecordQualityScore("go", 100) },
		func() { RecordSmellsDetected("deep_nesting", "go", 1) },
		func() { RecordSecretsDetected("aws", "critical", ".env", 1) },
		func() { RecordDuration("check_file", "go", 1.5) },
		func() { RecordError("test_error") },
	}

	for i, fn := range recordings {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("recording %d panicked with IsEnabled=false: %v", i, r)
				}
			}()
			fn()
		}()
	}
}
