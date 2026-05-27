# Primitive Obsession — Replace with Value Object

## What
Using primitive types (string, int, float) to represent domain concepts that deserve their own types. For example, using `string` for an email address or `float64` for money.

## Why It's Harmful
- **No validation**: A `string email` can hold any string. An `Email` type validates on construction.
- **No behavior**: A `float64 price` can't format itself. A `Money` type can.
- **Confusion**: Adding inches to centimeters — same type, different meaning.

## Step-by-Step Refactoring

### 1. Identify Primitive Obsession
Look for:
- String fields with implicit format rules (email, phone, URL, ID)
- Numeric fields with units or constraints (money, percentage, age, quantity)
- Repeated validation logic scattered across the codebase

### 2. Create a Value Object
```go
// BEFORE: Primitive Obsession
type User struct {
    Email string // must be valid email
    Age   int    // must be >= 0, <= 150
}

func (u *User) Validate() error {
    if !strings.Contains(u.Email, "@") {
        return errors.New("invalid email")
    }
    if u.Age < 0 || u.Age > 150 {
        return errors.New("invalid age")
    }
    return nil
}

// AFTER: Value Objects
type Email struct {
    value string
}

func NewEmail(s string) (Email, error) {
    if !strings.Contains(s, "@") {
        return Email{}, errors.New("invalid email")
    }
    return Email{value: s}, nil
}

func (e Email) String() string { return e.value }

type Age struct {
    value int
}

func NewAge(n int) (Age, error) {
    if n < 0 || n > 150 {
        return Age{}, errors.New("invalid age")
    }
    return Age{value: n}, nil
}

type User struct {
    Email Email
    Age   Age
}
// No Validate() needed — can't construct invalid User
```

### 3. Add Domain Behavior
```go
func (e Email) Domain() string { ... }
func (e Email) IsPersonal() bool { ... }
func (m Money) Add(other Money) Money { ... }
func (m Money) Format(currency string) string { ... }
```

## Verification
- [ ] No raw primitives representing domain concepts
- [ ] Validation happens at construction time (can't create invalid objects)
- [ ] Domain behavior lives on the value object
- [ ] Re-run `analyze_code_health` — Primitive Obsession percentage should decrease
