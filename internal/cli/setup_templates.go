package cli

import "github.com/ailinter/ailinter/internal/parser"

const opencodeAgentConfig = `---
model: deepseek/deepseek-v4-flash
mode: subagent
description: Dedicated ailinter code scanner. Use for running ` + "`ailinter check`" + ` on files/directories, analyzing code quality, scanning for secrets, detecting vulnerabilities. Delegate all ailinter scans here.
permission:
  bash: allow
  read: allow
---

You are the ailinter code scanner. Your job is to run security and quality scans and return concise results.

## Workflow

1. Run ` + "`ailinter check <path> --format problems --no-secrets`" + ` via bash — code quality + vulnerabilities
2. Run ` + "`ailinter check <path> --format problems --secrets-only`" + ` via bash — secrets only (problems format omits raw values)
   (problems format: ~3x smaller than JSON, line-oriented, no parse failures, score line per file)
3. Parse output: ` + "`# <path> score=<N> label=<tier>...`" + ` lines for scores, subsequent lines for findings
4. Return a concise structured summary (≤ 30 lines) with scores, issues, vulnerabilities, secrets, and tier breakdown

## Rules
- Never output raw JSON — only the summary
- If score < 75, flag as STOP & REFACTOR
- Return the result and stop — do not attempt to fix issues
`

var opencodeSkill = `---
name: ailinter
description: Use when analyzing source code files for quality issues, scanning for hardcoded secrets, checking if a file is safe for AI modification, getting refactoring strategies, listing Git hotspots, or managing ailinter configuration. Trigger keywords: ailinter, analyze, lint, code quality, refactor, secrets, assess, hotspots, code health.
---

# ailinter — AI Linter & Code Quality MCP

ailinter is a local MCP server that gives AI coding assistants visibility into code quality, secrets, and refactoring guidance before and after edits.

## When to Use Each Tool

### 1. analyze_code — Before and After Every File Edit

Call ` + "`analyze_code(file_path)`" + ` on any file BEFORE writing changes, and again AFTER changes are applied. It returns a quality score (0-100) with tier classification:

` + "\n" + parser.TierReferenceTable() + "\n" + `
**Rule:** Always run ` + "`analyze_code`" + ` before committing. If score dropped, fix the regression.

### 2. scan_for_secrets(content) — On All Generated Code

Call ` + "`scan_for_secrets(content)`" + ` on every AI-generated code block before committing. It scans for 150+ secret patterns (AWS keys, Stripe tokens, GitHub PATs, private keys, JWT, etc.).

**Rule:** Never commit code with detected secrets. Rewrite to use environment variables.

### 3. get_refactoring_strategy(smell_name) — The Refactoring Loop 🔧

When ` + "`analyze_code`" + ` reports code smells, call ` + "`get_refactoring_strategy(smell_name)`" + ` for exact step-by-step refactoring instructions with before/after examples and verification checks.

**Always follow this loop when smells are found:**
1. Call ` + "`get_refactoring_strategy(\"smell_name\")`" + ` for each detected smell
2. Refactor in 3-5 small, reviewable steps
3. Re-run ` + "`analyze_code`" + ` after each step — confirm score improves
4. Do NOT proceed until score is 80+

Available patterns: ` + "`deep_nesting`" + `, ` + "`brain_method`" + `, ` + "`bumpy_road`" + `, ` + "`complex_conditional`" + `, ` + "`god_class`" + `, ` + "`long_parameter_list`" + `, ` + "`primitive_obsession`" + `, ` + "`duplicated_code`" + `.

### 4. assess_file(file_path) — Quick Safety Check

Call ` + "`assess_file(file_path)`" + ` for a quick "Go Ahead / Proceed with Care / Stop & Refactor" classification before modifying a file.

### 5. list_hotspots(repo_path) — Find Priority Refactoring Targets

Call ` + "`list_hotspots(repo_path)`" + ` to find files that change frequently AND have low quality scores.

### 6. set_config / get_config — Configuration

Use ` + "`set_config(key, value)`" + ` to configure ailinter. Valid keys: ` + "`language`" + `, ` + "`repo_path`" + `, ` + "`enabled_tools`" + `, ` + "`read_only`" + `, ` + "`disable_git`" + `. Use ` + "`get_config`" + ` to view current settings.

## Workflow: The Refactoring Loop

**This is the most important ailinter pattern.** Whenever analyze_code finds issues:

` + "```" + `
User: "Add caching to the getUser function"
Agent:
  1. Run analyze_code("src/users.go") → Score: 42, At Risk, 3 issues
  2. ALERT: Must refactor before adding features
  3. Call get_refactoring_strategy("deep_nesting") → Guard Clauses + Extract Method
  4. Refactor in 3 small steps, re-running analyze_code after each: 42 → 61 → 85 → 97
  5. NOW safe to add caching feature
  6. Run analyze_code after edit → Score: 96 (no regression)
  7. Run scan_for_secrets(content) on new code → clean
  8. Suggest commit
` + "```" + `

**Rule:** If analyze_code reports any issues with score < 80, you MUST call get_refactoring_strategy() before writing new code. Never skip the refactoring loop.
`

