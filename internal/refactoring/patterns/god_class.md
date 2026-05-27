# God Class — Extract Class (Single Responsibility Principle)

## What
A file or class that has grown too large (>500-1000 lines), containing too many responsibilities. Also known as Brain Class or Large Class.

## Why It's Harmful
- **Context window loss**: Large files exceed token limits, reducing AI accuracy.
- **Change risk**: Modifying one responsibility can accidentally break another.
- **Coordination bottleneck**: Multiple developers need to edit the same file.

## Step-by-Step Refactoring

### 1. Identify Responsibilities
List the distinct "topics" the class handles:
- Data storage / persistence
- Business rules / validation
- Formatting / presentation
- External API communication
- Event handling

### 2. Extract One Class at a Time
Start with the most independent responsibility:

```go
// BEFORE: God Class (500+ lines)
type OrderManager struct {
    db    *sql.DB
    email *EmailService
    cache *redis.Client
}

func (m *OrderManager) CreateOrder(...) { ... }
func (m *OrderManager) ValidateOrder(...) { ... }
func (m *OrderManager) SendConfirmation(...) { ... }
func (m *OrderManager) CacheOrder(...) { ... }
func (m *OrderManager) GenerateReport(...) { ... }

// AFTER: Split by responsibility
type OrderRepository struct { db *sql.DB }
func (r *OrderRepository) Save(...) { ... }

type OrderValidator struct {}
func (v *OrderValidator) Validate(...) { ... }

type OrderNotifier struct { email *EmailService }
func (n *OrderNotifier) SendConfirmation(...) { ... }
```

### 3. Use Composition
The original class becomes a coordinator that delegates to extracted classes.

### 4. Keep Extracting
Continue until each class has a single, well-defined responsibility.

## Verification
- [ ] No file exceeds the size threshold
- [ ] Each class has one clear purpose
- [ ] Classes communicate through well-defined interfaces
- [ ] Re-run `analyze_code_health` — File Bloat smell should be resolved
