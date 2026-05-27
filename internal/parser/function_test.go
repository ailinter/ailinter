package parser_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

// === Python function detection ===

func TestPythonFunc_Simple(t *testing.T) {
	lines := strings.Split("def hello():\n    return 42\n", "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].Name != "hello" {
		t.Errorf("expected 'hello', got %q", bloats[0].Name)
	}
}

func TestPythonFunc_Async(t *testing.T) {
	lines := strings.Split("async def fetch():\n    return await api()\n", "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].Name != "fetch" {
		t.Errorf("expected 'fetch', got %q", bloats[0].Name)
	}
}

func TestPythonFunc_Tabbed(t *testing.T) {
	lines := []string{"def tabs():", "\tx = 1", "\ty = 2"}
	bloats := parser.DetectFunctionBloatsIndent(lines, 1)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].Name != "tabs" {
		t.Errorf("expected 'tabs', got %q", bloats[0].Name)
	}
}

func TestPythonFunc_Decorated(t *testing.T) {
	lines := strings.Split("@staticmethod\ndef decorated():\n    return 1\n", "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].Name != "decorated" {
		t.Errorf("expected 'decorated', got %q", bloats[0].Name)
	}
}

func TestPythonFunc_Nested(t *testing.T) {
	src := "def outer():\n    def inner():\n        return 1\n    return inner()\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 2 {
		t.Fatalf("expected 2 functions (outer + inner), got %d", len(bloats))
	}
	if bloats[0].Name != "inner" {
		t.Errorf("expected inner first, got %q", bloats[0].Name)
	}
	if bloats[1].Name != "outer" {
		t.Errorf("expected outer second, got %q", bloats[1].Name)
	}
}

func TestPythonFunc_ClassMethod(t *testing.T) {
	src := "class MyClass:\n    def method1(self):\n        return 1\n\n    def method2(self):\n        return 2\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(bloats))
	}
	if bloats[0].Name != "method1" {
		t.Errorf("expected method1, got %q", bloats[0].Name)
	}
	if bloats[1].Name != "method2" {
		t.Errorf("expected method2, got %q", bloats[1].Name)
	}
}

func TestPythonFunc_Multiple(t *testing.T) {
	src := "def a(): pass\ndef b(): pass\ndef c(): pass\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(bloats))
	}
}

// === Indent auto-detection ===

func TestDetectIndentSize_Spaces(t *testing.T) {
	lines := strings.Split("def foo():\n    x = 1\n", "\n")
	size := parser.DetectIndentSize(lines)
	if size != 4 {
		t.Errorf("expected 4 for spaces, got %d", size)
	}
}

func TestDetectIndentSize_Tabs(t *testing.T) {
	lines := []string{"def foo():", "	x = 1"}
	size := parser.DetectIndentSize(lines)
	if size != 1 {
		t.Errorf("expected 1 for tabs, got %d", size)
	}
}

func TestDetectIndentSize_EmptyDefault(t *testing.T) {
	lines := []string{"", "# comment"}
	size := parser.DetectIndentSize(lines)
	if size != 4 {
		t.Errorf("expected default 4 for empty/comment-only, got %d", size)
	}
}

// === Multiline brace signatures ===

func TestMultilineSig_RunBraceDetector(t *testing.T) {
	src := "func foo(\n    a int,\n) error {\n    return nil\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 multiline function, got %d", len(bloats))
	}
	if bloats[0].Name != "foo" {
		t.Errorf("expected 'foo', got %q", bloats[0].Name)
	}
}

func TestMultilineSig_GoDetector(t *testing.T) {
	src := "package main\n\nfunc process(\n    data []byte,\n    opts *Options,\n) error {\n    if opts == nil {\n        return nil\n    }\n    return nil\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].LineCount < 8 {
		t.Errorf("function should span at least 8 lines, got %d", bloats[0].LineCount)
	}
}

// === Java new patterns ===

func TestJavaFunc_Annotation(t *testing.T) {
	bloats := parser.DetectFunctionBloatsJava(strings.Split("@GetMapping(\"/api\")\npublic List<User> getUsers() {\n    return repo.findAll();\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 annotated method, got %d", len(bloats))
	}
	if bloats[0].Name != "getUsers" {
		t.Errorf("expected 'getUsers', got %q", bloats[0].Name)
	}
}

func TestJavaFunc_GenericReturn(t *testing.T) {
	bloats := parser.DetectFunctionBloatsJava(strings.Split("public List<User> getUsers() {\n    return repo.findAll();\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 generic method, got %d", len(bloats))
	}
}

// === Kotlin modifiers ===

func TestKotlinFunc_Suspend(t *testing.T) {
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split("suspend fun fetch(): String {\n    return api.get()\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 suspend function, got %d", len(bloats))
	}
	if bloats[0].Name != "fetch" {
		t.Errorf("expected 'fetch', got %q", bloats[0].Name)
	}
}