const opencodeMCPConfig = `{
  "mcp": {
    "ailinter": {
      "type": "local",
      "command": ["ailinter", "mcp"],
      "enabled": true,
      "env": {
        "AILINTER_MCP_CLIENT": "cursor"
      }
    }
  }
}
`

const claudeMCPConfig = `{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"],
      "env": {
        "AILINTER_MCP_CLIENT": "claude"
      }
    }
  }
}
`

var claudeInstructions = `# CLAUDE.md

## Code Quality with ailinter

This project uses ailinter for code quality and security scanning. When making changes:

### Before and After Every Edit
- Run the ailinter ` + "`analyze_code`" + ` tool on the file BEFORE and AFTER changes
- **Score 80-100 (Go Ahead)**: Safe to modify freely
- **Score 60-79 (Proceed with Care)**: Use guard clauses, small changes, re-check
- **Score 40-59 (Needs Work)**: Significant issues — refactor incrementally in small steps
- **Score <40 (Stop & Refactor)**: **Refactor FIRST** — call ` + "`get_refactoring_strategy`" + ` before adding features

### The Refactoring Loop (Mandatory When Score < 80)
1. Call ` + "`analyze_code(file)`" + ` — know your baseline score and smells
2. For EACH smell detected, call ` + "`get_refactoring_strategy(\"smell_name\")`" + ` for exact steps
3. Refactor in 3-5 small steps, one smell at a time
4. Re-run ` + "`analyze_code`" + ` after each step — confirm score improves
5. Do NOT add new features until score is 80+

### Secret Detection
- Run ` + "`scan_for_secrets`" + ` on all generated code
- Never commit hardcoded secrets — use environment variables
- If secrets found, rewrite before suggesting commit

### Quick Safety Check
- Use ` + "`assess_file`" + ` for a quick Go Ahead / Proceed with Care / Stop & Refactor check

### Git Hotspots
- Use ` + "`list_hotspots`" + ` to find frequently-changed low-quality files

### CLI Commands
` + "```bash" + `
ailinter check <file|dir>           # Full check (quality + secrets + vulns)
ailinter check --no-secrets <path>  # Skip secrets (for AI context safety)
ailinter check --format problems <path>  # Machine-parseable output
` + "```" + `

## Quality Score Reference

` + "\n" + parser.TierReferenceTable() + "\n" + `
## Available MCP Tools

| Tool | Purpose |
|------|---------|
| analyze_code | Full structural analysis: quality score (0-100), issues, severity, locations |
| scan_for_secrets | Secret detection: AWS keys, API tokens, private keys, JWT, etc. |
| get_refactoring_strategy | Pattern lookup: returns exact steps + examples for each issue |
| assess_file | Quick classification: Go Ahead / Proceed with Care / Stop & Refactor |
| list_hotspots | Frequently-changed files with low quality scores |
`

var cursorRules = `---
description: ailinter — Code Quality & Security
globs: **/*.go,**/*.py,**/*.js,**/*.ts,**/*.tsx,**/*.java,**/*.rs,**/*.rb,**/*.c,**/*.cpp,**/*.h,**/*.hpp,**/*.cs,**/*.swift,**/*.kt,**/*.kts,**/*.php,**/*.sh,**/*.bash,**/*.tf,**/*.yaml,**/*.yml,**/*.toml,**/*.xml,**/*.html,**/*.css,**/*.sql
alwaysApply: true
---

# ailinter — Code Quality & Security

## Before Every Edit
- Call the ailinter ` + "`analyze_code`" + ` tool on the file
- **Score 80-100 (Go Ahead)**: Safe to modify
- **Score 60-79 (Proceed with Care)**: Use guard clauses, re-check after
- **Score 40-59 (Needs Work)**: Significant issues — refactor incrementally
- **Score <40 (Stop & Refactor)**: **MUST refactor first** — call ` + "`get_refactoring_strategy`" + ` before touching code

## The Refactoring Loop (Critical Path)
When ` + "`analyze_code`" + ` finds issues:
1. Call ` + "`get_refactoring_strategy(\"smell_name\")`" + ` for each detected smell
2. Refactor in 3-5 small steps
3. Re-run ` + "`analyze_code`" + ` after each step — verify score improves
4. Repeat until score is 80+

## After Every Edit
- Call ` + "`analyze_code`" + ` again to confirm no regression
- If score dropped, fix the issue before marking as done

## Secret Detection
- Call ` + "`scan_for_secrets`" + ` on all generated code before suggesting commit
- Never commit hardcoded credentials — use env vars

## Quick Check
- ` + "`assess_file`" + ` for a fast tier classification (includes per-smell refactoring recommendations)

Available smells: deep_nesting, brain_method, bumpy_road, complex_conditional, god_class, long_parameter_list, primitive_obsession, duplicated_code
`

const cursorMCPConfig = `{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"],
      "env": {
        "AILINTER_MCP_CLIENT": "cursor"
      }
    }
  }
}
`

