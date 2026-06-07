# MCP Integration Guide

Connect AILINTER to any AI coding assistant that supports the Model Context Protocol (MCP). Each integration takes under 60 seconds.

---

## Quick Reference

| Agent | Config File | Type | Setup Command |
|-------|-------------|------|---------------|
| Claude Code | `~/.claude/settings.json` | MCP stdio | `ailinter init --agent claude` |
| Cursor | `.cursor/mcp.json` (project) | MCP stdio | `ailinter init --agent cursor` |
| Cline | `cline_mcp_settings.json` | MCP stdio | Manual |
| OpenCode | `opencode.json` (project) | MCP stdio | `ailinter init --agent opencode` |
| Windsurf | `.windsurf/mcp.json` (project) | MCP stdio | Manual |
| Continue | `~/.continue/config.json` | MCP stdio | Manual |
| Copilot | `.github/copilot-instructions.md` | Instructions | `ailinter init --agent copilot` |

> **One-command setup:** `ailinter init --agent all` configures all supported agents at once.

---

## Universal Config

Every MCP-compatible tool uses the same config block — just placed in a different file:

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"]
    }
  }
}
```

If `ailinter` is not in your AI tool's PATH, use the full path:

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "/usr/local/bin/ailinter",
      "args": ["mcp"]
    }
  }
}
```

Find it with `which ailinter`.

---

## 1. Claude Code

**Config file:** `~/.claude/settings.json`

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"]
    }
  }
}
```

Or one-command setup:
```bash
ailinter init --agent claude
```

**Verify:**
```bash
claude mcp check ailinter
# → "MCP server 'ailinter' is connected and ready"
```

**What Claude can do:**
- Call `analyze_code` automatically before editing files
- Ask: "What's the quality score of src/main.go?"
- Ask: "Check this file for secrets"
- Follow the refactoring loop: detect → strategy → fix → verify

---

## 2. Cursor

**Config file:** `.cursor/mcp.json` (project root)

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"]
    }
  }
}
```

Or one-command setup:
```bash
ailinter init --agent cursor
```

**Verify:**
1. Restart Cursor (or "Developer: Reload Window")
2. Look for "ailinter" in the MCP server list (bottom-right status bar)

**Cursor Rules (automatic checks):**

Create `.cursor/rules/ailinter.mdc`:

```markdown
---
description: Check code quality before AI edits
globs: *.go, *.py, *.js, *.ts, *.java
alwaysApply: true
---

Before modifying any file:
1. Run `ailinter assess <filepath>` to check if it's safe to edit
2. If score < 60 ("Stop & Refactor"), refactor the file first
3. If score < 80 ("Proceed with Care"), make small changes and re-check
After editing:
4. Run `ailinter assess <filepath>` again to verify quality hasn't regressed
5. If secrets or vulnerabilities are found, fix them before committing
```

---

## 3. Cline

**Config file:** `cline_mcp_settings.json`

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"]
    }
  }
}
```

**Verify:**
1. Save the config file
2. Restart Cline
3. Type: "What tools do you have?" — AILINTER's 7 MCP tools should appear

---

## 4. OpenCode

**Config file:** `opencode.json` (project root)

```json
{
  "mcp": {
    "ailinter": {
      "type": "local",
      "command": ["ailinter", "mcp"],
      "enabled": true
    }
  },
  "skills": {
    "paths": [".opencode/skills"]
  }
}
```

Or one-command setup:
```bash
ailinter init --agent opencode
```

This also creates `.opencode/skills/ailinter/SKILL.md` — an instruction file that teaches your AI assistant how to use all 7 MCP tools effectively.

**Verify:**
```bash
opencode mcp list
# → ✓ ailinter — local command: ailinter mcp
```

---

## 5. Windsurf

**Config file:** `.windsurf/mcp.json` (project root)

```json
{
  "mcpServers": {
    "ailinter": {
      "command": "ailinter",
      "args": ["mcp"]
    }
  }
}
```

**Verify:**
1. Save `.windsurf/mcp.json`
2. Restart Windsurf
3. Open the MCP panel (plug icon in status bar)
4. "ailinter" should show as connected

**Windsurf Rules:**

Create `.windsurf/rules/ailinter.md`:

```markdown
# AILINTER Quality Gate