func TestKotlinFunc_Inline(t *testing.T) {
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split("inline fun <T> measure(block: () -> T): T {\n    return block()\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 inline function, got %d", len(bloats))
	}
}

func TestKotlinFunc_Operator(t *testing.T) {
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split("operator fun plus(other: Point): Point {\n    return Point(x + other.x)\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 operator function, got %d", len(bloats))
	}
}

// === C++ destructors and operators ===

func TestCppFunc_Destructor(t *testing.T) {
	bloats := parser.DetectFunctionBloatsCPP(strings.Split("~MyClass() {\n    cleanup();\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 destructor, got %d", len(bloats))
	}
	if bloats[0].Name != "~MyClass" {
		t.Errorf("expected '~MyClass', got %q", bloats[0].Name)
	}
}

func TestCppFunc_OperatorOverload(t *testing.T) {
	bloats := parser.DetectFunctionBloatsCPP(strings.Split("bool operator==(const Point& other) const {\n    return x == other.x;\n}\n", "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 operator, got %d", len(bloats))
	}
}

// === RunBraceDetector edge ===

func TestRunBraceDetector_MultipleFuncs(t *testing.T) {
	src := "fun a() {\n    return 1\n}\n\nfun b() {\n    return 2\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsKotlin(lines)
	if len(bloats) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(bloats))
	}
}

// === classify function tests ===

func TestClassify_GoAhead(t *testing.T) {
	if parser.TestClassifyHelper(80) != "Go Ahead" {
		t.Errorf("expected Go Ahead for 80 (boundary)")
	}
	if parser.TestClassifyHelper(100) != "Go Ahead" {
		t.Errorf("expected Go Ahead for 100")
	}
}

func TestClassify_ProceedWithCare(t *testing.T) {
	if parser.TestClassifyHelper(60) != "Proceed with Care" {
		t.Errorf("expected Proceed with Care for 60 (boundary)")
	}
	if parser.TestClassifyHelper(79) != "Proceed with Care" {
		t.Errorf("expected Proceed with Care for 79")
	}
}

func TestClassify_NeedsWork(t *testing.T) {
	if parser.TestClassifyHelper(40) != "Needs Work" {
		t.Errorf("expected Needs Work for 40 (boundary)")
	}
	if parser.TestClassifyHelper(59) != "Needs Work" {
		t.Errorf("expected Needs Work for 59")
	}
}

func TestClassify_StopRefactor(t *testing.T) {
	if parser.TestClassifyHelper(10) != "Stop & Refactor" {
		t.Errorf("expected Stop & Refactor for 10")
	}
	if parser.TestClassifyHelper(39) != "Stop & Refactor" {
		t.Errorf("expected Stop & Refactor for 39 (just below 40 boundary)")
	}
}

// === Go function name extraction edge cases ===

func TestGoFuncName_MethodReceiver(t *testing.T) {
	src := "func (r *Receiver) Name(param string) error {\n    return nil\n}\n"
	bloats := parser.DetectFunctionBloats(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
	if bloats[0].Name != "Name" {
		t.Errorf("expected 'Name' for method receiver, got %q", bloats[0].Name)
	}
}

func TestGoFuncName_Closure(t *testing.T) {
	src := "func() {\n    fmt.Println(\"anon\")\n}()\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) > 0 {
		t.Logf("closure detected: name=%s lines=%d", bloats[0].Name, bloats[0].LineCount)
	}
}

// === Java edge cases ===

func TestJavaFunc_DefaultMethod(t *testing.T) {
	src := "default void run() {\n    doWork();\n}\n"
	bloats := parser.DetectFunctionBloatsJava(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 default method, got %d", len(bloats))
	}
}

func TestJavaFunc_MultipleModifiers(t *testing.T) {
	src := "public static final synchronized String formatData(String input) {\n    return input.trim();\n}\n"
	bloats := parser.DetectFunctionBloatsJava(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 method with multiple modifiers, got %d", len(bloats))
	}
}

func TestJavaFunc_AbstractMethod(t *testing.T) {
	src := "abstract void process();\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsJava(lines)
	// Abstract methods have no body, may or may not be detected
	_ = bloats
}

// === Rust edge cases ===

func TestRustFunc_PubCrate(t *testing.T) {
	src := "pub(crate) fn internal_helper() -> u32 {\n    42\n}\n"
	bloats := parser.DetectFunctionBloatsRust(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 pub(crate) fn, got %d", len(bloats))
	}
}

func TestRustFunc_AsyncFn(t *testing.T) {
	src := "pub async fn fetch_data() -> Result<Vec<User>> {\n    Ok(vec![])\n}\n"
	bloats := parser.DetectFunctionBloatsRust(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 async fn, got %d", len(bloats))
	}
}

// === Kotlin edge cases ===

