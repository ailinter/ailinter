# Changelog

## [v1.0.0] — 2026-06-08

**Stable release** — API-stable, feature-complete, production-ready. Consolidates v0.8.7 through v0.9.0 development and ships the v1.0 milestone.

### Added

#### VS Code Extension (CodeScene-grade UX)
- **Full-featured extension** in `vscode-extension/` repo (`v0.2.0`): diagnostics in Problems panel, inline decorations (gutter icons + highlights), CodeLens with function-level scores and ▲/▼ delta, hover guidance with refactoring strategy previews, Quick Fix lightbulb (Get Refactoring Strategy / Replace Secret / Suppress Warning)
- **Sidebar view** — project quality tree with per-directory breakdown, file count, issue stats, top hotspots, top code smells
- **Status bar** — quality score with delta indicator; click opens file quality summary
- **Webview panel** — rich documentation browser + refactoring strategy viewer
- **Git merge-base delta tracking** — compares current score against `main` branch baseline, shows improvement/regression
- **Automated scanning** — on save, on file open, periodic git change polling (9s interval), external file system watcher (1s debounced)
- **Concurrency-limited executor** — queues scans to prevent resource exhaustion
- **Code Quality Delta dashboard** — webview-based delta visualization
- **Walkthrough** — Getting Started guide with 5-step onboarding
- **Configuration** — 5 settings: binary path, enable/disable, scan-on-open, quality threshold, gutter icons, CodeLens

#### Refactoring Strategies — 24 Patterns with Go Examples
- **19 real smells** covered in initial release: brain_method, bumpy_road, complex_conditional, complex_method, data_class, deep_nesting, duplicated_code, excessive_comments, file_bloat, global_data, god_class, lazy_element, long_method, long_parameter_list, long_scope_variable, long_switch, low_cohesion, message_chains, parallel_inheritance, paragraph_of_code, primitive_obsession, refused_bequest, shotgun_surgery
- **24 total patterns** in `internal/refactoring/patterns/` — each with step-by-step strategy, before/after Go code examples, and verification steps
- **`get-refactoring-strategy` MCP tool** updated to discover all 24 patterns dynamically
- **Pattern lookup by smell name** via `ailinter check --refactoring-strategy <smell>` and MCP `get_refactoring_strategy` tool

#### SARIF Output with Embedded Refactoring Guidance
- **SARIF v2.1.0 output format** (`--format sarif`) — GitHub Code Scanning compatible
- **Rich rule metadata** — each SARIF result includes embedded refactoring strategy, severity, and classification
- **Stable rule names** — deterministic, version-independent identifiers for all quality/secrets/vulnerability findings
- **Security-severity as string** (per SARIF spec), repo-relative URIs for cross-workspace compatibility
- **GitHub Code Scanning upload** — automated SARIF upload in `pr-check.yml` workflow
- **PR comment integration** — posts full quality report as PR comment with collapsible details

#### GitHub Actions PR Check
- **Composite action** `.github/actions/ailinter-check/` — reusable across any repository
  - Installs ailinter binary, runs scan, returns `quality-score`, `secrets-found`, `issues-found` as step outputs
  - Configurable `quality-threshold`, `fail-on-secrets`, `version`, `scan-path`
  - Works on Linux and macOS runners (x86_64 and arm64)
- **PR check workflow** (`pr-check.yml`):
  - Runs on every PR to main/master
  - Dogfoods the composite action
  - Builds ailinter from source (for latest SARIF features)
  - Generates SARIF report and uploads to GitHub Code Scanning
  - Posts PR comment with quality summary + collapsible details
- **CI workflow** (`ci.yml`): lint (go vet, gofmt), test (race detection, coverage gate), binary build, cross-build matrix
- **Release workflow** (`release.yml`): automated goreleaser on tags, Homebrew tap update

#### Diff-Aware Analysis (`--diff`)
- **`--diff` flag** on `ailinter check` — scan only lines changed relative to a git ref (`main`, `HEAD~1`, `HEAD`)
- **`git.ChangedFiles()` / `git.ChangedLines()`** — resolves git ref, computes diff hunks, returns affected files and line ranges
- **`executeCheckDiff()`** — diff-aware execution path: skip unchanged files entirely
- **`checkDirectoryDiff()`** — per-directory diff scanning
- **`diffModeAnnotate()`** — marks findings as `[diff-only]` in output
- **`smellInRanges()`** — filters quality issues to only those in changed line ranges (reduces noise in PR scans)
- **`executeCheckDiff()`** handles edge cases: no changes detected, binary-only diffs, non-git repos

