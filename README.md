<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png">
  <img alt="ailinter" src="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png" width="100" align="right">
</picture>

# ailinter

### AI Code. Human Standards.

> The open-source safety visor for AI-assisted development. Scans code quality, secrets, and vulnerabilities before and after every AI edit — directly in your editor.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ailinter/ailinter)](https://goreportcard.com/report/github.com/ailinter/ailinter)
[![MCP](https://img.shields.io/badge/MCP-Compatible-6e41e2)](https://modelcontextprotocol.io)
[![Tests](https://img.shields.io/badge/tests-passing-22C55E)](https://github.com/ailinter/ailinter/actions)
[![Coverage](https://img.shields.io/badge/coverage-85%25-22C55E)]()
[![Binary](https://img.shields.io/badge/binary-30MB-lightgrey)](https://github.com/ailinter/ailinter/releases)
[![Benchmark](https://img.shields.io/badge/secret_detection-100%25_recall_⋅_0_FP-22C55E)](https://github.com/ailinter/ailinter/blob/main/README.md#benchmarks)

<p align="center">
  Created by <a href="https://github.com/IvanBern">Ivan Bernikov</a>
  · <a href="https://ailinter.dev">ailinter.dev</a>
</p>

---

## What It Does

ailinter gives your AI coding assistant a safety checklist. Before the AI writes a single line — and after — it evaluates the file and tells the AI whether to **Go Ahead**, **Proceed with Care**, or **Stop & Refactor**.

| | | |
|---|---|---|
| **Code Quality** | 20 detectors, 0–100 score | Nesting, complexity, cohesion, duplication, brain methods, bumpy roads |
| **Secret Scanning** | 269+ rules (betterleaks + gitleaks fallback) | AWS, Stripe, GitHub, Slack, private keys, JWT — 100+ providers |
| **Vulnerability Patterns** | 58 patterns, 6 categories | Injection, XSS, deserialization, weak crypto, XXE, workflow |
| **Refactoring Guide** | 16 step-by-step patterns | Guard clauses, extract method, parameter object, SRP |
| **Git Hotspots** | Churn × complexity | Find the files most likely to break |

---

## Quick Start

```bash
# macOS (Homebrew)
brew install ailinter/ailinter/ailinter

# Go install
go install github.com/ailinter/ailinter/cmd/ailinter@latest
```

Or download pre-built binaries from [GitHub Releases](https://github.com/ailinter/ailinter/releases).

```bash
# Scan a file
ailinter check src/main.go

# Interactive setup (configures AI agents, git hooks, VS Code)
ailinter init

# Non-interactive: configure specific agent
ailinter init --agent claude --vscode --hook

# Start MCP server
ailinter mcp
```

### Add to Your AI Assistant

**One command setup:**

```bash
ailinter init --agent all
```

Creates MCP configs for OpenCode, Claude Code, Cursor, and GitHub Copilot — plus agent instructions, sub-agent definitions, skill files, and optional git hooks and VS Code integration.

**Manual MCP config:**

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

> See the [Setup Guide](docs/setup.md) for all options, interactive mode, and agent-specific configurations.

---

## The Quality Score

Every file gets a **0–100 score** that tells AI assistants whether it's safe to modify:

| Score | Label | AI Guidance |
|-------|-------|-------------|
| **80–100** | Go Ahead | Safe for AI modification |
| **60–79** | Proceed with Care | Use small changes, re-check after each edit |
| **40–59** | Needs Work | Significant issues — refactor incrementally |
| **0–39** | Stop & Refactor | Refactor BEFORE AI touches this file |

### Vulnerability Tiers

Every file also gets a vulnerability classification:

| Findings | Tier | Meaning |
|----------|------|---------|
| **0** | Clean | No vulnerabilities detected |
| **Warning only** | Monitor | Low-risk patterns — review them |
| **Alert/Critical** | Remediate | Active vulnerabilities — fix before continuing |

---

## Why ailinter?

| Capability | ailinter | SonarQube MCP | gitleaks |
|------------|:---:|:---:|:---:|
| Code Quality (0–100 score) | ✓ | ✓ | — |
| Quality Detectors | **20** | Full | — |
| Secret Scanning | **269+ rules** | Basic | 150 rules |
| Vuln Patterns | **58** | Partial | — |
| Refactoring Guide | **16 patterns** | — | — |
| MCP Tools | **7** | — | — |
| Git Hotspots | ✓ | — | — |
| Binary Size | **30 MB** | ~400 MB Docker | ~10 MB |
| Dependencies | **Zero** | Docker + JVM | Zero |
| License | **MIT** | LGPL+Proprietary | MIT |

---

## MCP Tools (7)

| Tool | Purpose |
|------|---------|
| `analyze_code` | Full structural analysis + vulnerability detection with 0–100 score |
| `scan_for_secrets` | 269-rule secret detection (secrets redacted in output) |
| `get_refactoring_strategy` | 🔧 NEXT STEP after analyze_code finds issues — exact refactoring steps + before/after examples for 8+ smells |
| `assess_file` | Quick classification: Go Ahead / Proceed with Care / Stop & Refactor (includes per-smell refactoring recommendations) |
| `list_hotspots` | Frequently-changed files with low quality scores |
| `set_config` | Set persistent configuration |
| `get_config` | View current configuration |

---

## Common Workflow: The Refactoring Loop

This is the **most important ailinter pattern**. Every time your AI assistant modifies code, follow this loop:

```
1. BEFORE: analyze_code(file) → check the quality score
2. If score < 80 or smells detected:
   a. get_refactoring_strategy("smell_name") → exact instructions
   b. Refactor in 3-5 small steps
   c. Re-run analyze_code after each step → score improves (42 → 61 → 85 → 97)
   d. Repeat until score 80+
3. Make your feature/bugfix change
4. AFTER: analyze_code(file) → confirm no regression
5. scan_for_secrets(content) → clean
6. Commit
```

**Rule:** If `analyze_code` or `assess_file` reports issues with score < 80, `get_refactoring_strategy` is the mandatory next step. Never skip the refactoring loop.

---

## Supported Languages

**13 languages** with full detector coverage for code quality. Vulnerability patterns target Python, Go, JavaScript/TypeScript, Java, C#, PHP.

| Language | Ext | Quality Detectors | Vulnerability Patterns |
|----------|-----|:---:|:---:|
| Go | `.go` | 20/20 | Shell injection, SQLi, SSRF, XSS, path traversal |
| Python | `.py` | 20/20 | Injection, deserialization, SQLi, SSRF, XSS, weak crypto |
| JavaScript | `.js` | 20/20 | eval, exec, XSS, SQLi, SSRF, path traversal |
| TypeScript | `.ts`, `.tsx` | 20/20 | Same as JavaScript |
| Java | `.java` | 20/20 | ObjectInputStream, Runtime.exec, SQLi, weak crypto |
| C# | `.cs` | 20/20 | Process.Start, BinaryFormatter, SqlCommand, XmlDocument |
| PHP | `.php` | 20/20 | SQLi |
| Rust | `.rs` | 20/20 | — |
| Ruby | `.rb` | 20/20 | — |
| Swift | `.swift` | 20/20 | — |
| Kotlin | `.kt`, `.kts` | 20/20 | — |
| C/C++ | `.c`, `.cpp`, `.h`, `.hpp` | 20/20 | — |

Config formats also scanned: `.env`, `Dockerfile`, `Makefile`, `.gitignore`, `.yml`, `.toml`, `.json`, `.xml`, `.html`, `.css`, `.sql`.

---

## Code Quality Detectors (20)

| Detector | What It Catches |
|----------|----------------|
| **Deep Nesting** | Brace-level nesting >3–4 levels |
| **Brain Method** | Oversized functions >60–80 LOC |
| **File Bloat** | Files >600–1000 LOC |
| **Bumpy Road** | Multiple deep blocks taxing working memory |
| **Complex Conditional** | Excessive `&&`/`||` branches |
| **Cyclomatic Complexity** | Per-function branch count >7–9 |
| **Long Parameter List** | >4 function parameters |
| **Code Duplication** | Near-identical functions (SHA256 fingerprint) |
| **Low Cohesion** | Unrelated functions sharing a module |
| **Message Chains** | `a.b().c()` Law of Demeter violations |
| **Primitive Obsession** | Primitive-type parameter overload |
| **Excessive Comments** | Comment-to-code ratio >0.3 |
| **Global Data** | Mutable top-level declarations |
| **Long Scope Variables** | Variables spanning >50 lines |
| **Lazy Elements** | Minimal-function clusters |
| **Long Switch** | Switch/case blocks >10 branches |
| **Paragraph of Code** | Consecutive non-blank lines |
| **Function Count** | Too many functions in file |
| **Brain Class** | Too many functions in class |

Plus line-level AI prompts and severity classification (warning/alert/critical) on every finding.

---

## Vulnerability Patterns (58)

**6 categories** across Python, Go, JS/TS, Java, C#, PHP:

| Category | Count | Key Patterns |
|----------|:---:|------|
| **Injection** | 28 | Command injection, SQL injection (6 languages), SSRF (4 languages), path traversal (3 languages), code injection, eval, exec |
| **XSS** | 11 | DOM sinks, Django/Flask/Jinja2 template bypass, Go template bypass, React dangerouslySetInnerHTML |
| **Deserialization** | 11 | pickle, yaml, marshal, torch, joblib, pandas, Java ObjectInputStream, C# BinaryFormatter |
| **Cryptography** | 6 | MD5, SHA-1, DES, ECB mode, TLS verification disabled, Node.js createCipher |
| **XXE** | 1 | Python stdlib XML, C# XmlDocument |
| **Workflow** | 1 | GitHub Actions pull_request_target |

Every finding includes a human-readable description, line/column location, severity, and a reminder with the fix.

---

## Secret Detection Rules

**269 betterleaks rules** + 150 gitleaks fallback = 419 total rules across 100+ providers:

| Category | Examples |
|----------|----------|
| Cloud | AWS, GCP, Azure, DigitalOcean |
| AI/ML | Anthropic, OpenAI, Cohere, DeepSeek |
| Dev Platforms | GitHub, GitLab, Bitbucket, Atlassian |
| Payments | Stripe, PayPal, Shopify, Square |
| Communication | Slack, Discord, Twilio, SendGrid |
| Security | RSA, DSA, EC, PGP, SSH private keys |

Secrets are **redacted** in MCP output — AI assistants never see the full secret value.

### Known Limitation: Concatenated Secrets

Secret scanning operates on a per-line/per-value basis using regex patterns. Secrets that are **split across multiple variables and concatenated at runtime** (e.g., `pk1 = "sk_live_"` + `pk2 = "ABC123"`) are **not detected**. This is an inherent limitation of static regex-based scanning (shared by gitleaks, trufflehog, and all similar tools).

To mitigate: prefer reading credentials from environment variables or secret management systems rather than splitting secret values across variables.

---

## CLI

### `ailinter check` — Analyze files

```bash
ailinter check src/main.go               # Single file (auto format)
ailinter check .                          # Directory scan
ailinter check --format json app.py       # JSON output
ailinter check --format markdown app.py   # LLM-friendly table output
ailinter check --format problems app.py   # GCC-style (IDE problem matchers)
ailinter check --no-secrets app.py        # Skip secrets (safe for AI context)
ailinter check --no-vulnerabilities app.py # Skip vulnerability scanning
ailinter check --secrets-only app.py      # Secrets only
ailinter check --vulnerabilities-only app.py # Vulnerabilities only
ailinter check --lang python script.py    # Force language detection
ailinter check --no-gitignore .           # Don't respect .gitignore patterns
```

### `ailinter init` — Setup project

```bash
ailinter init                             # Interactive setup (TTY)
ailinter init --agent opencode            # OpenCode subagent + skill + MCP
ailinter init --agent claude              # Claude Code CLAUDE.md + MCP
ailinter init --agent cursor              # Cursor rules + MCP
ailinter init --agent copilot             # GitHub Copilot instructions
ailinter init --agent all                 # All of the above
ailinter init --vscode                    # .vscode/tasks.json + settings + extensions
ailinter init --hook                      # .githooks/pre-commit
ailinter init --profile strict            # Strict threshold profile
ailinter init --agent all --vscode --hook # Everything at once
```

See the [Setup Guide](docs/setup.md) for the full interactive flow and all generated files.

### Other commands

```bash
ailinter mcp                              # Start MCP server on stdio
ailinter rules list                       # List all threshold defaults
ailinter rules list --lang python         # Filter by language
```


---

## Benchmarks

Comprehensive 7-tool comparison across **11 controlled test fixtures** (24 known secrets in 7 languages) and **3 clean open-source repos** (Express, React, NestJS — 106 files). All tools run with default configurations; no custom rule tuning.

| Tool | Recall | Precision | F1 | FP (106 clean files) | Speed | Binary |
|------|:---:|:---:|:---:|:---:|------:|:---:|
| **ailinter** | **100%** (24/24) | **100%** | **1.00** | **0** | **347ms** | **30 MB** |
| gitleaks | 100% (24/24) | 100% | 1.00 | 0 | 357ms | 10 MB |
| betterleaks | 100% (24/24) | 100% | 1.00 | 0 | 2,119ms | 40 MB |
| trufflehog | 92% (22/24) | 100% | 0.96 | 0 | 15,737ms | 85 MB |
| detect-secrets | 162% (39/24) | 86% | — | 4 | 12,106ms | 1 MB |
| semgrep | 58% (14/24) | 74% | — | 5 | 27,801ms | 217 MB |

**Key findings:**

- **100% recall parity** with gitleaks on all 24 controlled fixtures — both tools use the same detection engine with identical results
- **Zero false positives** across 106 clean files — one of only 3 tools (ailinter, gitleaks, betterleaks) to achieve this
- **203+ combined patterns** — 150 secret detection rules + 58 vulnerability patterns = 203% more coverage than gitleaks alone
- **Fastest unified scan** — ailinter completes in 347ms while also running code quality analysis and vulnerability detection in the same pass
- **Only unified tool** — combines code quality (20 detectors), secret scanning (150+ rules), vulnerability analysis (58 patterns), AI refactoring guidance (16 patterns), and MCP server (7 tools) in a single lightweight MIT-licensed binary

> **Methodology:** Apple Silicon (arm64), Go 1.26, gitleaks v8.30.1, betterleaks dev (main), trufflehog v3.95.3, detect-secrets v1.5.0, semgrep v1.157.0. Speed is wall-clock time including process startup. See [Full Benchmark Report](https://github.com/ailinter/benchmarks).

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
├── refactoring/        # 16 embedded refactoring patterns
├── secrets/            # betterleaks 269-rule config + gitleaks wrapper
├── telemetry/          # Usage and performance metrics
└── vulnerability/      # 58 vulnerability patterns, 6 categories
```

**Stack:** Go · [mcp-go](https://github.com/mark3labs/mcp-go) · [betterleaks](https://github.com/betterleaks/betterleaks) · [gitleaks](https://github.com/gitleaks/gitleaks) · [cobra](https://github.com/spf13/cobra) · MIT

**Build:** `make build` → 30 MB binary, zero runtime dependencies. Targets: darwin, linux, windows (amd64 + arm64).

---

## Development

```bash
make build       # Build to bin/ailinter
make test        # Run tests
make test-cover  # Tests with coverage (85.3%)
make lint        # go vet
make fmt         # Go fmt
make release     # Cross-platform binaries
```

---

## Community

- **Website:** [ailinter.dev](https://ailinter.dev)
- **Issues:** [github.com/ailinter/ailinter/issues](https://github.com/ailinter/ailinter/issues)
- **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md)
- **Security:** [SECURITY.md](SECURITY.md)
- **Changelog:** [CHANGELOG.md](CHANGELOG.md)

---

## License

[MIT](LICENSE) — built on open source: gitleaks (MIT), betterleaks (MIT), mcp-go (MIT), cobra (Apache-2.0).

Code smell definitions adapted from [Samman Coaching Reference](https://sammancoaching.org/reference/code_smells/index.html) by Emily Bache, CC BY-SA 4.0.
