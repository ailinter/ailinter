# Lazy Element — Inline Function or Delete

## What
A function or method that is too small to justify its existence — typically fewer than 3 meaningful lines. It adds indirection without abstraction.

## Why It's Harmful
- **Indirection without benefit**: Each function call is a mental jump. If the body is as clear as the name, the function just adds overhead.
- **Token waste**: AI models consume tokens reading function signatures, docstrings, and call sites. A 1-line getter costs ~30 tokens per call site.
- **Call stack noise**: Debugging through 15 tiny functions is harder than reading 3 well-structured blocks.

## Step-by-Step Refactoring

### 1. Identify Trivial Functions
```go
// BEFORE: Lazy elements
func (u *User) GetName() string {
    return u.name
}

func (u *User) SetName(n string) {
    u.name = n
}

func isNotEmpty(s string) bool {
    return s != ""
}

func add(a, b int) int {
    return a + b
}
```

### 2. Inline When the Body Is Clearer Than the Call
```go
// AFTER: Inline the trivial cases
// Remove GetName/SetName — use u.Name directly
type User struct {
    Name string  // exported field, no getter needed
}

// Remove isNotEmpty — use s != "" directly
if name != "" {  // clearer than if isNotEmpty(name)
    // ...
}

// Remove add — use a + b directly
result := a + b  // clearer than result := add(a, b)
```

### 3. Keep Functions That Enforce Invariants
Not all short functions are lazy elements. Keep them when they:
- Validate or transform: `func NewEmail(s string) (Email, error)`
- Encapsulate computation that might change: `func (p *Price) WithTax() float64`
- Are part of an interface contract: `func (h *HealthCheck) Status() string`

```go
// KEEP: Validates invariant
func NewPercentage(v float64) (Percentage, error) {
    if v < 0 || v > 100 {
        return 0, fmt.Errorf("percentage out of range: %v", v)
    }
    return Percentage(v), nil
}
```

### 4. Collapse Single-Use Private Helpers
If a private function is called exactly once and is trivial, inline it:
```go
// BEFORE
func validateEmail(e string) error {
    return checkFormat(e)
}
func checkFormat(e string) error {
    return checkDomain(e)
}

// AFTER
func validateEmail(e string) error {
    if !strings.Contains(e, "@") {
        return errors.New("invalid format")
    }
    if !strings.Contains(e, ".") {
        return errors.New("invalid domain")
    }
    return nil
}
```

## Verification
- [ ] No functions shorter than 3 meaningful lines (excluding constructors and interface stubs)
- [ ] Remaining short functions enforce invariants or hide future change
- [ ] No single-use private helpers that are trivial
- [ ] Re-run `analyze_code` — Lazy Element smell should be resolved
