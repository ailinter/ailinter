# Changelog

## [Unreleased] — v0.5.0-dev

### Added (Phase 1 — Complete)

#### Code Quality Radar (17 detectors)
- **Deep nesting detection** — brace-counting with warning/alert thresholds
- **Brain Method detection** — per-language function length analysis (Go, C++, Java, Rust, Ruby, Swift, Kotlin, Python)
- **File bloat detection** — God Class risk with 3-tier severity (warn/alert/critical)
- **Bumpy Road detection** — multiple deeply-nested blocks segmentation
- **Complex conditional detection** — `&&`/`||` branch counting in if/while
- **Long parameter list detection** — function signature parameter counting
- **Cyclomatic complexity** — per-function branch counting (CC >= 9 warns)
- **Message chain detection** — Law of Demeter `a.b().c()` patterns
- **Primitive obsession detection** — primitive-type overload in signatures
- **Lazy element detection** — functions shorter than 3 lines
- **Paragraph of code detection** — consecutive non-blank lines > 15
- **Code duplication detection** — SHA256-normalized fingerprints with similarity scoring
- **Cohesion analysis** — shared-type analysis with low cohesion detection
- **Excessive comments detection** — comment-to-code ratio > 30%
- **Global data detection** — mutable top-level declarations
- **Long scope variable detection** — variables declared far from last use
- **Long switch detection** — case count threshold analysis

#### Secret Scanning (Tier 1)
- **269 detection rules** from betterleaks (evolved gitleaks engine) — 2× broader coverage than gitleaks 150-rule default
- Embedded at compile time, falls back to gitleaks default if config parsing fails
- 100+ providers covered: cloud, AI/ML, dev platforms, payments, comms, databases, private keys
- New rule types: Cerebras, DeepSeek, Perplexity, XAI, GitHub fine-grained PAT, Cursor, Databricks, Vercel, etc.
- Supporting config file scanning: `.env`, `.env.*`, `.properties`, `.ini`, `.cfg`, `.conf`, `Dockerfile`, `Makefile`
- Multi-provider coverage: AWS, GCP, Azure, GitHub, GitLab, Stripe, Slack, etc.
- Entropy-based severity classification (critical >= 4.5, alert >= 3.5, warning < 3.5)
- Secret redaction in output (first 4 + last 4 characters)
- AI prompt injection with `os.Getenv` remediation guidance
- Custom rule set support via `NewScannerConfig(tomlString)` in Go API

#### Refactoring Guide
- **8 embedded patterns**: deep_nesting, brain_method, bumpy_road, complex_conditional, god_class, long_parameter_list, primitive_obsession, duplicated_code
- Step-by-step instructions with before/after examples
- Pattern lookup by smell name via MCP tool

#### MCP Server (7 tools)
- `analyze_code` — Full structural analysis with quality score
- `scan_for_secrets` — Gitleaks-based secret detection
- `get_refactoring_strategy` — Pattern lookup by smell name
- `assess_file` — Quick Go Ahead (95+) / Proceed with Care (75-94) / Stop & Refactor (<75) classification
- `set_config` — Persistent config management (8 keys)
- `get_config` — View current configuration
- `list_hotspots` — Git log analysis for frequently-changed low-quality files

#### CLI Commands
- `ailinter check <file|dir>` — Analyze files with multiple output formats (human, json, markdown, problems)
- `ailinter mcp` — Start MCP server on stdio
- `ailinter init` — Bootstrap project (.ailinter.toml, AGENTS.md, optional .vscode/tasks.json)
- `ailinter rules list` — Display default language thresholds

#### Language Support
- **12 languages** with custom thresholds: Go, Python, JavaScript, TypeScript, C/C++, Java, Rust, Ruby, Swift, Kotlin, C#
- **33 source file extensions** for directory scanning
- Per-language thresholds (nesting depth, cyclomatic complexity, function LOC, file LOC, max arguments)

#### Infrastructure
- Cross-platform builds: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- Makefile with build, test, coverage, benchmark, release targets
- 15 test files with 141 test functions + 2 benchmarks
- 13 test fixture files across 7 scenarios

### Fixed (v0.5.0-dev)

#### Function Detection
- **Python**: Now correctly detects functions using indentation-based boundaries (previously used Go's brace detector, missing all Python functions). Supports `def`, `async def`, `@decorators`, class methods, nested functions, and auto-detection of tabs vs spaces.
- **Multiline signatures**: All brace-based detectors (Go, C++, Java, Rust, Swift, Kotlin) now correctly handle function signatures where the opening `{` is on a different line from the function keyword.
- **Java**: Added support for annotated methods (`@GetMapping`, any `@` prefix), generic return types (`List<User>`), `default` methods, `synchronized`, and proper word-boundary matching (fixed `strings.Contains` false positives that could match comment text).
- **Kotlin**: Added support for `suspend`, `inline`, `operator`, `infix`, `tailrec`, `external`, and annotation-stripped function definitions.
- **C++**: Added support for destructors (`~ClassName`), operator overloads (`operator==`, `operator+`), and constructor initializer lists.

#### Scoring
- Action-oriented tiers: Go Ahead (95+), Proceed with Care (75-94), Stop & Refactor (<75)
- Labels are direct instructions LLMs follow: "Go Ahead" = safe, "Proceed with Care" = guarded, "Stop & Refactor" = hard constraint

### Planned (v0.6.0)

- SAST lite: SQL injection, XSS, weak crypto, insecure permissions detection
- IaC security: 10 rules for Terraform/CF/Docker/K8s misconfigurations
- Dependency SCA: hallucinated package detection + OSV.dev CVE lookup
- OWASP Top 10, CWE, NIST SSDF classification on all security findings
- AI hallucination heuristics for secrets (unknown-format detection, multi-file correlation)
- Pre-commit hook installation
- Diff-aware analysis (analyze changed files vs base branch)
