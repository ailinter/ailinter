#!/bin/sh
set -eu

# ailinter installer
# Usage: curl -fsSL https://raw.githubusercontent.com/ailinter/ailinter/main/scripts/install.sh | sh

REPO="ailinter/ailinter"
VERSION="${VERSION:-latest}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin)  OS="darwin" ;;
    linux)   OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

SUFFIX=""
[ "$OS" = "windows" ] && SUFFIX=".exe"

# Download URL
if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/$REPO/releases/latest/download/ailinter-$OS-$ARCH$SUFFIX"
else
    URL="https://github.com/$REPO/releases/download/$VERSION/ailinter-$OS-$ARCH$SUFFIX"
fi

# Install path
if [ "$OS" = "windows" ]; then
    DEST="$HOME/bin/ailinter.exe"
else
    DEST="$HOME/.local/bin/ailinter"
    mkdir -p "$HOME/.local/bin"
fi

echo "ailinter: installing $VERSION for $OS/$ARCH..."
echo "  Downloading $URL"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "$DEST"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$URL" -O "$DEST"
else
    echo "Error: curl or wget required"
    exit 1
fi

chmod +x "$DEST"

echo "  Installed to $DEST"
echo "  Run 'ailinter --version' to verify"

# Check if in PATH
if ! command -v ailinter >/dev/null 2>&1; then
    echo ""
    echo "Note: $HOME/.local/bin is not in your PATH."
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
