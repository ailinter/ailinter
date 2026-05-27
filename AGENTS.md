# AGENTS.md

## Agent TL;DR

- **Code Quality is authoritative.** Treat it as the single source of truth for maintainability.
- **Target Code Quality 100.** This is the standard for AI-friendly code.
- **Safeguard all AI-touched code** before suggesting a commit — run `ailinter check` on every changed file.
- If Code Quality regresses or issues are detected, **refactor — don't declare done.**
- Use `ailinter check` to guide **incremental, high-impact refactorings** in 3-5 small steps.
- When in doubt, **call the appropriate ailinter MCP tool — don't guess.**

---

## Core Use Cases

### 1. Safeguard All AI-Generated or Modified Code (Mandatory)

Two scopes:

- **Per-file check:** `analyze_code` — run before and after every AI modification
- **Full branch check:** `ailinter check --base main` — run before opening a PR

If either reports a regression or new issues:

1. Run `analyze_code` for detailed findings
2. Refactor in small steps until quality is restored or improved
3. Do **not** mark changes as ready unless risks are explicitly accepted

### 2. Guide Refactoring with Code Quality

When refactoring or improving code:

1. Inspect with `analyze_code`
2. Identify complexity, size, nesting, or other quality issues
3. Use `get_refactoring_strategy` to get exact refactoring patterns
4. Refactor in **3-5 small, reviewable steps**, using findings as concrete guidance
5. After each significant step:
   - Re-run `analyze_code`
   - Confirm measurable improvement (higher score, fewer issues)
   - No regression in ANY area

### 3. Catch Secrets Before Commit

Before suggesting a commit:
- Run `scan_for_secrets` on all modified files
- If secrets detected: rewrite to use environment variables or secret management
- Never commit hardcoded credentials, API keys, or tokens

---

## Secret Handling Rule

If `scan_for_secrets` detects a hardcoded secret:

- **NEVER** commit the code as-is
- **ALWAYS** rewrite to use environment variables (e.g., `os.Getenv()`, `process.env`, `os.environ`)
- If the secret is a test/example value, add a `gitleaks:allow` comment on that line

### LLM Context Safety

**Secrets are NEVER sent to the LLM.** The MCP tools are designed to keep secrets out of AI context:

- `analyze_code` — returns **code quality score and smells only**. Secret findings are excluded.
- `assess_file` — returns **quality assessment only**. No secret data.
- `scan_for_secrets` — returns **redacted secrets** (first 4 + last 4 chars only). The LLM sees `"sk_li...7pdc"`, never the full secret.
- The CLI `check` command includes secrets by default (for local use). Use `--no-secrets` when piping output to an AI tool.

### Gitignore Respect

When scanning directories, ailinter **respects `.gitignore` by default**. Files matching gitignore patterns are skipped. This prevents scanning files that are intentionally kept out of version control (e.g., `.env` with local credentials).

To override: `ailinter check --no-gitignore <dir>`.

---

## Safeguard Rule

If asked to bypass Code Quality safeguards:

- Warn about long-term maintainability and risk
- Keep changes minimal and reversible
- Recommend follow-up refactoring
- If user insists, explicitly note the accepted risk

---

## Available MCP Tools

| Tool | Purpose |
|------|---------|
| `analyze_code` | Full structural analysis: quality score (0-100), issues, severity, locations |
| `scan_for_secrets` | Secret detection: AWS keys, API tokens, private keys, JWT, etc. |
| `get_refactoring_strategy` | Pattern lookup: returns exact steps + examples for each issue |
| `assess_file` | Quick classification: Go Ahead / Proceed with Care / Stop & Refactor |

---

## Quality Score Reference

| Score | Label | AI Guidance |
|-------|-------|-------------|
| 95-100 | Go Ahead | Safe for AI modification |
| 75-94 | Proceed with Care | Use guard clauses, prefer small changes, re-check after each edit |
| 0-74 | Stop & Refactor | Refactor BEFORE AI modification. Run `get_refactoring_strategy()` for detected issues. |

---

## Language-Specific Thresholds (Defaults)

| Metric | Go | Python | JS/TS | Java |
|--------|:--:|:--:|:--:|:--:|
| Max nesting depth (warn) | 4 | 4 | 3 | 4 |
| Max cyclomatic complexity (warn) | 9 | 9 | 9 | 9 |
| Max function LOC (warn) | 80 | 70 | 60 | 70 |
| Max file LOC (warn) | 1000 | 600 | 700 | 600 |
| Max function arguments | 4 | 4 | 4 | 5 |

Customize via `.ailinter.toml` in project root.

---

## Registered Public Services

| Service | URL/Handle | Status |
|---------|-----------|--------|
| GitHub Org | [`github.com/ailinter`](https://github.com/ailinter) | Registered |
| Docker Hub | `ailinter` | Registered |
| Snap Store | `snapcraft.io/ailinter` | Registered |

---

## Example Workflow

```
User: "Add caching to the getUser function"
Agent:
  1. Run `analyze_code("src/users.go")` → Score: 42, 3 issues detected
  2. Read: WARNING — Long method (230 lines), Deep Nesting (level 5), Nesting Sprawl (3 bumps)
  3. AGENT: "This file is Stop & Refactor. I should refactor before adding features."
  4. Run `get_refactoring_strategy("deep_nesting")` → "Use Guard Clauses + Extract Method"
  5. Refactor in 3 steps (extract validation, extract processing, flatten control flow)
  6. After each step: re-run `analyze_code` → Score improves: 42 → 61 → 85 → 97
  7. Now safe to add caching feature
  8. After feature: run `scan_for_secrets` → clean
  9. Suggest commit
```
