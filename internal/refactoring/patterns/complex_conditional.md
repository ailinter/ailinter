# Complex Conditional — Decompose Conditional

## What
A single `if`/`while` condition with many `&&` and `||` operators that makes the intent hard to understand.

## Why It's Harmful
- **Hard to parse**: Multiple boolean operators require mental "stack" to evaluate.
- **Poor readability**: The business rule is buried in boolean logic.
- **Error-prone**: Easy to misplace parentheses or invert logic.

## Step-by-Step Refactoring

### 1. Extract Each Part to a Boolean Variable
```go
// BEFORE
if user.Age > 18 && user.HasVerifiedEmail && !user.IsBanned && (user.Subscription == "premium" || user.PurchaseTotal > 1000) {
    applyDiscount(user)
}

// AFTER: Decomposed
isAdult := user.Age > 18
isVerified := user.HasVerifiedEmail
isActive := !user.IsBanned
isEligible := user.Subscription == "premium" || user.PurchaseTotal > 1000

if isAdult && isVerified && isActive && isEligible {
    applyDiscount(user)
}
```

### 2. Extract Complex Conditions to Methods
```go
// BETTER: Self-documenting method
if user.IsEligibleForDiscount() {
    applyDiscount(user)
}
```

### 3. Use Guard Clauses to Flatten
```go
if !user.IsAdult() {
    return
}
if !user.IsVerified() {
    return
}
if user.IsBanned {
    return
}
applyDiscount(user)
```

## Verification
- [ ] Each `if` condition has <= 2 boolean operators
- [ ] Condition reads like natural language
- [ ] Complex rules extracted to named methods
- [ ] Re-run `analyze_code_health` — Complex Conditional smell should be resolved
