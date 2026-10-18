package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cliunrepacker/internal/system"
	"cliunrepacker/internal/ui"
	"cliunrepacker/internal/unpacker"
	"cliunrepacker/internal/verifier"
)

func printHelp() {
	helpText := fmt.Sprintf(`%sFitGirl Repack Cross-Platform CLI Tool (Go)%s

Usage: fitgirl-cli [command] [options]

Available subcommands:
  menu                  Launch interactive terminal menu (default if no args)
  check                 Run system diagnostics
  info                  Inspect repack installer and archives
  verify-bins           Verify fg-*.bin files against MD5/fitgirl-bins.md5
  unpack                Headlessly unpack the repack
  verify                Verify unpacked directory against fitgirl.md5

Options:
  -h, --help            Show this help message and exit

Run 'fitgirl-cli [command] --help' for options specific to a command.
`, ui.Bold, ui.Reset)
	fmt.Print(helpText)
}

func defaultMenuHandlers() ui.MenuHandlers {
	return ui.MenuHandlers{
		InspectRepack:      unpacker.InspectRepack,
		VerifyBins:         verifier.VerifyBins,
		UnpackRepack:       unpacker.UnpackRepack,
		VerifyInstallation: verifier.VerifyInstallation,
	}
}

func main() {
	if len(os.Args) <= 1 {
		ui.RunInteractiveMenu(".", defaultMenuHandlers())
		return
	}

	command := os.Args[1]

	switch command {
	case "menu", "wizard":
		ui.RunInteractiveMenu(".", defaultMenuHandlers())

	case "check":
		info := system.RunSystemCheck()
		data, err := json.MarshalIndent(info, "", "  ")
		if err == nil {
			fmt.Println(string(data))
		}

	case "info":
		fs := flag.NewFlagSet("info", flag.ExitOnError)
		dir := fs.String("dir", ".", "Repack directory (default: current directory)")
		_ = fs.Parse(os.Args[2:])
		absDir, _ := filepath.Abs(*dir)
		unpacker.InspectRepack(absDir)

	case "verify-bins":
		fs := flag.NewFlagSet("verify-bins", flag.ExitOnError)
		dir := fs.String("dir", ".", "Repack directory (default: current directory)")
		_ = fs.Parse(os.Args[2:])
		absDir, _ := filepath.Abs(*dir)
		verifier.VerifyBins(absDir)

	case "unpack":
		fs := flag.NewFlagSet("unpack", flag.ExitOnError)
		dir := fs.String("dir", ".", "Repack directory (default: current directory)")
		var output string
		fs.StringVar(&output, "output", "", "Destination directory for unpacked files")
		fs.StringVar(&output, "o", "", "Destination directory for unpacked files (shorthand)")
		components := fs.String("components", "", "Inno Setup components to install (optional)")
		limitRAM := fs.Bool("limit-ram", true, "Enable and enforce 2 GB RAM limit during extraction (default: true)")
		verify := fs.Bool("verify", false, "Automatically verify files after unpacking")

		_ = fs.Parse(os.Args[2:])

		if output == "" {
			fmt.Printf("%s[!] Error: -output / -o is required for unpacking.%s\n", ui.Red, ui.Reset)
			fs.Usage()
			os.Exit(1)
		}

		absDir, _ := filepath.Abs(*dir)
		absOut, _ := filepath.Abs(output)
		unpacker.UnpackRepack(absDir, absOut, *components, *limitRAM, *verify)

	case "verify":
		fs := flag.NewFlagSet("verify", flag.ExitOnError)
		var target string
		fs.StringVar(&target, "target", "", "Unpacked game directory to verify")
		fs.StringVar(&target, "t", "", "Unpacked game directory to verify (shorthand)")
		md5File := fs.String("md5", "", "Custom path to fitgirl.md5 file")

		_ = fs.Parse(os.Args[2:])

		if target == "" {
			fmt.Printf("%s[!] Error: -target / -t is required for verification.%s\n", ui.Red, ui.Reset)
			fs.Usage()
			os.Exit(1)
		}

		absTarget, _ := filepath.Abs(target)
		absMD5 := *md5File
		if absMD5 != "" {
			absMD5, _ = filepath.Abs(absMD5)
		}
		verifier.VerifyInstallation(absTarget, absMD5)

	case "-h", "--help", "help":
		printHelp()

	default:
		fmt.Printf("%sUnknown command: %s%s\n\n", ui.Red, command, ui.Reset)
		printHelp()
		os.Exit(1)
	}
}
