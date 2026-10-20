#!/usr/bin/env bash
# =========================================================================
# FitGirl Repack CLI - macOS Native Terminal Launcher (.command)
# Double-clickable in macOS Finder or executable via Terminal.
# =========================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

printf "\033]0;FitGirl Repack CLI (Go)\007"

ARCH="$(uname -m)"
case "$ARCH" in
    arm64|aarch64)
        GOARCH="arm64"
        ;;
    *)
        GOARCH="amd64"
        ;;
esac

DIST_BIN="$SCRIPT_DIR/dist/fitgirl-cli-darwin-${GOARCH}"
BIN=""

if [ -f "$DIST_BIN" ] && [ ! -x "$DIST_BIN" ]; then
    chmod +x "$DIST_BIN" 2>/dev/null || true
fi

if [ -x "$DIST_BIN" ]; then
    BIN="$DIST_BIN"
elif [ -x "$SCRIPT_DIR/fitgirl-cli" ]; then
    BIN="$SCRIPT_DIR/fitgirl-cli"
elif command -v go >/dev/null 2>&1; then
    echo "Compiling fitgirl-cli for macOS (${GOARCH})..."
    mkdir -p "$SCRIPT_DIR/dist"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$DIST_BIN" .
    BIN="$DIST_BIN"
fi

if [ -z "$BIN" ]; then
    echo "=================================================================="
    echo " Error: fitgirl-cli binary not found."
    echo " Please compile it with: ./build.sh or go build -o dist/fitgirl-cli-darwin-${GOARCH} ."
    echo "=================================================================="
    echo
    read -n 1 -s -r -p "Press any key to close..."
    echo
    exit 1
fi

"$BIN" "$@"
EXIT_CODE=$?

if [ $# -eq 0 ]; then
    echo
    read -n 1 -s -r -p "Press any key to exit..."
    echo
fi

exit $EXIT_CODE
