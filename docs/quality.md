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

## Supported Languages

Go, Python, JavaScript, TypeScript, Java, C#, PHP, Rust, Ruby, Swift, Kotlin, C/C++ — all with full 20-detector coverage.
