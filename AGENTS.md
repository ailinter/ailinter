# AGENTS.md

## Agent TL;DR

- **Code Quality is authoritative.** Treat it as the single source of truth for maintainability.
- **Target Code Quality 100.** This is the standard for AI-friendly code.
- **Safeguard all AI-touched code** before suggesting a commit — delegate to `@ailinter` sub-agent.
- **When issues are found: the refactoring loop.** Call `get_refactoring_strategy()` → refactor in 3-5 steps → re-run `analyze_code` to confirm improvement. **Never skip this loop.**
- Use `@ailinter` sub-agent to guide **incremental, high-impact refactorings** in 3-5 small steps.
- When in doubt, **call the appropriate ailinter MCP tool — don't guess.**

---

## Sub-Agent Delegation (Ailinter Scans)

**Delegate all ailinter scanning to the `@ailinter` sub-agent.** This keeps scan output out of the main conversation context and provides consistent, parsed results.

### When to invoke `@ailinter`:

| Scenario | Command |
|----------|---------|
| Scan changed files before commit | `@ailinter scan <file1> <file2> ...` |
| Full project health check | `@ailinter scan .` |
| Per-file analysis before editing | `@ailinter analyze <file>` |
| Secret scan on new code | `@ailinter secrets <file>` |
| Vulnerability assessment | `@ailinter vulns <file>` |
| Quick safety check | `@ailinter assess <file>` |

### What `@ailinter` does:

1. Runs `ailinter check <paths> --format problems --no-secrets` — code quality + vulnerabilities + **go vet/staticcheck/gofmt/misspell/ineffassign by default**
2. Runs `ailinter check <paths> --format problems --secrets-only` — secrets only
3. Parses line-oriented results: code quality scores, vulnerability findings, secret detections
4. Returns a concise structured summary (≤ 30 lines)
4. Flags regressions from baseline scores

**The sub-agent handles all scanning.** You receive only the summary — no raw JSON dumps, no false-positive noise from embedded test fixtures, no betterleaks.toml self-matches.

---

## Core Use Cases

### 1. Safeguard All AI-Generated or Modified Code (Mandatory)

Two scopes:

- **Per-file check:** `analyze_code` — run before and after every AI modification
- **Full branch check:** `ailinter check --base main` — run before opening a PR

If either reports a regression or new issues:

1. Run `analyze_code` for detailed findings
2. Call `get_refactoring_strategy("<smell_name>")` to get exact steps for each detected issue
3. Refactor in 3-5 small steps until quality is restored or improved
4. Re-run `analyze_code` after each step to verify improvement
5. Do **not** mark changes as ready unless risks are explicitly accepted

### 2. Guide Refactoring with Code Quality

When `analyze_code` or `assess_file` reports code quality issues below 80, follow the **refactoring loop**:

1. Inspect with `analyze_code` — identify the specific smells
2. For EACH detected smell, call `get_refactoring_strategy("<smell_name>")` — returns exact before/after code examples, step-by-step instructions, and verification steps
3. Supported smells: `deep_nesting`, `brain_method`, `bumpy_road`, `complex_conditional`, `god_class`, `long_parameter_list`, `primitive_obsession`, `duplicated_code`
4. Refactor in **3-5 small, reviewable steps**, using the strategy as concrete guidance
5. After each significant step:
   - Re-run `analyze_code`
   - Confirm measurable improvement (higher score, fewer issues)
   - No regression in ANY area
6. Repeat the loop until score reaches 80+

### 3. Catch Go Vet Issues Before Commit (NEW)

**Mandatory for Go test files.** Before suggesting a commit:

1. Run `go vet ./<package>/...` on any modified Go package — catch `sync.Once` copy violations, `loopclosure`, nilness, etc.
2. All AI-generated test files MUST pass `go vet` before commit
3. `ailinter check --meta-lint` now runs `go vet` + `staticcheck` + `gofmt` + `misspell` + `ineffassign` **by default** — no need to pass `--meta-lint` explicitly

Example of what `go vet` catches (DO NOT write this pattern):
```go
// WRONG — copies sync.Once (which contains sync.noCopy):
origOnce := initOnce  // "assignment copies lock value"

// RIGHT — just reset with direct assignment:
initOnce = sync.Once{}
```