Before editing any file:
1. Run `ailinter assess <filepath>` to get the current quality score
2. If score < 60, do not edit — explain quality issues and ask user to refactor
3. If score < 80, make small, incremental changes and re-check after each

After editing:
1. Run `ailinter assess <filepath>` to verify no regression
2. Run `ailinter secrets <filepath>` if edit touched credential-like code
```

---

## 6. Continue

**Config file:** `~/.continue/config.json`

```json
{
  "experimental": {
    "mcpServers": {
      "ailinter": {
        "command": "ailinter",
        "args": ["mcp"]
      }
    }
  }
}
```

> **Note:** Continue's MCP support is experimental. If the above doesn't work, try:
> ```json
> {
>   "mcpServers": {
>     "ailinter": {
>       "type": "local",
>       "command": "ailinter",
>       "args": ["mcp"]
>     }
>   }
> }
> ```

**What Continue can do:**
- "Check this file for secrets"
- "What's the code quality score?"
- "Refactor this function following AILINTER's strategy"

---

## 7. GitHub Copilot

**Config file:** `.github/copilot-instructions.md` (project root)

Copilot doesn't natively run MCP servers, but the instructions file guides its behavior:

```markdown
## Code Quality with AILINTER

This project uses AILINTER for code quality assurance.

### Before Writing Code
Run `ailinter assess <file>` to check if the file is safe to edit.
- Score < 60: Explain quality issues, suggest refactoring first
- Score < 80: Make small changes, re-check

### Before Committing
- No hardcoded secrets: `ailinter check --secrets-only <file>`
- No vulnerabilities: `ailinter check --vulnerabilities-only <file>`
- Quality score 80+: `ailinter check <file>`

### Score Reference
| 80-100 | Go Ahead | Safe to modify |
| 60-79 | Proceed with Care | Small changes |
| 40-59 | Needs Work | Refactor first |
| 0-39 | Stop & Refactor | Must fix first |
```

Or one-command setup:
```bash
ailinter init --agent copilot
```

**Verify:**
1. Save `.github/copilot-instructions.md`
2. In VS Code Copilot Chat, ask: "What code quality tool does this project use?"
3. Copilot should reference AILINTER

---

## The 7 MCP Tools

Once connected, your AI assistant has access to these tools:

| Tool | What It Does | Typical Response |
|------|-------------|-----------------|
| `analyze_code` | Full quality + vulnerability scan | Score (0–100), issues list |
| `scan_for_secrets` | Check for hardcoded secrets (redacted) | Clean or list of findings |
| `assess_file` | Quick check: "Go Ahead / Care / Stop" | One-word assessment + score |
| `get_refactoring_strategy` | Step-by-step fix for a code smell | Strategy name + steps |
| `list_hotspots` | Files with high churn × low quality | File list with scores |
| `set_config` | Change AILINTER settings | Confirmation |
| `get_config` | View current configuration | Config object |

---

## Troubleshooting MCP Connections

| Symptom | Cause | Fix |
|---------|-------|-----|
| "Command not found" | `ailinter` not in AI tool's PATH | Use full path in config |
| MCP server won't start | Config file has syntax error | Validate JSON with `python3 -m json.tool` |
| Server starts but no tools | Wrong command/args | Must be `"command": "ailinter", "args": ["mcp"]` |
| Copilot ignores instructions | Session not refreshed | Reload VS Code, clear Copilot session |
| Cursor "Connection error" | Binary not found at config path | `which ailinter` → use that full path |

### Universal Debug

```bash
# 1. Verify AILINTER is installed
ailinter version

# 2. Test MCP server directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ailinter mcp

# 3. Find the binary path
which ailinter
```

### Path Resolution

Most MCP connection issues come from the AI tool not finding `ailinter` in its PATH. The AI tool's PATH may differ from your shell's PATH.

```bash
which ailinter
# → /usr/local/bin/ailinter  (use this full path in config)
```

---

## Security Notes

- **Secrets are redacted**: `scan_for_secrets` returns only the first 4 and last 4 characters. Your AI never sees the full value.
- **Read-only mode**: Set `read_only = true` in `.ailinter.toml` under `[mcp]` to disable `set_config`.
- **No data leaves your machine**: AILINTER runs entirely locally. Telemetry is opt-out with `AILINTER_NO_TELEMETRY=1`.
