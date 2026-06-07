# CI Integration

Integrate AILINTER into your CI pipeline to block PRs with low quality scores or hardcoded secrets.

> **This page is under construction.** For now, see the [README](https://github.com/ailinter/ailinter#-ci-integration) for CI setup.

## GitHub Actions

```yaml
# .github/workflows/ailinter.yml
name: AILINTER Quality Gate
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install AILINTER
        run: go install github.com/ailinter/ailinter/cmd/ailinter@latest
      - name: Run quality check
        run: ailinter check . --format problems
      - name: Enforce gate (score >= 80)
        run: ailinter check . --format json | jq -e '.score >= 80'
```

## Quality Gate

Set a threshold that blocks PRs when code quality drops below a minimum score. The recommended minimum for most projects is 80.

## Secret Blocking

Add `--secrets-only` as a separate step to fail the build if any secrets are detected.