### 4. Catch Staticcheck Issues Before Commit (NEW)

**Mandatory before every commit.** `go vet` and `go test` do NOT catch unused functions.

1. Run `staticcheck ./...` on the entire repo — catches unused code (U1000), deprecated APIs (SA1019), and subtle bugs
2. Install: `go install honnef.co/go/tools/cmd/staticcheck@latest`
3. The pre-commit hook (`scripts/pre-commit.sh`) now runs staticcheck on changed packages automatically
4. All AI-generated code MUST pass staticcheck before commit

### 5. Catch Secrets Before Commit

Before suggesting a commit:
- Run `scan_for_secrets` on all modified files
- If secrets detected: rewrite to use environment variables or secret management
- Never commit hardcoded credentials, API keys, or tokens

## Template & Generated Code Safety

### HTML Template JS Escaping

Go's `html/template` package auto-escapes ALL content, including JSON intended for JavaScript
blocks. This silently breaks generated HTML.

```go
// WRONG — html/template escapes the JSON, turning [ into &amp;#91;
type data struct {
    ElementsJSON string  // {{.ElementsJSON}} → escaped string, not JS
}

// RIGHT — use template.JS to mark content as safe JavaScript
type data struct {
    ElementsJSON template.JS  // {{.ElementsJSON}} → raw JS, unescaped
}
data := data{ElementsJSON: template.JS(jsonBytes)}
```

**Always test generated HTML** by opening it in a browser before pushing. A test that
checks `strings.Contains(html, "const DATA")` is NOT sufficient — it must verify the
content is valid JavaScript, not an escaped string.

### CDN Library Versioning

Always pin exact versions for CDN-loaded libraries. Unpinned imports (e.g.,
`cytoscape-fcose/cytoscape-fcose.js` without a version tag) break silently
when the upstream API changes.

```html
<!-- WRONG — latest version may be incompatible -->
<script src="https://unpkg.com/cytoscape-fcose/cytoscape-fcose.js"></script>

<!-- RIGHT — pinned to a tested version -->
<script src="https://unpkg.com/cytoscape@3.30/dist/cytoscape.min.js"></script>
```

Prefer built-in functionality over plugins. Cytoscape's built-in `cose` layout avoids
the external `fcose` plugin entirely.

### Persistence Path Safety

Never hardcode `~/.ailinter/` or `os.UserHomeDir()` for persistence paths.
Use workspace-relative or configurable paths:

```go
// WRONG — not portable, pollutes home directory
func snapshotPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".ailinter", "knowledge", "snapshot.json")
}

// RIGHT — workspace-relative, portable
func (g *Graph) snapshotPath() string {
    return filepath.Join(g.KnowledgeDir, "snapshot.json")
}
```

### Logic Bugs in Persistence

When writing `NeedsRebuild()` or similar freshness checks, verify the return value carefully:

```go
// WRONG — always returns true after the loop
for ... {
    if stale { return true }
}
return true  // ← BUG

// RIGHT — return false if nothing is stale
return false
```

Also truncate timestamps to second precision — many filesystems don't preserve
nanosecond mtimes, causing false-positive staleness:

```go
if info.ModTime().Truncate(time.Second).After(mtime.Truncate(time.Second)) {
    return true
}
```

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
| `get_refactoring_strategy` | 🔧 NEXT STEP after analyze_code finds issues — exact refactoring steps + before/after examples for 8+ smells |
| `assess_file` | Quick classification: Go Ahead / Proceed with Care / Stop & Refactor (includes per-smell refactoring recommendations) |

---

## Quality Score Reference

| Score | Label | AI Guidance |
|-------|-------|-------------|
| 80-100 | Go Ahead | Safe for AI modification |
| 60-79 | Proceed with Care | Use guard clauses, prefer small changes, re-check after each edit |
| 40-59 | Needs Work | Significant issues — refactor incrementally in small steps |
| 0-39 | Stop & Refactor | Refactor BEFORE AI modification. Run `get_refactoring_strategy()` for detected issues. |

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
