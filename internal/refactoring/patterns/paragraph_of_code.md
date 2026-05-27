# Paragraph of Code — Break into Logical Sections

## What
A long stretch of consecutive non-blank lines (>20) without visual breaks. The code reads as a wall of text — no separation between conceptual steps.

## Why It's Harmful
- **Scanning failure**: Developers scan code by looking at blank-line boundaries. A 50-line block forces line-by-line reading.
- **AI chunking issues**: LLMs process code in attention windows. A monolithic block reduces the model's ability to separate independent steps.
- **Hidden bugs**: A missing `if` or mis-scoped variable is hard to spot in a wall of text. Visual grouping exposes structural errors.

## Step-by-Step Refactoring

### 1. Identify Logical Groups Within the Paragraph
```go
// BEFORE: Monolithic paragraph (28 consecutive non-blank lines)
func ProcessOrder(order *Order) error {
    if order == nil {
        return errors.New("nil order")
    }
    if order.Customer == nil {
        return errors.New("missing customer")
    }
    total := 0.0
    for _, item := range order.Items {
        if item.Quantity <= 0 {
            return fmt.Errorf("invalid quantity for %s", item.Name)
        }
        total += item.Price * float64(item.Quantity)
    }
    if total > order.Customer.CreditLimit {
        return errors.New("exceeds credit limit")
    }
    order.Total = total
    order.Status = "confirmed"
    if err := order.Save(); err != nil {
        return fmt.Errorf("save failed: %w", err)
    }
    if err := SendConfirmation(order.Customer.Email, order); err != nil {
        log.Printf("confirmation email failed: %v", err)
    }
    return nil
}
```

### 2. Add Blank Lines Between Conceptual Steps
```go
// AFTER: Logical sections with blank-line separation
func ProcessOrder(order *Order) error {
    // === Validation section ===
    if order == nil {
        return errors.New("nil order")
    }
    if order.Customer == nil {
        return errors.New("missing customer")
    }

    // === Price calculation section ===
    total := 0.0
    for _, item := range order.Items {
        if item.Quantity <= 0 {
            return fmt.Errorf("invalid quantity for %s", item.Name)
        }
        total += item.Price * float64(item.Quantity)
    }

    // === Credit check section ===
    if total > order.Customer.CreditLimit {
        return errors.New("exceeds credit limit")
    }

    // === Persist section ===
    order.Total = total
    order.Status = "confirmed"
    if err := order.Save(); err != nil {
        return fmt.Errorf("save failed: %w", err)
    }

    // === Notification section ===
    if err := SendConfirmation(order.Customer.Email, order); err != nil {
        log.Printf("confirmation email failed: %v", err)
    }
    return nil
}
```

### 3. Extract Sections into Named Functions
Once sections are identified, extract the largest ones:

```go
func ProcessOrder(order *Order) error {
    if err := validateOrder(order); err != nil {
        return err
    }

    total, err := calculateTotal(order.Items)
    if err != nil {
        return err
    }

    if err := checkCredit(order.Customer, total); err != nil {
        return err
    }

    if err := finalizeOrder(order, total); err != nil {
        return err
    }

    notifyCustomer(order.Customer, order)
    return nil
}
```

## Verification
- [ ] No consecutive block exceeds 20 non-blank lines
- [ ] Blank lines separate distinct conceptual steps
- [ ] Large sections extracted into named functions
- [ ] Re-run `analyze_code` — Paragraph of Code smell should be resolved
