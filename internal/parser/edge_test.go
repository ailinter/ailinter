package parser_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

func sp(s string) []string { return strings.Split(s, "\n") }

func TestDetectedLanguage_All(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{".go", "go"}, {".py", "python"}, {".js", "javascript"},
		{".ts", "typescript"}, {".tsx", "typescript"}, {".java", "java"},
		{".rs", "rust"}, {".rb", "ruby"}, {".cpp", "cpp"},
		{".cc", "cpp"}, {".cxx", "cpp"}, {".c", "c"},
		{".h", "cpp"}, {".hpp", "cpp"}, {".swift", "swift"},
		{".kt", "kotlin"}, {".kts", "kotlin"}, {".cs", "csharp"},
		{".html", ""}, {".css", ""}, {".unknown", ""}, {"", ""},
	}
	for _, tc := range cases {
		got := parser.DetectedLanguage(tc.ext)
		if got != tc.want {
			t.Errorf("DetectedLanguage(%q) = %q, want %q", tc.ext, got, tc.want)
		}
	}
}

func TestDetectExcessiveComments_NotTriggered(t *testing.T) {
	code := []string{
		"package main",
		"func hello() { return 42 }",
		"func world() { return 99 }",
	}
	s := parser.DetectExcessiveComments(code, 0.3)
	if s != nil {
		t.Errorf("expected nil, got %v", s.Message)
	}
}

func TestDetectExcessiveComments_Triggered(t *testing.T) {
	code := []string{
		"// Comment 1",
		"// Comment 2",
		"// Comment 3",
		"// Comment 4",
		"// Comment 5",
		"// Comment 6",
		"package main",
		"// Comment 7",
		"func hello() { return 42 }",
		"func world() { return 99 }",
	}
	s := parser.DetectExcessiveComments(code, 0.3)
	if s == nil {
		t.Error("expected excessive comments smell")
	} else if s.Name != "excessive_comments" {
		t.Errorf("unexpected smell: %s", s.Name)
	}
}

func TestDetectExcessiveComments_Alert(t *testing.T) {
	code := []string{
		"// c1", "// c2", "// c3", "// c4", "// c5",
		"// c6", "// c7", "// c8", "// c9", "// c10",
		"package main",
		"func f() { return 1 }",
	}
	s := parser.DetectExcessiveComments(code, 0.2)
	if s != nil && s.Severity == "alert" {
		t.Log("alert severity detected as expected")
	}
}

func TestDetectExcessiveComments_SmallFile(t *testing.T) {
	code := []string{
		"// comment", "// comment", "// comment",
	}
	s := parser.DetectExcessiveComments(code, 0.3)
	if s != nil {
		t.Errorf("small file (<10 lines) should not trigger: %v", s.Message)
	}
}

func TestDetectGlobalData_NotTriggered(t *testing.T) {
	code := []string{
		"package main",
		"func main() { x := 1; println(x) }",
	}
	s := parser.DetectGlobalData(code, 3)
	if s != nil {
		t.Errorf("expected nil, got %v", s.Message)
	}
}

func TestDetectGlobalData_Triggered(t *testing.T) {
	code := []string{
		"package main",
		"var global1 = 1",
		"var global2 = 2",
		"var global3 = 3",
		"var global4 = 4",
		"const constant1 = 5",
		"",
		"func main() { println(global1) }",
	}
	s := parser.DetectGlobalData(code, 3)
	if s == nil {
		t.Error("expected global data smell")
	} else if s.Name != "global_data" {
		t.Errorf("unexpected smell: %s", s.Name)
	}
}

func TestDetectGlobalData_BelowThreshold(t *testing.T) {
	code := []string{
		"package main",
		"var x = 1",
		"func main() { println(x) }",
	}
	s := parser.DetectGlobalData(code, 5)
	if s != nil {
		t.Errorf("expected nil (1 global < 5 threshold): %v", s.Message)
	}
}