#### `install-hook` + `init` Commands
- **`ailinter install-hook`** — auto-installs pre-commit hook into current git repo:
  - Supports standard repos and git worktrees
  - Backs up existing hooks (`pre-commit.backup`)
  - `--force` flag to overwrite existing backup
  - Hook runs: go vet → staticcheck → gofmt → ailinter quality + secrets + vulnerability scan
- **`ailinter init`** — interactive setup wizard:
  - Agent setup for OpenCode, Claude Code, Cursor, Copilot
  - Non-interactive mode for CI/scripted environments
  - Generates `.ailinter.toml`, AGENTS.md, VS Code settings, git hooks
  - Profile-based configuration (default, strict, minimal)
  - VS Code settings file generation with ailinter integration

#### Git Merge-Base Delta Analysis
- **`git.ChangedFiles(repoRoot, ref)`** — files changed relative to any git ref
- **`git.ChangedLines(repoRoot, ref, file)`** — line-level changes via hunk parsing
- **`git.ChangedFilesStaged(repoRoot, ref)`** — staged changes only (pre-commit)
- **`LineRange`** type with `Start`/`End` for precise hunk boundaries
- **`ParseHunkHeaders()`** — parses unified diff hunk headers (`@@ -a,b +c,d @@`)
- **VS Code integration** — delta dashboard compares current scores against merge-base
- **`--diff HEAD`** for pre-commit scanning, `--diff main` for PR analysis

#### CLI Coverage 80%+
- `internal/cli` package coverage: **80%+ line coverage** across all major files
  - `check.go`: 85%+ avg — all scan paths, output formats, diff mode
  - `output_human.go`, `output_md.go`, `output_problems.go`, `output_sarif.go` — near 100% for all render paths
  - `init.go`: 90%+ — non-interactive and agent setup paths
  - `setup.go`: 75%+ — config writing, file generation, VS Code integration
  - `install_hook.go`: 80% command registration
  - `report.go`: 80% — report generation
  - 269 total functions, only interactive/terminal-dependent paths uncovered

#### Click-Tracking Worker
- **Cloudflare Worker** (`telemetry-worker/src/links.ts`) for campaign click tracking:
  - Redirect-based click measurement (`/r/<campaign>` → landing page)
  - Logs clicks to R2 (`clicks/YYYY-MM-DD.jsonl`) with: timestamp, campaign, IP, country, city, region, timezone, user-agent, referer
  - UTM parameter injection for campaign attribution
  - Configurable landing URL via KV or env vars
  - Zero-dependency, ~2KB worker

#### API Versioning + Semver Contract
- **`internal/version` package** — semantic versioning with ldflags injection
  - `Version`, `Commit`, `BuildDate` set at build time
  - `Semver()` — strips `v` prefix for comparison
  - `APIVersion()` — returns major API compatibility version (`v0` / `v1`)
  - `IsPrerelease()` — detects `-dev`, `-alpha`, `-beta`, `-rc` suffixes
  - `String()` — full version string with Go version, OS/arch
  - `Short()` — compact `ailinter version v1.0.0`
- **Semver contract**: MAJOR bump for breaking CLI/MCP API changes, MINOR for new features, PATCH for bug fixes
- **Backward-compatible output formats**: `auto`, `human`, `json`, `markdown`, `problems`, `sarif` all stable
- **MCP tool interface** stable: `analyze_code`, `scan_for_secrets`, `assess_file`, `get_refactoring_strategy`, `get_config`, `set_config`, `list_hotspots`