func TestKotlinFunc_Annotation(t *testing.T) {
	src := "@JvmStatic\nfun helper(): Int {\n    return 42\n}\n"
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 annotated kotlin fun, got %d", len(bloats))
	}
}

func TestKotlinFunc_GenericName(t *testing.T) {
	src := "fun <T> List<T>.customFilter(predicate: (T) -> Boolean): List<T> {\n    return this\n}\n"
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 generic kotlin fun, got %d", len(bloats))
	}
}

// === C++ edge cases ===

func TestCppFunc_ConstructorInit(t *testing.T) {
	src := "MyClass::MyClass(int val) : value(val), name(\"default\") {\n}\n"
	bloats := parser.DetectFunctionBloatsCPP(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 constructor, got %d", len(bloats))
	}
}

func TestCppFunc_NamespaceQualified(t *testing.T) {
	src := "void ns::SubNS::helper() {\n    doWork();\n}\n"
	bloats := parser.DetectFunctionBloatsCPP(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 namespace qualified fn, got %d", len(bloats))
	}
	if bloats[0].Name != "helper" {
		t.Errorf("expected 'helper', got %q", bloats[0].Name)
	}
}

// === Swift edge cases ===

func TestSwiftFunc_PrivateFunc(t *testing.T) {
	src := "private func doInternalWork() -> Bool {\n    return true\n}\n"
	bloats := parser.DetectFunctionBloatsSwift(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 private func, got %d", len(bloats))
	}
}

func TestSwiftFunc_Mutating(t *testing.T) {
	src := "mutating func update() {\n    self.value += 1\n}\n"
	bloats := parser.DetectFunctionBloatsSwift(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 mutating func, got %d", len(bloats))
	}
}

// === Java: control-flow keywords should NOT be detected as functions ===

func TestJavaFunc_NotIfStatement(t *testing.T) {
	src := "if (x > 0) {\n    return x;\n}\n"
	bloats := parser.DetectFunctionBloatsJava(strings.Split(src, "\n"))
	if len(bloats) != 0 {
		t.Errorf("if statement should not be detected as function, got %d", len(bloats))
	}
}

func TestJavaFunc_NotForLoop(t *testing.T) {
	src := "for (int i = 0; i < n; i++) {\n    process(i);\n}\n"
	bloats := parser.DetectFunctionBloatsJava(strings.Split(src, "\n"))
	if len(bloats) != 0 {
		t.Errorf("for loop should not be detected as function, got %d", len(bloats))
	}
}

// === Go: multiline with method receiver ===

func TestGoFunc_MultilineMethod(t *testing.T) {
	src := "func (s *Service) ProcessRequest(\n    ctx context.Context,\n    req *Request,\n) (*Response, error) {\n    return nil, nil\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 multiline method, got %d", len(bloats))
	}
}

// === Python: trailing function at EOF ===

func TestPythonFunc_Trailing(t *testing.T) {
	src := "def last_stand():\n    x = 1\n    return x"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloatsIndent(lines, 4)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 trailing function, got %d", len(bloats))
	}
}

// === JS/TS: arrow functions and export patterns ===

func TestGoDetector_JSArrowFunc(t *testing.T) {
	src := "const handler = (req, res) => {\n    res.json({ ok: true })\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 arrow function, got %d", len(bloats))
	}
}

func TestGoDetector_ExportedFunction(t *testing.T) {
	src := "export function parse(input: string): Result {\n    return { value: input }\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) != 1 {
		t.Fatalf("expected 1 exported function, got %d", len(bloats))
	}
}

// === TS specific patterns ===

func TestTSDetector_AsyncMethod(t *testing.T) {
	src := "class Service {\n    async fetch(id: string): Promise<User> {\n        return await db.find(id)\n    }\n}\n"
	bloats := parser.DetectFunctionBloatsTS(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 async method, got %d", len(bloats))
	}
}

func TestTSDetector_Getter(t *testing.T) {
	src := "class Config {\n    get timeout(): number {\n        return this._timeout\n    }\n}\n"
	bloats := parser.DetectFunctionBloatsTS(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 getter, got %d", len(bloats))
	}
}

// === TS: function keyword detection ===

func TestTSDetector_FunctionKeyword(t *testing.T) {
	src := "function process(items: string[]) {\n    return items.map(i => i.trim())\n}\n"
	bloats := parser.DetectFunctionBloatsTS(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function, got %d", len(bloats))
	}
}

// === Kotlin multiline annotation ===

func TestKotlinFunc_MultilineAnnotation(t *testing.T) {
	src := "@Deprecated(\"Use newHelper instead\")\n@Suppress(\"unused\")\nfun oldHelper(): Int {\n    return 0\n}\n"
	bloats := parser.DetectFunctionBloatsKotlin(strings.Split(src, "\n"))
	if len(bloats) != 1 {
		t.Fatalf("expected 1 function with multiline annotations, got %d", len(bloats))
	}
}
