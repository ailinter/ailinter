# ailinter Roadmap

**Open-source AI linter and safety visor for AI-assisted development**

> *Last updated: July 2026*

This roadmap describes the current state of ailinter and the planned evolution over the next 12 months. It reflects the priorities of the core maintainers and the community. Milestones and timelines are best-effort targets, not guarantees — priorities shift based on user feedback and contributor availability.

---

## Current State (July 2026)

### ✅ What's Shipped

| Capability | Details |
|---|---|
| **CLI** | Zero-config Go binary — `ailinter check .` scans code quality, secrets, and vulnerabilities in one command |
| **Code Quality Radar** | 20 detectors across 13 languages — deep nesting, brain method, god class, long parameter list, primitive obsession, duplicated code, complex conditional, and more |
| **Quality Scoring** | 0–100 per-file score with AI-guidance labels (Go Ahead / Proceed with Care / Needs Work / Stop & Refactor) |
| **Secret Scanning** | 269+ betterleaks rules + 150 gitleaks fallback rules covering 100+ providers (AWS, Azure, GCP, OpenAI, Stripe, GitHub, SSH keys, and more). Secrets are redacted in MCP output. |
| **Vulnerability Detection** | 58 patterns across 6 categories (Injection, XSS, Deserialization, Weak Crypto, XXE, Workflow Attacks) covering Python, Go, JS/TS, Java, C#, PHP |
| **Refactoring Strategies** | 24 embedded refactoring patterns covering 19 code smells — actionable before/after code examples for AI agents |
| **MCP Server** | 7 Model Context Protocol tools (`analyze_code`, `assess_file`, `scan_for_secrets`, `get_refactoring_strategy`, `list_hotspots`, `get_config`, `set_config`) — compatible with Claude, Cursor, Cline, OpenCode, Windsurf, Continue.dev, Copilot |
| **VS Code Extension** | Live on the marketplace — inline diagnostics, status bar score, CodeLens annotations, Delta Dashboard for code health tracking over time |
| **Install Methods** | Homebrew tap, Go install, Docker, prebuilt binaries (5 platforms: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64) |
| **CI/CD Integration** | GitHub Actions workflow, SARIF v2.1.0 output for GitHub Code Scanning, pre-commit hook with `go vet` + `staticcheck` + `gofmt` + `misspell` + `ineffassign` |
| **Documentation** | README, AGENTS.md, CONTRIBUTING.md, CHANGELOG.md, CODE_OF_CONDUCT.md, SECURITY.md, ailinter.dev/docs |
| **Go Metalinters** | Embedded `go vet`, `staticcheck`, `gofmt`, `misspell`, `ineffassign` — zero extra setup |

### 📊 Key Metrics

- **⭐ GitHub Stars:** ~8
- **👥 Contributors:** 3
- **📦 Releases:** 8 (v1.0.0 shipped June 2026)
- **✅ Commits:** 88
- **🔧 Open Issues:** 0
- **🖥️ VS Code Installs:** Growing

### 🔮 What's Coming Next

This roadmap is organized into three time horizons.

---

## Short-Term Goals (Q3 2026) — July–September 2026

### Python Support Expansion

- **Python vulnerability scanner**: Expand the 58-pattern vulnerability detector with Python-specific rules (Django/Flask SSRF, SQLAlchemy injection, Pickle deserialization, Jinja2 SSTI, `yaml.load()` unsafe patterns)
- **Python code quality detectors**: Add Python-specific smell variants (PEP 8 violations, unused imports, overly broad exception handlers)
- **Python package scanning**: Scan `requirements.txt` / `pyproject.toml` for known-vulnerable dependencies via an embedded advisory database

### CI/CD Integration Depth

- **GitHub Actions marketplace action**: Publish a reusable `ailinter/action` for one-step scan in any workflow
- **GitLab CI template**: Provide a ready-to-use `.gitlab-ci.yml` include
- **Pre-commit hooks repository**: Add ailinter to the pre-commit hooks registry (`pre-commit-hooks.yaml`)
- **Merge gate enforcement**: Document patterns for blocking PRs on quality regressions (branch protection rules + status checks)

### Community & Outreach

- **Blog post series**: "Why AI code needs a safety visor" — publish on ailinter.dev and cross-post to Dev.to, Medium
- **HN / Reddit launch**: Submit to Hacker News, r/programming, r/devtools, r/golang
- **Good first issues**: Label and triage 10+ beginner-friendly issues for new contributors
- **Discord server**: Launch a community Discord for users, contributors, and discussions

### Platform Polish

- **Telemetry opt-in dashboard**: Implement optional usage insights dashboard for maintainers (fully opt-in, no data by default)
- **Performance profiling**: Benchmark and optimize scan times — target <50ms per file for typical Go/Python files
- **Configuration file schema**: Publish JSON Schema for `.ailinter.toml` to enable IDE autocompletion

---

## Mid-Term Goals (Q4 2026) — October–December 2026

### Multi-Language Expansion

- **JavaScript / TypeScript support**: 
  - Vulnerability patterns for Node.js (path traversal, prototype pollution, unsafe `eval()`, NoSQL injection)
  - Code quality detectors for JS/TS (callback hell detection, async/await misuse, excessive `any` usage)
  - ESLint-compatible output format for teams migrating from ESLint
