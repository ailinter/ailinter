# Long Scope Variable — Move Declaration Closer to Use

## What
A variable declared far from where it's actually used (>50 lines between declaration and last use). The variable "lives" across a large span of code, forcing readers to keep it in working memory.

## Why It's Harmful
- **Working memory strain**: Readers must hold the variable in mind across dozens of lines. Each intervening line is a chance to forget or confuse it.
- **Unintended reuse**: A variable still in scope may be accidentally modified or read after its intended lifetime, causing subtle bugs.
- **AI context fragmentation**: LLMs lose track of variables when declaration and use are separated by many tokens. The model may "hallucinate" a different value.

## Step-by-Step Refactoring

### 1. Identify Long-Scope Variables
```go
// BEFORE: Variable declared at line 10, last used at line 85
func ProcessReport(data []Record) (*Report, error) {
    var report *Report          // line 10 — declared far from use
    var err error
    stats := computeStats(data)

    // ... 30 lines of stat processing ...

    // ... 20 lines of formatting logic ...

    // ... 15 lines of validation ...

    report = buildReport(stats) // line 80 — first meaningful use
    if err = report.Validate(); err != nil {
        return nil, err
    }
    return report, nil
}
```

### 2. Move Declaration to First Use
```go
// AFTER: Declaration at point of use
func ProcessReport(data []Record) (*Report, error) {
    stats := computeStats(data)

    // ... 30 lines of stat processing ...

    // ... 20 lines of formatting logic ...

    // ... 15 lines of validation ...

    report := buildReport(stats)  // declared and used here
    if err := report.Validate(); err != nil {
        return nil, err
    }
    return report, nil
}
```

### 3. Break Long Functions to Reduce Scope Naturally
If moving the declaration isn't possible because the variable is used in multiple distant locations, the function itself is too long:

```go
// BEFORE: Variable spans entire function
func ProcessOrder(order *Order) error {
    var total float64          // declared here

    total += calculateItems(order.Items)     // used at line 50
    total += calculateTax(total)             // used at line 80
    total += calculateShipping(order.Address) // used at line 120

    order.Total = total
    return order.Save()
}

// AFTER: Split into pipeline — each function has its own scope
func ProcessOrder(order *Order) error {
    itemsTotal := calculateItems(order.Items)
    taxTotal := calculateTax(itemsTotal)
    shippingTotal := calculateShipping(order.Address)

    order.Total = itemsTotal + taxTotal + shippingTotal
    return order.Save()
}
```

### 4. Use Block Scoping
```go
// BEFORE: Variable used as temporary but lives too long
func validate(data []Record) error {
    var tmp string
    for _, r := range data {
        tmp = r.Format()
        if len(tmp) > 100 {
            return errors.New("too long")
        }
    }
    // tmp still in scope here, but shouldn't be used
    return nil
}

// AFTER: Limit scope with block
func validate(data []Record) error {
    for _, r := range data {
        if len(r.Format()) > 100 {
            return errors.New("too long")
        }
    }
    return nil
}
```

## Verification
- [ ] No variable has a declaration-to-last-use span > 50 lines
- [ ] Temporary variables live only within their logical block
- [ ] Long functions are split when variables force large scopes
- [ ] Re-run `analyze_code` — Long Scope Variable smell should be resolved
