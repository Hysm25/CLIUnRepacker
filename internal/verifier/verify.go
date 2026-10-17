package verifier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cliunrepacker/internal/ui"
)

// VerifyMD5List verifies a list of entries against baseDir with progress output
func VerifyMD5List(entries []MD5Entry, baseDir string) (int, int, []string) {
	passed := 0
	failed := 0
	var errors []string
	totalFiles := len(entries)

	fmt.Printf("Total files to verify: %s%d%s\n\n", ui.Bold, totalFiles, ui.Reset)

	for idx, entry := range entries {
		target := filepath.Join(baseDir, filepath.FromSlash(entry.RelPath))
		prefix := fmt.Sprintf("[%d/%d] %s", idx+1, totalFiles, entry.RelPath)

		stat, err := os.Stat(target)
		if os.IsNotExist(err) {
			failed++
			msg := fmt.Sprintf("MISSING: %s", entry.RelPath)
			errors = append(errors, msg)
			fmt.Printf("  %s[✗] %s (File missing)%s\n", ui.Red, prefix, ui.Reset)
			continue
		}

		actualHash, err := CalculateMD5WithProgress(target, 4*1024*1024)
		if err != nil {
			failed++
			msg := fmt.Sprintf("READ_ERROR: %s (%v)", entry.RelPath, err)
			errors = append(errors, msg)
			fmt.Printf("  %s[✗] %s (Read error: %v)%s\n", ui.Red, prefix, err, ui.Reset)
			continue
		}

		if strings.EqualFold(actualHash, entry.ExpectedHash) {
			passed++
			fmt.Printf("  %s[✓] %s (%s) - OK%s\n", ui.Green, prefix, ui.FormatBytes(stat.Size()), ui.Reset)
		} else {
			failed++
			msg := fmt.Sprintf("MISMATCH: %s (got %s, expected %s)", entry.RelPath, actualHash, entry.ExpectedHash)
			errors = append(errors, msg)
			fmt.Printf("  %s[✗] %s - Checksum mismatch!%s\n", ui.Red, prefix, ui.Reset)
		}
	}

	return passed, failed, errors
}

// VerifyBins verifies fg-*.bin files in repackDir against MD5/fitgirl-bins.md5
func VerifyBins(repackDir string) {
	ui.PrintHeader("Verify BIN Files (Pre-Install Check)")

	md5File := filepath.Join(repackDir, "MD5", "fitgirl-bins.md5")
	if _, err := os.Stat(md5File); os.IsNotExist(err) {
		fmt.Printf("%s[!] Checksum file '%s' not found.%s\n", ui.Red, md5File, ui.Reset)
		return
	}

	entries, err := ParseMD5File(md5File)
	if err != nil {
		fmt.Printf("%s[!] Failed to parse '%s': %v%s\n", ui.Red, md5File, err, ui.Reset)
		return
	}
	if len(entries) == 0 {
		fmt.Printf("%s[!] No valid checksum entries found in '%s'.%s\n", ui.Yellow, md5File, ui.Reset)
		return
	}

	fmt.Printf("Loaded checksum definition: %s%s%s\n", ui.Cyan, filepath.Base(md5File), ui.Reset)
	passed, failed, _ := VerifyMD5List(entries, repackDir)

	fmt.Printf("\n%sBIN Verification Summary:%s\n", ui.Bold, ui.Reset)
	fmt.Printf("  %sPassed: %d%s\n", ui.Green, passed, ui.Reset)
	if failed > 0 {
		fmt.Printf("  %sFailed: %d%s\n", ui.Red, failed, ui.Reset)
		fmt.Printf("\n%s[!] Warning: Corrupted BIN files will cause decompression failure! Re-download bad parts.%s\n", ui.Red, ui.Reset)
	} else {
		fmt.Printf("  %sAll archive BIN files are healthy and ready for unpacking!%s\n", ui.Green, ui.Reset)
	}
}

// VerifyInstallation verifies extracted game files against fitgirl.md5
func VerifyInstallation(targetDir string, customMD5Path string) {
	ui.PrintHeader("Verify Installed Game Files")

	md5Path := customMD5Path
	if md5Path == "" {
		candidates := []string{
			filepath.Join(targetDir, "_Redist", "fitgirl.md5"),
			filepath.Join(targetDir, "fitgirl.md5"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				md5Path = c
				break
			}
		}
	}

	if md5Path == "" {
		fmt.Printf("%s[!] Could not locate 'fitgirl.md5' in '%s'.%s\n", ui.Red, targetDir, ui.Reset)
		return
	}
	if _, err := os.Stat(md5Path); os.IsNotExist(err) {
		fmt.Printf("%s[!] Checksum file '%s' does not exist.%s\n", ui.Red, md5Path, ui.Reset)
		return
	}

	entries, err := ParseMD5File(md5Path)
	if err != nil {
		fmt.Printf("%s[!] Failed to parse '%s': %v%s\n", ui.Red, md5Path, err, ui.Reset)
		return
	}
	if len(entries) == 0 {
		fmt.Printf("%s[!] No entries found in '%s'.%s\n", ui.Yellow, md5Path, ui.Reset)
		return
	}

	fmt.Printf("Verifying target files in '%s' using '%s'...\n", targetDir, filepath.Base(md5Path))
	passed, failed, errors := VerifyMD5List(entries, targetDir)

	fmt.Printf("\n%sInstallation Verification Summary:%s\n", ui.Bold, ui.Reset)
	fmt.Printf("  %sPassed: %d%s\n", ui.Green, passed, ui.Reset)
	if failed > 0 {
		fmt.Printf("  %sFailed: %d%s\n", ui.Red, failed, ui.Reset)
		maxErr := 10
		if len(errors) < maxErr {
			maxErr = len(errors)
		}
		for _, err := range errors[:maxErr] {
			fmt.Printf("    • %s\n", err)
		}
	} else {
		fmt.Printf("  %sAll installed files are verified authentic and intact!%s\n", ui.Green, ui.Reset)
	}
}
