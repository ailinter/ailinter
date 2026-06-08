# Shotgun Surgery — Consolidate Scattered Changes

## What

A single change (e.g., adding a field to a struct) requires editing many files scattered across the codebase. The opposite of Divergent Change — one logical change touches many places.

## Why It's Harmful

- **High regression risk**: Editing 10 files for one feature means 10 opportunities for bugs.
- **Knowledge fragmentation**: The logic for one concept is spread across the codebase — no single place to look.
- **AI failure mode**: LLMs generate inconsistent changes across files, missing some locations.

## Step-by-Step Refactoring

### 1. Find the Scattered Code

Trace every place a change must be made and identify common patterns:

```go
// BEFORE: Validation logic scattered across 5 handlers
// order_handler.go
func validateOrderInput(input CreateOrderInput) error {
    if input.Amount <= 0 { return errors.New("amount must be positive") }
    return nil
}

// user_handler.go
func validateUserInput(input CreateUserInput) error {
    if input.Age <= 0 { return errors.New("age must be positive") }
    return nil
}

// payment_handler.go
func validatePaymentInput(input PaymentInput) error {
    if input.Amount <= 0 { return errors.New("amount must be positive") }
    return nil
}
```

### 2. Consolidate into One Place

```go
// AFTER: All validation in one package
package validation

func Positive(name string, n int) error {
    if n <= 0 {
        return fmt.Errorf("%s must be positive", name)
    }
    return nil
}

func NonEmpty(name, value string) error {
    if strings.TrimSpace(value) == "" {
        return fmt.Errorf("%s must not be empty", name)
    }
    return nil
}

// order_handler.go
func validateOrderInput(input CreateOrderInput) error {
    if err := validation.Positive("amount", input.Amount); err != nil {
        return err
    }
    return nil
}

// payment_handler.go
func validatePaymentInput(input PaymentInput) error {
    if err := validation.Positive("amount", input.Amount); err != nil {
        return err
    }
    return nil
}
```

### 3. Use a Struct for Grouped State

When the same data fields appear across files, consolidate into a single struct:

```go
type Paging struct {
    Offset int
    Limit  int
}

func (p Paging) Validate() error {
    if p.Offset < 0 { return errors.New("offset must be >= 0") }
    if p.Limit <= 0 || p.Limit > 100 { return errors.New("limit must be 1-100") }
    return nil
}
```

## Verification

- [ ] Each logical change touches ≤ 3 files
- [ ] Cross-cutting concerns have a single home package
- [ ] Repeated validation/transformation patterns are unified
- [ ] Re-run `analyze_code` — Shotgun Surgery smell should decrease
