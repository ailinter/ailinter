package analyzer_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

// BenchmarkAnalyze measures performance of the full analysis pipeline.
func BenchmarkAnalyze(b *testing.B) {
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("test.go", src, "go", analyzer.DefaultThresholds("go"))
	}
}

// BenchmarkLargeFile tests analysis of a ~500-line file.
func BenchmarkLargeFile(b *testing.B) {
	src := "package main\n"
	for i := 0; i < 500; i++ {
		src += fmt.Sprintf("var x%d = %d\n", i, i)
	}
	for i := 0; i < 10; i++ {
		src += fmt.Sprintf(`
func Func%d() {
	if x > 0 {
		if x < 100 {
			for _, v := range items {
				if v > 0 {
					process(v)
				}
			}
		}
	}
}
`, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("test.go", src, "go", analyzer.DefaultThresholds("go"))
	}
}

// BenchmarkCSComparison runs ailinter on testdata fixtures and compares expected smells.
func TestBenchmarkCSComparison(t *testing.T) {
	if os.Getenv("BENCHMARK_CS") == "" {
		t.Skip("set BENCHMARK_CS=1 to run CS comparison benchmark")
	}

	tests := []struct {
		file     string
		minScore int
		maxScore int
		want     []string // expected smell names
	}{
		{file: "../../testdata/healthy/simple_func.go", minScore: 95, maxScore: 100, want: nil},
		{file: "../../testdata/deep_nested/depth_5.go", minScore: 90, maxScore: 100, want: []string{"deep_nesting"}},
		{file: "../../testdata/bumpy_road/two_bumps.go", minScore: 90, maxScore: 100, want: []string{"bumpy_road"}},
		{file: "../../testdata/complex_conditional/many_and_or.go", minScore: 90, maxScore: 100, want: []string{"complex_conditional"}},
		{file: "../../testdata/brain_method/long_func_150.go", minScore: 90, maxScore: 100, want: []string{"brain_method"}},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Skipf("testdata not available: %v", err)
			}
			result := analyzer.Analyze(tt.file, string(data), "go", analyzer.DefaultThresholds("go"))
			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("score %d not in [%d, %d]", result.Score, tt.minScore, tt.maxScore)
			}
			if tt.want != nil {
				names := smellNames(result.Smells)
				for _, w := range tt.want {
					if !contains(names, w) {
						t.Errorf("expected smell %q not found (got: %v)", w, names)
					}
				}
			}
			t.Logf("%s: score=%d smells=%v", tt.file, result.Score, smellNames(result.Smells))
		})
	}
}

func smellNames(smells []analyzer.Smell) []string {
	names := make([]string, len(smells))
	for i, s := range smells {
		names[i] = s.Name
	}
	return names
}

func contains(xs []string, x string) bool {
	for _, item := range xs {
		if item == x {
			return true
		}
	}
	return false
}

// TestAllFixtureFiles runs ailinter on every test fixture and checks basic invariants.
func TestAllFixtureFiles(t *testing.T) {
	files := []string{
		"../../testdata/healthy/simple_func.go",
		"../../testdata/healthy/well_structured.py",
		"../../testdata/deep_nested/depth_3.go",
		"../../testdata/deep_nested/depth_5.go",
		"../../testdata/deep_nested/depth_5.py",
		"../../testdata/bumpy_road/two_bumps.go",
		"../../testdata/bumpy_road/three_bumps.go",
		"../../testdata/complex_conditional/simple_if.go",
		"../../testdata/complex_conditional/many_and_or.go",
		"../../testdata/brain_method/long_func_150.go",
		"../../testdata/brain_method/long_func_90.py",
		"../../testdata/long_switch/long.go",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Skipf("file not found: %v", err)
			}
			lang := detectLang(file)
			result := analyzer.Analyze(file, string(data), lang, analyzer.DefaultThresholds(lang))

			if result.Score < 1.0 || result.Score > 100 {
				t.Errorf("score %d out of range [1.0, 100]", result.Score)
			}
			if result.LinesOfCode <= 0 {
				t.Error("lines of code should be > 0")
			}
		})
	}
}

func detectLang(file string) string {
	switch {
	case containsExt(file, ".py"):
		return "python"
	case containsExt(file, ".js"):
		return "javascript"
	case containsExt(file, ".ts"):
		return "typescript"
	default:
		return "go"
	}
}

func containsExt(file, ext string) bool {
	return len(file) >= len(ext) && file[len(file)-len(ext):] == ext
}
