package unpacker

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"cliunrepacker/internal/system"
	"cliunrepacker/internal/ui"
	"cliunrepacker/internal/verifier"
)

// UnpackRepack executes headless extraction with live status tracking
func UnpackRepack(repackDir string, outDir string, components string, limitRAM bool, verifyAfter bool) {
	ui.PrintHeader("Repack Unpacker")

	setupExe := filepath.Join(repackDir, "setup.exe")
	if _, err := os.Stat(setupExe); os.IsNotExist(err) {
		fmt.Printf("  %s[!] 'setup.exe' was not found in '%s'%s\n", ui.Red, repackDir, ui.Reset)
		return
	}

	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		absOutDir = outDir
	}

	if err := os.MkdirAll(absOutDir, 0755); err != nil {
		fmt.Printf("  %s[!] Failed to create output directory '%s': %v%s\n", ui.Red, absOutDir, err, ui.Reset)
		return
	}

	fmt.Printf("  %sRepack Directory:%s %s\n", ui.Dim, ui.Reset, repackDir)
	fmt.Printf("  %sDestination:%s      %s\n", ui.Dim, ui.Reset, absOutDir)
	if components != "" && components != "text" {
		fmt.Printf("  %sComponents:%s       %s\n", ui.Dim, ui.Reset, components)
	}

	ramStatus := fmt.Sprintf("%sDisabled%s", ui.Dim, ui.Reset)
	if limitRAM {
		ramStatus = fmt.Sprintf("%sEnabled & Enforced%s (2 GB memory ceiling)", ui.Green, ui.Reset)
	}
	fmt.Printf("  %s2GB RAM Limit:%s    %s\n", ui.Dim, ui.Reset, ramStatus)

	// Check free disk space
	if freeSpace, err := system.GetFreeDiskSpace(absOutDir); err == nil {
		fmt.Printf("  %sAvailable Disk:%s   %s\n", ui.Dim, ui.Reset, ui.FormatBytes(int64(freeSpace)))
		if freeSpace < 5*1024*1024*1024 {
			fmt.Printf("  %s[!] Warning: Less than 5 GB free disk space available.%s\n", ui.Yellow, ui.Reset)
		}
	}

	sysInfo := system.RunSystemCheck()

	// 1. Prefer native Linux extraction for Inno Setup container files (avoiding Wine where possible)
	if sysInfo.Innoextract != "" {
		fmt.Printf("  %sNative Extractor:%s %s (Extracting container payload without Wine)\n", ui.Dim, ui.Reset, filepath.Base(sysInfo.Innoextract))
		extractCmd := exec.Command(sysInfo.Innoextract, "-e", setupExe, "-d", absOutDir, "-m")
		extractCmd.Dir = repackDir
		if err := extractCmd.Run(); err == nil {
			// Innoextract places files under app/; promote them to destination root
			appDir := filepath.Join(absOutDir, "app")
			if stat, err := os.Stat(appDir); err == nil && stat.IsDir() {
				_ = filepath.WalkDir(appDir, func(p string, d fs.DirEntry, err error) error {
					if err == nil && !d.IsDir() {
						rel, _ := filepath.Rel(appDir, p)
						dest := filepath.Join(absOutDir, rel)
						_ = os.MkdirAll(filepath.Dir(dest), 0755)
						_ = os.Rename(p, dest)
					}
					return nil
				})
				_ = os.RemoveAll(appDir)
			}
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		args := []string{
			fmt.Sprintf(`/DIR="%s"`, absOutDir),
		}
		if components != "" && components != "text" {
			args = append(args, fmt.Sprintf(`/COMPONENTS="%s"`, components))
		}
		args = append(args, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/SP-")
		cmd = exec.Command(setupExe, args...)
	} else {
		wineBin := sysInfo.Wine
		if wineBin == "" {
			fmt.Printf("\n  %s[!] Error: 'wine' is required to run the decompression pipeline on %s.%s\n", ui.Red, runtime.GOOS, ui.Reset)
			fmt.Println("  Please install Wine via your package manager before running extraction.")
			return
		}

		wineTarget := "Z:" + filepath.FromSlash(filepath.Clean(absOutDir))
		args := []string{
			setupExe,
			"/DIR=" + wineTarget,
		}
		if components != "" && components != "text" {
			args = append(args, "/COMPONENTS="+components)
		}
		args = append(args, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/SP-")

		if limitRAM && sysInfo.Prlimit != "" {
			// Enforce 2 GB (2147483648 bytes) memory limit on process tree
			prlimitArgs := append([]string{"--data=2147483648", wineBin}, args...)
			cmd = exec.Command(sysInfo.Prlimit, prlimitArgs...)
		} else {
			cmd = exec.Command(wineBin, args...)
		}
		cmd.Env = append(os.Environ(), "WINEDEBUG=-all")
	}

	cmd.Dir = repackDir

	fmt.Printf("\n  %sLaunching decompression pipeline (FreeArc / SREP / LOLZ)...%s\n\n", ui.Cyan, ui.Reset)
	tStart := time.Now()

	if err := cmd.Start(); err != nil {
		fmt.Printf("  %s[!] Failed to launch unpacker: %v%s\n", ui.Red, err, ui.Reset)
		return
	}

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinIdx := 0

	// Channel to signal process completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	var procErr error
	running := true

	for running {
		select {
		case procErr = <-done:
			running = false
		case <-ticker.C:
			spinChar := spinner[spinIdx%len(spinner)]
			spinIdx++
			elapsed := time.Since(tStart).Seconds()

			// Count current files and bytes in absOutDir
			var fileCount int
			var writtenBytes int64
			_ = filepath.WalkDir(absOutDir, func(path string, d fs.DirEntry, err error) error {
				if err == nil && d != nil && !d.IsDir() {
					fileCount++
					if info, err := d.Info(); err == nil {
						writtenBytes += info.Size()
					}
				}
				return nil
			})

			status := fmt.Sprintf("\r  %s%s%s Elapsed: %s%s%s | Extracted: %s%d files%s (%s)",
				ui.Cyan, spinChar, ui.Reset,
				ui.Yellow, ui.FormatSeconds(elapsed), ui.Reset,
				ui.White, fileCount, ui.Reset,
				ui.FormatBytes(writtenBytes),
			)
			ui.ClearLine()
			os.Stdout.WriteString(status)
		}
	}

	ui.ClearLine()
	totalDuration := time.Since(tStart).Seconds()

	if procErr == nil {
		fmt.Printf("  %s[✓] Unpacking completed successfully in %s!%s\n", ui.Green, ui.FormatSeconds(totalDuration), ui.Reset)
	} else {
		fmt.Printf("  %s[!] Installer process exited: %v (%s elapsed)%s\n", ui.Yellow, procErr, ui.FormatSeconds(totalDuration), ui.Reset)
	}

	// Post-unpack verification if requested
	if verifyAfter {
		md5Candidates := []string{
			filepath.Join(absOutDir, "_Redist", "fitgirl.md5"),
			filepath.Join(absOutDir, "fitgirl.md5"),
		}
		var foundMD5 string
		for _, m := range md5Candidates {
			if _, err := os.Stat(m); err == nil {
				foundMD5 = m
				break
			}
		}

		if foundMD5 != "" {
			fmt.Printf("\n  %sFound post-install checksum file:%s %s\n", ui.Dim, ui.Reset, filepath.Base(foundMD5))
			verifier.VerifyInstallation(absOutDir, foundMD5)
		} else {
			fmt.Printf("\n  %sNote: No 'fitgirl.md5' file detected in output directory for post-verification.%s\n", ui.Dim, ui.Reset)
		}
	}
}
