# Brain Method — Extract Method Decomposition

## What
A single function that is too long (typically >50-100 lines) and tries to do too many things. Also known as Long Function or God Function.

## Why It's Harmful
- **Hard to understand**: The reader must hold the entire function's logic in working memory.
- **Hard to test**: Testing the whole function requires setting up all its dependencies.
- **AI amplification**: AI generates worse code when working within large, complex functions.

## Step-by-Step Refactoring

### 1. Identify Logical Sections
Read through the function and mark where the "topic" changes:
- Input validation
- Data transformation
- Business logic
- Side effects (DB writes, API calls)
- Error handling

### 2. Extract Each Section
```python
# BEFORE: Brain Method (100+ lines)
def handle_order(order_data):
    # validate (20 lines)
    # transform (30 lines)
    # apply discounts (20 lines)
    # save to database (20 lines)
    # send confirmation (10 lines)
    pass

# AFTER: Composed of small functions
def handle_order(order_data):
    order = validate_order(order_data)
    order = transform_order(order)
    order = apply_discounts(order)
    save_order(order)
    send_confirmation(order)
```

### 3. Name Each Function Well
A good function name communicates WHAT it does without needing comments. If you struggle to name it, the boundary might be wrong.

### 4. Keep Functions Small
Target: each function does ONE thing. Most extracted functions should be 5-15 lines.

## Verification
- [ ] No function exceeds the threshold (e.g., 70 lines for Python, 80 for Go)
- [ ] The main function reads as a high-level outline
- [ ] Each extracted function is independently testable
- [ ] Re-run `analyze_code_health` — score should improve
