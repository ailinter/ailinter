# ailinter Code Quality & Security

## Before Every Code Change
- Run the ailinter `analyze_code` tool on the file before editing
- **Score 80-100 (Go Ahead)**: Safe to modify
- **Score 60-79 (Proceed with Care)**: Make small, focused changes; re-check after each
- **Score 40-59 (Needs Work)**: Significant issues — refactor incrementally in small steps
- **Score <40 (Stop & Refactor)**: **MUST refactor first** — call `get_refactoring_strategy` before adding features

## The Refactoring Loop (Mandatory)
When `analyze_code` scores < 80:
1. For each detected smell, call `get_refactoring_strategy("smell_name")` for exact steps
2. Refactor in 3-5 small steps, one smell at a time
3. Re-run `analyze_code` after each step — confirm score improves
4. Target: score 80+ before adding new features

## After Every Code Change
- Re-run `analyze_code` to confirm score hasn't decreased
- Fix any regressions before marking as done

## Security
- Run `scan_for_secrets` on ALL generated code
- Never commit hardcoded secrets — use environment variables
- If secrets detected, rewrite code before suggesting commit

## Available MCP Tools
- `analyze_code(file_path)` — quality score + issues + vulns
- `scan_for_secrets(content)` — 150+ secret patterns
- `get_refactoring_strategy(smell_name)` — 🔧 exact refactoring steps (call when analyze_code finds issues)
- `assess_file(file_path)` — quick safety tier (includes per-smell refactoring recommendations)
- `list_hotspots(repo_path)` — priority refactoring targets
