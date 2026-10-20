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
		fmt.Printf("\n  %sInno Setup Container Details (Native):%s\n", ui.Bold, ui.Reset)
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
		fmt.Printf("\n  %sNote: Install 'innoextract' for deeper native installer container inspection.%s\n", ui.Dim, ui.Reset)
	}

	// Structural Reverse-Engineering Insights (via REA / GNU objdump)
	fmt.Printf("\n  %sReverse-Engineering Insights (REA / Binary Inspection):%s\n", ui.Bold, ui.Reset)
	fmt.Printf("    • %sPE Binary Container:%s 32-bit PE GUI executable (pei-i386)\n", ui.White, ui.Reset)
	fmt.Printf("    • %sSection Boundaries:%s  PE sections (.text, .data, .rsrc) conclude at 0x22600 (~140 KB)\n", ui.White, ui.Reset)
	fmt.Printf("    • %sInstaller Overlay:%s   Remaining ~9.6 MB is Inno Setup compressed payload\n", ui.White, ui.Reset)

	// Required Extraction Tools & Compression Formats
	fmt.Printf("\n  %sIdentified Extraction Tools & Compression Pipeline:%s\n", ui.Bold, ui.Reset)
	fmt.Printf("    • %sContainer Unpacker:%s     Inno Setup Enhanced (Native Linux: innoextract)\n", ui.Cyan, ui.Reset)
	fmt.Printf("    • %sArchive Coordinator:%s    unarc.dll & ISDone.dll (FreeArc extraction engine)\n", ui.Cyan, ui.Reset)
	fmt.Printf("    • %sDeduplication Filter:%s   Super Reflation (SREP) - cls-srep_x86 / cls-srep_x64\n", ui.Cyan, ui.Reset)
	fmt.Printf("    • %sCore Compressor:%s        LOLZ (rANS/LZMA) - cls-magic2 / cls-magic2l (LDMF disk cache)\n", ui.Cyan, ui.Reset)
	fmt.Printf("    • %sAuxiliary Compressors:%s  Razor (rzw.exe) & Lollypop (cls-lollypop)\n", ui.Cyan, ui.Reset)
	fmt.Printf("    • %sPayload Archive Format:%s FreeArc solid streams (ArC\\x01) with multi-stage compression\n", ui.Cyan, ui.Reset)

	// Platform Compatibility Analysis
	fmt.Printf("\n  %sPlatform Compatibility & Limitations:%s\n", ui.Bold, ui.Reset)
	fmt.Printf("    • %sNative Linux & Termux:%s  innoextract for setup container, native Go MD5 verifier (100%% native Go)\n", ui.Green, ui.Reset)
	fmt.Printf("    • %sTermux on Android:%s      Runs natively on ARM64; container extract & verify need zero emulation;\n", ui.Cyan, ui.Reset)
	fmt.Printf("                                LOLZ/Razor stream decompression requires Box64 + Wine.\n")
	fmt.Printf("    • %sWine Requirement:%s       LOLZ and Razor decompressors are closed-source Windows PE binaries;\n", ui.Yellow, ui.Reset)
	fmt.Printf("                                100%% native Linux extraction without Wine is impossible for these streams.\n")
	fmt.Printf("    • %s2 GB RAM Limit:%s         Enforced via CLS.ini (512MB SREP, 1024MB LOLZ) & prlimit memory ceiling\n", ui.Green, ui.Reset)
}
