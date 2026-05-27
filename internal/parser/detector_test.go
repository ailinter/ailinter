package parser_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

func TestBumpyRoad_Detected(t *testing.T) {
	lines := strings.Split("for _, it := range items {\nif it > 0 {\nif it < 100 {\nif it != 50 {\nc += it\n}\n}\n}\n}\n// surface\nfor _, it := range items {\nif it < 0 {\nif it > -100 {\nif it != -50 {\nc -= it\n}\n}\n}\n}", "\n")
	s := parser.DetectBumpyRoadSmell(lines, 2, 2)
	if s == nil {
		t.Fatal("expected bumpy_road smell")
	}
	t.Logf("Bumpy road: %s", s.Message)
}

func TestBumpyRoad_NotDetected(t *testing.T) {
	lines := strings.Split("a := 1\nb := 2\nreturn a + b", "\n")
	s := parser.DetectBumpyRoadSmell(lines, 2, 2)
	if s != nil {
		t.Errorf("flat function should not trigger bumpy_road: %s", s.Message)
	}
}

func TestComplexConditional_Detected(t *testing.T) {
	lines := strings.Split("if a > 0 && b > 0 && c > 0 || d > 0 {", "\n")
	smells := parser.DetectComplexConditional(lines, 2, 5)
	if len(smells) == 0 {
		t.Fatal("expected complex_conditional smell")
	}
	if len(smells) > 1 {
		t.Errorf("expected 1 smell, got %d", len(smells))
	}
}

func TestComplexConditional_NotDetected(t *testing.T) {
	lines := strings.Split("if a > 0 {", "\n")
	smells := parser.DetectComplexConditional(lines, 2, 5)
	if len(smells) > 0 {
		t.Error("simple condition should not trigger")
	}
}

func TestLongParameterList_Detected(t *testing.T) {
	lines := strings.Split("func ManyParams(a string, b int, c float64, d bool, e string) error {", "\n")
	smells := parser.DetectLongParameterList(lines, 4, 7)
	if len(smells) == 0 {
		t.Fatal("expected long_parameter_list")
	}
}

func TestLongParameterList_NotDetected(t *testing.T) {
	lines := strings.Split("func TwoParams(a string, b int) error {", "\n")
	smells := parser.DetectLongParameterList(lines, 4, 7)
	if len(smells) > 0 {
		t.Error("2-param function should not trigger")
	}
}

func TestCountBranches(t *testing.T) {
	lines := strings.Split("if a > 0 {\nfor _, v := range items {\nif v > 0 && v < 100 {\nreturn v\n}\n}\n}\nreturn 0", "\n")
	counts := parser.CountBranches(lines)
	if len(counts) == 0 {
		t.Fatal("expected branch counts")
	}
	t.Logf("Branch counts: %+v", counts)
}

func TestGoFunctionDetection(t *testing.T) {
	src := "func Hello() {\n\tfmt.Println(\"hi\")\n}\nfunc World() {\n\treturn\n}\n"
	bloats := parser.DetectFunctionBloats(strings.Split(src, "\n"))
	if len(bloats) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(bloats))
	}
	if bloats[0].Name != "Hello" {
		t.Errorf("expected Hello, got %s", bloats[0].Name)
	}
	if bloats[1].Name != "World" {
		t.Errorf("expected World, got %s", bloats[1].Name)
	}
}

func TestJSFunctionDetection(t *testing.T) {
	src := "function greet(name) {\n\tconsole.log('hi')\n}\nconst add = (a, b) => {\n\treturn a + b\n}\n"
	bloats := parser.DetectFunctionBloatsTS(strings.Split(src, "\n"))
	if len(bloats) < 2 {
		t.Fatalf("expected at least 2 JS functions, got %d", len(bloats))
	}
	t.Logf("JS functions found: %d", len(bloats))
	for _, b := range bloats {
		t.Logf("  %s: %d lines at L%d", b.Name, b.LineCount, b.LineStart)
	}
}

func TestCppFunctionDetection(t *testing.T) {
	src := "int add(int a, int b) {\n\treturn a + b;\n}\nvoid print(const char* msg) {\n\tprintf(\"%s\", msg);\n}\n"
	bloats := parser.DetectFunctionBloatsCPP(strings.Split(src, "\n"))
	if len(bloats) != 2 {
		t.Fatalf("expected 2 C++ functions, got %d", len(bloats))
	}
}

