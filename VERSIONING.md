# Versioning Policy

AILINTER follows [Semantic Versioning 2.0.0](https://semver.org/).

## API Surface

The v1.0 API surface includes:

- **CLI flags and commands**: `ailinter check`, `ailinter mcp`, `ailinter init`, `ailinter install-hook`
- **Output formats**: `--format problems`, `--format json`, `--format sarif` (SARIF from v1.0)
- **MCP tool schemas**: Tool names, parameter names, return types
- **Configuration file**: `.ailinter.toml` schema
- **Exit codes**: 0 = clean, 1 = findings, 2 = error

## Breaking Changes

A **breaking change** is any change that:

1. Removes or renames a CLI flag, command, or MCP tool
2. Changes the output format (field names, structure)
3. Changes exit code semantics
4. Removes or renames a `.ailinter.toml` key
5. Drops support for a language, OS, or architecture

## Non-Breaking Changes

These are **not** breaking changes and are allowed in minor/patch releases:

- Adding new CLI flags, commands, or MCP tools
- Adding new fields to existing output formats
- Adding new secret rules, vulnerability patterns, or code quality detectors
- Improving performance
- Bug fixes

## Compatibility Guarantees

| Version | Guarantee |
|---------|-----------|
| v1.x    | CLI, MCP, config, and output formats are stable. Minor versions add features; patch versions fix bugs. |
| v0.x    | No stability guarantees. Breaking changes may occur in any release. |

## Build Metadata

The following build-time values are embedded in every release binary via `ldflags`:

| Variable | Source | Example |
|----------|--------|---------|
| `version.Version` | `git describe --tags --always --dirty` | `v1.0.0` |
| `version.Commit` | `git rev-parse --short HEAD` | `a1b2c3d` |
| `version.BuildDate` | `date -u +"%Y-%m-%dT%H:%M:%SZ"` | `2026-06-07T12:00:00Z` |

View them at runtime with:

```bash
ailinter --version
```
