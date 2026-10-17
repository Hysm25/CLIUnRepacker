package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cliunrepacker/internal/system"
)

// MenuHandlers defines the operational callbacks dispatched by the menu
type MenuHandlers struct {
	InspectRepack      func(repackDir string)
	VerifyBins         func(repackDir string)
	UnpackRepack       func(repackDir string, outDir string, components string, verifyAfter bool)
	VerifyInstallation func(targetDir string, md5Path string)
}

// WizardHandlers provides backward compatibility with main.go
type WizardHandlers = MenuHandlers

// RunInstallerWizard provides backward compatibility with main.go
func RunInstallerWizard(repackDir string, handlers WizardHandlers) {
	RunInteractiveMenu(repackDir, handlers)
}

// InteractiveMenu provides backward compatibility
func InteractiveMenu(repackDir string, handlers MenuHandlers) {
	RunInteractiveMenu(repackDir, handlers)
}

// RunInteractiveMenu launches the clean, responsive terminal interface
func RunInteractiveMenu(initialRepackDir string, handlers MenuHandlers) {
	currentRepackDir, err := filepath.Abs(initialRepackDir)
	if err != nil {
		currentRepackDir = initialRepackDir
	}

	for {
		PrintBanner()
		meta := LoadRepackMeta(currentRepackDir)
		PrintRepackSummary(meta, currentRepackDir)

		fmt.Println("  " + Bold + "Actions:" + Reset)
		fmt.Printf("    %s[1]%s %s%-24s%s %sExtract game files to destination%s\n", Cyan, Reset, Bold, "Unpack Repack", Reset, Dim, Reset)
		fmt.Printf("    %s[2]%s %s%-24s%s %sPre-installation MD5 check on archives%s\n", Cyan, Reset, Bold, "Verify BIN Files", Reset, Dim, Reset)
		fmt.Printf("    %s[3]%s %s%-24s%s %sPost-installation integrity verification%s\n", Cyan, Reset, Bold, "Verify Installed Game", Reset, Dim, Reset)
		fmt.Printf("    %s[4]%s %s%-24s%s %sView setup metadata, archives & compression%s\n", Cyan, Reset, Bold, "Inspect Repack Details", Reset, Dim, Reset)
		fmt.Printf("    %s[5]%s %s%-24s%s %sCheck OS, Wine, RAM & storage space%s\n", Cyan, Reset, Bold, "System Diagnostics", Reset, Dim, Reset)
		fmt.Printf("    %s[6]%s %s%-24s%s %sSelect a different repack folder%s\n", Cyan, Reset, Bold, "Change Directory", Reset, Dim, Reset)
		fmt.Printf("    %s[0]%s %s%-24s%s %sExit application%s\n", Cyan, Reset, Bold, "Exit", Reset, Dim, Reset)

		fmt.Println()
		choice := strings.ToLower(PromptInput("Select option [0-6] (q to exit)", ""))

		switch choice {
		case "1":
			promptUnpack(currentRepackDir, meta, handlers)
		case "2":
			handlers.VerifyBins(currentRepackDir)
			Pause()
		case "3":
			defaultTarget := filepath.Join(currentRepackDir, "unpacked_game")
			target := PromptInput("Installed game directory to verify", defaultTarget)
			absTarget, err := filepath.Abs(target)
			if err == nil {
				target = absTarget
			}
			handlers.VerifyInstallation(target, "")
			Pause()
		case "4":
			handlers.InspectRepack(currentRepackDir)
			Pause()
		case "5":
			showSystemDiagnostics(currentRepackDir)
			Pause()
		case "6":
			newDir := PromptInput("Enter repack directory path", currentRepackDir)
			absNewDir, err := filepath.Abs(newDir)
			if err == nil {
				if stat, err := os.Stat(absNewDir); err == nil && stat.IsDir() {
					currentRepackDir = absNewDir
					fmt.Printf("\n%s[✓] Repack directory updated to: %s%s\n", Green, currentRepackDir, Reset)
				} else {
					fmt.Printf("\n%s[!] Directory does not exist or is not accessible: %s%s\n", Red, newDir, Reset)
				}
			}
			Pause()
		case "0", "q", "exit", "quit":
			fmt.Printf("\n%sGoodbye!%s\n\n", Cyan, Reset)
			return
		default:
			if choice != "" {
				fmt.Printf("\n%s[!] Invalid option: %s. Please enter a number between 0 and 6.%s\n", Yellow, choice, Reset)
			}
		}
	}
}