- **Rust support**:
  - Safety detectors for unsafe blocks, `unwrap()` proliferation, memory safety patterns
  - Integration points with `cargo clippy` and `cargo audit`
  - Refactoring strategies for common Rust code smells

### IDE Ecosystem Expansion

- **JetBrains plugin**: Publish ailinter plugin for IntelliJ IDEA, GoLand, PyCharm, WebStorm — live inline diagnostics and on-save scanning
- **Neovim / Vim plugin**: Native Lua plugin with diagnostic integration via `vim.diagnostics`
- **VS Code extension v2**: 
  - Inline fix suggestions (quick-fix lightbulb actions)
  - Refactoring preview pane showing before/after diff
  - Per-user configuration UI in Settings panel

### Performance & Scalability

- **Parallel file scanning** (Goroutine pool): Use all available cores for multi-file scans — target 10x throughput improvement for large monorepos
- **Incremental scanning**: Track file modification times and only re-scan changed files — integrate with `git diff` for CI speedups
- **LRU result cache**: Cache scan results for unchanged files during interactive usage
- **Memory optimization**: Profile and reduce heap allocations during scanning — target <50MB RSS for typical projects

### MCP Server v2

- **Context-aware analysis**: Pass file relationships and imports to enable cross-file smell detection
- **Batch analysis tool**: New MCP tool `analyze_batch` for scanning multiple files in a single call
- **Custom threshold tool**: Let users configure quality thresholds per-file via MCP

---

## Long-Term Vision (H1 2027) — January–June 2027

### Custom Rule Engine

- **User-defined rules**: DSL for writing custom code quality rules (e.g., "flag any function with more than 3 database calls" or "detect deprecated package usage")
- **Rule marketplace**: Community-contributed rule repository where users can publish and discover rules
- **Semantic-aware analysis**: AST-based pattern matching beyond regex — detect structural code issues with context awareness
- **Multi-file rules**: Rules that analyze relationships across files (e.g., "every handler in `routes/` must have a corresponding test in `tests/`")

### Team Collaboration Features

- **Quality dashboards**: Per-team and per-project quality trend charts with historical data
- **Baseline management**: Track quality baselines per branch and alert on regressions across PR history
- **Review annotations**: Allow team leads to annotate flagged issues with context (false positive, accepted risk, triaged)
- **Policy-as-code**: Team-level configuration files that enforce minimum quality standards across all contributors

### Enterprise Readiness

- **SAML / SSO integration**: Enterprise single sign-on for hosted dashboard and team management
- **RBAC**: Role-based access control for team settings, rule configuration, and policy management
- **Audit logging**: Full audit trail of all scans, config changes, and policy violations
- **On-premise deployment**: Single-tenant deployment option with Docker Compose or Kubernetes helm chart
- **SLA monitoring**: Built-in health checks and uptime monitoring for enterprise deployments

### Platform Expansion

- **GitHub App**: Installable GitHub App that automatically scans PRs and posts quality reports as check runs and PR comments
- **GitLab integration**: Native GitLab MR scanning with inline comments on findings
- **Pre-commit cloud**: Optional SaaS tier for teams that want centralized configuration and dashboards without self-hosting
- **OpenTelemetry integration**: Export scan metrics to existing observability stacks (Datadog, Grafana, New Relic)

### AI Agent Ecosystem

- **Claude Code native plugin**: Tailored integration for Claude Code CLI with structured output and guided refactoring
- **OpenAI Codex CLI support**: Native Codex CLI integration for end-to-end AI safety
- **Multi-agent orchestration**: Reference architecture for teams running multiple AI coding agents with unified ailinter enforcement
- **Continuous improvement loop**: Automated feedback from scan results back into detection rules (LLM-assisted rule refinement)

---

## Community Goals

We aim to hit the following community milestones over the next 12 months:

| Goal | Target | Timeline |
|---|---|---|
| **GitHub Stars** | 100+ | H1 2027 |
| **Contributors** | 10+ active | Q4 2026 |
| **Integrations** | 5+ CI/CD + IDE plugins | Q4 2026 |
| **Open Issues** | Healthy triage (avg response <48h) | Ongoing |
| **Discord Members** | 200+ | H1 2027 |
| **GitHub Sponsors** | 5+ recurring sponsors | H1 2027 |
| **Adoption** | 1,000+ monthly Docker pulls | Q4 2026 |

### How You Can Help

- ⭐ **Star the repo** — [github.com/ailinter/ailinter](https://github.com/ailinter/ailinter)
- 🐛 **Report bugs** — open an issue with reproduction steps
- 📝 **Write docs** — improve installation guides, add language-specific examples
- 🔧 **Contribute code** — see [CONTRIBUTING.md](CONTRIBUTING.md) and look for `good first issue` labels
- 💬 **Join the discussion** — Discord link coming soon
- 🐦 **Follow us** — [@ailinter_dev](https://x.com/ailinter_dev)

---

## How This Roadmap Is Maintained

This roadmap is a living document. It is reviewed and updated quarterly by the core maintainers based on:

- User feedback and feature requests (GitHub issues, Discord, Twitter)
- Community contribution velocity and interest areas
- Shifts in the AI-assisted development ecosystem
- Maintainer bandwidth and sponsorship funding

**Propose changes** by opening a PR or discussion on GitHub.

---

*Built with ❤️ by [Ivan Bernikov](https://github.com/IvanBern) and contributors. MIT licensed.*
