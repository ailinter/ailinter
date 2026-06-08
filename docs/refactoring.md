# Refactoring Strategies

AILINTER provides step-by-step refactoring guidance for **24 code smells** — from pure smells (deep nesting, brain method) to data-oriented smells (primitive obsession, data class) to inheritance smells (refused bequest, parallel inheritance).

Call `get_refactoring_strategy("smell_name")` via MCP to get actionable instructions with before/after examples and verification steps.

---

## Complete Smell Coverage

### Pure Smells

| Smell | Strategy | Description |
|-------|----------|-------------|
| Deep Nesting | Guard Clauses + Extract Method | Flatten deeply nested conditionals |
| Brain Method | Extract Method + SRP | Break monolithic methods into focused ones |
| Bumpy Road | Extract Method + Flatten | Smooth readability by extracting mid-method blocks |
| Complex Conditional | Guard Clauses + Strategy Pattern | Simplify compound conditionals |
| God Class | Extract Class + SRP | Split classes that do too much |
| Long Parameter List | Parameter Object Pattern | Group related parameters |
| Primitive Obsession | Type Wrapper + Value Object | Encapsulate primitives in domain types |
| Duplicated Code | Extract Method + Template Method | Unify repeated logic |
| Long Method | Extract Method | Decompose methods that exceed length thresholds |
| Complex Method | Extract + Simplify | Reduce cyclomatic complexity |

### Data-Oriented Smells

| Smell | Strategy | Description |
|-------|----------|-------------|
| Data Class | Move Logic In | Move behavior into data containers |
| Low Cohesion | Extract Class | Split classes with unrelated responsibilities |
| Message Chains | Hide Delegate | Encapsulate navigation chains |
| Lazy Element | Inline Element | Remove unnecessary abstraction layers |
| Global Data | Encapsulate | Wrap global state in access-controlled modules |
| Long Scope Variable | Reduce Scope | Narrow variable lifetimes |
| Magic Number | Named Constant | Replace literals with named constants |
| Excessive Comments | Self-Documenting | Make code read well without comments |

### Inheritance & Structural Smells

| Smell | Strategy | Description |
|-------|----------|-------------|
| Shotgun Surgery | Move + Combine | Consolidate scattered changes from one trigger |
| Refused Bequest | Replace Delegation | Replace inheritance with composition |
| Parallel Inheritance | Strategy Pattern | Eliminate parallel class hierarchies |
| Long Switch | Replace with Map | Replace switch statements with lookup tables |
| Paragraph of Code | Extract Method | Extract logical blocks into named methods |
| File Bloat | Extract Module | Split oversized files into focused modules |

---

## Usage

### Via MCP (from AI Assistant)

```javascript
// Your AI assistant calls this automatically when quality issues are found
get_refactoring_strategy("deep_nesting")
// Returns: step-by-step instructions with before/after examples
```

### Via CLI

```bash
# List all available strategies
ailinter rules list

# Get strategy for a specific smell (embedded in check output)
ailinter check file.go --format markdown
```

### Via the Refactoring Loop

The recommended workflow:

```
1. analyze_code(file) → score
2. If score < 80 or smells detected:
   a. get_refactoring_strategy("smell_name") → exact instructions
   b. Refactor in 3–5 small steps
   c. Re-run analyze_code after each step
   d. Repeat until score ≥ 80
3. Make your feature change
4. analyze_code(file) → confirm no regression
5. scan_for_secrets(content) → clean
6. Commit
```

Each strategy includes:
- **Before/after code examples** — exact patterns to transform
- **Step-by-step instructions** — 3-5 incremental refactoring steps
- **Verification steps** — how to confirm the refactoring is correct
- **Edge case handling** — what to watch out for during the refactoring
