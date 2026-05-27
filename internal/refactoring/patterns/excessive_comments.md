# Excessive Comments — Replace Comments with Clear Names

## What
Comments that explain what the code does, when the code itself could be renamed to be self-documenting. A comment ratio above 30% of total lines indicates the code is not expressive enough.

## Why It's Harmful
- **Lies over time**: Comments rot. Code gets refactored, comments don't. The compiler checks code correctness, not comment accuracy.
- **Signal dilution**: Important comments (why, not what) get buried in noise. Developers learn to ignore all comments.
- **Token waste for AI**: Each comment line costs tokens without adding information the code already provides. Redundant comments consume ~15% of AI context budgets.

## Step-by-Step Refactoring

### 1. Delete Redundant Comments (What = Code)
```go
// BEFORE: Redundant comments
// Check if the user is authenticated
if user.IsAuthenticated {
    // Redirect to dashboard
    http.Redirect(w, r, "/dashboard", 302)
}

// AFTER: Self-documenting code
if user.IsAuthenticated {
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

### 2. Replace What-Comments with Better Names
```go
// BEFORE: Comments explain what
func process(d []Data) {
    // Filter out inactive records
    var active []Data
    for _, item := range d {
        if item.Status == "active" {
            active = append(active, item)
        }
    }

    // Sort by priority descending
    sort.Slice(active, func(i, j int) bool {
        return active[i].Priority > active[j].Priority
    })
}

// AFTER: Extract to named functions
func process(records []Data) {
    active := filterActive(records)
    sortByPriorityDesc(active)
}

func filterActive(records []Data) []Data { ... }
func sortByPriorityDesc(records []Data) { ... }
```

### 3. Keep Why-Comments, Delete What-Comments
```go
// GOOD: Explains why — non-obvious business rule
// Must use explicit JOIN here; the ORM's lazy loading
// causes N+1 queries when Order has >1000 items.
orders := db.Exec("SELECT ... FROM orders JOIN ...")

// BAD: Explains what — the code already says this
// Loop through all orders
for _, order := range orders { ... }
```

### 4. Delete Commented-Out Code
```go
// BEFORE: Zombie code
// func oldImplementation() {
//     // deprecated, remove after Q3
// }

// AFTER: Delete it — git history preserves it
```

### 5. Replace Section Headers with Functions
```go
// BEFORE: Section header comments
func handleRequest(r *Request) {
    // === Validation ===
    if r.Name == "" { return }
    // === Authorization ===
    if !r.HasPermission { return }
    // === Processing ===
    result := computeResult(r)
    // === Response ===
    writeResponse(result)
}

// AFTER: Extract to named functions
func handleRequest(r *Request) {
    if err := validate(r); err != nil { return }
    if err := authorize(r); err != nil { return }
    result := computeResult(r)
    writeResponse(result)
}
```

## Verification
- [ ] Comment ratio below 30% of total lines
- [ ] Every remaining comment explains why, not what
- [ ] No commented-out code blocks
- [ ] Function and variable names are expressive enough to replace section headers
- [ ] Re-run `analyze_code` — Excessive Comments smell should be resolved
