# Too Many Functions — Split by Responsibility

## What
A file contains more functions or methods than one reader can hold in context. The detector counts named functions, including C# expression-bodied members and one-line methods.

## Why It's Harmful
- **Discovery cost**: finding the right function means scanning a long catalogue.
- **Hidden coupling**: unrelated functions share imports and helpers, so changes ripple.
- **AI context waste**: a model must read the whole file to find one function.

## Step-by-Step Refactoring

### 1. Group by Responsibility
List the functions and group them by the data they operate on or the domain they serve. A group with three or more functions is a candidate module.

### 2. Extract a File or Class per Group
Move each group to a file or class named after its domain. Keep public entry points in the original file and delegate to the new unit.

### 3. Keep Expression-Bodied Members Together
C# expression-bodied members are still methods. Group them with the class they belong to; do not split a cohesive class only to lower the count.

## Verification
- [ ] Each file has one responsibility
- [ ] The original public API is unchanged
- [ ] Tests still pass
- [ ] Re-run the scan — `function_count` should clear
