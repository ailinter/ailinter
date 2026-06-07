# Secret Scanning

AILINTER detects hardcoded secrets using 269+ rules from betterleaks with a gitleaks fallback — covering 100+ providers.

> **This page is under construction.** For now, see the [README](https://github.com/ailinter/ailinter#secret-detection-rules).

## Coverage

| Category | Examples |
|----------|----------|
| Cloud | AWS, GCP, Azure, DigitalOcean |
| AI/ML | Anthropic, OpenAI, Cohere, DeepSeek |
| Dev Platforms | GitHub, GitLab, Bitbucket, Atlassian |
| Payments | Stripe, PayPal, Shopify, Square |
| Communication | Slack, Discord, Twilio, SendGrid |
| Security | RSA, DSA, EC, PGP, SSH private keys |

## AI Safety

Secrets are **redacted** in MCP output — AI assistants see only the first 4 and last 4 characters.

## Known Limitation

Secrets split across multiple variables and concatenated at runtime are not detected by any static scanner (AILINTER, Gitleaks, TruffleHog). Use environment variables or secret management systems instead.
