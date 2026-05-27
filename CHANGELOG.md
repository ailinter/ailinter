# Changelog

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
