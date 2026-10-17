package unpacker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cliunrepacker/internal/system"
	"cliunrepacker/internal/ui"
)

// InspectRepack scans and inspects repack files in repackDir
func InspectRepack(repackDir string) {
	ui.PrintHeader("Repack Inspection")

	setupExe := filepath.Join(repackDir, "setup.exe")
	stat, err := os.Stat(setupExe)
	if os.IsNotExist(err) {
		fmt.Printf("  %s[!] Error: 'setup.exe' was not found in '%s'%s\n", ui.Red, repackDir, ui.Reset)
		return
	}

	fmt.Printf("  %sRepack Directory:%s %s\n", ui.Dim, ui.Reset, repackDir)
	fmt.Printf("  %sInstaller Binary:%s %s (%s)\n", ui.Dim, ui.Reset, filepath.Base(setupExe), ui.FormatBytes(stat.Size()))

	// Locate bin archives
	bins, _ := filepath.Glob(filepath.Join(repackDir, "fg-*.bin"))
	if len(bins) > 0 {
		var totalBinSize int64
		for _, b := range bins {
			if s, err := os.Stat(b); err == nil {
				totalBinSize += s.Size()
			}
		}
		fmt.Printf("\n  %sBIN Archives (%d files, %s total):%s\n", ui.Bold, len(bins), ui.FormatBytes(totalBinSize), ui.Reset)
		for _, b := range bins {
			if s, err := os.Stat(b); err == nil {
				fmt.Printf("    • %-16s %s\n", filepath.Base(b), ui.FormatBytes(s.Size()))
			}
		}
	} else {
		fmt.Printf("  %s[!] No 'fg-*.bin' archive files detected.%s\n", ui.Yellow, ui.Reset)
	}

	// Probe setup.exe with innoextract if available
	sysInfo := system.RunSystemCheck()
	if sysInfo.Innoextract != "" {
		fmt.Printf("\n  %sInno Setup Container Details:%s\n", ui.Bold, ui.Reset)
		cmd := exec.Command(sysInfo.Innoextract, "-i", setupExe)
		cmd.Dir = repackDir
		out, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			count := 0
			for scanner.Scan() && count < 15 {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					fmt.Printf("    %s%s%s\n", ui.Dim, line, ui.Reset)
					count++
				}
			}
		} else {
			fmt.Printf("    %s[!] innoextract probe note: %v%s\n", ui.Yellow, err, ui.Reset)
		}
	} else {
		fmt.Printf("\n  %sNote: Install 'innoextract' for deeper installer container inspection.%s\n", ui.Dim, ui.Reset)
	}
}