#### Benchmark Badge (203% Recall vs Gitleaks)
- **SecretBench benchmark badge** in README: ![SecretBench](https://img.shields.io/badge/SecretBench-203%25_recall_vs_Gitleaks-7c3aed)
- **Published benchmark section** in README with:
  - Full comparison table: ailinter vs gitleaks vs trufflehog vs detect-secrets vs Semgrep
  - Metrics: recall (203%), precision (46%+), F1, detection speed, binary size
  - Methodology: 15 real-world repos, 1,259 commits, 15,084 labeled true secrets
  - Academic citation: SecretBench (MSR 2023) / FPSecretBench (ESEM 2023)
- **Multi-tool benchmark**: 7 tools compared across recall, speed, coverage
- **Secret badge** in header: `secret_detection: 100% recall`

#### Infrastructure
- **Dockerfile** — multi-stage build (builder → distroless), 30 MB binary
- **Docker HEALTHCHECK** — `ailinter version` health probe
- **Pre-commit hook** — `scripts/pre-commit.sh` with 3-step quality gate: go vet → staticcheck/go fmt → ailinter self-check
- **Makefile** — `test-quick` target, `precommit` alias, cross-platform builds, release automation
- **Telemetry improvements**:
  - Flag tracking (enabled/disabled tools, thresholds)
  - MCP client auto-detection (OpenCode, Claude, Cursor, Copilot, custom)
  - First-run detection (unique installs)
  - Directory scan metrics (files found, skipped, by language)
  - Go version in resource attributes
- **CLI update**: file exclusion config for vulnerability scanner, auto-skip `testdata/` directories

#### Documentation
- **VS Code Extension walkthrough** — 5-step interactive Getting Started guide (install, scan, understand scores, Quick Fix, sidebar)
- **AGENTS.md** — post-retro lessons, lifecycle ownership rules, staticcheck enforcement
- **CHANGELOG** — comprehensive release notes for all versions
- **README** — SecretBench benchmark section, updated language support table, installation guide
- **Smell documentation** — 24 refactoring pattern markdown files with Go examples
- **Benchmark report** — comprehensive comparison in `ops/roadmap/benchmark-report.md`
- **CLI coverage report** — `ops/roadmap/cli-coverage-80.md` with per-function breakdown

### Changed
- **Version**: v0.8.6 → v1.0.0 (stable, semver-contract API)
- **Refactoring strategies**: expanded from 8 → 24 patterns (3× coverage)
- **SARIF output**: enriched with refactoring guidance, stable rule names, security-severity compliance
- **CI/CD**: actions upgraded to v6/v7 (checkout, setup-go, upload-artifact), Node.js 24-compatible
- **MCP server**: client auto-detection, dynamic refactoring strategy discovery, self-documenting tools
- **Quality gate**: staticcheck added to pre-commit hook, `Tests: true` includes test files in analysis
- **Pre-commit**: `--meta-lint` enabled by default, go vet + staticcheck + gofmt + ineffassign + misspell
- **Code coverage upload**: non-fatal when Code Quality not enabled in GitHub
- **Binary size**: documented as 30 MB self-contained Go binary
- **Docker**: multi-stage build, proper HEALTHCHECK command
- **Telemetry**: enhanced with flag tracking, MCP client detection, first-run, directory scans, Go version

### Fixed
- **SARIF security-severity**: now serialized as string per SARIF v2.1.0 specification
- **SARIF paths**: repo-relative URIs instead of absolute paths
- **Testdata directories**: auto-skipped in scans (was scanning fixture files with false positives)
- **Code coverage upload**: non-fatal fallback when GitHub Code Quality feature not enabled
- **Docker HEALTHCHECK**: corrected `--version` syntax
- **`sync.Once` copy violation**: removed lock copy in telemetry tests (`telemetry_init_test.go`)
- **Gofmt compliance**: formatting fixed across 5+ files
- **`Tests: true` behavior**: now correctly includes test files in analysis (was excluding them)
- **Staticcheck enforcement**: pre-commit hook now catches U1000 (unused functions) that go vet misses
- **CLI edge cases**: binary detection, permission denied, symlink loops, non-git repos

### v1.0.0 Milestone Summary

| Metric | Value |
|--------|-------|
| Code Quality detectors | 20 types across 13 languages |
| Secret scanning rules | 269 (betterleaks engine) |
| Vulnerability patterns | 58 across 8 languages, 6 categories |
| Refactoring strategies | 24 patterns with Go examples |
| MCP tools | 7 (stable API) |
| CLI commands | 8 (check, mcp, init, install-hook, rules, telemetry, report, version) |
| Output formats | 6 (auto, human, json, markdown, problems, sarif) |
| VS Code extension | v0.2.0 (full-featured) |
| GitHub Actions | Composite action + 3 workflows |
| Binary size | 30 MB (self-contained, no runtime deps) |
| Platforms | darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64 |
| Secret detection | 203% recall vs Gitleaks (SecretBench) |
| CLI coverage | 80%+ |
| API | Versioned with semver contract |
| Telemetry | OTLP/HTTP → Cloudflare Worker → R2 → DuckDB |
| Install | Homebrew, Go install, GitHub Releases, Docker, VS Code |

---

## [v0.8.6] — 2026-05-28

### Added
- **Telemetry improvements**: flag tracking, MCP client detection, first-run detection, directory scan metrics, Go version in resource attributes
- **Pre-commit hook**: `scripts/pre-commit.sh` with 3-step quality gate (go vet → go fmt → ailinter self-check)
- **Makefile targets**: `precommit` target aliased to `lint check`

### Changed
- **Metalinter**: `Tests: true` now includes test files (was excluding them); `--meta-lint` enabled by default
- **Version**: `v0.8.5` → `v0.8.6`

### Gaps (In Progress — deferred to v0.9)
- **Benchmark report**: README badge + comparison table (203% recall vs gitleaks) — see `benchmark-report.md`
- **CLI coverage**: `internal/cli` at 74.8% → target 80% — see `cli-coverage-80.md`
- **Demo GIFs**: 5 terminal recordings for landing page/README — see `demo-content.md`

### Fixed
- **Metalinter bug**: `Tests: true` flag now correctly includes test files in analysis
- **sync.Once copy violation**: removed `sync.Once` copy in telemetry tests (`telemetry_init_test.go`)
- **Gofmt compliance**: formatting fixed across 5 files

## [v0.8.5] — 2026-05-28

### Added
- **Metalinter package**: Multi-linter integration with go vet, staticcheck, misspell, ineffassign
- **Token estimator**: Estimate token cost of code analysis for LLM context window management
- **Function coverage gate**: 70% function coverage + 80% line coverage enforcement in CI

### Changed
- Improved quality scores — analyzer 99, scanner 100 across core packages
- Updated dependencies: viper v1.21, go-toml v3.1, isatty v0.0.22, mergo v1.0.2
- go.mod: go 1.25.5 directive
- Version from `0.0.0-dev` → `v0.8.5`

### Fixed
- Telemetry test: fixed TestInit_ResourceError to correctly test disabled path
- CI coverage exclusion: exclude scripts/ and cmd/ from test coverage
- Release workflow: proper contents:write permission, single aggregated release

## [v0.8.1] — 2026-05-28

### Added
- **Anonymous telemetry** via OpenTelemetry OTLP/HTTP
  - 9 metric instruments: CLI invocations, MCP tool calls, files analyzed, quality scores, smells/secrets detected, duration, errors, installations
  - Cloudflare Worker → R2 → DuckDB backend
  - Opt-out via `AILINTER_NO_TELEMETRY=1` or `go build -tags no_telemetry`
  - `ailinter telemetry` command shows what is collected
- **Proper versioning**: `debug.ReadBuildInfo()` embeds version for `go install @v0.8.1`, ldflags for release builds
- **Pre-built binaries** on GitHub Releases for all platforms
- **Homebrew tap**: `brew install ailinter/ailinter/ailinter`

### Changed
- Version from `0.5.0-dev` → `v0.8.1`
- Go module directive: `go 1.25`

## [v0.7.3] — 2026-05-27

### Added
- **Vulnerability scanner**: 37 security patterns across 8 languages, 6 categories (injection, XSS, deserialization, weak crypto, XXE, path traversal)
- Most-vulnerable identification tracking per language
- `--no-vulnerabilities` flag, MCP vuln data integration

### Fixed
- C# Process.Start regression
- Docstring cleanup, workflow fixes, JS SQLi pattern, Java SHA-1 detection

## [v0.7.2] — 2026-05-27

### Fixed
- Bug fixes: docstrings, workflow, JavaScript SQLi, Java SHA-1

## [v0.7.0] — 2026-05-26

### Added
- **Vulnerability scanner**: 25 security patterns, 6 categories
- `--no-vulnerabilities` flag
- MCP vulnerability data in `analyze_code` output

## [v0.6.0] — 2026-05-25

### Added
- **Vulnerability scanner**: initial security pattern detection
- **MCP server improvements**: error handling, path validation

## [v0.5.0] — 2026-05-24

### Added

#### Code Quality Radar (17 detectors)
- Deep nesting, brain method, file bloat, bumpy road detection
- Complex conditional, long parameter list, cyclomatic complexity
- Message chain, primitive obsession, lazy element, paragraph of code
- Code duplication, cohesion analysis, excessive comments
- Global data, long scope variable, long switch detection

#### Secret Scanning
- 269 detection rules (betterleaks engine)
- 100+ providers: cloud, AI/ML, dev platforms, payments, databases
- Config file scanning (`.env`, `.toml`, `Dockerfile`, etc.)
- Entropy-based severity classification
- Secret redaction in output

#### Refactoring Guide
- 8 embedded patterns: deep_nesting, brain_method, bumpy_road, etc.
- Step-by-step instructions with before/after examples
- Pattern lookup by smell name via MCP tool

#### MCP Server (7 tools)
- `analyze_code`, `scan_for_secrets`, `get_refactoring_strategy`, `assess_file`
- `set_config`, `get_config`, `list_hotspots`

#### CLI Commands
- `ailinter check`, `ailinter mcp`, `ailinter init`, `ailinter rules list`

#### Language Support
- 12 languages with custom thresholds
- 33 source file extensions for directory scanning

#### Infrastructure
- Cross-platform builds: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- Makefile with build, test, coverage, benchmark, release targets
- CI: lint, test (80% coverage gate), build, cross-build, benchmark
