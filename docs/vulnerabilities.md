# Vulnerability Patterns

AILINTER detects **58 vulnerability patterns** across **6 categories** — covering Python, Go, JavaScript/TypeScript, Java, C#, PHP, and workflows.

## Complete Pattern Catalog

| Category | Count | Key Patterns |
|----------|:-----:|--------------|
| **Injection** | 28 | Command injection (exec, spawn, subprocess), SQL injection (raw queries, string concatenation), SSRF (request URL from user input), path traversal (zip slip, file access), eval/exec of dynamic code |
| **XSS** | 11 | DOM sinks (innerHTML, dangerouslySetInnerHTML), template bypass, markdown injection, React ref injection, Go template auto-escape bypass |
| **Deserialization** | 11 | pickle, yaml/ruamel, ObjectInputStream (Java), BinaryFormatter (C#), jQuery $.extend, prototype pollution, JSON.parse on unsanitized input |
| **Cryptography** | 6 | MD5/SHA-1 for security, DES/3DES, ECB mode, TLS certificate verification disabled |
| **XXE** | 1 | Python stdlib XML parsers (lxml, ElementTree without resolve_entities=False), C# XmlDocument |
| **Workflow** | 1 | GitHub Actions pull_request_target — full token access from forks, supply-chain attack vector |

## Per-Language Coverage

| Language | Patterns | Highlights |
|----------|:--------:|------------|
| Python | 17 | subprocess shell injection, pickle, yaml load, eval, SSRF |
| JavaScript/TypeScript | 22 | child_process.exec, innerHTML, prototype pollution, crypto weak, eval |
| Go | 7 | crypto/des, crypto/md5, tls.InsecureSkipVerify, template injection, exec.Command shell |
| Java | 5 | Runtime.exec, ObjectInputStream, DES/ECB, yaml.load, SSRF |
| C# | 6 | Process.Start, BinaryFormatter, DES, XmlDocument XXE, XSS |
| PHP | 2 | eval, shell_exec |
| Workflows | 2 | CI workflow injection, pull_request_target |

## Language Detection

AILINTER detects language from file extensions:

- `.py`, `.pyi`, `.ipynb` → Python
- `.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs`, `.mts`, `.cts`, `.vue`, `.svelte` → JS/TS
- `.go` → Go
- `.java` → Java
- `.cs` → C#
- `.php` → PHP
- `.yml`, `.yaml` → Workflows (GitHub Actions)
- `.md`, `.mdx`, `.txt`, `.rst` → Documents (subset of patterns)

## Example Findings

| Rule ID | Severity | Description |
|---------|:--------:|-------------|
| `child_process_exec` | 🔴 critical | Shell string interpolation enables command injection |
| `sql_injection_concat` | 🔴 critical | SQL query built with string concatenation |
| `md5_hash` | 🟡 warning | MD5 is not cryptographically secure |
| `pull_request_target` | 🔴 critical | Full repo token access from any fork |
| `prototype_pollution` | 🟠 alert | Unsafe object merge from user input |
| `deserialize_pickle` | 🔴 critical | pickle.loads on untrusted data = RCE |

## Integration with SARIF

When running with `--format sarif`, each vulnerability finding includes:

- **Rule ID** — unique identifier for the pattern
- **Severity** — critical / alert / warning
- **Description** — what the pattern is and why it matters
- **Reminder** — actionable fix guidance
- **File path + line + column** — exact location
- **Refactoring metadata** — link to the associated refactoring strategy
