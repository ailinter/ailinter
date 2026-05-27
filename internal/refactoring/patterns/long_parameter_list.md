# Long Parameter List — Introduce Parameter Object

## What
A function that takes too many arguments (4+). Hard to understand, hard to call correctly, and hard to extend.

## Why It's Harmful
- **Cognitive load**: The caller must remember the order and meaning of many parameters.
- **Brittle**: Adding a new parameter requires changing every call site.
- **Primitive obsession**: Often, groups of parameters belong together as a concept.

## Step-by-Step Refactoring

### 1. Identify Parameter Groups
Look for parameters that always appear together or represent a single concept.

```go
// BEFORE: Long parameter list
func CreateUser(name string, email string, street string, city string, zip string, country string, plan string, billingEmail string) error { ... }

// Parameters that form groups:
// - name, email → UserIdentity
// - street, city, zip, country → Address
// - plan, billingEmail → Subscription
```

### 2. Create Parameter Objects
```go
type UserIdentity struct {
    Name  string
    Email string
}

type Address struct {
    Street  string
    City    string
    Zip     string
    Country string
}

type Subscription struct {
    Plan         string
    BillingEmail string
}

// AFTER: Clean signature
func CreateUser(identity UserIdentity, address Address, sub Subscription) error { ... }
```

### 3. Use Options Pattern (for Go)
```go
type CreateUserOptions struct {
    Name         string
    Email        string
    Address      Address
    Subscription Subscription
}

func CreateUser(opts CreateUserOptions) error { ... }
```

## Verification
- [ ] Function has <= 4 parameters
- [ ] Parameter groups extracted into meaningful objects
- [ ] Call sites are readable and clear
- [ ] Re-run `analyze_code_health` — Long Parameter List smell should be resolved
