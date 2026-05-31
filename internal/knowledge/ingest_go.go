package knowledge

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// funcInfo tracks a parsed function for call graph resolution.
type funcInfo struct {
	pkg  string
	name string
	recv string // receiver type, empty for functions
	file string
	line int
}

// fileIngestState holds the computed state for ingesting a single Go file.
type fileIngestState struct {
	graph      *Graph
	fset       *token.FileSet
	relPath    string
	pkgName    string
	isTestFile bool
	pkgNodeID  string
	fileNodeID string
}

// IngestGoCodebase walks the given directory and ingests all .go files into the graph.
func IngestGoCodebase(graph *Graph, rootDir string) error {
	Logf("ingesting Go codebase from %s", rootDir)

	pkgFuncs := make(map[string][]funcInfo)
	allFuncs := ingestGoFiles(graph, rootDir, pkgFuncs)
	buildCallGraph(graph, rootDir, pkgFuncs)

	Logf("ingested %d functions from Go codebase", len(allFuncs))
	return nil
}

// ingestGoFiles walks rootDir and ingests all .go files, returning the function list.
func ingestGoFiles(graph *Graph, rootDir string, pkgFuncs map[string][]funcInfo) []funcInfo {
	var allFuncs []funcInfo
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			if err == nil && info.IsDir() && skipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		funcs, ingestErr := ingestGoFile(graph, rootDir, path)
		if ingestErr != nil {
			Logf("warning: could not ingest %s: %v", relativePath(rootDir, path), ingestErr)
			return nil
		}
		allFuncs = append(allFuncs, funcs...)
		for _, f := range funcs {
			pkgFuncs[f.pkg] = append(pkgFuncs[f.pkg], f)
		}
		graph.SourceFiles[path] = info.ModTime()
		return nil
	})
	return allFuncs
}

// skipDir checks if a directory should be skipped during walk.
func skipDir(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" || base == "testdata"
}

// ingestGoFile parses a single .go file and adds its nodes to the graph.
func ingestGoFile(graph *Graph, rootDir, path string) ([]funcInfo, error) {
	relPath := relativePath(rootDir, path)

	_, fset, f, err := parseGoFile(path)
	if err != nil {
		return nil, err
	}

	isTestFile := strings.HasSuffix(path, "_test.go")
	state := &fileIngestState{
		graph:      graph,
		fset:       fset,
		relPath:    relPath,
		pkgName:    f.Name.Name,
		isTestFile: isTestFile,
	}
	addPkgNode(state, path, strings.Contains(relPath, "internal/"))
	addFileNode(state)
	addImportNodes(graph, f, state.fileNodeID)
	funcs := addFuncNodes(state, f)

	return funcs, nil
}

// relativePath computes the relative path from rootDir or falls back to the absolute path.
func relativePath(rootDir, path string) string {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return path
	}
	return relPath
}

// parseGoFile reads and parses a .go file, returning the file set, AST, and any error.
func parseGoFile(path string) ([]byte, *token.FileSet, *ast.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read file: %w", err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.ParseComments|parser.AllErrors)
	if err != nil && f == nil {
		return nil, nil, nil, fmt.Errorf("parse file: %w", err)
	}
	return data, fset, f, nil
}

// addPkgNode adds a package node to the graph if it doesn't already exist.
func addPkgNode(state *fileIngestState, path string, isInternal bool) {
	pkgNodeID := slug("pkg", state.pkgName)
	if _, exists := state.graph.GetNode(pkgNodeID); exists {
		state.pkgNodeID = pkgNodeID
		return
	}
	state.pkgNodeID = pkgNodeID
	pkgNodeLabel := state.pkgName
	if isInternal {
		pkgNodeLabel = "internal/" + state.pkgName
	}
	state.graph.AddNode(Node{
		ID:    pkgNodeID,
		Type:  NodePackage,
		Label: pkgNodeLabel,
		Properties: map[string]interface{}{
			"path":    state.relPath,
			"file":    path,
			"is_test": state.isTestFile,
		},
	})
}

