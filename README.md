# FitGirl Repack CLI Unpacker & Inspector (Go)

A high-performance, cross-platform interactive terminal application and headless orchestrator written in **Go** for inspecting, pre-verifying, headlessly unpacking, and post-verifying **FitGirl Repacks** across **Windows**, **Linux**, and **macOS**.

---

## About

> I created this project to address the challenges of handling FitGirl repacks on Linux and Unix environments. This project has been validated and tested with **The Legend of Zelda: Link's Awakening [FitGirl Repack]**. Compatibility with other repacks may vary depending on their specific compressor plugins.

---

## 1. Reverse-Engineering Insights (via [REA](https://github.com/morluto/rea))

Deep structural analysis of `setup.exe` and associated archives (`fg-*.bin`) using [REA](https://github.com/morluto/rea) (`rea inspect-with-objdump`), GNU objdump, and Inno Setup container parsing revealed:

### 1.1 Binary Container Architecture
* `setup.exe` is a 32-bit PE GUI executable (`i386`) built with Inno Setup (Unicode Enhanced Edition).
* REA section inspection (`rea inspect-with-objdump setup.exe section_headers`) demonstrates the PE sections (`.text`, `.data`, `.rsrc`, `.idata`, `.tls`) conclude at file offset `0x22600` (~140 KB).
* Everything following offset `0x22600` (over 9.6 MB) is an appended compressed Inno Setup payload overlay containing setup scripts and extraction tools.

### 1.2 Identified Extraction Tools & Decompression Pipeline
When executed, the installer coordinates decompression using an orchestrated pipeline:
* **Inno Setup Engine (`setup.exe`)**: Container installer and orchestrator. Compatible with native Linux **`innoextract`**.
* **`unarc.dll` & `ISDone.dll`**: FreeArc extraction coordinator and Inno Setup callback bridge.
* **`cls-srep_x64.exe` / `_x86.exe` (`CLS-srep.dll`)**: Super Reflation (SREP) deduplication filter by Bulat Ziganshin.
* **`cls-magic2_x64.exe` / `_x86.exe` & `cls-magic2l_*` (`cls-magic2.dll`)**: ProFrager's LOLZ (rANS/LZMA) decompressor. The `magic2l` variant implements LDMF (Low-memory Disk Memory File) disk swapping.
* **`rz.exe` / `rzw.exe`**: Razor decompressor by Bulat Ziganshin.
* **`cls-lollypop_x64.exe` / `_x86.exe`**: Lollypop compression codec.
* **`facompress.dll`**: FreeArc compression library helper.

### 1.3 Identified Compression Formats & Archive Payload
* **Container Overlay**: Inno Setup compressed payload format.
* **Archive Files (`fg-01.bin`, `fg-02.bin`)**: FreeArc solid archive stream containers (`ArC\x01`).
* **Solid Stream Formats**:
  * Multi-layer preprocessed streams: SREP (deduplication) + LOLZ (`magic2` / `magic2l` rANS) + Razor (`rzw`) + FreeArc internal LZMA (`lzma:mfbt4:d1m`).

### 1.4 Native Linux Tools vs. Wine Requirements
* **Prefer Native Linux Tools**:
  * **Inno Setup Unpacking**: Native Linux **`innoextract`** is prioritized over running the bundled `setup.exe` via Wine. It extracts game metadata, icons, and checksum files natively in under a second with zero Wine overhead.
  * **Checksum Verification**: Native Go MD5 engine (`internal/verifier/`) verifies archives and game files without Wine or `QuickSFV.EXE`.
* **Where Wine Remains Required**:
  * LOLZ (`cls-magic2`, `cls-magic2l`) and Razor (`rzw`) codecs are **closed-source Windows-exclusive binaries** (`.exe`). There are no source code or native Linux ports available.
  * Therefore, unpacking the FreeArc solid streams in `fg-*.bin` containing LOLZ/Razor data fundamentally requires Windows binary execution (via `Wine` on Linux and macOS).
  * Wine execution is isolated with `WINEDEBUG=-all` and targeted directly at the decompression pipeline.

### 1.5 2 GB RAM Limit Architecture
The CLI supports and enforces the 2 GB RAM limit during extraction:
* **Configuration Layer (`CLS.ini`)**:
  * `[Srep] Memory=512m` (reduced from 2048 MB).
  * `[magic2] MaxMemoryUsage=1024m` & `[magic2l] MaxMemoryUsage=1024m`.
  * `ldmfMaxMemoryUsage=1024m` with `ldmfTempPath` configured to swap decompressor dictionary to disk (`.ldmf` files) rather than exhausting host RAM.
* **OS-Level Enforcement**:
  * On Linux, the extraction process tree is constrained with `prlimit --data=2147483648` (2 GB heap ceiling) to guarantee host system stability.

---

## 2. Key Features

* **High Performance**: Native Go implementation with buffered 4 MB I/O chunks and optimized hashing.
* **Native Linux Extraction**: Prefers native `innoextract` and native Go hashing over Wine whenever compatible.
* **2 GB RAM Limit**: Enabled and enforced out-of-the-box (recommended for systems with $\le$ 8 GB RAM).
* **100% Self-Contained**: Statically linked binaries with zero runtime dependencies (no Python, no Node.js required).
* **Cross-Platform**:
  * **Windows** (`x86_64`)
  * **Linux** (`x86_64`, `ARM64 / aarch64` including **Android Termux**)
  * **macOS** (`Apple Silicon ARM64`, `Intel x86_64`)
* **Dual Operation Modes**:
  * **Interactive TUI**: Styled terminal menu with ANSI colors and easy number navigation.
  * **Headless CLI Subcommands**: Full scripting support for CI/CD (`check`, `info`, `verify-bins`, `unpack`, `verify`).
* **Real-Time Progress Tracking**: Adaptive single-line progress bar with byte count, percentage, and read speed (`MB/s`).

---

## 3. Directory Structure

```
CLIUnRepacker/
├── build.sh                  # Cross-platform build & release packaging script
├── fitgirl-cli.sh            # Universal Linux / macOS / Termux launcher script
├── fitgirl-cli.bat           # Windows Command Prompt launcher
├── fitgirl-cli.command       # macOS Finder double-clickable launcher
├── dist/                     # Cross-compiled static binaries:
│   ├── fitgirl-cli-linux-amd64
│   ├── fitgirl-cli-linux-arm64       # Linux ARM64 & Android Termux
│   ├── fitgirl-cli-windows-amd64.exe
│   ├── fitgirl-cli-darwin-arm64
│   └── fitgirl-cli-darwin-amd64
├── internal/
│   ├── system/               # OS & Termux diagnostics, REA & prlimit detection, memory checks
│   ├── ui/                   # ANSI colors, banner, interactive menu loop
│   ├── unpacker/             # Native innoextract & headless unpacker & repack inspection
│   └── verifier/             # QuickSFV / MD5 parser and progress verifier
├── main.go                   # CLI entrypoint and subcommand router
└── go.mod
```

---

## 4. Usage

### 4.1 Interactive Terminal UI

Run without arguments to start the interactive menu:

```bash
# On Linux / macOS / Termux (auto-detects architecture):
./fitgirl-cli.sh

# Or directly execute pre-compiled binary:
./dist/fitgirl-cli-linux-amd64
# (or on ARM64 / Termux):
./dist/fitgirl-cli-linux-arm64

# On Windows:
fitgirl-cli.bat
```

The interactive menu provides:
* **Header & Repack Status**: Repack metadata, archive sizes, and environment diagnostics.
* **Unpack Repack**: Destination path setup, disk calculation, 2GB low-memory enforcement toggle, and post-install verification.
* **Verify BIN Files**: Real-time progress bar checking `fg-*.bin` files against MD5 checksums.
* **Verify Installed Game**: Checks extracted game files against `fitgirl.md5`.
* **Inspect Repack**: Instant view of installer binary, archive sizes, REA reverse-engineering report, and compression formats.
* **System Diagnostics**: Displays host RAM, Wine status, innoextract availability, REA presence, and prlimit enforcer.

---

### 4.2 CLI Subcommands (Scripting & Automation)

#### System Diagnostics
```bash
./fitgirl-cli.sh check
```
Returns system OS, architecture, detected environment (including Android / Termux), detected `innoextract`, `wine`, `rea`, `prlimit`, and physical RAM in JSON format:
```json
{
  "os": "linux",
  "arch": "amd64",
  "is_termux": true,
  "environment": "Termux (Android)",
  "innoextract": "/data/data/com.termux/files/usr/bin/innoextract",
  "wine": "/data/data/com.termux/files/usr/bin/wine",
  "prlimit": "/data/data/com.termux/files/usr/bin/prlimit",
  "mem_total_gb": 7.72,
  "mem_avail_gb": 4.05
}
```

#### Inspect Repack Metadata
```bash
./fitgirl-cli.sh info [--dir /path/to/repack]
```

#### Pre-Install BIN Checksum Verification
```bash
./fitgirl-cli.sh verify-bins [--dir /path/to/repack]
```

#### Headless Extraction (with 2 GB RAM Limit)
```bash
./fitgirl-cli.sh unpack --output /path/to/destination [--limit-ram=true] [--verify]
```
Options:
* `--output, -o`: Target installation directory (required).
* `--dir`: Repack directory (default: current directory).
* `--limit-ram`: Enforce 2 GB RAM limit during extraction (default: `true`).
* `--components`: Inno Setup components to extract (optional).
* `--verify`: Automatically run post-installation MD5 check after extraction.

#### Post-Install Game Verification
```bash
./fitgirl-cli.sh verify --target /path/to/destination [--md5 /path/to/fitgirl.md5]
```

---

## 5. Technical Limitations & Platform Compatibility

### 5.1 General Limitations
1. **Proprietary Codecs (LOLZ & Razor)**:
   * Because ProFrager's LOLZ (`cls-magic2_x86.exe` / `cls-magic2_x64.exe`) and Dmitry Shelokovskiy / Bulat Ziganshin's Razor (`rzw.exe`) are closed-source Windows binaries without Linux native source or ports, FreeArc solid stream decompression cannot run 100% natively on Linux without Wine.
2. **Wine WoW64 IPC**:
   * FitGirl repacks use an IPC shared memory segment (`Init_MapFile_` with prefix `Global\`) and anonymous pipes between `unarc.dll`, `CLS-srep.dll`, and subprocesses. On modern 64-bit Wine, these require stable WoW64 thread management.
3. **RAM and CPU Demands**:
   * LOLZ decompression is CPU-intensive. When the 2 GB RAM limit is enforced, dictionary state spills to disk (`.ldmf` files), saving memory at the cost of additional disk I/O. Ensure at least 5 GB of free disk space in the destination drive.

### 5.2 Android & Termux Support Guide

CLIUnRepacker includes native **ARM64 (`aarch64`)** binaries statically compiled for Linux and Termux on Android.

#### Prerequisites & Dependencies in Termux:
Install native tools via Termux package manager:
```bash
pkg update
pkg install innoextract
```
*(Optional: If building from source inside Termux, install Go via `pkg install golang`.)*

#### Critical Storage Note:
> [!IMPORTANT]
> Always clone and execute CLIUnRepacker within Termux's internal home directory (`~/` or `$HOME`), **not** on external shared storage (`/sdcard` or `/storage/emulated/0`).
> Android mounts external storage with the `noexec` permission flag, which prohibits executing any compiled binaries or scripts directly.

#### Running in Termux:
```bash
# Ensure execution permissions
chmod +x fitgirl-cli.sh dist/fitgirl-cli-linux-arm64

# Launch interactive menu
./fitgirl-cli.sh

# Or run headless subcommands
./fitgirl-cli.sh check
./fitgirl-cli.sh verify-bins
./fitgirl-cli.sh info
```

#### What Runs 100% Natively in Termux:
* **Pre-installation MD5 Verification (`verify-bins`)**: Statically linked native ARM64 Go hashing engine runs at maximum hardware speed with zero dependencies.
* **Repack Container Inspection (`info`)**: Parses Inno Setup structure using native `innoextract`.
* **Inno Setup Payload Extraction**: Native `innoextract` extracts setup scripts, redistributables, and game assets without Wine.
* **Post-installation Game Verification (`verify`)**: Verifies unpacked files against `fitgirl.md5` natively.

#### Full Game Decompression on ARM64 Android:
Because FreeArc LOLZ (`magic2`) and Razor (`rzw`) decompressor filters are closed-source x86/x64 Windows executables, running full FreeArc stream extraction on ARM64 Android requires an x86 translation layer such as:
* **PRoot-Distro** with **Box64 / Box86 + Wine**
* **Mobox** or **Winlator** container environments

---

## 6. Building from Source & Packaging

### 6.1 Using the Build Script

A cross-platform build and packaging script `build.sh` is provided in the repository root:

```bash
# Build all cross-platform targets into dist/
./build.sh

# Build only for Linux ARM64 (Termux):
./build.sh linux-arm64

# Build only for Linux x86_64:
./build.sh linux-amd64

# Create complete release packages (.tar.gz and .zip archives in dist/):
./build.sh package

# Clean build artifacts:
./build.sh clean
```

### 6.2 Manual Cross-Compilation Commands

To manually cross-compile statically linked binaries (`CGO_ENABLED=0`):

```bash
# Linux (x86_64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-linux-amd64 .

# Linux & Termux (ARM64 / aarch64)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-linux-arm64 .

# Windows (x86_64)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-windows-amd64.exe .

# macOS (Apple Silicon ARM64)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-darwin-arm64 .

# macOS (Intel x86_64)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-darwin-amd64 .
```
