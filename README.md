# FitGirl Repack CLI Unpacker & Inspector (Go)

A high-performance, cross-platform interactive terminal application and headless orchestrator written in **Go** for inspecting, pre-verifying, headlessly unpacking, and post-verifying **FitGirl Repacks** across **Windows**, **Linux**, and **macOS**.

---

## About

> I'm a hobbyist developer who created this project because I had trouble handling FitGirl repacks on Linux. This project has only been tested with The Legend of Zelda: Link's Awakening [FitGirl Repack], so compatibility with other repacks is not guaranteed.

---

## 1. Reverse-Engineering Insights (via [REA](https://github.com/morluto/rea))

Deep structural analysis of `setup.exe` and associated archives (`fg-*.bin`) using [REA](https://github.com/morluto/rea), GNU objdump, and Inno Setup container parsing revealed:

1. **Binary Container**:
   * `setup.exe` is a 32-bit PE GUI executable (`i386`) packed with Inno Setup (Unicode Enhanced Edition).
   * REA section analysis (`rea inspect-with-objdump section_headers`) demonstrates the PE sections (`.text`, `.data`, `.rsrc`) conclude at offset `0x22600` (~140 KB). Everything following (over 9.6 MB) is the compressed Inno Setup payload overlay.

2. **Extracted Decompression Engine**:
   When executed, Inno Setup materializes a specialized extraction pipeline into `%TEMP%`:
   * `ISDone.dll` & `unarc.dll`: FreeArc unarc extraction coordinator.
   * `cls-magic2_x64.exe` / `_x86.exe`: ProFrager's LOLZ (rANS) decompressor.
   * `cls-srep_x64.exe` / `_x86.exe`: Super Reflation (SREP) deduplication filter.
   * `CLS.ini` & `arc.ini`: Compressor pipelines and memory usage limits.

3. **Archive Payload (`fg-*.bin`)**:
   * The `.bin` archives are FreeArc (`ArC\x01`) containers with multi-stage solid compression streams (SREP + LOLZ).
   * **Cross-Platform Compatibility**: Because LOLZ and Razor are closed-source Windows-exclusive binaries without Linux/macOS native ports, 100% native decompression without Windows binary execution is fundamentally impossible.
   * **Orchestration Solution**: This Go CLI provides native high-speed pre/post MD5 verification and container inspection, while dispatching headless, silent extraction via native Windows calls on Windows, or headless compatibility runtime (`Wine`) with low-memory parameters on Linux and macOS.

---

## 2. Key Features

* **High Performance**: Native Go implementation with buffered 4 MB I/O chunks and optimized hashing.
* **100% Self-Contained**: Statically linked binaries with zero runtime dependencies (no Python, no Node.js required).
* **Cross-Platform**:
  * **Windows** (`x86_64`)
  * **Linux** (`x86_64`, `ARM64`)
  * **macOS** (`Apple Silicon ARM64`, `Intel x86_64`)
* **Dual Operation Modes**:
  * **Interactive TUI**: Styled terminal menu with ANSI colors and easy number navigation.
  * **Headless CLI Subcommands**: Full scripting support for CI/CD and automated setups (`check`, `info`, `verify-bins`, `unpack`, `verify`).
* **Real-Time Progress Tracking**: Dynamic animated progress bar with byte count, percentage, and read speed (`MB/s`).
* **Headless Unpacker**: Silent installation bypassing GUI wizard prompts and skipping DirectX/VC++ web downloads (`/COMPONENTS="text"`).

---

## 3. Directory Structure

```
CLIUnRepacker/
├── fitgirl-cli               # Pre-compiled native Linux executable
├── fitgirl-cli.exe           # Pre-compiled native Windows executable
├── fitgirl-cli.sh            # Universal Linux launcher script
├── fitgirl-cli.bat           # Windows Command Prompt launcher
├── fitgirl-cli.command       # macOS Finder double-clickable launcher
├── dist/                     # Cross-compiled binaries:
│   ├── fitgirl-cli-linux-amd64
│   ├── fitgirl-cli-windows-amd64.exe
│   ├── fitgirl-cli-darwin-arm64
│   └── fitgirl-cli-darwin-amd64
├── internal/
│   ├── system/               # OS diagnostics, memory, and free disk checks
│   ├── ui/                   # ANSI colors, banner, interactive menu loop
│   ├── unpacker/             # Headless unpacker & repack inspection
│   └── verifier/             # QuickSFV / MD5 parser and progress verifier
├── main.go                   # CLI entrypoint and subcommand router
└── go.mod
```

---

## 4. Usage

### 4.1 Interactive Terminal UI

Simply run without arguments to start the interactive menu:

```bash
# On Linux / macOS
./fitgirl-cli

# Or using the launcher
./fitgirl-cli.sh

# On Windows
fitgirl-cli.exe
# Or double-click fitgirl-cli.bat
```

The interactive interface provides a clean, responsive, and intuitive terminal menu:

* **Header & Repack Status**: Dynamic terminal width detection with aligned repack metadata, archive sizes, and environment diagnostics.
* **Unpack Repack**: Streamlined unpack setup with destination path, free disk calculation, automatic 2GB low-memory recommendation, and post-installation verification toggle.
* **Verify BIN Files**: Real-time adaptive single-line progress bar with byte tracking, percentage, and read speed (`MB/s`).
* **Verify Installed Game**: Check integrity of extracted game files against `fitgirl.md5`.
* **Inspect Repack**: Instant view of installer executable, archive split sizes, and Inno Setup container details.
* **System Diagnostics**: Check host memory, Wine status, innoextract availability, and available storage.
* **Change Directory**: Seamlessly switch between repack folders.

---

### 4.2 CLI Subcommands (Scripting & Automation)

#### System Diagnostics
```bash
./fitgirl-cli check
```
Returns system OS, architecture, detected `innoextract`, `wine`, and physical RAM in JSON format.

#### Inspect Repack Metadata
```bash
./fitgirl-cli info [--dir /path/to/repack]
```

#### Pre-Install BIN Checksum Verification
```bash
./fitgirl-cli verify-bins [--dir /path/to/repack]
```
Verifies `fg-*.bin` files against `MD5/fitgirl-bins.md5` with active progress bar and speed tracking.

#### Headless Extraction
```bash
./fitgirl-cli unpack --output /path/to/destination [--verify]
```
Options:
* `--output, -o`: Target installation directory (required).
* `--dir`: Repack directory (default: current directory).
* `--components`: Inno Setup components to extract (default: `text`).
* `--verify`: Automatically run post-installation MD5 check after extraction.

#### Post-Install Game Verification
```bash
./fitgirl-cli verify --target /path/to/destination [--md5 /path/to/fitgirl.md5]
```

---

## 5. Building from Source

To build for your current host system:

```bash
go build -ldflags="-s -w" -o fitgirl-cli .
```

To cross-compile for all supported platforms:

```bash
# Linux (x86_64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-linux-amd64 .

# Windows (x86_64)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-windows-amd64.exe .

# macOS (Apple Silicon ARM64)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-darwin-arm64 .

# macOS (Intel x86_64)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/fitgirl-cli-darwin-amd64 .
```
