#!/usr/bin/env bash
# =========================================================================
# FitGirl Repack CLI - Linux / macOS / Termux Launcher
# Native Go binary launcher with automatic architecture and environment detection
# =========================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
cd "$SCRIPT_DIR"

# Detect OS and Arch to find pre-built binary in dist/
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    linux*|android*)
        OS="linux"
        ;;
    darwin*)
        OS="darwin"
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        GOARCH="amd64"
        ;;
    arm64|aarch64|armv8*|armv9*)
        GOARCH="arm64"
        ;;
    *)
        GOARCH="amd64"
        ;;
esac

DIST_BIN="$SCRIPT_DIR/dist/fitgirl-cli-${OS}-${GOARCH}"

# Ensure executable permissions on dist binary if present
if [ -f "$DIST_BIN" ] && [ ! -x "$DIST_BIN" ]; then
    chmod +x "$DIST_BIN" 2>/dev/null || true
fi

# Prefer pre-compiled binary in dist/
if [ -x "$DIST_BIN" ]; then
    exec "$DIST_BIN" "$@"
fi

# Fallback: check for local binary in root if present
if [ -x "$SCRIPT_DIR/fitgirl-cli" ]; then
    exec "$SCRIPT_DIR/fitgirl-cli" "$@"
fi

# If Go is installed, build into dist/ on the fly
if command -v go >/dev/null 2>&1; then
    echo "Building fitgirl-cli for ${OS}/${GOARCH} from source..."
    mkdir -p "$SCRIPT_DIR/dist"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$DIST_BIN" "$SCRIPT_DIR"
    exec "$DIST_BIN" "$@"
fi

echo "Error: Could not locate a suitable fitgirl-cli binary for ${OS}-${GOARCH}." >&2
echo "Expected binary: $DIST_BIN" >&2
echo "Please compile with: ./build.sh or go build -o dist/fitgirl-cli-${OS}-${GOARCH} ." >&2
exit 1
