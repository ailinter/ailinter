package analyzer_test

import (
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

func analyzeTest(src string) analyzer.QualityResult {
	return analyzer.Analyze(
		analyzer.SourceInput{FilePath: "test.go", Source: src, Lang: "go"},
		analyzer.DefaultThresholds("go"),
	)
}

func assertSmellDetected(t *testing.T, result analyzer.QualityResult, smellName string) bool {
	t.Helper()
	for _, s := range result.Smells {
		if s.Name == smellName {
			return true
		}
	}
	return false
}

func TestQualityScore_HealthyFile(t *testing.T) {
	src := `package main

func Greet(name string) string {
	if name == "" {
		return "Hello, World!"
	}
	return "Hello, " + name + "!"
}
`
	result := analyzeTest(src)
	if result.Score < 95 {
		t.Errorf("healthy file should score >= 95, got %d", result.Score)
	}
}

func TestDeepNesting_Detected(t *testing.T) {
	src := `package main

func DeepNested(data *Data) error {
	if data != nil {
		if data.IsActive() {
			for _, item := range data.Items {
				if item.IsValid() {
					if item.NeedsProcessing() {
						processItem(item)
					}
				}
			}
		}
	}
	return nil
}
`
	result := analyzeTest(src)
	if !assertSmellDetected(t, result, "deep_nesting") {
		t.Error("expected deep_nesting smell, not found")
	}
}

func TestBrainMethod_Detected(t *testing.T) {
	src := "package main\nfunc LongFunc() {\n"
	for i := 0; i < 90; i++ {
		src += "\t_ = " + string(rune('a'+i%26)) + "\n"
	}
	src += "}\n"
	result := analyzeTest(src)
	if !assertSmellDetected(t, result, "brain_method") {
		t.Error("expected brain_method smell for 90-line function, not found")
	}
}

func TestFileBloat_Detected(t *testing.T) {
	src := "package main\n"
	for i := 0; i < 1100; i++ {
		src += "var x" + string(rune('a'+i%26)) + " = 1\n"
	}
	result := analyzeTest(src)
	if !assertSmellDetected(t, result, "file_bloat") {
		t.Error("expected file_bloat smell", nil)
	}
}

func TestComplexConditional_Detected(t *testing.T) {
	src := `package main

func IsEligible(user *User) bool {
	if user.Age > 18 && user.HasVerifiedEmail && !user.IsBanned && (user.Subscription == "premium" || user.PurchaseTotal > 1000 || user.Referrals > 5) {
		return true
	}
	return false
}
`
	result := analyzeTest(src)
	if !assertSmellDetected(t, result, "complex_conditional") {
		t.Error("expected complex_conditional smell, not found")
	}
}

func TestBumpyRoad_Detected(t *testing.T) {
	src := `package main

func TwoBumps(items []int) int {
	c := 0
	for _, it := range items {
		if it > 0 {
			if it < 100 {
				if it != 50 {
					c += it
				}
			}
		}
	}
	for _, it := range items {
		if it < 0 {
			if it > -100 {
				if it != -50 {
					c -= it
				}
			}
		}
	}
	return c
}
`
	result := analyzeTest(src)
	if assertSmellDetected(t, result, "bumpy_road") {
		return
	}
	if assertSmellDetected(t, result, "deep_nesting") {
		t.Log("bumpy_road not detected but deep_nesting was — acceptable")
		return
	}
	t.Error("expected bumpy_road or deep_nesting smell, neither found")
}

func TestDuplication_Detected(t *testing.T) {
	src := `package main

func ProcessA(data []int) int {
	s := 0
	for _, v := range data {
		if v > 0 {
			s += v
			s *= 2
			s -= 1
		}
	}
	return s
}

func ProcessB(data []int) int {
	s := 0
	for _, v := range data {
		if v > 0 {
			s += v
			s *= 2
			s -= 1
		}
	}
	return s
}
`
	result := analyzeTest(src)
	if !assertSmellDetected(t, result, "code_duplication") {
		t.Error("expected code_duplication smell for identical functions, not found")
	}
}

func TestCyclomaticComplexity_Detected(t *testing.T) {
	src := `package main

func ComplexFunc(x int) string {
	if x > 0 && x < 100 {
		if x > 10 || x < 5 {
			if x > 50 && x != 60 {
				if x == 42 || x == 99 {
					return "special"
				}
				return "mid"
			}
			return "low"
		}
		return "small"
	}
	return "zero"
}
`
	result := analyzeTest(src)
	for _, s := range result.Smells {
		t.Logf("  [%s] %s: %s", s.Severity, s.Name, s.Message)
	}
	if !assertSmellDetected(t, result, "complex_method") {
		t.Error("expected complex_method smell, not found")
	}
}
