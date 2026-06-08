# CLI Reference

Complete reference for all AILINTER CLI commands.

> Complete CLI reference for ailinter v1.0.0. See the [README](https://github.com/ailinter/ailinter#cli) for a quick-start overview.

## Commands Overview

| Command | Description |
|---------|-------------|
| `ailinter check` | Analyze files for quality, secrets, and vulnerabilities |
| `ailinter init` | Setup project: agents, hooks, VS Code |
| `ailinter mcp` | Start MCP server on stdio |
| `ailinter rules list` | List all threshold defaults |
| `ailinter version` | Print version information |

## `ailinter check`

```
ailinter check [files...] [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--format` | Output format: `table` (default), `json`, `markdown`, `problems`, `sarif` |
| `--no-secrets` | Skip secret scanning |
| `--no-vulnerabilities` | Skip vulnerability scanning |
| `--secrets-only` | Only scan for secrets |
| `--vulnerabilities-only` | Only scan for vulnerabilities |
| `--lang` | Force language detection |
| `--no-gitignore` | Don't respect `.gitignore` patterns |
| `--estimate-tokens` | Estimate AI token cost |
| `--diff <ref>` | Diff-aware: scan only lines changed relative to a git ref |
| `--output <file>` | Write output to file (used with `--format sarif`) |

## `ailinter init`

```
ailinter init [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--agent` | Agent to configure: `claude`, `cursor`, `copilot`, `opencode`, `all` |
| `--vscode` | Generate VS Code tasks, settings, extensions |
| `--hook` | Install pre-commit git hook |
| `--profile` | Threshold profile: `default`, `strict` |

## `ailinter rules`

```
ailinter rules list [--lang <language>]
```