func promptUnpack(repackDir string, meta RepackMeta, handlers MenuHandlers) {
	PrintHeader("Unpack Repack")

	setupExe := filepath.Join(repackDir, "setup.exe")
	if _, err := os.Stat(setupExe); os.IsNotExist(err) {
		fmt.Printf("%s[!] Cannot unpack: 'setup.exe' was not found in '%s'%s\n", Red, repackDir, Reset)
		Pause()
		return
	}

	outDir := filepath.Join(repackDir, "unpacked_game")
	sysInfo := system.RunSystemCheck()

	limitRAM := sysInfo.MemTotalGB > 0 && sysInfo.MemTotalGB <= 8.5
	postVerify := true
	components := "text"

	for {
		PrintHeader("Unpack Configuration")
		fmt.Printf("  %sTarget Repack:%s %s%s%s\n", Dim, Reset, Bold, meta.AppName, Reset)
		fmt.Printf("  %sSource Path:%s   %s\n", Dim, Reset, repackDir)
		fmt.Printf("  %sDestination:%s   %s\n", Dim, Reset, outDir)

		if freeSpace, err := system.GetFreeDiskSpace(outDir); err == nil {
			fmt.Printf("  %sAvailable Disk:%s %s\n", Dim, Reset, FormatBytes(int64(freeSpace)))
		}

		ramText := fmt.Sprintf("%sDisabled%s", Dim, Reset)
		if limitRAM {
			ramText = fmt.Sprintf("%sEnabled%s (recommended for <= 8GB RAM)", Green, Reset)
		}
		fmt.Printf("  %s2GB RAM Limit:%s %s\n", Dim, Reset, ramText)

		verifyText := fmt.Sprintf("%sEnabled%s", Green, Reset)
		if !postVerify {
			verifyText = fmt.Sprintf("%sDisabled%s", Dim, Reset)
		}
		fmt.Printf("  %sPost-Verify:%s   %s\n", Dim, Reset, verifyText)
		fmt.Printf("  %sComponents:%s    %s (core game files)\n\n", Dim, Reset, components)

		fmt.Printf("  %s[Enter]%s Start Unpacking   %s[C]%s Configure Options   %s[B]%s Back\n",
			Cyan, Reset, Cyan, Reset, Cyan, Reset)
		action := strings.ToLower(PromptInput("Choice", ""))

		if action == "" || action == "y" || action == "yes" || action == "start" {
			handlers.UnpackRepack(repackDir, outDir, components, postVerify)
			Pause()
			return
		} else if action == "c" {
			fmt.Println()
			outDir = PromptInput("Destination directory", outDir)
			limitRAM = PromptConfirm("Limit installer to 2GB RAM usage", limitRAM)
			postVerify = PromptConfirm("Verify files integrity after installation", postVerify)
			components = PromptInput("Inno Setup components to extract", components)
		} else if action == "b" || action == "back" || action == "q" {
			return
		}
	}
}

func showSystemDiagnostics(repackDir string) {
	PrintHeader("System Diagnostics")

	sysInfo := system.RunSystemCheck()

	fmt.Printf("  %sOperating System:%s   %s (%s)\n", Dim, Reset, sysInfo.OS, sysInfo.Arch)
	if sysInfo.MemTotalGB > 0 {
		fmt.Printf("  %sHost Memory:%s        %.2f GB total (%.2f GB available)\n",
			Dim, Reset, sysInfo.MemTotalGB, sysInfo.MemAvailGB)
	}

	if sysInfo.Wine != "" {
		fmt.Printf("  %sWine Compatibility:%s %s%s%s (Ready)\n", Dim, Reset, Green, sysInfo.Wine, Reset)
	} else if sysInfo.OS != "windows" {
		fmt.Printf("  %sWine Compatibility:%s %sNot found in PATH%s (Required for unpack on Linux/macOS)\n",
			Dim, Reset, Yellow, Reset)
	}

	if sysInfo.Innoextract != "" {
		fmt.Printf("  %sInnoextract Tool:%s   %s%s%s (Available)\n", Dim, Reset, Green, sysInfo.Innoextract, Reset)
	} else {
		fmt.Printf("  %sInnoextract Tool:%s   %sNot found%s (Optional for container inspection)\n",
			Dim, Reset, Dim, Reset)
	}

	if freeSpace, err := system.GetFreeDiskSpace(repackDir); err == nil {
		fmt.Printf("  %sStorage Free Space:%s %s on current drive\n", Dim, Reset, FormatBytes(int64(freeSpace)))
	}
}