func TestRustFunctionDetection(t *testing.T) {
	src := "fn main() {\n\tprintln!(\"hi\");\n}\npub fn greet(name: &str) -> String {\n\tformat!(\"Hello {}\", name)\n}\n"
	bloats := parser.DetectFunctionBloatsRust(strings.Split(src, "\n"))
	if len(bloats) < 2 {
		t.Fatalf("expected at least 2 Rust functions, got %d", len(bloats))
	}
}

func TestRubyFunctionDetection(t *testing.T) {
	src := "class Foo\n  def bar\n    puts 'hi'\n  end\n  def baz(x)\n    x * 2\n  end\nend\n"
	bloats := parser.DetectFunctionBloatsRuby(strings.Split(src, "\n"))
	if len(bloats) < 2 {
		t.Fatalf("expected at least 2 Ruby functions, got %d", len(bloats))
	}
	for _, b := range bloats {
		t.Logf("Ruby: %s L%d (%d lines)", b.Name, b.LineStart, b.LineCount)
	}
}

func TestSwiftFunctionDetection(t *testing.T) {
	src := "func greet(name: String) {\n\tprint(\"hi\")\n}\nprivate func add(a: Int, b: Int) -> Int {\n\treturn a + b\n}\n"
	bloats := parser.DetectFunctionBloatsSwift(strings.Split(src, "\n"))
	if len(bloats) < 2 {
		t.Fatalf("expected at least 2 Swift functions, got %d", len(bloats))
	}
}

func TestKotlinFunctionDetection(t *testing.T) {
	src := "fun greet(name: String) {\n\tprintln(\"hi\")\n}\noverride fun toString(): String = \"foo\"\n"
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split(src, "\n"))
	if len(bloats) < 1 {
		t.Fatalf("expected at least 1 Kotlin function, got %d", len(bloats))
	}
	t.Logf("Kotlin functions: %d", len(bloats))
}

func TestLongSwitch_Detected(t *testing.T) {
	code := `func foo(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	case 10:
		return "ten"
	case 11:
		return "eleven"
	default:
		return "unknown"
	}
}`
	lines := strings.Split(code, "\n")
	smells := parser.DetectLongSwitch(lines, 10, 20)
	if len(smells) == 0 {
		t.Fatal("expected long_switch smell")
	}
	s := smells[0]
	t.Logf("Long switch: %s (line %d-%d)", s.Message, s.LineStart, s.LineEnd)
	if !strings.Contains(s.Message, "12 branches") {
		t.Errorf("expected 12 branches, got: %s", s.Message)
	}
}

func TestLongSwitch_NotDetected(t *testing.T) {
	code := `func foo(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	default:
		return "other"
	}
}`
	lines := strings.Split(code, "\n")
	smells := parser.DetectLongSwitch(lines, 10, 20)
	if len(smells) > 0 {
		t.Errorf("small switch should not trigger: %s", smells[0].Message)
	}
}

func TestLongSwitch_PythonMatch(t *testing.T) {
	code := `def handle(code):
    match code:
        case 200:
            return "OK"
        case 301:
            return "Moved"
        case 400:
            return "Bad Request"
        case 401:
            return "Unauthorized"
        case 403:
            return "Forbidden"
        case 404:
            return "Not Found"
        case 500:
            return "Server Error"
        case 502:
            return "Bad Gateway"
        case 503:
            return "Unavailable"
        case 504:
            return "Timeout"
`
	lines := strings.Split(code, "\n")
	smells := parser.DetectLongSwitch(lines, 10, 20)
	if len(smells) == 0 {
		t.Fatal("expected long_switch smell for Python match")
	}
	t.Logf("Python match: %s", smells[0].Message)
}

func TestLongSwitch_JS(t *testing.T) {
	code := `function getColor(n) {
    switch (n) {
        case 1: return "red";
        case 2: return "green";
        case 3: return "blue";
        case 4: return "yellow";
        case 5: return "purple";
        case 6: return "orange";
        case 7: return "pink";
        case 8: return "brown";
        case 9: return "black";
        case 10: return "white";
        case 11: return "gray";
        default: return "unknown";
    }
}`
	lines := strings.Split(code, "\n")
	smells := parser.DetectLongSwitch(lines, 10, 20)
	if len(smells) == 0 {
		t.Fatal("expected long_switch smell for JS switch")
	}
	t.Logf("JS switch: %s", smells[0].Message)
}
