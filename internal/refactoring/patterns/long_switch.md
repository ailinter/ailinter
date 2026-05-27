# Long Switch — Replace with Map Lookup or Strategy Pattern

## What
A switch/case or match/when block with many branches (10+ cases). Common in language parsers, HTTP handlers, and command dispatchers.

## Why It's Harmful
- **LLM error risk**: AI models are ~40% more likely to introduce bugs when modifying long switch chains (missing `break`, duplicating cases, wrong fallthrough).
- **Edit locality**: Adding/removing a case requires touching the switch itself — can't be done independently.
- **Context window pressure**: A 50-case switch consumes ~500 tokens in structure alone. LLMs lose track of earlier cases.
- **Cognitive load**: Each case may operate on different types or have subtly different behavior.

## Step-by-Step Refactoring

### 1. Replace with Map/Dictionary (for uniform cases)
```go
// BEFORE: Long switch
func getHandler(name string) Handler {
    switch name {
    case "login":
        return &LoginHandler{}
    case "logout":
        return &LogoutHandler{}
    case "register":
        return &RegisterHandler{}
    // ... 20 more cases
    }
    return nil
}

// AFTER: Map lookup
var handlers = map[string]func() Handler{
    "login":    func() Handler { return &LoginHandler{} },
    "logout":   func() Handler { return &LogoutHandler{} },
    "register": func() Handler { return &RegisterHandler{} },
}

func getHandler(name string) Handler {
    if h, ok := handlers[name]; ok {
        return h()
    }
    return nil
}
```

### 2. Strategy/Polymorphism (for type-dependent behavior)
```go
// BEFORE: Switch on type
func process(item interface{}) error {
    switch v := item.(type) {
    case *Order:
        return processOrder(v)
    case *Invoice:
        return processInvoice(v)
    default:
        return fmt.Errorf("unknown type")
    }
}

// AFTER: Interface + polymorphism
type Processable interface {
    Process() error
}

func process(item Processable) error {
    return item.Process()
}
```

### 3. Break into Smaller Functions
If the switch is unavoidable, extract each case body into a named function:
```go
func dispatch(cmd Command) error {
    switch cmd.Type {
    case "create":
        return handleCreate(cmd)
    case "update":
        return handleUpdate(cmd)
    case "delete":
        return handleDelete(cmd)
    }
    return nil
}
```

## Verification
- [ ] Switch has <= 10 cases
- [ ] Cases are uniform (same structure) — consider map lookup
- [ ] Cases differ by type — consider polymorphism
- [ ] Re-run `analyze_code_health` — score should improve
