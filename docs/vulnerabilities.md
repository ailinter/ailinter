# Vulnerability Patterns

AILINTER detects 58 vulnerability patterns across 6 categories.

> **This page is under construction.** For now, see the [README](https://github.com/ailinter/ailinter#vulnerability-patterns-58).

## Categories

| Category | Count | Key Patterns |
|----------|:-----:|--------------|
| Injection | 28 | Command injection, SQL injection, SSRF, path traversal, eval |
| XSS | 11 | DOM sinks, template bypass, dangerouslySetInnerHTML |
| Deserialization | 11 | pickle, yaml, ObjectInputStream, BinaryFormatter |
| Cryptography | 6 | MD5, SHA-1, DES, ECB, TLS bypass |
| XXE | 1 | Python stdlib XML, C# XmlDocument |
| Workflow | 1 | GitHub Actions pull_request_target |

## Languages Covered

Python, Go, JavaScript/TypeScript, Java, C#, PHP.
