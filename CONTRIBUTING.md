# Contributing to ailinter

## Development Setup

```bash
git clone https://github.com/ailinter/ailinter.git
cd ailinter
make build        # Builds to bin/ailinter
make test         # Run all tests
make lint         # Go vet
```

Requirements: **Go 1.25+**

## Code Conventions

- Follow Go idioms: `gofmt`, `go vet`, table-driven tests
- No empty interface (`interface{}`) in new code — prefer generics or concrete types
- Error messages start lowercase
- Package names are single word, lowercase
- Test functions use descriptive names: `TestFeature_Scenario`

## Test Requirements

- **Coverage threshold: 80%+** on line, branch, and function coverage
- Every new feature must include unit tests
- Every new detector must include test fixtures in `testdata/`
- Run `make test-cover` before submitting

## Pull Request Checklist

- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] `make test-cover` shows >=80% line coverage
- [ ] `make bench` shows no regressions
- [ ] New features include tests and testdata fixtures
- [ ] `ailinter check .` on changed files shows no regressions
- [ ] No hardcoded secrets (`ailinter check --no-secrets .` is clean)

## Project Architecture

```
cmd/ailinter/           # CLI entry point (cobra commands)
internal/
  analyzer/             # Orchestrator: runs all detectors, computes 0-100 score
  cli/                  # check, init, mcp commands + output formatters
  config/               # Persistent JSON config at ~/.config/ailinter/config.json
  mcp/                  # MCP stdio server + 7 tool handlers
  parser/               # 17 code smell detectors + git hotspot analysis
  refactoring/          # Embedded .md patterns + lookup engine
  secrets/              # Gitleaks v8 wrapper (150+ rules)
testdata/               # Language-specific test fixtures
```

## Adding a New Detector

1. Create detector function in `internal/parser/`
2. Add threshold fields to `parser.Thresholds` in `thresholds.go`
3. Wire into `analyzer.Analyze()` in `analyzer.go`
4. Add test fixture in `testdata/<smell_name>/`
5. Add unit test in `internal/parser/`
6. Add integration test in `internal/analyzer/`
7. Add refactoring pattern in `internal/refactoring/patterns/`

## Adding a New Language

1. Add extension mapping in `parser.DetectedLanguage()` in `types.go`
2. Add threshold profile in `parser.DefaultThresholds()` in `thresholds.go`
3. If language needs function detection, implement in `parser/bloat_langs.go`
4. Add `isSourceFile` extension in `cli/check.go`
5. Add test fixture files in `testdata/`
6. Add language-specific tests in `parser/detector_test.go`

## Release Process

```bash
git tag -a v0.5.0 -m "v0.5.0"
git push origin v0.5.0
make release    # Cross-compile for all platforms
```

## Questions?

Open an issue on GitHub or start a discussion.
