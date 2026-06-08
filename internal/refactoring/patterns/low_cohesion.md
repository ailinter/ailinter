# Low Cohesion — Extract Unrelated Functions

## What

A package or file contains functions that serve unrelated purposes. For example, a `utils.go` with string helpers, file I/O, math operations, and HTTP helpers all mixed together.

## Why It's Harmful

- **Discovery failure**: Developers and AI models miss reusable code because it's in an unrelated file.
- **Context waste**: Opening a 500-line "misc" file consumes tokens on unrelated code.
- **Testing friction**: Tests for unrelated functions share the same file, making `go test -run` filtering awkward.

## Step-by-Step Refactoring

### 1. Identify Cohesion Groups

Group functions by the types they operate on and the domain they belong to:

```go
// BEFORE: Low cohesion — all in utils.go
package util

func ReverseString(s string) string { ... }
func ReadFile(path string) ([]byte, error) { ... }
func Sum(nums []float64) float64 { ... }
func ParseJSON(data []byte, v any) error { ... }
func WriteCSV(w io.Writer, rows [][]string) error { ... }
```

### 2. Split by Responsibility

```go
// AFTER: Split into focused files
package strutil   // strings.go
func Reverse(s string) string { ... }
func Capitalize(s string) string { ... }

package fileutil   // files.go
func Read(path string) ([]byte, error) { ... }
func Exists(path string) bool { ... }

package mathutil   // math.go
func Sum(nums []float64) float64 { ... }
func Mean(nums []float64) float64 { ... }
```

### 3. Name Packages After Domain

A package name should tell you what it does: `strutil`, `fileutil`, `httputil`. Avoid generic names like `util`, `helpers`, or `common`.

## Verification

- [ ] Each package has a single domain focus
- [ ] No file contains functions from unrelated domains
- [ ] Package name describes the domain, not "misc" terminology
- [ ] Re-run `analyze_code` — Low Cohesion score should improve
