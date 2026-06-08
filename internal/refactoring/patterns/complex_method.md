# Complex Method — Reduce Cyclomatic Complexity

## What

A function with high cyclomatic complexity (CC) — too many decision paths through if/else, switch, for loops, and boolean operators. **Go warns at CC ≥ 9, alerts at CC ≥ 20.**

## Why It's Harmful

- **Untestable**: Exponential path combinations make full coverage impossible.
- **Cognitive overload**: More than ~10 decisions exceeds working memory capacity.
- **AI-unfriendly**: LLMs lose track of preconditions when functions have >10 branches.

## Which Approach Should I Use?

| Your Function Has... | Use This Pattern |
|---|---|
| Explicit state transitions (pending→approved→shipped) | **State Machine** |
| Same variable tested against many values (switch on ext/type) | **Table Lookup** |
| Mixed validation + business logic + side effects | **Extract by Responsibility** |

---

## Pattern 1: State Machine

Replace conditional state transitions with an explicit state type and transition table:

```go
// BEFORE: if-else state ramps — CC = 5
func (o *Order) Process(action string) error {
    if o.Status == "pending" && action == "approve" {
        o.Status = "approved"
    } else if o.Status == "approved" && action == "ship" {
        o.Status = "shipped"
    } else if o.Status == "shipped" && action == "deliver" {
        o.Status = "delivered"
    } else {
        return fmt.Errorf("invalid transition: %s → %s", o.Status, action)
    }
    return nil
}

// AFTER: state machine — CC = 1
type OrderState int

const (
    StatePending  OrderState = iota
    StateApproved
    StateShipped
    StateDelivered
)

type transition struct {
    from  OrderState
    event string
}

var table = map[transition]OrderState{
    {StatePending, "approve"}:  StateApproved,
    {StateApproved, "ship"}:    StateShipped,
    {StateShipped, "deliver"}:  StateDelivered,
}

func (o *Order) Process(event string) error {
    next, ok := table[transition{o.Status, event}]
    if !ok {
        return fmt.Errorf("invalid: %v on %s", o.Status, event)
    }
    o.Status = next
    return nil
}
```

## Pattern 2: Table Lookup

Replace switch-on-value with a lookup map:

```go
// BEFORE: switch on ext — CC = 7
func processFile(path string) error {
    switch strings.ToLower(filepath.Ext(path)) {
    case ".go":
        return processGo(path)
    case ".py":
        return processPython(path)
    case ".js", ".ts":
        return processJS(path)
    case ".md":
        return processMarkdown(path)
    default:
        return fmt.Errorf("unsupported: %s", path)
    }
}

// AFTER: map lookup — CC = 3
var analyzers = map[string]func(string) error{
    ".go": processGo,
    ".py": processPython,
    ".js": processJS,
    ".ts": processJS,
    ".md": processMarkdown,
}

func processFile(path string) error {
    fn, ok := analyzers[strings.ToLower(filepath.Ext(path))]
    if !ok {
        return fmt.Errorf("unsupported: %s", path)
    }
    return fn(path)
}
```

## Pattern 3: Extract by Responsibility

Separate validation, discount logic, and persistence into their own functions:

```go
// BEFORE: processOrder CC=16 — validation, discounts, tax, DB all mixed
func processOrder(order *Order, cfg *Config, db *DB) error {
    if order == nil { return errors.New("nil order") }
    if !order.IsValid() { return errors.New("invalid") }
    // ... 50 lines of discount, tax, save logic ...

// AFTER: Composed from small focused functions — CC = 4
func processOrder(order *Order, cfg *Config, db *DB) error {
    if err := validateOrder(order); err != nil {
        return err
    }
    total, err := calculateTotal(order.Items, cfg)
    if err != nil {
        return err
    }
    if err := saveOrder(db, order, total); err != nil {
        return err
    }
    sendConfirmation(order.Email, total) // non-critical
    return nil
}
```

## Cyclomatic Complexity Reference

| Language | Warning (≥) | Alert (≥) |
|----------|-------------|-----------|
| Go       | 9           | 20        |
| Python   | 9           | 20        |
| TypeScript | 9          | 20        |
| JavaScript | 7          | 15        |

Configure via `.ailinter.toml`: `[cyclomatic_complexity] weight = 1.0 warning = 9 alert = 20`.

## Verification

- [ ] Each function CC ≤ 10
- [ ] State machines have explicit state types and transition tables
- [ ] Switch/value chains replaced with lookup maps
- [ ] No single function handles both validation and business logic
- [ ] All existing tests pass without modification
