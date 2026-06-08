# CI Integration

Integrate AILINTER into your CI pipeline to block PRs with low quality scores, hardcoded secrets, or vulnerabilities. v1.0 supports **SARIF output** for GitHub Code Scanning, **quality gates** for score enforcement, and **diff-aware analysis** for fast incremental scans.

---

## GitHub Code Scanning (SARIF)

Upload AILINTER results directly to the GitHub Security tab using the SARIF v2.1.0 format:

```yaml
# .github/workflows/ailinter-sarif.yml
name: AILINTER SARIF
on:
  pull_request:
  push:
    branches: [main]
jobs:
  analyze:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      contents: read
    steps:
      - uses: actions/checkout@v4
      - name: Install AILINTER
        run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - name: Run AILINTER with SARIF output
        run: ailinter check . --format sarif --output results.sarif
      - name: Upload SARIF to GitHub
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
          category: ailinter
```

This surfaces all findings — quality, secrets, and vulnerabilities — in the GitHub **Security > Code Scanning** tab. Each finding includes severity, rule ID, description, and refactoring guidance.

---

## Quality Gate

Block PRs when code quality drops below a threshold. The recommended minimum is 80:

```yaml
# .github/workflows/ailinter-gate.yml
name: AILINTER Quality Gate
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install AILINTER
        run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - name: Quality gate (score >= 80, no secrets)
        run: |
          ailinter check . --format problems
          ailinter check . --format json | jq -e '.score >= 80'
```

| Exit Code | Meaning |
|:---------:|---------|
| 0 | All checks passed (score ≥ threshold, no secrets, no vulnerabilities) |
| 1 | Score below threshold, secrets found, or vulnerabilities detected |

---

## Secret Blocking

Add a dedicated secrets scan step to fail the build if any hardcoded credentials are detected:

```yaml
- name: Scan for secrets
  run: ailinter check . --secrets-only --format problems
```

Secrets are detected using 269+ rules covering 100+ providers (AWS, GitHub, Stripe, OpenAI, and more).

---

## Diff-Aware Analysis

For large repositories, scan only the lines that changed in a PR relative to a base branch:

```bash
ailinter check . --diff main
```

This is ideal for CI where you want fast feedback without rescanning the entire codebase. Only files that changed are analyzed, and within those files, only the changed lines trigger quality/vulnerability findings.

Example with pull request base:

```yaml
- name: Diff-aware scan
  run: ailinter check . --diff ${{ github.event.pull_request.base.ref }}
```

---

## Combined Workflow

Run both SARIF upload and quality gate in a single workflow:

```yaml
name: AILINTER CI
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
      - name: Install AILINTER
        run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - name: Generate SARIF
        run: ailinter check . --format sarif --output results.sarif --diff ${{ github.event.pull_request.base.ref }}
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
      - name: Enforce quality gate
        run: |
          ailinter check . --format json --diff ${{ github.event.pull_request.base.ref }} | jq -e '.score >= 80'
```

---

## CI Platform Comparison

| Platform | Integration Method | Status |
|----------|-------------------|--------|
| **GitHub Actions** | SARIF upload + quality gate | ✅ Documented above |
| **GitLab CI** | Custom job with exit code | ✅ (use `--format json`, check exit code) |
| **CircleCI** | Custom step with `jq` gate | ✅ (same pattern) |
| **Jenkins** | Shell step with SARIF archive | ✅ |
| **Bitbucket Pipelines** | Custom step | ✅ |

All platforms can use the JSON output (`--format json`) and check the exit code. For SARIF, pipe the output to a file and upload it using your CI platform's SARIF upload mechanism.
