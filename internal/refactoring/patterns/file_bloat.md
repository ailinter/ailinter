# File Bloat — Extract Modules by Responsibility

## What
A source file that has grown too large (600+ lines for Python/JS, 1000+ for Go/Java). Combines too many distinct responsibilities in a single file.

## Why It's Harmful
- **Token overflow**: AI models have fixed context windows. A 2000-line file consumes ~6000 tokens just to read, leaving little room for reasoning.
- **Merge conflicts**: Large files have more concurrent edits. Each line touched by multiple developers increases conflict probability.
- **Poor discoverability**: Functions buried deep in a large file are hard to find and easy to duplicate accidentally.

## Step-by-Step Refactoring

### 1. Identify Cohesive Groups
Scan for functions that share common data or domain concepts:

```go
// BEFORE: 1200-line single file
// user_service.go — contains auth, profiles, notifications, billing
func (s *Service) Authenticate(...) { ... }     // auth concern
func (s *Service) GetProfile(...) { ... }        // profile concern
func (s *Service) SendWelcomeEmail(...) { ... }  // notification concern
func (s *Service) ChargeCustomer(...) { ... }    // billing concern
func (s *Service) UpdateAvatar(...) { ... }      // profile concern
func (s *Service) GenerateInvoice(...) { ... }   // billing concern
```

### 2. Extract into Separate Files by Concern
```go
// AFTER: auth.go (~80 lines)
type AuthService struct { db *sql.DB }
func (s *AuthService) Authenticate(...) { ... }

// AFTER: profile.go (~120 lines)
type ProfileService struct { db *sql.DB }
func (s *ProfileService) GetProfile(...) { ... }
func (s *ProfileService) UpdateAvatar(...) { ... }

// AFTER: billing.go (~150 lines)
type BillingService struct { gateway PaymentGateway }
func (s *BillingService) ChargeCustomer(...) { ... }
func (s *BillingService) GenerateInvoice(...) { ... }

// AFTER: notification.go (~60 lines)
type NotificationService struct { email EmailClient }
func (s *NotificationService) SendWelcomeEmail(...) { ... }
```

### 3. Wire Dependencies at the Top Level
```go
type App struct {
    Auth         *AuthService
    Profile      *ProfileService
    Billing      *BillingService
    Notification *NotificationService
}
```

## Verification
- [ ] No file exceeds the language-specific LOC threshold
- [ ] Each file has a single, clear responsibility
- [ ] File name matches its primary domain concept
- [ ] Re-run `analyze_code` — File Bloat smell should be resolved
