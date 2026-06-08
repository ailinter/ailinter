# Refused Bequest — Replace Inheritance with Delegation

## What

A type inherits from a base type but only uses a small fraction of its fields or methods. The type "refuses the bequest" of the parent. In Go, this appears as embedded structs where most inherited methods are irrelevant.

## Why It's Harmful

- **Misleading API**: The type exposes methods that don't make sense for its domain.
- **Breakage risk**: Changes to the parent's interface ripple to all children, even those that don't use it.
- **Liskov violation**: The child can't genuinely substitute for the parent.

## Step-by-Step Refactoring

### 1. Identify Refused Bequest

Look for embedded types where only one field or method is actually used:

```go
// BEFORE: ReportService inherits everything from BaseService
type BaseService struct {
    db      *sql.DB
    cache   *redis.Client
    logger  *log.Logger
    metrics *metrics.Client
}

func (b *BaseService) DB() *sql.DB { return b.db }
func (b *BaseService) Cache() *redis.Client { return b.cache }
func (b *BaseService) Log() *log.Logger { return b.logger }

// ReportService only needs db — refuses cache, logger, metrics
type ReportService struct {
    *BaseService
}

func (s *ReportService) GenerateReport() (*Report, error) {
    rows, err := s.DB().Query("SELECT ...")
    // never uses s.Cache(), s.Log(), s.Metrics()
}
```

### 2. Replace with Delegation

```go
// AFTER: Delegation — only what you need
type ReportService struct {
    db *sql.DB
}

func NewReportService(db *sql.DB) *ReportService {
    return &ReportService{db: db}
}

func (s *ReportService) GenerateReport() (*Report, error) {
    rows, err := s.db.Query("SELECT ...")
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()
    // ... build report
}
```

### 3. Extract Interface When Needed

If multiple services share a subset of behavior, extract just that subset:

```go
type Querier interface {
    Query(query string, args ...any) (*sql.Rows, error)
}
```

## Verification

- [ ] No embedded types where most fields/methods are unused
- [ ] Each type has only the dependencies it actually uses
- [ ] Interfaces are narrow and client-defined
- [ ] Re-run `analyze_code` — Refused Bequest smell should be resolved
