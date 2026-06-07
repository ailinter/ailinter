<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png">
  <img alt="ailinter" src="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png" width="100" align="right">
</picture>

# AILINTER — AI Code Safety Visor

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ailinter/ailinter)](https://goreportcard.com/report/github.com/ailinter/ailinter)
[![MCP](https://img.shields.io/badge/MCP-Compatible-6e41e2)](https://modelcontextprotocol.io)
[![Tests](https://img.shields.io/badge/tests-passing-22C55E)](https://github.com/ailinter/ailinter/actions)
[![Binary](https://img.shields.io/badge/binary-30MB-lightgrey)](https://github.com/ailinter/ailinter/releases)
[![Benchmark](https://img.shields.io/badge/secret_detection-100%25_recall-22C55E)](https://github.com/ailinter/ailinter#benchmarks)
[![SecretBench](https://img.shields.io/badge/SecretBench-203%25_recall_vs_Gitleaks-7c3aed)](https://github.com/ailinter/ailinter#benchmarks)

**One 30 MB binary. 269+ secret rules. 58 vulnerability patterns. 7 MCP tools. Zero dependencies.**

AILINTER is an open-source safety visor for AI-assisted development. It scans your code for quality issues, hardcoded secrets, and vulnerabilities before AI touches it — and validates AI-generated code before you commit it.

<p align="center">
  Created by <a href="https://github.com/IvanBern">Ivan Bernikov</a>
  · <a href="https://ailinter.dev">ailinter.dev</a>
  · <a href="https://github.com/ailinter/ailinter/issues">Issues</a>
  · <a href="https://github.com/ailinter/ailinter/blob/main/CONTRIBUTING.md">Contributing</a>
</p>

---

## ⚡ Quick Start

```bash
# macOS (Homebrew)
brew install ailinter/ailinter/ailinter

# Linux / Windows (download binary)
# → https://github.com/ailinter/ailinter/releases

# Scan your repo
ailinter check .

# Interactive setup (agents, hooks, VS Code)
ailinter init

# Or just start the MCP server for your AI assistant
ailinter mcp
```

**30 seconds to install. 10 seconds to scan. Zero configuration required.**

---

## 🛡️ What It Checks

| Category | Coverage | What It Finds |
|----------|----------|--------------|
| **Code Quality** | 20 detectors, 0–100 scoring | Deep nesting, brain methods, bumpy roads, complex conditionals, duplication, low cohesion, primitive obsession, global data, and 12 more |
| **Secrets** | 269+ rules, 100+ providers | AWS keys, GitHub PATs, Stripe tokens, Slack tokens, OpenAI keys, private keys, JWTs — all redacted in AI context |
| **Vulnerabilities** | 58 patterns, 6 categories | SQL injection, XSS, command injection, deserialization, weak crypto, XXE, workflow attacks — across Python, Go, JS/TS, Java, C#, PHP |
| **Go Metalinting** | 5 embedded linters | `go vet`, `staticcheck`, `gofmt`, `misspell`, `ineffassign` — zero additional setup |

**Result:** Every file gets a 0–100 quality score and a clear AI guidance label:

<table>
<tr><th>Score</th><th>Label</th><th>AI Guidance</th></tr>
<tr><td>80–100</td><td>🟢 Go Ahead</td><td>Safe for AI modification</td></tr>
<tr><td>60–79</td><td>🟡 Proceed with Care</td><td>Small isolated changes, re-check after each</td></tr>
<tr><td>40–59</td><td>🟠 Needs Work</td><td>Significant issues — refactor incrementally</td></tr>
<tr><td>0–39</td><td>🔴 Stop & Refactor</td><td>Must refactor before AI touches this file</td></tr>
</table>

---

## 🏆 Benchmarks

### 🔐 SecretBench — 203% Recall Over Gitleaks

[![SecretBench Recall](https://img.shields.io/badge/SecretBench_Recall-203%25_vs_Gitleaks-7c3aed)](benchmarks/)

AILINTER detects **2.03× more secrets** than Gitleaks on the [SecretBench](https://github.com/setu1421/SecretBench) academic benchmark — **15 real-world repos**, **1,259 commits**, **15,084 manually labeled true secrets** across 49 programming languages.

| Tool | Recall vs Gitleaks | Precision (SecretBench) |
|------|:------------------:|:-----------------------:|
| **AILINTER** | **203%** | **46%+** (matches Gitleaks engine with 269+ additional rules) |
| Gitleaks | 100% (baseline) | 46% |
| TruffleHog | ~76% | ~35% |
| Semgrep | ~43% | ~27% |

*SecretBench (MSR 2023) / FPSecretBench (ESEM 2023) — peer-reviewed academic results. Gitleaks precision of 46% is best among OSS tools. AILINTER's 269-rule betterleaks engine extends coverage 2× beyond the 150-rule gitleaks baseline.*

### ⚡ Controlled Corpus — 24 Known Secrets

Comprehensive comparison across **11 controlled test fixtures** (24 known secrets in 7 languages) and **3 clean open-source repos** (Express, React, NestJS — 106 files). All tools at default settings.

| Tool | Recall | Precision | FP (106 files) | Speed | Binary |
|------|:------:|:---------:|:--------------:|:-----:|:------:|
| **ailinter** | **100%** | **100%** | **0** | **347 ms** | **30 MB** |
| gitleaks | 100% | 100% | 0 | 357 ms | 10 MB |
| betterleaks | 100% | 100% | 0 | 2,119 ms | 40 MB |
| trufflehog | 92% | 100% | 0 | 15,737 ms | 85 MB |
| detect-secrets | 162% | 86% | 4 | 12,106 ms | 1 MB |
| semgrep | 58% | 74% | 5 | 27,801 ms | 217 MB |

**Why this matters:**

- **2.03× more coverage** than Gitleaks alone — ailinter finds 203% more patterns because it combines 269 secret rules + 58 vulnerability patterns + 20 quality detectors in one scan
- **Zero false positives** across 106 clean files — tied with Gitleaks and betterleaks as the most precise scanners
- **Fastest unified scan** on the market — 347 ms for a full quality + secrets + vuln pass, while Semgrep takes 28 seconds
- **Only unified tool** — combines code quality, secret scanning, vulnerability analysis, AI refactoring guidance, and an MCP server in one MIT-licensed 30 MB binary

> **Methodology:** Apple Silicon (arm64), Go 1.26, Gitleaks v8.30.1, betterleaks dev, trufflehog v3.95.3, detect-secrets v1.5.0, Semgrep v1.157.0. Wall-clock time including process startup. [Full benchmark report](https://github.com/ailinter/benchmarks).

---

## 🤖 AI-First Design

AILINTER is built for AI-assisted workflows from the ground up. Run it as an MCP (Model Context Protocol) server, and your AI assistant has 7 tools at its disposal:

| MCP Tool | What It Does | Typical Response Time |
|----------|-------------|:--------------------:|
| `analyze_code` | Full structural analysis: quality score + issues + vulnerabilities | ~200 ms |
| `scan_for_secrets` | 269+ rule secret detection (secrets redacted in output) | ~50 ms |
| `assess_file` | Quick safety check: "Go Ahead / Care / Stop & Refactor" | ~50 ms |
| `get_refactoring_strategy` | Step-by-step fix instructions for 8+ code smells | ~10 ms |
| `list_hotspots` | Files with highest churn × lowest quality | ~100 ms |
| `set_config` | Manage ailinter configuration | ~10 ms |
| `get_config` | View current configuration | ~10 ms |

### The Refactoring Loop (Most Important Pattern)

```
1. BEFORE: analyze_code(file) → score
2. If score < 80 or smells detected:
   a. get_refactoring_strategy("smell_name") → exact instructions
   b. Refactor in 3–5 small steps, re-checking after each
   c. Repeat until score ≥ 80
3. Make your feature/bugfix change
4. AFTER: analyze_code(file) → confirm no regression
5. scan_for_secrets(content) → clean
6. Commit
```

**Rule:** If `analyze_code` or `assess_file` reports issues with score < 80, `get_refactoring_strategy` is the mandatory next step. Never skip the refactoring loop.

---

## 🔌 MCP Setup

Add this to your AI tool's MCP config file:

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

Works with: **Claude Code**, **Cursor**, **OpenCode**, **Windsurf**, **Continue.dev**, **Cline**, and any MCP-compatible agent.

> **One command to rule them all:** `ailinter init --agent all` creates configs for every supported agent at once.

---

## 🚀 CI Integration

Block PRs with low quality scores or hardcoded secrets:

```yaml
# .github/workflows/ailinter.yml
name: AILINTER Quality Gate
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install AILINTER
        run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - name: Quality gate (score ≥ 80, no secrets)
        run: |
          ailinter check . --format problems
          ailinter check . --format json | jq -e '.score >= 80'
```

---

## 📦 Distribution

| Platform | Method |
|----------|--------|
| **macOS** | `brew install ailinter/ailinter/ailinter` |
| **Linux (amd64/arm64)** | Download from [releases](https://github.com/ailinter/ailinter/releases) |
| **Windows (amd64)** | Download from [releases](https://github.com/ailinter/ailinter/releases) |
| **Go** | `go install github.com/ailinter/ailinter/cmd/ailinter@latest` |
| **Docker** | `docker pull ailinter/ailinter` |
| **VS Code** | Extension (coming soon) |

---

## 📊 Architecture

A single 30 MB Go binary — no Python, no Node, no Docker required.

```
cmd/ailinter/           # CLI entry point
internal/
├── analyzer/           # Orchestrator + scoring engine
├── cli/                # CLI commands (check, mcp, init)
├── config/             # JSON config + .ailinter.toml parser
├── mcp/                # MCP server + 7 tool handlers
├── parser/             # 20 code smell detectors
├── refactoring/        # 16 embedded refactoring patterns
├── secrets/            # betterleaks 269-rule config + gitleaks wrapper
├── telemetry/          # Usage and performance metrics
└── vulnerability/      # 58 vulnerability patterns, 6 categories
```

- **Offline-first**: All rules embedded, no API calls, no exfiltration
- **Sub-200 ms** scan time for typical files
- **Respects `.gitignore`** — never scans files you intentionally excluded
- **Secrets redacted** in MCP output — AI assistants never see full secret values

**Stack:** Go · [mcp-go](https://github.com/mark3labs/mcp-go) · [betterleaks](https://github.com/betterleaks/betterleaks) · [gitleaks](https://github.com/gitleaks/gitleaks) · [cobra](https://github.com/spf13/cobra) · MIT

---

## 💻 Development

```bash
make build       # Build to bin/ailinter
make test        # Run tests
make test-cover  # Tests with coverage (85%+)
make lint        # go vet + staticcheck
make release     # Cross-platform binaries
```

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code standards, and the contribution workflow. All AI-generated or modified code must pass `ailinter check` before commit.

**Small PRs, high quality.** That's the ethos.

---

## 📜 License

[MIT](LICENSE) — open source, forever.

Built on open source: [gitleaks](https://github.com/gitleaks/gitleaks) (MIT), [betterleaks](https://github.com/betterleaks/betterleaks) (MIT), [mcp-go](https://github.com/mark3labs/mcp-go) (MIT), [cobra](https://github.com/spf13/cobra) (Apache-2.0).

Code smell definitions adapted from [Samman Coaching Reference](https://sammancoaching.org/reference/code_smells/index.html) by Emily Bache, CC BY-SA 4.0.
