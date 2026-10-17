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
BIN=""

if [ -x "$SCRIPT_DIR/fitgirl-cli" ]; then
    BIN="$SCRIPT_DIR/fitgirl-cli"
elif [ "$ARCH" = "arm64" ] && [ -x "$SCRIPT_DIR/dist/fitgirl-cli-darwin-arm64" ]; then
    BIN="$SCRIPT_DIR/dist/fitgirl-cli-darwin-arm64"
elif [ -x "$SCRIPT_DIR/dist/fitgirl-cli-darwin-amd64" ]; then
    BIN="$SCRIPT_DIR/dist/fitgirl-cli-darwin-amd64"
elif command -v go >/dev/null 2>&1; then
    echo "Compiling fitgirl-cli for macOS..."
    go build -ldflags="-s -w" -o "$SCRIPT_DIR/fitgirl-cli" .
    BIN="$SCRIPT_DIR/fitgirl-cli"
fi

if [ -z "$BIN" ]; then
    echo "=================================================================="
    echo " Error: fitgirl-cli binary not found."
    echo " Please compile it with: go build -o fitgirl-cli ."
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
