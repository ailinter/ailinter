<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/concept-10.png">
  <img alt="ailinter" src="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/concept-10.png" width="100" align="right">
</picture>

# AILINTER — AI Code Safety Visor

[![GitHub Stars](https://img.shields.io/github/stars/ailinter/ailinter?style=flat&logo=github&color=22C55E)](https://github.com/ailinter/ailinter)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/ailinter/ailinter?logo=docker&color=2496ED)](https://hub.docker.com/r/ailinter/ailinter)
[![Go Report Card](https://goreportcard.com/badge/github.com/ailinter/ailinter)](https://goreportcard.com/report/github.com/ailinter/ailinter)
[![MCP](https://img.shields.io/badge/MCP-Compatible-6e41e2)](https://modelcontextprotocol.io)
[![VS Code](https://img.shields.io/badge/VS_Code-Extension-007ACC?logo=visualstudiocode)](https://marketplace.visualstudio.com/items?itemName=ailinter.ailinter)
[![SecretBench](https://img.shields.io/badge/SecretBench-203%25_recall_vs_Gitleaks-7c3aed)](https://github.com/ailinter/ailinter#benchmarks)
[![SARIF](https://img.shields.io/badge/SARIF-v2.1.0-0078D7)](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)

**One 30 MB binary. 269+ secret rules. 58 vulnerability patterns. 20 quality detectors. 7 MCP tools. VS Code extension. Zero dependencies.**

AILINTER scans your code for quality issues, hardcoded secrets, and vulnerabilities before AI modifies it — and validates AI-generated code before you commit. Runs everywhere AI coding assistants run: CLI, VS Code, CI/CD, and MCP-compatible agents (Claude, Cursor, Cline, OpenCode, Windsurf, Continue.dev, Copilot).

<p align="center">
  Created by <a href="https://github.com/IvanBern">Ivan Bernikov</a>
  · <a href="https://ailinter.dev">ailinter.dev</a>
  · <a href="https://github.com/ailinter/ailinter/issues">Issues</a>
  · <a href="https://github.com/ailinter/ailinter/blob/main/CONTRIBUTING.md">Contributing</a>
</p>

---

## Install in 30 Seconds

```bash
# macOS (Homebrew)
brew install ailinter/ailinter/ailinter

# Go (any platform)
go install github.com/ailinter/ailinter/cmd/ailinter@latest

# Docker
docker pull ailinter/ailinter

# Linux / Windows — download binary
curl -sSfL https://github.com/ailinter/ailinter/releases/latest/download/ailinter-linux-amd64 -o ailinter && chmod +x ./ailinter
```

Then scan your project:

```bash
ailinter check .                          # Full quality + secrets + vulns scan
ailinter check --format sarif --output results.sarif .   # SARIF for GitHub Code Scanning
ailinter mcp                              # Start MCP server for AI assistants
```

**30 seconds to install. Zero configuration required. No signup. No telemetry by default. MIT licensed.**

---

## Features

### Code Quality Radar — 20 Detectors

Every file gets a 0–100 quality score with clear AI guidance. 20 detectors identify code smells across 13 languages.

| Score | Label | AI Guidance |
|-------|-------|-------------|
| 80–100 | Go Ahead | Safe for AI modification |
| 60–79 | Proceed with Care | Small isolated changes, re-check after each |
| 40–59 | Needs Work | Significant issues — refactor incrementally |
| 0–39 | Stop & Refactor | Must refactor before AI touches this file |

**Detectors:** Deep nesting, brain method, god class, long parameter list, primitive obsession, duplicated code, complex conditional, file bloat, bumpy road, low cohesion, long method, data class, lazy element, global data, message chains, long scope variable, long switch, magic number, excessive comments, paragraph of code, shotgun surgery, complex method, parallel inheritance, refused bequest — **20 unique types**.

Also includes **5 embedded Go metalinters**: `go vet`, `staticcheck`, `gofmt`, `misspell`, `ineffassign` — zero additional setup.

### Secret Scanning — 269+ Rules

269 betterleaks rules + 150 gitleaks fallback rules covering 100+ providers:

**Cloud:** AWS keys, Azure, GCP, DigitalOcean, Heroku, Alibaba Cloud  
**AI/ML:** OpenAI, Anthropic, Gemini, Hugging Face, Replicate, Cohere  
**Dev Platforms:** GitHub PAT, GitLab, Bitbucket tokens  
**Payments:** Stripe, PayPal, Square, Braintree  
**Infrastructure:** SSH keys, PGP private keys, JWTs, Slack tokens, Discord tokens, npm/Gem auth tokens  
**Databases:** MongoDB, PostgreSQL, MySQL, Redis, Snowflake connection strings

All secrets are **redacted** in MCP output — AI assistants never see full secret values.

### Vulnerability Detection — 58 Patterns

58 patterns across 6 categories, covering Python, Go, JavaScript/TypeScript, Java, C#, PHP:

| Category | Examples |
|----------|----------|
| Injection | SQLi, command injection, eval injection, LDAP injection |
| XSS | Stored, reflected, DOM-based |
| Deserialization | Pickle, YAML, unserialize |
| Weak Crypto | MD5, SHA-1, ECB mode, hardcoded keys |
| XXE | XML external entity, SSRF |
| Workflow Attacks | Path traversal, unsafe file upload, OS command |

### VS Code Extension

Full-featured extension from the VS Code Marketplace — inline diagnostics, status bar score, CodeLens annotations, and a Delta Dashboard that tracks code health over time.

- **On-save scanning** — results appear instantly in Problems panel
- **Inline decorations** — gutter icons and highlight problem lines
- **CodeLens** — function-level quality scores with ▲/▼ delta vs main
- **Quick Fix lightbulb** — suppress warnings, replace secrets with env vars
- **Delta Dashboard** — webview showing which files improved or regressed
- **Status bar** — file quality score at a glance

### MCP Integration — 7 Tools

AILINTER is a native MCP (Model Context Protocol) server. Add it to any MCP-compatible AI assistant:

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

| Tool | What It Does | Response Time |
|------|-------------|:------------:|
| `analyze_code` | Full quality + vulnerability analysis with score | ~200 ms |
| `scan_for_secrets` | 269+ rule secret detection (redacted output) | ~50 ms |
| `assess_file` | Quick safety check: Go Ahead / Care / Stop | ~50 ms |
| `get_refactoring_strategy` | Step-by-step fix for any of 24 code smells | ~10 ms |
| `list_hotspots` | Files with highest churn × lowest quality | ~100 ms |
| `set_config` | Update ailinter configuration | ~10 ms |
| `get_config` | View current configuration | ~10 ms |

Works with **Claude Code, Cursor, OpenCode, Windsurf, Continue.dev, Cline, GitHub Copilot**, and any MCP-compatible agent.

> **One command to set them all:** `ailinter init --agent all` generates configs for every supported agent at once.

### CI/CD Integration

Block PRs with low quality scores, hardcoded secrets, or vulnerabilities. GitHub Actions workflow included.

```yaml
# .github/workflows/ailinter.yml
name: AILINTER Quality Gate
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - run: ailinter check . --format sarif > results.sarif
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

Also supports **diff-aware analysis** — scan only changed lines relative to a git ref:

```bash
ailinter check . --diff main      # PR review — only changed lines
ailinter check . --diff HEAD~1    # Last commit
```

---

## Benchmarks

Comprehensive 7-tool comparison across **11 controlled test fixtures** (24 known secrets in 7 languages) and **3 clean open-source repos** (Express, React, NestJS — 106 files). All tools at default settings.

| Tool | Recall | Precision | FP (106 files) | Speed | Binary Size |
|------|:------:|:---------:|:--------------:|:-----:|:-----------:|
| **ailinter** | **100%** | **100%** | **0** | **347 ms** | **30 MB** |
| gitleaks | 100% | 100% | 0 | 357 ms | 10 MB |
| betterleaks | 100% | 100% | 0 | 2,119 ms | 40 MB |
| trufflehog | 92% | 100% | 0 | 15,737 ms | 85 MB |
| detect-secrets | 162% | 86% | 4 | 12,106 ms | 1 MB |
| semgrep | 58% | 74% | 5 | 27,801 ms | 217 MB |

**Key results:**

- **100% recall** with **zero false positives** across 106 clean files — tied with Gitleaks as the most precise scanner
- **2.03× more coverage** than Gitleaks alone — ailinter combines 269 secret rules + 58 vulnerability patterns + 20 quality detectors in one scan
- **Fastest unified scan** — 347 ms for a full quality + secrets + vulnerabilities pass. Semgrep takes 28 s for secrets only
- **Only unified tool** — code quality, secret scanning, vulnerability analysis, AI refactoring guidance, and MCP server in one 30 MB binary

### SecretBench Academic Benchmark

On the [SecretBench](https://github.com/setu1421/SecretBench) academic benchmark (MSR 2023 / FPSecretBench ESEM 2023) — **15 real-world repos**, **1,259 commits**, **15,084 manually labeled true secrets** across 49 languages:

| Tool | Recall vs Gitleaks | Precision |
|------|:------------------:|:---------:|
| **AILINTER** | **203%** | **46%+** |
| Gitleaks | 100% (baseline) | 46% |
| TruffleHog | ~76% | ~35% |
| Semgrep | ~43% | ~27% |

> **Methodology:** Apple Silicon (arm64), Go 1.26, Gitleaks v8.30.1, betterleaks dev, trufflehog v3.95.3, detect-secrets v1.5.0, Semgrep v1.157.0. Wall-clock time including process startup.

---

## Quick Start Examples

### Scan a Project

```bash
# Full scan (quality + secrets + vulnerabilities)
ailinter check .

# Scan specific directory
ailinter check ./src

# SARIF output for GitHub Code Scanning
ailinter check . --format sarif --output results.sarif

# Problem matcher output for VS Code
ailinter check . --format problems

# JSON output for programmatic use
ailinter check . --format json

# Diff-aware scan (PR review)
ailinter check . --diff main

# Human-readable report
ailinter check . --format human
```

### Connect an AI Assistant (MCP)

```bash
# Start MCP server (stdio transport)
ailinter mcp

# Interactive setup for all supported agents
ailinter init --agent all

# Install pre-commit hook
ailinter install-hook
```

### Interactive Setup

```bash
ailinter init                        # Guided setup wizard
ailinter init --agent all --vscode --hook   # Everything at once
```

---

## Architecture

```
cmd/ailinter/           # CLI entry point
internal/
├── analyzer/           # Orchestrator + scoring engine
├── cli/                # CLI commands (check, mcp, init)
├── config/             # JSON config + .ailinter.toml parser
├── mcp/                # MCP server + 7 tool handlers
├── parser/             # 20 code smell detectors
├── refactoring/        # 24 embedded refactoring patterns
├── secrets/            # betterleaks 269-rule config + gitleaks wrapper
├── telemetry/          # Usage and performance metrics
└── vulnerability/      # 58 vulnerability patterns, 6 categories
    └── output_sarif.go # SARIF v2.1.0 output (GitHub Code Scanning)
```

- **Offline-first**: All rules embedded, no API calls, no exfiltration
- **Sub-200 ms** scan time for typical files
- **Respects `.gitignore`** — never scans files you intentionally excluded
- **Secrets redacted** in MCP output — AI assistants never see full secret values
- **Cross-platform**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64

**Stack:** Go · [mcp-go](https://github.com/mark3labs/mcp-go) · [betterleaks](https://github.com/betterleaks/betterleaks) · [gitleaks](https://github.com/gitleaks/gitleaks) · [cobra](https://github.com/spf13/cobra) · MIT

---

## Links

| Resource | URL |
|----------|-----|
| GitHub Repository | [github.com/ailinter/ailinter](https://github.com/ailinter/ailinter) |
| Issues / Bug Reports | [github.com/ailinter/ailinter/issues](https://github.com/ailinter/ailinter/issues) |
| VS Code Extension | [marketplace.visualstudio.com/items?itemName=ailinter.ailinter](https://marketplace.visualstudio.com/items?itemName=ailinter.ailinter) |
| Docker Hub | [hub.docker.com/r/ailinter/ailinter](https://hub.docker.com/r/ailinter/ailinter) |
| Documentation | [ailinter.dev/docs](https://ailinter.dev/docs) |
| Homebrew Tap | [github.com/ailinter/homebrew-ailinter](https://github.com/ailinter/homebrew-ailinter) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| License | [LICENSE](LICENSE) |

---

## Development

```bash
make build       # Build to bin/ailinter
make test        # Run tests
make test-cover  # Tests with coverage (80%+ line, 70%+ function)
make lint        # go vet + staticcheck
make release     # Cross-platform binaries (5 targets)
```

---

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code standards, and the contribution workflow. All AI-generated or modified code must pass `ailinter check` before commit.

**Small PRs, high quality.**

---

## License

[MIT](LICENSE) — open source, forever.

Built on open source: [gitleaks](https://github.com/gitleaks/gitleaks) (MIT), [betterleaks](https://github.com/betterleaks/betterleaks) (MIT), [mcp-go](https://github.com/mark3labs/mcp-go) (MIT), [cobra](https://github.com/spf13/cobra) (Apache-2.0).

Code smell definitions adapted from [Samman Coaching Reference](https://sammancoaching.org/reference/code_smells/index.html) by Emily Bache, CC BY-SA 4.0.
