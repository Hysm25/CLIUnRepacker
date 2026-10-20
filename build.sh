#!/usr/bin/env bash
# =========================================================================
# FitGirl Repack CLI - Cross-Platform Build & Packaging Script
# Compiles statically linked native binaries for Linux, Termux (ARM64),
# Windows, and macOS.
# =========================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

DIST_DIR="$SCRIPT_DIR/dist"
mkdir -p "$DIST_DIR"

LDFLAGS="-s -w"

# Helper to build a single target
build_target() {
    local target_os="$1"
    local target_arch="$2"
    local ext="$3"
    local out_name="fitgirl-cli-${target_os}-${target_arch}${ext}"
    local out_path="$DIST_DIR/$out_name"

    echo "==> Building ${target_os}/${target_arch} -> dist/${out_name}..."
    GOOS="$target_os" GOARCH="$target_arch" CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o "$out_path" .
    if [ -z "$ext" ]; then
        chmod +x "$out_path"
    fi
}

build_all() {
    echo "==> Compiling all cross-platform targets..."
    # Linux (x86_64)
    build_target "linux" "amd64" ""
    # Linux / Termux (ARM64 / aarch64)
    build_target "linux" "arm64" ""
    # Windows (x86_64)
    build_target "windows" "amd64" ".exe"
    # macOS Apple Silicon (ARM64)
    build_target "darwin" "arm64" ""
    # macOS Intel (x86_64)
    build_target "darwin" "amd64" ""
    echo "==> All binaries successfully compiled to dist/"
}

package_releases() {
    build_all
    echo "==> Creating release packages..."

    local tmp_pkg
    tmp_pkg="$(mktemp -d)"

    # Package Linux amd64
    mkdir -p "$tmp_pkg/linux-amd64/dist"
    cp "$DIST_DIR/fitgirl-cli-linux-amd64" "$tmp_pkg/linux-amd64/dist/"
    cp "$SCRIPT_DIR/fitgirl-cli.sh" "$tmp_pkg/linux-amd64/"
    cp "$SCRIPT_DIR/README.md" "$tmp_pkg/linux-amd64/"
    tar -czf "$DIST_DIR/fitgirl-cli-linux-amd64.tar.gz" -C "$tmp_pkg/linux-amd64" .
    echo "    Created dist/fitgirl-cli-linux-amd64.tar.gz"

    # Package Linux / Termux arm64
    mkdir -p "$tmp_pkg/linux-arm64/dist"
    cp "$DIST_DIR/fitgirl-cli-linux-arm64" "$tmp_pkg/linux-arm64/dist/"
    cp "$SCRIPT_DIR/fitgirl-cli.sh" "$tmp_pkg/linux-arm64/"
    cp "$SCRIPT_DIR/README.md" "$tmp_pkg/linux-arm64/"
    tar -czf "$DIST_DIR/fitgirl-cli-linux-arm64.tar.gz" -C "$tmp_pkg/linux-arm64" .
    echo "    Created dist/fitgirl-cli-linux-arm64.tar.gz"

    # Package macOS arm64
    mkdir -p "$tmp_pkg/darwin-arm64/dist"
    cp "$DIST_DIR/fitgirl-cli-darwin-arm64" "$tmp_pkg/darwin-arm64/dist/"
    cp "$SCRIPT_DIR/fitgirl-cli.command" "$tmp_pkg/darwin-arm64/"
    cp "$SCRIPT_DIR/fitgirl-cli.sh" "$tmp_pkg/darwin-arm64/"
    cp "$SCRIPT_DIR/README.md" "$tmp_pkg/darwin-arm64/"
    tar -czf "$DIST_DIR/fitgirl-cli-darwin-arm64.tar.gz" -C "$tmp_pkg/darwin-arm64" .
    echo "    Created dist/fitgirl-cli-darwin-arm64.tar.gz"

    # Package macOS amd64
    mkdir -p "$tmp_pkg/darwin-amd64/dist"
    cp "$DIST_DIR/fitgirl-cli-darwin-amd64" "$tmp_pkg/darwin-amd64/dist/"
    cp "$SCRIPT_DIR/fitgirl-cli.command" "$tmp_pkg/darwin-amd64/"
    cp "$SCRIPT_DIR/fitgirl-cli.sh" "$tmp_pkg/darwin-amd64/"
    cp "$SCRIPT_DIR/README.md" "$tmp_pkg/darwin-amd64/"
    tar -czf "$DIST_DIR/fitgirl-cli-darwin-amd64.tar.gz" -C "$tmp_pkg/darwin-amd64" .
    echo "    Created dist/fitgirl-cli-darwin-amd64.tar.gz"

    # Package Windows amd64 (.zip)
    mkdir -p "$tmp_pkg/windows-amd64/dist"
    cp "$DIST_DIR/fitgirl-cli-windows-amd64.exe" "$tmp_pkg/windows-amd64/dist/"
    cp "$SCRIPT_DIR/fitgirl-cli.bat" "$tmp_pkg/windows-amd64/"
    cp "$SCRIPT_DIR/README.md" "$tmp_pkg/windows-amd64/"
    if command -v zip >/dev/null 2>&1; then
        (cd "$tmp_pkg/windows-amd64" && zip -r -q "$DIST_DIR/fitgirl-cli-windows-amd64.zip" .)
    else
        python3 -c "import shutil; shutil.make_archive('$DIST_DIR/fitgirl-cli-windows-amd64', 'zip', '$tmp_pkg/windows-amd64')"
    fi
    echo "    Created dist/fitgirl-cli-windows-amd64.zip"

    rm -rf "$tmp_pkg"
    echo "==> Packaging complete! Artifacts located in dist/"
}

clean() {
    echo "==> Cleaning build artifacts..."
    rm -rf "$DIST_DIR"/*
    rm -f "$SCRIPT_DIR/fitgirl-cli" "$SCRIPT_DIR/fitgirl-cli.exe"
    echo "==> Clean complete."
}

# Subcommand dispatch
case "${1:-all}" in
    all)
        build_all
        ;;
    linux-amd64)
        build_target "linux" "amd64" ""
        ;;
    linux-arm64)
        build_target "linux" "arm64" ""
        ;;
    windows-amd64)
        build_target "windows" "amd64" ".exe"
        ;;
    darwin-arm64)
        build_target "darwin" "arm64" ""
        ;;
    darwin-amd64)
        build_target "darwin" "amd64" ""
        ;;
    package)
        package_releases
        ;;
    clean)
        clean
        ;;
    help|-h|--help)
        echo "Usage: $0 [all|linux-amd64|linux-arm64|windows-amd64|darwin-arm64|darwin-amd64|package|clean]"
        ;;
    *)
        echo "Unknown target: $1" >&2
        echo "Usage: $0 [all|linux-amd64|linux-arm64|windows-amd64|darwin-arm64|darwin-amd64|package|clean]" >&2
        exit 1
        ;;
esac
