# ailinter Setup Guide

Set up ailinter for your AI coding assistant in one command.

## Quick Setup (Interactive)

```bash
ailinter init
```

When run in a terminal, this launches an interactive setup that walks you through:

1. **Threshold profile** — Default (balanced), Strict (catches more), or Relaxed (fewer warnings)
2. **AI agents** — Select which tools to configure (OpenCode, Claude Code, Cursor, GitHub Copilot)
3. **Git hooks** — Optionally set up a pre-commit hook
4. **VS Code** — Add tasks, settings, and extension recommendations

All files are created idempotently (existing files are skipped, not overwritten).

## Non-Interactive Setup (CI/Scripts)

```bash
# Configure specific agents
ailinter init --agent opencode
ailinter init --agent claude
ailinter init --agent cursor
ailinter init --agent copilot

# All agents at once
ailinter init --agent all

# With VS Code integration
ailinter init --agent all --vscode

# With git pre-commit hook
ailinter init --agent opencode --hook

# Strict thresholds
ailinter init --profile strict

# Everything
ailinter init --agent all --vscode --hook --profile strict
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--agent <name>` | _(none)_ | AI agent: `opencode`, `claude`, `cursor`, `copilot`, `all` |
| `--vscode` | false | Create `.vscode/tasks.json`, `settings.json`, `extensions.json` |
| `--hook` | false | Create `.githooks/pre-commit` for pre-commit scanning |
| `--profile <name>` | `default` | Threshold profile: `default`, `strict`, `relaxed` |
| `--no-agents` | false | Skip AGENTS.md creation |

## Generated Files

### Always Created
| File | Description |
|------|-------------|
| `.ailinter.toml` | Configuration with thresholds for your language |
| `AGENTS.md` | AI agent instructions for this project (unless `--no-agents`) |

### `--agent opencode`
| File | Purpose |
|------|---------|
| `opencode.json` | MCP server config |
| `.opencode/agent/ailinter.md` | Sub-agent definition — delegates scans to `@ailinter` |
| `.opencode/skills/ailinter/SKILL.md` | Skill definition — teaches AI how to use ailinter tools |

### `--agent claude`
| File | Purpose |
|------|---------|
| `.claude/settings.json` | MCP server config for Claude Code |
| `CLAUDE.md` | Project instructions for Claude |

### `--agent cursor`
| File | Purpose |
|------|---------|
| `.cursor/mcp.json` | MCP server config for Cursor |
| `.cursor/rules/ailinter.mdc` | Always-applied rule — code quality checks before/after edits |

### `--agent copilot`
| File | Purpose |
|------|---------|
| `.github/copilot-instructions.md` | Instructions for GitHub Copilot Chat |

### `--vscode`
| File | Purpose |
|------|---------|
| `.vscode/tasks.json` | Problem matcher tasks (`ailinter: check current file`, `ailinter: check workspace`) |
| `.vscode/settings.json` | Enable ailinter on save |
| `.vscode/extensions.json` | Recommend the ailinter extension |

### `--hook`
| File | Purpose |
|------|---------|
| `.githooks/pre-commit` | Runs `ailinter check` on staged files before commit |

To enable the hook:

```bash
git config core.hooksPath .githooks
```

## Threshold Profiles

| Profile | Nesting | Function LOC | CC Warning | File LOC | Dup Min Lines |
|---------|:-------:|:------------:|:----------:|:--------:|:-------------:|
| **default** | warn 3 | warn 70 | — | warn 500 | — |
| **strict** | warn 2 | warn 50 | warn 6 | warn 400 | 3 lines |
| **relaxed** | warn 5 | warn 120 | warn 12 | warn 800 | 8 lines |

Customize thresholds in `.ailinter.toml` after initialization.

## MCP Server Connection

After setup, your AI assistant connects to ailinter via the MCP protocol. The config points to:

```json
{
  "command": "ailinter",
  "args": ["mcp"]
}
```

## Next Steps

1. Run `ailinter check .` to scan your codebase
2. Review the `.ailinter.toml` thresholds and adjust for your project
3. Configure git hooks: `git config core.hooksPath .githooks`
4. Open your project in your AI editor — the agent instructions are ready