// addFileNode adds a file or test node to the graph and links it to the package.
func addFileNode(state *fileIngestState) {
	fid := slug("file", state.relPath)
	state.fileNodeID = fid
	if state.isTestFile {
		state.graph.AddNode(Node{
			ID:    fid,
			Type:  NodeTest,
			Label: state.relPath,
			Properties: map[string]interface{}{
				"path":    state.relPath,
				"package": state.pkgName,
				"is_test": true,
			},
		})
		if pkgNode, exists := state.graph.GetNode(state.pkgNodeID); exists && pkgNode.Type == NodePackage {
			state.graph.AddEdge(fid, state.pkgNodeID, EdgeTests, nil)
		}
	} else {
		state.graph.AddNode(Node{
			ID:    fid,
			Type:  NodeFile,
			Label: state.relPath,
			Properties: map[string]interface{}{
				"path":    state.relPath,
				"package": state.pkgName,
				"is_test": false,
			},
		})
	}
	state.graph.AddEdge(fid, state.pkgNodeID, EdgeContains, nil)
}

// addImportNodes adds import nodes and edges for all imports in the file.
func addImportNodes(graph *Graph, f *ast.File, fileNodeID string) {
	for _, imp := range f.Imports {
		impPath := strings.Trim(imp.Path.Value, "\"")
		impPkg := extractPackageName(impPath)
		if impPkg == "" {
			continue
		}
		ensureImportNode(graph, impPath, impPkg)
		graph.AddEdge(fileNodeID, slug("pkg", impPkg), EdgeImports, map[string]interface{}{
			"import_path": impPath,
		})
	}
}

// ensureImportNode adds an import package node if it doesn't already exist.
func ensureImportNode(graph *Graph, impPath, impPkg string) {
	impNodeID := slug("pkg", impPkg)
	if _, exists := graph.GetNode(impNodeID); exists {
		return
	}
	isExternal := !strings.Contains(impPath, "github.com/ailinter") && !strings.HasPrefix(impPath, "internal/")
	graph.AddNode(Node{
		ID:    impNodeID,
		Type:  NodePackage,
		Label: impPkg,
		Properties: map[string]interface{}{
			"import_path": impPath,
			"is_external": isExternal,
		},
	})
}

// addFuncNodes adds function and method nodes from the parsed file to the graph.
func addFuncNodes(state *fileIngestState, f *ast.File) []funcInfo {
	var funcs []funcInfo
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcs = append(funcs, addFuncNode(state, funcDecl))
	}
	return funcs
}

// addFuncNode adds a single function/method node to the graph and returns its info.
func addFuncNode(state *fileIngestState, funcDecl *ast.FuncDecl) funcInfo {
	recv := extractReceiver(funcDecl)
	funcName := funcDecl.Name.Name
	line := state.fset.Position(funcDecl.Pos()).Line

	funcID, funcLabel := funcIdentifiers(state.pkgName, recv, funcName)

	state.graph.AddNode(Node{
		ID:    funcID,
		Type:  NodeFunction,
		Label: funcLabel,
		Properties: map[string]interface{}{
			"package":     state.pkgName,
			"receiver":    recv,
			"name":        funcName,
			"file":        state.relPath,
			"line":        line,
			"is_test":     state.isTestFile,
			"is_exported": ast.IsExported(funcName),
		},
	})
	state.graph.AddEdge(state.pkgNodeID, funcID, EdgeContains, nil)
	state.graph.AddEdge(state.fileNodeID, funcID, EdgeContains, nil)

	return funcInfo{
		pkg:  state.pkgName,
		name: funcName,
		recv: recv,
		file: state.relPath,
		line: line,
	}
}

// funcIdentifiers builds the node ID and label for a function/method node.
func funcIdentifiers(pkgName, recv, funcName string) (string, string) {
	if recv != "" {
		recvClean := strings.TrimPrefix(recv, "*")
		return slug("func", pkgName, recvClean, funcName),
			fmt.Sprintf("%s.(%s).%s", pkgName, recv, funcName)
	}
	return slug("func", pkgName, funcName),
		fmt.Sprintf("%s.%s", pkgName, funcName)
}

// extractReceiver extracts the receiver type string from a function declaration.
func extractReceiver(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return ""
	}
	return formatExpr(funcDecl.Recv.List[0].Type)
}

