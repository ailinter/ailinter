# Bumpy Road — Extract Method Per Bump

## What
A function with multiple distinct blocks of deeply nested logic, each acting like a "speed bump" that forces the reader to slow down. Identified by indentation profiling — regions of depth >= 2 separated by shallow areas.

## Why It's Harmful
- **Missing abstractions**: Each bump represents a logical chunk that should be its own function.
- **Working memory tax**: Switching between different nested contexts requires mental "gear changes."
- **Higher refactoring cost**: More bumps = more missing abstractions = more work to fix.

## Step-by-Step Refactoring

### 1. Identify the Bumps
Run `analyze_code_health` to get the exact line ranges of each bump.

### 2. Extract Each Bump into Its Own Method
```go
// BEFORE: Bumpy Road (3 bumps)
func processReport(data []Record) *Report {
    r := &Report{}
    
    // Bump 1: Validation (depth 3)
    for _, d := range data {
        if d.IsActive {
            if d.HasRequiredFields() {
                // validation logic
            }
        }
    }
    
    // Bump 2: Aggregation (depth 3)
    for _, d := range data {
        if d.Type == "sale" {
            if d.Amount > 0 {
                // aggregation logic
            }
        }
    }
    
    // Bump 3: Formatting (depth 2)
    if r.Total > 1000 {
        if r.Currency == "USD" {
            // formatting logic
        }
    }
    return r
}

// AFTER: Flat orchestration
func processReport(data []Record) *Report {
    r := &Report{}
    validateRecords(data, r)
    aggregateSales(data, r)
    formatOutput(r)
    return r
}
```

### 3. Name Each Extracted Function
Give each function a name that describes what the bump was doing.

### 4. Verify Each Extraction
Run `analyze_code_health` after each extraction. The score should improve incrementally.

## Verification
- [ ] Zero bumps remaining in the main function
- [ ] Each extracted function has a clear, single responsibility
- [ ] The main function reads as high-level orchestration
- [ ] Re-run `analyze_code_health` — Bumpy Road smell should be gone
