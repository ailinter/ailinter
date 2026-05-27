<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png">
  <img alt="ailinter" src="https://raw.githubusercontent.com/ailinter/ops/main/branding/logo/icon.png" width="100" align="right">
</picture>

# ailinter

### AI Code. Human Standards.

> The open-source safety visor for AI-assisted development. Scans code quality, secrets, and vulnerabilities before and after every AI edit — directly in your editor.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-Compatible-6e41e2)](https://modelcontextprotocol.io)
[![Tests](https://img.shields.io/badge/tests-passing-22C55E)](https://github.com/ailinter/ailinter/actions)
[![Coverage](https://img.shields.io/badge/coverage-85%25-22C55E)](https://github.com/ailinter/ailinter)
[![Binary Size](https://img.shields.io/badge/binary-15MB-lightgrey)](https://github.com/ailinter/ailinter/releases)

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
# Install
go install github.com/ailinter/ailinter/cmd/ailinter@latest

# Scan a file
ailinter check src/main.go

# Initialize a project
ailinter init

# Start MCP server
ailinter mcp
```

### Add to Your AI Assistant

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

> Config: OpenCode `opencode.json` · Claude `.claude/mcp.json` · Cursor `.cursor/mcp.json` · VS Code `.vscode/mcp.json`

---

## The Quality Score

Every file gets a **0–100 score** that tells AI assistants whether it's safe to modify:

| Score | Label | AI Guidance |
|-------|-------|-------------|
| **95–100** | Go Ahead | Safe for AI modification |
| **75–94** | Proceed with Care | Use small changes, re-check after each edit |
| **0–74** | Stop & Refactor | Refactor BEFORE AI touches this file |

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
| Binary Size | **15 MB** | ~400 MB Docker | ~10 MB |
| Dependencies | **Zero** | Docker + JVM | Zero |
| License | **MIT** | LGPL+Proprietary | MIT |

---

## MCP Tools (7)

| Tool | Purpose |
|------|---------|
| `analyze_code` | Full structural analysis + vulnerability detection with 0–100 score |
| `scan_for_secrets` | 269-rule secret detection (secrets redacted in output) |
| `get_refactoring_strategy` | Exact step-by-step refactoring instructions with before/after examples |
| `assess_file` | Quick classification: Go Ahead / Proceed with Care / Stop & Refactor |
| `list_hotspots` | Frequently-changed files with low quality scores |
| `set_config` | Set persistent configuration |
| `get_config` | View current configuration |

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

---

## CLI

```bash
ailinter check src/main.go               # Single file (auto format)
ailinter check .                          # Directory scan
ailinter check --format json app.py       # JSON output
ailinter check --format markdown app.py   # LLM-friendly
ailinter check --format problems app.py   # GCC-style (IDE integration)
ailinter check --no-secrets app.py        # Skip secrets
ailinter check --no-vulnerabilities app.py # Skip vulnerabilities
ailinter check --secrets-only app.py      # Secrets only
ailinter check --vulnerabilities-only app.py # Vulns only
ailinter check --lang python script.py    # Force language
ailinter check --no-gitignore .           # Don't respect .gitignore
ailinter init                             # Create .ailinter.toml + AGENTS.md
ailinter mcp                              # Start MCP server
ailinter rules list                       # List all threshold defaults
ailinter rules list --lang python         # Filter by language
```

---

## Benchmarks

### Secret Detection — 31-File Multi-Language Corpus

| Tool | Rules | Secrets Found | False Positives | Speed |
|------|:---:|:---:|:---:|------:|
| **ailinter** | **269** | **71** | 0 on React/NestJS | 186ms |
| gitleaks | 150 | 35 (baseline) | 0 on all repos | 196ms |
| trufflehog | 800+ | 15 | 0 on all repos | 11,292ms |

---

## Architecture

```
cmd/ailinter/           # CLI entry point
internal/
├── analyzer/           # Orchestrator + scoring engine
├── cli/                # CLI commands (check, mcp, init)
├── config/             # JSON config + .ailinter.toml parser
├── mcp/                # MCP server + 7 tool handlers
├── parser/             # 17 code smell detectors
├── refactoring/        # 16 embedded refactoring patterns
├── secrets/            # betterleaks 269-rule config + gitleaks wrapper
├── telemetry/          # Usage and performance metrics
└── vulnerability/      # 58 vulnerability patterns, 6 categories
```

**Stack:** Go · [mcp-go](https://github.com/mark3labs/mcp-go) · [betterleaks](https://github.com/betterleaks/betterleaks) · [gitleaks](https://github.com/gitleaks/gitleaks) · [cobra](https://github.com/spf13/cobra) · MIT

**Build:** `make build` → 15 MB binary, zero runtime dependencies. Targets: darwin, linux, windows (amd64 + arm64).

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