var copilotInstructions = `# ailinter Code Quality & Security

## Before Every Code Change
- Run the ailinter ` + "`analyze_code`" + ` tool on the file before editing
- **Score 80-100 (Go Ahead)**: Safe to modify
- **Score 60-79 (Proceed with Care)**: Make small, focused changes; re-check after each
- **Score 40-59 (Needs Work)**: Significant issues — refactor incrementally in small steps
- **Score <40 (Stop & Refactor)**: **MUST refactor first** — call ` + "`get_refactoring_strategy`" + ` before adding features

## The Refactoring Loop (Mandatory)
When ` + "`analyze_code`" + ` scores < 80:
1. For each detected smell, call ` + "`get_refactoring_strategy(\"smell_name\")`" + ` for exact steps
2. Refactor in 3-5 small steps, one smell at a time
3. Re-run ` + "`analyze_code`" + ` after each step — confirm score improves
4. Target: score 80+ before adding new features

## After Every Code Change
- Re-run ` + "`analyze_code`" + ` to confirm score hasn't decreased
- Fix any regressions before marking as done

## Security
- Run ` + "`scan_for_secrets`" + ` on ALL generated code
- Never commit hardcoded secrets — use environment variables
- If secrets detected, rewrite code before suggesting commit

## Available MCP Tools
- ` + "`analyze_code(file_path)`" + ` — quality score + issues + vulns
- ` + "`scan_for_secrets(content)`" + ` — 150+ secret patterns
- ` + "`get_refactoring_strategy(smell_name)`" + ` — 🔧 exact refactoring steps (call when analyze_code finds issues)
- ` + "`assess_file(file_path)`" + ` — quick safety tier (includes per-smell refactoring recommendations)
- ` + "`list_hotspots(repo_path)`" + ` — priority refactoring targets
`

const vscodeSettings = `{
  "ailinter.enable": true,
  "ailinter.run": "onSave"
}
`

const vscodeExtensions = `{
  "recommendations": [
    "ailinter.ailinter"
  ]
}
`

const gitPreCommitHook = `#!/bin/sh
# ailinter pre-commit hook
# Runs code quality + vulnerability scan on staged files.
# Secrets scanning is skipped in hooks (secrets should be checked manually or via CI).

STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(go|py|js|ts|tsx|java|rs|rb|c|cpp|h|hpp|cs|swift|kt|kts|php|sh|bash|tf|yaml|yml|toml|xml|html|css|sql)$' || true)

if [ -z "$STAGED_FILES" ]; then
    exit 0
fi

echo "ailinter: checking staged files..."
FAILED=0

for FILE in $STAGED_FILES; do
    if [ -f "$FILE" ]; then
        ailinter check --format problems --no-secrets "$FILE"
        if [ $? -ne 0 ]; then
            FAILED=1
        fi
    fi
done

if [ $FAILED -ne 0 ]; then
    echo ""
    echo "ailinter found issues. Review the findings above."
    echo "To skip this check: git commit --no-verify"
    exit 1
fi

echo "ailinter: all clear."
exit 0
`

const strictConfig = `# ailinter configuration (strict thresholds)
# Generated by: ailinter init

extends = "default"

[rules]
deep_nesting = { weight = 1.0, warning = 2, alert = 4 }
brain_method = { weight = 1.0, warning_lines = 50, alert_lines = 150 }
file_bloat = { weight = 1.0, warning_lines = 400, alert_lines = 1000 }
complex_conditional = { weight = 1.0, branches_warning = 2, branches_alert = 4 }
bumpy_road = { weight = 1.0, bumps_warning = 1, bump_depth = 2 }
long_parameter_list = { weight = 1.0, warning = 3, alert = 5 }
cyclomatic_complexity = { weight = 1.0, warning = 6, alert = 10 }
excessive_comments = { ratio = 0.2 }
global_data = { warning = 3 }
long_scope_variable = { min_lines = 30 }
duplicated_code = { weight = 1.0, min_lines = 3 }

[gitleaks]
extend = "default"

[mcp]
enabled_tools = ["*"]
read_only = false
`

const relaxedConfig = `# ailinter configuration (relaxed thresholds)
# Generated by: ailinter init

extends = "default"

[rules]
deep_nesting = { weight = 1.0, warning = 5, alert = 7 }
brain_method = { weight = 1.0, warning_lines = 120, alert_lines = 300 }
file_bloat = { weight = 1.0, warning_lines = 800, alert_lines = 3000 }
complex_conditional = { weight = 1.0, branches_warning = 3, branches_alert = 7 }
bumpy_road = { weight = 1.0, bumps_warning = 3, bump_depth = 3 }
long_parameter_list = { weight = 1.0, warning = 6, alert = 10 }
cyclomatic_complexity = { weight = 1.0, warning = 12, alert = 20 }
excessive_comments = { ratio = 0.4 }
global_data = { warning = 10 }
long_scope_variable = { min_lines = 80 }
duplicated_code = { weight = 1.0, min_lines = 8 }

[gitleaks]
extend = "default"

[mcp]
enabled_tools = ["*"]
read_only = false
`
