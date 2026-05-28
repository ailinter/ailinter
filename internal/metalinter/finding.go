package metalinter

import "fmt"

// Finding represents a single issue detected by an embedded meta-linter.
type Finding struct {
	Tool     string `json:"tool"`     // "govet", "staticcheck", "gofmt", "misspell", "ineffassign"
	Code     string `json:"code"`     // e.g. "SA1000", "S1017", "formatting"
	Severity string `json:"severity"` // "error", "warning", "info"
	File     string `json:"file_path"`
	Line     int    `json:"line_start"`
	Column   int    `json:"col_start"`
	Message  string `json:"message"`
	Category string `json:"category"` // "bug", "style", "formatting", "unused", "performance"
}

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"

	CategoryBug         = "bug"
	CategoryStyle       = "style"
	CategoryFormatting  = "formatting"
	CategoryUnused      = "unused"
	CategoryPerformance = "performance"
	CategoryCorrectness = "correctness"
)

// ProblemsFormat returns a GCC-style string for IDE problem matchers.
func (f Finding) ProblemsFormat() string {
	col := f.Column
	if col < 1 {
		col = 1
	}
	return fmt.Sprintf("%s:%d:%d: [%s] %s (%s)", f.File, f.Line, col, f.Tool, f.Message, f.Code)
}
