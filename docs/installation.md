# Installation

AILINTER v1.0.0 is a single 30 MB Go binary with zero runtime dependencies. Install it on macOS, Linux, or Windows — or install the VS Code extension for inline editor integration.

---

## macOS (Homebrew)

```bash
brew install ailinter/ailinter/ailinter
```

Verify:
```bash
ailinter version
# → ailinter v1.0.0 (30 MB, zero dependencies)
```

## Linux (amd64 / arm64)

### Option 1: Download from Releases

```bash
# Latest version — replace with actual URL from releases page
curl -sSL https://github.com/ailinter/ailinter/releases/latest/download/ailinter-linux-amd64.tar.gz | tar xz
sudo mv ailinter /usr/local/bin/
```

### Option 2: Go Install

```bash
go install github.com/ailinter/ailinter/cmd/ailinter@latest
```

> Ensure `$GOPATH/bin` is in your `$PATH`:
> ```bash
> echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
> source ~/.bashrc
> ```

## Windows

### Option 1: Download from Releases

1. Download `ailinter-windows-amd64.exe` from [releases](https://github.com/ailinter/ailinter/releases)
2. Rename to `ailinter.exe`
3. Add to a directory in your `PATH`

### Option 2: Go Install

```powershell
go install github.com/ailinter/ailinter/cmd/ailinter@latest
```

## Any Platform (Go Install)

Requires Go 1.23+:

```bash
go install github.com/ailinter/ailinter/cmd/ailinter@latest
```

This works on macOS, Linux, Windows, and any OS that Go supports.

## Docker

```bash
docker pull ailinter/ailinter

# Scan your code
docker run -v $(pwd):/code ailinter/ailinter check /code
```

## VS Code Extension

Install from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=ailinter.ailinter):

1. Open VS Code
2. Go to Extensions (`Cmd+Shift+X` / `Ctrl+Shift+X`)
3. Search for "AILINTER"
4. Click **Install**

Or install via command line:

```bash
code --install-extension ailinter.ailinter
```

Once installed:
- **Inline diagnostics** appear in the Problems panel when you open a file
- **Status bar** shows the current file's quality score and issue count
- **Run on save** automatically checks files on save (configurable)
- **Problem matcher** maps ailinter output to VS Code's diagnostic system

No manual MCP config needed — the extension handles the connection.

---

## Verify Installation

Run these commands to confirm AILINTER is working:

```bash
# Check version
ailinter version

# Quick self-test
echo 'print("hello")' > /tmp/test.py
ailinter check /tmp/test.py
# → Should show a quality score
rm /tmp/test.py
```

## Upgrading

```bash
# Homebrew
brew upgrade ailinter/ailinter/ailinter

# Go install
go install github.com/ailinter/ailinter/cmd/ailinter@latest

# Binary download
# Re-download from https://github.com/ailinter/ailinter/releases
```

## Uninstalling

```bash
# Homebrew
brew uninstall ailinter/ailinter/ailinter

# Go install
rm $(go env GOPATH)/bin/ailinter
```

## Platform Matrix

| Platform | amd64 | arm64 | Install Methods |
|----------|:-----:|:-----:|----------------|
| macOS | ✓ | ✓ | Homebrew, Go install, binary |
| Linux | ✓ | ✓ | Go install, binary, Docker |
| Windows | ✓ | — | Go install, binary |
| Docker | ✓ | ✓ | `docker pull ailinter/ailinter` |

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `ailinter: command not found` | Ensure install path is in `$PATH`. For Go install: `export PATH=$PATH:$(go env GOPATH)/bin` |
| `brew install` fails | Run `brew update` first. If tap is missing, run `brew tap ailinter/ailinter` |
| Docker permissions | Make sure the mounted volume is readable by the container UID |
| Binary "cannot execute" | You may have downloaded the wrong architecture. Use `uname -m` to check |
