# Magic Number — Replace with Named Constants

## What

Using literal values (numbers, strings) directly in code without a named constant to explain their meaning. Common examples: retry limits, timeouts, buffer sizes, status codes.

## Why It's Harmful

- **Unreadable**: A bare `3` or `5000` has no semantic meaning — the reader must infer intent.
- **Hard to change**: Updating a magic number requires finding every occurrence across the codebase.
- **AI-unfriendly**: LLMs propagate magic numbers verbatim from context, making generated code harder to maintain.

## Step-by-Step Refactoring

### 1. Identify All Magic Numbers

Look for literal values in function bodies — any `int`, `float64`, `time.Duration`, or `string` that isn't `0`, `1`, or `""`:

```go
// BEFORE: Magic numbers everywhere
func fetch(url string) ([]byte, error) {
    retries := 3
    timeout := 5000
    for i := 0; i < retries; i++ {
        resp, err := http.Get(url)
        if err != nil {
            if i < retries-1 {
                time.Sleep(time.Duration(i*1000) * time.Millisecond)
            }
            continue
        }
        return resp, nil
    }
    return nil, errors.New("failed after retries")
}
```

### 2. Extract into Named Constants

```go
// AFTER: Named constants
const (
    maxRetries    = 3
    baseTimeout   = 5 * time.Second
    backoffBaseMs = 1000
)

func fetch(url string) ([]byte, error) {
    client := &http.Client{Timeout: baseTimeout}
    for i := range maxRetries {
        resp, err := client.Get(url)
        if err == nil {
            return io.ReadAll(resp.Body)
        }
        if i < maxRetries-1 {
            time.Sleep(time.Duration((i+1)*backoffBaseMs) * time.Millisecond)
        }
    }
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### 3. Group Related Constants

Use `const` blocks with a common prefix or `iota` for related values:

```go
const (
    maxRetries   = 3
    maxBatchSize = 100
    maxWorkers   = 10
)
```

## Verification

- [ ] No unexplained literal values in function bodies
- [ ] Constants are grouped in `const` blocks with clear names
- [ ] Re-run `analyze_code` — Magic Number smell should be resolved
