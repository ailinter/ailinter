# Benchmarks

Comprehensive 7-tool comparison across 11 controlled test fixtures and 3 clean open-source repos.

> **This page is under construction.** See the [README](https://github.com/ailinter/ailinter#-benchmarks) for current benchmark data.

## Quick Results

| Tool | Recall | Precision | Speed | Binary Size |
|------|:------:|:---------:|:-----:|:-----------:|
| **ailinter** | **100%** | **100%** | **347 ms** | **30 MB** |
| gitleaks | 100% | 100% | 357 ms | 10 MB |
| trufflehog | 92% | 100% | 15,737 ms | 85 MB |
| semgrep | 58% | 74% | 27,801 ms | 217 MB |

## Key Findings

- **100% recall** with zero false positives
- **2.03× more patterns** than Gitleaks (secrets + vulns + quality in one pass)
- **80× faster** than Semgrep
- **Full benchmark report:** [github.com/ailinter/benchmarks](https://github.com/ailinter/benchmarks)
