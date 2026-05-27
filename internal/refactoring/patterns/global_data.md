# Global Data — Encapsulate Variable (Replace with Accessor)

## What
Mutable variables declared at the package or module scope, accessible from anywhere. Includes global singletons, module-level state maps, and unexported-but-mutable top-level variables.

## Why It's Harmful
- **Hidden coupling**: Any function that reads a global depends on every function that writes to it. Dependencies are invisible.
- **Test isolation impossible**: Tests cannot run in parallel when they share global state. Test order matters.
- **AI hallucination risk**: LLMs assume local reasoning is sufficient. With globals, the correct behavior depends on execution history the model cannot see.

## Step-by-Step Refactoring

### 1. Identify Mutable Globals
```go
// BEFORE: Mutable global data
var config *AppConfig           // mutable pointer
var activeUsers = map[string]*User{}  // mutable map
var requestCount int            // mutable counter

func HandleRequest(r *Request) {
    requestCount++              // hidden side effect
    user := activeUsers[r.UserID]
    if user == nil {
        user = loadUser(r.UserID)
        activeUsers[r.UserID] = user  // hidden mutation
    }
}
```

### 2. Encapsulate in a Struct with Methods
```go
// AFTER: Encapsulated state
type RequestContext struct {
    config       *AppConfig
    activeUsers  map[string]*User
    requestCount int
    mu           sync.RWMutex
}

func NewRequestContext(cfg *AppConfig) *RequestContext {
    return &RequestContext{
        config:      cfg,
        activeUsers: make(map[string]*User),
    }
}

func (rc *RequestContext) HandleRequest(r *Request) {
    rc.mu.Lock()
    rc.requestCount++
    rc.mu.Unlock()

    user := rc.getOrLoadUser(r.UserID)
    // ...
}

func (rc *RequestContext) getOrLoadUser(id string) *User {
    rc.mu.RLock()
    user := rc.activeUsers[id]
    rc.mu.RUnlock()
    if user == nil {
        rc.mu.Lock()
        user = loadUser(id)
        rc.activeUsers[id] = user
        rc.mu.Unlock()
    }
    return user
}
```

### 3. For Configuration — Make It Immutable
```go
// AFTER: Immutable config passed via constructor
type AppConfig struct {
    Port    int
    DBURL   string
    Timeout time.Duration
}

func NewServer(cfg AppConfig) *Server {
    return &Server{cfg: cfg}
}
// cfg never changes after construction
```

### 4. For Counters — Inject via Interface
```go
type Metrics interface {
    IncrementRequestCount()
}

// Tests use a mock Metrics; production uses PrometheusMetrics
```

## Verification
- [ ] No mutable package-level variables
- [ ] State is owned by a specific struct instance
- [ ] Configuration is immutable after construction
- [ ] Tests can be safely run in parallel (`go test -parallel`)
- [ ] Re-run `analyze_code` — Global Data smell should be resolved
