# Duplicated Code — Extract Method / Pull Up

## What
The same or very similar code appears in multiple places. Known as the DRY (Don't Repeat Yourself) violation.

## Why It's Harmful
- **Change amplification**: Fixing a bug requires finding and updating every copy.
- **Divergence risk**: Copies drift apart over time as they're modified independently.
- **AI blind spot**: AI models often copy-paste without recognizing existing abstractions.

## Step-by-Step Refactoring

### 1. Identify Duplicates
Look for blocks of code that are identical or very similar (75%+ similarity).

### 2. Extract the Common Logic
```python
# BEFORE: Duplicated validation
def create_user(data):
    if not data.get("name"):
        raise ValueError("name required")
    if not data.get("email"):
        raise ValueError("email required")
    if "@" not in data["email"]:
        raise ValueError("invalid email")
    # ... create user

def update_user(data):
    if not data.get("name"):
        raise ValueError("name required")
    if not data.get("email"):
        raise ValueError("email required")
    if "@" not in data["email"]:
        raise ValueError("invalid email")
    # ... update user

# AFTER: Extracted validator
def validate_user_data(data):
    if not data.get("name"):
        raise ValueError("name required")
    if not data.get("email"):
        raise ValueError("email required")
    if "@" not in data["email"]:
        raise ValueError("invalid email")

def create_user(data):
    validate_user_data(data)
    # ... create user

def update_user(data):
    validate_user_data(data)
    # ... update user
```

### 3. Handle Variations with Parameters
If duplicates differ slightly, parameterize the extracted function:
```python
def validate_field(data, field, validator=None):
    if field not in data:
        raise ValueError(f"{field} required")
    if validator and not validator(data[field]):
        raise ValueError(f"invalid {field}")
```

### 4. For Class Hierarchies, Use Pull Up Method
If duplicates exist across subclasses, pull the common method into the parent class.

## Verification
- [ ] Each business rule exists in exactly one place
- [ ] Changes to a rule require editing only one function
- [ ] Extracted functions are well-named and independently testable
- [ ] Re-run `analyze_code_health` — DRY violations should be resolved
