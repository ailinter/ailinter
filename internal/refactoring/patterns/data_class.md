# Data Class — Move Behavior to Data (Tell Don't Ask)

## What

A struct that only holds data with getters/setters, while the logic that operates on that data lives in a separate service. Violates the "Tell Don't Ask" principle — code asks the data for its values, then decides what to do.

## Why It's Harmful

- **Anemic domain model**: Logic scattered across services duplicates validation and business rules.
- **Coupling**: Every service that manipulates the data must know its internal representation.
- **AI-unfriendly**: LLMs generate inconsistent business logic when the data and its behavior are separated.

## Step-by-Step Refactoring

### 1. Identify Data Clumps

Look for structs that only hold data, while another type acts on it:

```go
// BEFORE: Data class anemic struct + external service
type User struct {
    Email string
    Role  string
}

type UserService struct {
    db *sql.DB
}

func (s *UserService) UpdateEmail(user *User, email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    if user.Role == "admin" && !strings.HasSuffix(email, "@company.com") {
        return errors.New("admin must use company email")
    }
    user.Email = email
    return s.db.SaveUser(user)
}
```

### 2. Move Behavior to the Data Type

```go
// AFTER: Tell, don't ask — User knows its own rules
type User struct {
    Email string
    Role  string
    db    *sql.DB
}

func (u *User) UpdateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    if u.Role == "admin" && !strings.HasSuffix(email, "@company.com") {
        return errors.New("admin must use company email")
    }
    u.Email = email
    return u.db.SaveUser(u)
}
```

### 3. Validate on Construction

Push validation into the constructor so invalid data can't be created:

```go
func NewUser(email, role string, db *sql.DB) (*User, error) {
    if !strings.Contains(email, "@") {
        return nil, errors.New("invalid email")
    }
    return &User{Email: email, Role: role, db: db}, nil
}
```

## Verification

- [ ] Structs have behavior methods, not just getters/setters
- [ ] Business rules live with the data they constrain
- [ ] External services orchestrate, not micromanage
- [ ] Re-run `analyze_code` — Data Class smell should decrease
