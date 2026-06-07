# AILINTER Documentation

Welcome to AILINTER — the open-source AI Code Safety Visor. This site covers installation, usage, integration, and reference materials.

## Quick Summary

```
One 30 MB binary · 20 quality detectors · 269+ secret rules · 58 vulnerability patterns · 7 MCP tools · Zero dependencies
```

AILINTER gives your AI coding assistant a safety checklist. Before the AI writes a line — and after — it evaluates the file and tells the AI whether to **Go Ahead**, **Proceed with Care**, or **Stop & Refactor**.

## Getting Started

| Page | What You'll Learn |
|------|-------------------|
| [Installation](installation.md) | All install methods: Homebrew, Go install, binary download, Docker |
| [Quick Start](https://github.com/ailinter/ailinter#-quick-start) | 30-second setup on the README |
| [MCP Integration](mcp.md) | Connect AILINTER to Claude, Cursor, Cline, OpenCode, Windsurf, Copilot, Continue |
| [CI Integration](ci.md) | GitHub Actions workflow, quality gates, secret blocking |

## Core Concepts

| Page | What You'll Learn |
|------|-------------------|
| [CLI Reference](cli.md) | All commands: `check`, `init`, `mcp`, `rules` |
| [Quality Scoring](quality.md) | The 0–100 score, 20 detectors, AI guidance tiers |
| [Secret Scanning](secrets.md) | 269+ rules, 100+ providers, redacted output |
| [Vulnerability Patterns](vulnerabilities.md) | 58 patterns across 6 categories |
| [Refactoring Strategies](refactoring.md) | 8+ code smells with step-by-step fixes |
| [Configuration](configuration.md) | `.ailinter.toml` reference, thresholds, profiles |

## Reference

| Page | Description |
|------|-------------|
| [Benchmarks](benchmarks.md) | 7-tool comparison, recall/precision, speed |
| [Contributing](contributing.md) | Development setup, code standards, PR workflow |
| [Setup Guide](setup.md) | Interactive setup, agent-specific configs |

## The Refactoring Loop

The most important workflow in AILINTER:

```
1. BEFORE: analyze_code(file) → score
2. If score < 80 → get_refactoring_strategy → refactor in 3–5 steps
3. Make your change
4. AFTER: analyze_code(file) → confirm no regression
5. scan_for_secrets(content) → clean
6. Commit
```

Never skip the refactoring loop. If `analyze_code` or `assess_file` reports issues with score < 80, `get_refactoring_strategy` is the mandatory next step.

## Quick Reference

### Install

```bash
# macOS
brew install ailinter/ailinter/ailinter

# Any platform (Go)
go install github.com/ailinter/ailinter/cmd/ailinter@latest
```

### Scan

```bash
ailinter check .                    # Full scan (quality + secrets + vulns)
ailinter check main.go              # Single file
ailinter check --secrets-only .     # Secrets only
ailinter check --format json .      # Machine-readable output
```

### MCP

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

### Init

```bash
ailinter init                        # Interactive setup
ailinter init --agent all            # All AI agents at once
ailinter init --agent all --vscode --hook  # Everything
```

## Community

- **Website:** [ailinter.dev](https://ailinter.dev)
- **GitHub:** [github.com/ailinter/ailinter](https://github.com/ailinter/ailinter)
- **Issues:** [github.com/ailinter/ailinter/issues](https://github.com/ailinter/ailinter/issues)
- **License:** MIT — open source, forever.
