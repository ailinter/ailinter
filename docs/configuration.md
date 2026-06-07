# Configuration

AILINTER can be configured via `.ailinter.toml` in your project root.

> **This page is under construction.**

## `.ailinter.toml`

```toml
extends = "default"

[rules]
deep_nesting = { weight = 1.0, warning = 3, alert = 5 }
brain_method = { weight = 1.5, warning_lines = 50 }
long_parameter_list = { warning = 3, alert = 6 }

[exclude]
files = ["path/to/generated/*", "vendor/**"]

[mcp]
read_only = false
```

## Profiles

```bash
ailinter init --profile default   # Balanced defaults (recommended)
ailinter init --profile strict    # Stricter thresholds
```