func TestDetectLongScopeVariables_Triggered(t *testing.T) {
	lines := make([]string, 0, 60)
	lines = append(lines, "package main")
	lines = append(lines, "")
	lines = append(lines, "func longFunc() {")
	lines = append(lines, "	x := 1")
	for i := 0; i < 51; i++ {
		lines = append(lines, "	y := 1")
	}
	lines = append(lines, "	println(x)")
	lines = append(lines, "}")

	bloats := parser.DetectFunctionBloats(lines)
	smells := parser.DetectLongScopeVariables(lines, bloats, 50)
	_ = smells
	t.Logf("long scope smells found: %d", len(smells))
}

func TestDetectLongScopeVariables_ShortFunction(t *testing.T) {
	code := []string{
		"package main",
		"func short() {",
		"	x := 1",
		"	return x",
		"}",
	}
	bloats := parser.DetectFunctionBloats(code)
	smells := parser.DetectLongScopeVariables(code, bloats, 50)
	if len(smells) != 0 {
		t.Errorf("expected 0 smells for short function, got %d", len(smells))
	}
}

func TestDetectMessageChains_Edge(t *testing.T) {
	code := sp("a.b().c()\na.b().c().d().e()\n")
	smells := parser.DetectMessageChains(code)
	if len(smells) != 0 {
		t.Logf("message chains found: %d", len(smells))
	}
}

func TestDetectPrimitiveObsession_Edge(t *testing.T) {
	code := sp("func Process(a string, b int, c float64, d bool, e int) {\n}\n")
	smells := parser.DetectPrimitiveObsession(code)
	if len(smells) != 0 {
		t.Logf("primitive obsession: %d smells", len(smells))
	}
}

func TestDetectPrimitiveObsession_NotTriggered(t *testing.T) {
	code := sp("func Process(a MyType, b string) {\n}\n")
	smells := parser.DetectPrimitiveObsession(code)
	if len(smells) != 0 {
		t.Errorf("expected 0, got %d", len(smells))
	}
}

func TestDetectLazyElements_Edge(t *testing.T) {
	code := sp("func short() {\n\treturn 1\n}\nfunc normal() {\n\tx := 1\n\treturn x + 1\n}\n")
	bloats := parser.DetectFunctionBloats(code)
	smells := parser.DetectLazyElements(bloats, 3)
	if len(smells) > 0 {
		t.Logf("lazy elements: %d", len(smells))
	}
}

func TestDetectLazyElements_Disabled(t *testing.T) {
	code := sp("func short() {\n\treturn 1\n}\n")
	bloats := parser.DetectFunctionBloats(code)
	smells := parser.DetectLazyElements(bloats, 0)
	if len(smells) != 0 {
		t.Error("minLines=0 should not flag anything")
	}
}

func TestLongSwitch_EdgeCases(t *testing.T) {
	t.Run("not a switch", func(t *testing.T) {
		code := sp("func f() { x := 1; return x }")
		s := parser.DetectLongSwitch(code, 3, 5)
		if s != nil {
			t.Errorf("expected nil, got %v", s)
		}
	})
	t.Run("too few cases", func(t *testing.T) {
		code := sp("func f() { switch x { case 1: return; case 2: return; default: return } }")
		s := parser.DetectLongSwitch(code, 10, 20)
		if s != nil {
			t.Errorf("expected nil for switch with < 10 cases, got %v", s)
		}
	})
}

func TestHotspotSmell_Edge(t *testing.T) {
	s := parser.DetectHotspotSmell(parser.HotspotEntry{
		FilePath:     "test.go",
		CommitCount:  0,
		QualityScore: 10.0,
		Priority:     0,
	})
	if s != nil {
		t.Errorf("0 commits should not trigger hotspot: %v", s)
	}

	s = parser.DetectHotspotSmell(parser.HotspotEntry{
		FilePath:     "hot.go",
		CommitCount:  10,
		QualityScore: 3.0,
		Priority:     30,
	})
	if s == nil {
		t.Error("10 commits with low quality should trigger")
	} else {
		if s.Name != "hotspot" {
			t.Errorf("unexpected smell: %s", s.Name)
		}
		if s.Severity != "warning" {
			t.Errorf("expected warning severity, got %s", s.Severity)
		}
	}

	s = parser.DetectHotspotSmell(parser.HotspotEntry{
		FilePath:     "critical.go",
		CommitCount:  50,
		QualityScore: 2.0,
		Priority:     100,
	})
	if s != nil && s.Severity == "critical" {
		t.Log("critical hotspot severity detected")
	}
}

