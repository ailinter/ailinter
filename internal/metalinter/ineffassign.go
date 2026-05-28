package metalinter

import ineffAssignPkg "github.com/gordonklaus/ineffassign/pkg/ineffassign"

// ineffassignAnalyzer is the embedded ineffassign analyzer.
var ineffassignAnalyzer = ineffAssignPkg.Analyzer