// buildCallGraph walks files to find CALLS edges between functions.
func buildCallGraph(graph *Graph, rootDir string, pkgFuncs map[string][]funcInfo) {
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		processFileCalls(graph, path, rootDir, pkgFuncs)
		return nil
	})
}

// processFileCalls parses a .go file and adds CALLS edges for all its functions.
func processFileCalls(graph *Graph, path, rootDir string, pkgFuncs map[string][]funcInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.ParseComments|parser.AllErrors)
	if err != nil || f == nil {
		return
	}
	relPath, _ := filepath.Rel(rootDir, path)
	pkgName := f.Name.Name

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		addFunctionCalls(graph, funcDecl, relPath, pkgName, pkgFuncs)
	}
}

// addFunctionCalls inspects a function body for all call expressions and adds CALLS edges.
func addFunctionCalls(graph *Graph, funcDecl *ast.FuncDecl, relPath, pkgName string, pkgFuncs map[string][]funcInfo) {
	callerID := callerID(funcDecl, pkgName)
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		addCallEdge(graph, call, callerID, relPath, pkgName, pkgFuncs)
		return true
	})
}

// callerID builds the graph node ID for a function.
func callerID(funcDecl *ast.FuncDecl, pkgName string) string {
	recv := extractReceiver(funcDecl)
	if recv != "" {
		recvClean := strings.TrimPrefix(recv, "*")
		return slug("func", pkgName, recvClean, funcDecl.Name.Name)
	}
	return slug("func", pkgName, funcDecl.Name.Name)
}

// addCallEdge resolves a call expression and adds a CALLS edge if the callee is known.
func addCallEdge(graph *Graph, call *ast.CallExpr, callerID, relPath, pkgName string, pkgFuncs map[string][]funcInfo) {
	calleeName := extractCallee(call)
	if calleeName == "" {
		return
	}
	// Check same-package functions
	if funcs, ok := pkgFuncs[pkgName]; ok {
		for _, cf := range funcs {
			if cf.name != calleeName {
				continue
			}
			calleeID := makeCalleeID(cf)
			if callerID != calleeID {
				graph.AddEdge(callerID, calleeID, EdgeCalls, map[string]interface{}{
					"caller_file": relPath,
				})
			}
			break
		}
	}
	// Check package-qualified calls like log.Printf
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	calleeID := slug("func", pkgIdent.Name, sel.Sel.Name)
	if _, exists := graph.GetNode(calleeID); exists {
		graph.AddEdge(callerID, calleeID, EdgeCalls, map[string]interface{}{
			"caller_file": relPath,
		})
	}
}

// makeCalleeID builds the graph node ID for a callee's funcInfo.
func makeCalleeID(cf funcInfo) string {
	if cf.recv != "" {
		recvClean := strings.TrimPrefix(cf.recv, "*")
		return slug("func", cf.pkg, recvClean, cf.name)
	}
	return slug("func", cf.pkg, cf.name)
}

// extractCallee extracts the function name from a call expression.
func extractCallee(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			if isStdPkg(ident.Name) {
				return fun.Sel.Name
			}
			return fun.Sel.Name
		}
		return fun.Sel.Name
	default:
		return ""
	}
}

func isStdPkg(name string) bool {
	std := []string{"fmt", "os", "io", "net", "http", "strings", "strconv", "time",
		"encoding", "json", "xml", "regexp", "sort", "sync", "errors", "log",
		"context", "bytes", "bufio", "math", "rand", "crypto", "flag"}
	for _, s := range std {
		if name == s {
			return true
		}
	}
	return false
}

// extractPackageName extracts the short package name from an import path.
func extractPackageName(importPath string) string {
	parts := strings.Split(importPath, "/")
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	if strings.HasPrefix(name, "v") && len(name) > 1 && len(parts) > 1 {
		name = parts[len(parts)-2]
	}
	return name
}

// formatExpr converts an AST expression to a string (for receiver types).
func formatExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatExpr(t.X)
	case *ast.IndexExpr:
		return formatExpr(t.X) + "[" + formatExpr(t.Index) + "]"
	case *ast.SelectorExpr:
		return formatExpr(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatExpr(t.Elt)
	case *ast.MapType:
		return "map[" + formatExpr(t.Key) + "]" + formatExpr(t.Value)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
