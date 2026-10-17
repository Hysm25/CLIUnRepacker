#!/usr/bin/env bash
# =========================================================================
# FitGirl Repack CLI - Linux / macOS Launcher
# Native Go binary launcher with automatic architecture detection
# =========================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check for local compiled binary
if [ -x "$SCRIPT_DIR/fitgirl-cli" ]; then
    exec "$SCRIPT_DIR/fitgirl-cli" "$@"
fi

# Detect OS and Arch to find pre-built binary in dist/
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        GOARCH="amd64"
        ;;
    arm64|aarch64)
        GOARCH="arm64"
        ;;
    *)
        GOARCH="amd64"
        ;;
esac

DIST_BIN="$SCRIPT_DIR/dist/fitgirl-cli-${OS}-${GOARCH}"

if [ -x "$DIST_BIN" ]; then
    exec "$DIST_BIN" "$@"
fi

# If Go is installed, build on the fly
if command -v go >/dev/null 2>&1; then
    echo "Building fitgirl-cli from source..."
    go build -o "$SCRIPT_DIR/fitgirl-cli" .
    exec "$SCRIPT_DIR/fitgirl-cli" "$@"
fi

echo "Error: Could not locate a suitable fitgirl-cli binary." >&2
echo "Please run: go build -o fitgirl-cli ." >&2
exit 1