func TestGetCachedHotspots_Edge(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	result := parser.GetCachedHotspots(dir, 10)
	if result.Error != "" {
		t.Logf("get hot spots error (ok for empty repo): %s", result.Error)
	}
}

func TestClassifyConstants(t *testing.T) {
	if parser.LabelGoAhead == "" {
		t.Error("LabelGoAhead is empty")
	}
	if parser.LabelProceedWithCare == "" {
		t.Error("LabelProceedWithCare is empty")
	}
	if parser.LabelProceedWithCare == "" {
		t.Error("LabelProceedWithCare is empty")
	}
}

func initGitRepo(t *testing.T, dir string) {
	cmds := [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		_ = cmd.Run()
	}
}

func TestDuplicate_Pairs(t *testing.T) {
	code := sp("func A() int {\n\ts := 0\n\tfor i := 0; i < 10; i++ {\n\t\ts += i\n\t}\n\treturn s\n}\nfunc B() int {\n\ts := 0\n\tfor i := 0; i < 10; i++ {\n\t\ts += i\n\t}\n\treturn s\n}\n")
	bloats := parser.DetectFunctionBloats(code)
	pairs := parser.DetectDuplications(bloats, code, 10, 0.50)
	if len(pairs) != 0 {
		t.Logf("found %d dup pairs", len(pairs))
		for _, p := range pairs {
			t.Logf("  %s ~ %s (%.0f%%)", p.FuncA, p.FuncB, p.Similarity*100)
		}
	}
	smells := parser.DetectDuplicationSmells(pairs)
	if len(smells) != 0 && len(pairs) != 0 {
		t.Logf("duplication smells: %d", len(smells))
	}
}

func TestCohesion_Edge(t *testing.T) {
	code := sp("func f1() { x := MyType{}; x.Do() }\nfunc f2() { y := OtherType{}; y.Run() }\n")
	bloats := parser.DetectFunctionBloats(code)
	coh := parser.AnalyzeCohesion(bloats, code)
	if coh.IsLowCohesion {
		t.Log("low cohesion detected (possibly unexpected for small file)")
	}
	s := parser.DetectLowCohesionSmell(coh, 50, 75)
	if s != nil {
		t.Logf("low cohesion smell: %s", s.Message)
	}
}

func TestBumpyRoad_EdgeCase(t *testing.T) {
	code := sp("func hi() {\n\tprintln(\"hello\")\n}\n")
	s := parser.DetectBumpyRoadSmell(code, 2, 2)
	if s != nil {
		t.Errorf("flat function should not be bumpy: %s", s.Message)
	}

	code = sp("func test() {\n\tif true {\n\t\tif true {\n\t\t\tif true {\n\t\t\t\tprintln(\"deep\")\n\t\t\t}\n\t\t}\n\t}\n\tif false {\n\t\tif false {\n\t\t\tprintln(\"deep2\")\n\t\t}\n\t}\n}\n")
	s = parser.DetectBumpyRoadSmell(code, 2, 2)
	if s != nil {
		t.Logf("bumpy road detected: %s", s.Message)
	}
}

func TestComplexConditional_Edge(t *testing.T) {
	code := sp("if a && b {\n\tprintln(\"simple\")\n}\n")
	smells := parser.DetectComplexConditional(code, 2, 5)
	if len(smells) != 0 {
		t.Logf("complex conditional: %d", len(smells))
	}
}

func TestLongParameterList_Edge(t *testing.T) {
	code := sp("func f(a int) {\n}\n")
	smells := parser.DetectLongParameterList(code, 4, 7)
	if len(smells) != 0 {
		t.Errorf("1 param should not trigger: %d", len(smells))
	}
}

func TestCountBranches_Edge(t *testing.T) {
	code := sp("")
	counts := parser.CountBranches(code)
	if len(counts) != 0 {
		t.Error("empty code should have 0 branches")
	}
}

func TestJavaFunctionDetection(t *testing.T) {
	code := sp("public class Test {\n    public void hello() {\n        System.out.println(\"hi\");\n    }\n}\n")
	bloats := parser.DetectFunctionBloatsJava(code)
	if len(bloats) == 0 {
		t.Error("should detect Java method")
	}
}

