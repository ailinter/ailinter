# Long Method — See Brain Method

Long Method is the same problem as Brain Method — a function that is too long (>50-100 lines) and does too many things. See the [Brain Method strategy ↗](brain_method.md) for full refactoring guidance.

**Quick summary:** Functions over 80 lines should be decomposed using Extract Method. Break into logical sections: validation, transformation, business logic, and side effects. Each extracted function should be 5-15 lines and do one thing well.

```go
// BEFORE: Long Method (80+ lines)
func handleOrder(order *Order) {
    // validate (20 lines)
    // compute (30 lines)
    // save (15 lines)
    // notify (15 lines)
}

// AFTER: Composed of small functions
func handleOrder(order *Order) {
    validateOrder(order)
    total := computeTotal(order)
    saveOrder(order, total)
    notifyCustomer(order)
}
```

For detailed steps, examples, and verification, see [Brain Method →](brain_method.md).
