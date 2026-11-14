# Code Quality Scoring

How AILINTER's 0–100 quality score works and what each detector catches.

> Reference for ailinter's quality scoring. For a quick overview, see the [README](https://github.com/ailinter/ailinter#-what-it-checks).

## Score Tiers

| Score | Label | AI Guidance |
|-------|-------|-------------|
| 80–100 | Go Ahead | Safe for AI modification |
| 60–79 | Proceed with Care | Small changes, re-check after each |
| 40–59 | Needs Work | Significant issues — refactor incrementally |
| 0–39 | Stop & Refactor | Refactor before AI touches this file |

## 20 Detectors

Deep Nesting, Brain Method, File Bloat, Bumpy Road, Complex Conditional, Cyclomatic Complexity, Long Parameter List, Code Duplication, Low Cohesion, Message Chains, Primitive Obsession, Excessive Comments, Global Data, Long Scope Variables, Lazy Elements, Long Switch, Paragraph of Code, Function Count, Brain Class.

## Language-Aware Analysis

Detectors use the language model, not one Go-shaped heuristic:

| Analysis | Languages and signal |
|---|---|
| Full structural detectors | Go, Python, JavaScript/TypeScript, Java, C#, Kotlin, Swift, Rust, Ruby, C/C++ |
| Function detection | Braces for C-like languages, indentation for Python, `def`/`end` for Ruby, expression bodies for C# |
| Cohesion | Shared types in typed languages; module symbols, `this.*` and class membership for JS/TS; `self.*` for Python; `@ivar` for Ruby |
| Nesting and bumps | Brace depth for C-like languages; indentation for Python; JS strings, template literals and comments are stripped first |
| Parameters and complexity | Language signatures, `and`/`or`/`elif`, `guard`, `when`, `match`, `=>` and `->` |
| Comments and switches | Python docstrings, HTML/SQL/Ruby comments; Kotlin/Rust/Ruby/Python switch forms |

PHP, Scala, shell and Terraform are detected and reported accurately, with line-level detectors and best-effort language constructs. They do not yet have every language-specific rule.

## Data and Markup

JSON, TOML, XML, YAML, properties/INI/ENV are scanned for secrets and vulnerabilities but skipped by structural code-quality detectors. HTML and CSS get comment analysis; tag and selector nesting are not measured.

## Why Scores Move

A file can score lower after upgrading AILINTER because a detector now sees an issue that the old language model missed. A drop is not automatically a regression: check the new smell and the refactoring strategy before treating it as one. False positives are reported as bugs, not worked around in the score.