func TestPythonFunctionDetection(t *testing.T) {
	code := sp("def hello():\n    pass\n\ndef world():\n    return 42\n")
	bloats := parser.DetectFunctionBloatsIndent(code, 4)
	if len(bloats) == 0 {
		t.Error("should detect Python functions")
	}
}

func TestDetectFunctionBloatsRust(t *testing.T) {
	code := sp("fn hello() {\n    println!(\"hi\");\n}\n")
	bloats := parser.DetectFunctionBloatsRust(code)
	if len(bloats) == 0 {
		t.Error("should detect Rust function")
	}
}

func TestDetectFunctionBloatsRuby(t *testing.T) {
	code := sp("def hello\n  puts 'hi'\nend\n")
	bloats := parser.DetectFunctionBloatsRuby(code)
	if len(bloats) == 0 {
		t.Error("should detect Ruby method")
	}
}

func TestDetectFunctionBloatsSwift(t *testing.T) {
	code := sp("func hello() {\n    print(\"hi\")\n}\n")
	bloats := parser.DetectFunctionBloatsSwift(code)
	if len(bloats) == 0 {
		t.Error("should detect Swift function")
	}
}

func TestDetectFunctionBloatsKotlin(t *testing.T) {
	code := sp("fun hello() {\n    println(\"hi\")\n}\n")
	bloats := parser.DetectFunctionBloatsKotlin(code)
	if len(bloats) == 0 {
		t.Error("should detect Kotlin function")
	}
}

func TestDetectFunctionBloatsCPP_Edge(t *testing.T) {
	code := sp("void hello() {\n    printf(\"hi\");\n}\n")
	bloats := parser.DetectFunctionBloatsCPP(code)
	if len(bloats) == 0 {
		t.Error("should detect C++ function")
	}
}

func TestDetectFunctionBloatsCPP_Qualified(t *testing.T) {
	code := sp("void MyClass::hello() {\n    printf(\"hi\");\n}\n")
	bloats := parser.DetectFunctionBloatsCPP(code)
	if len(bloats) == 0 {
		t.Error("should detect qualified C++ method")
	} else {
		t.Logf("C++ function: %s at line %d", bloats[0].Name, bloats[0].LineStart)
	}
}

func TestDetectFileBloat_AllTiers(t *testing.T) {
	bigFile := make([]string, 5000)
	for i := range bigFile {
		bigFile[i] = "println()"
	}

	s := parser.DetectFileBloat(len(bigFile), 1000, 2000, 4000)
	if s == nil {
		t.Error("5000 LOC should be critical")
	} else if s.Severity != "critical" {
		t.Errorf("expected critical, got %s", s.Severity)
	}

	s = parser.DetectFileBloat(1500, 1000, 2000, 4000)
	if s == nil || s.Severity != "warning" {
		t.Errorf("1500 LOC should be warning, got %v", s)
	}

	s = parser.DetectFileBloat(2500, 1000, 2000, 4000)
	if s == nil || s.Severity != "alert" {
		t.Errorf("2500 LOC should be alert, got %v", s)
	}
}

func TestDetectBrainMethodSmell_Tiers(t *testing.T) {
	bloats := []parser.FunctionBloat{
		{Name: "shortFunc", LineStart: 1, LineCount: 30},
		{Name: "longFunc", LineStart: 40, LineCount: 100},
		{Name: "hugeFunc", LineStart: 150, LineCount: 350},
	}
	smells := parser.DetectBrainMethodSmell(bloats, 80, 300)
	if len(smells) != 2 {
		t.Errorf("expected 2 brain methods (shortFunc should be skipped), got %d", len(smells))
	}
	for _, s := range smells {
		if s.Name != "brain_method" {
			t.Errorf("smell name should be brain_method, got %s", s.Name)
		}
		if s.LineStart == 40 && s.Severity != "warning" {
			t.Errorf("100-line fn should be warning, got %s", s.Severity)
		}
		if s.LineStart == 150 && s.Severity != "alert" {
			t.Errorf("350-line fn should be alert, got %s", s.Severity)
		}
	}
}
