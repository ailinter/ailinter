# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.5.x   | Yes |
| < 0.5   | No |

## Reporting a Vulnerability

**Do not open a public issue.** Email security@ailinter.dev with details.

We aim to respond within 48 hours and resolve within 7 days.

### What to Include

- Steps to reproduce
- Affected version
- Any proof-of-concept code (if available)

## What ailinter Scans For

ailinter itself is built with security-first principles:

- **Secrets**: We scan our own codebase with ailinter before every commit. No API keys, tokens, or credentials are committed.
- **Vulnerabilities**: We run the same 25 vulnerability patterns on our source that we ship to you.
- **Dependencies**: We pin all Go module versions and review dependency changes.

## CI Security

Our GitHub Actions CI pipeline:

- Runs `ailinter check` on all changed files
- Requires passing checks before merge
- Does not log secrets or credentials
- Uses read-only repository tokens by default

## Acknowledgments

We follow the [Secure Software Development Framework (SSDF)](https://csrc.nist.gov/projects/ssdf) practices and align with OWASP Top 10 recommendations.
