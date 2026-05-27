# Deep Nesting — Replace Nested Conditional with Guard Clauses

## What
Code has many conditionals and loops nested inside one another (3+ levels deep). Also known as the Arrow Anti-Pattern due to the visual shape of deep indentation.

## Why It's Harmful
- **Cognitive overload**: Working memory limited to ~3-4 items. Each nesting level requires holding additional context.
- **High defect correlation**: CodeScene research shows nested complexity is a stronger predictor of defects than cyclomatic complexity.
- **AI failure risk**: AI models are 60% more likely to introduce defects in deeply nested code.

## Step-by-Step Refactoring

### 1. Invert Conditions (Guard Clauses)
```go
// BEFORE: Deep nesting
func process(order *Order) error {
    if order != nil {
        if order.IsValid() {
            if order.HasItems() {
                // business logic
                return nil
            }
            return errors.New("no items")
        }
        return errors.New("invalid order")
    }
    return errors.New("nil order")
}

// AFTER: Guard clauses (flat)
func process(order *Order) error {
    if order == nil {
        return errors.New("nil order")
    }
    if !order.IsValid() {
        return errors.New("invalid order")
    }
    if !order.HasItems() {
        return errors.New("no items")
    }
    // business logic
    return nil
}
```

### 2. Extract Inner Blocks to Methods
If an inner block has meaningful logic, extract it:
```go
func process(order *Order) error {
    if order == nil {
        return errors.New("nil order")
    }
    return processValidOrder(order)
}
```

### 3. Use Early Returns
Each condition becomes an early return. The "happy path" becomes the main flow.

## Verification
- [ ] Maximum nesting depth <= 2
- [ ] No nested `if` inside another `if`
- [ ] Main flow reads top-to-bottom without mental stack
- [ ] Re-run `analyze_code_health` — score should improve
