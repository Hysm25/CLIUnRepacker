package unpacker

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cliunrepacker/internal/system"
	"cliunrepacker/internal/ui"
	"cliunrepacker/internal/verifier"
)

// UnpackRepack executes headless extraction with live status tracking
func UnpackRepack(repackDir string, outDir string, components string, verifyAfter bool) {
	ui.PrintHeader("Headless Repack Unpacking")

	setupExe := filepath.Join(repackDir, "setup.exe")
	if _, err := os.Stat(setupExe); os.IsNotExist(err) {
		fmt.Printf("%s[!] 'setup.exe' not found in '%s'.%s\n", ui.Red, repackDir, ui.Reset)
		return
	}

	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		absOutDir = outDir
	}

	if err := os.MkdirAll(absOutDir, 0755); err != nil {
		fmt.Printf("%s[!] Failed to create output directory '%s': %v%s\n", ui.Red, absOutDir, err, ui.Reset)
		return
	}

	if components == "" {
		components = "text"
	}

	fmt.Printf("%sRepack Directory:%s %s\n", ui.Bold, ui.Reset, repackDir)
	fmt.Printf("%sOutput Directory:%s %s\n", ui.Bold, ui.Reset, absOutDir)
	fmt.Printf("%sComponents:%s       %s (isolates game payload, skipping redistributable downloads)\n", ui.Bold, ui.Reset, components)

	// Check free disk space
	if freeSpace, err := system.GetFreeDiskSpace(absOutDir); err == nil {
		fmt.Printf("%sFree Disk Space:%s  %s\n", ui.Bold, ui.Reset, ui.FormatBytes(int64(freeSpace)))
		if freeSpace < 5*1024*1024*1024 {
			fmt.Printf("%s[!] Warning: Less than 5 GB free space available. Decompression may require more room.%s\n", ui.Yellow, ui.Reset)
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(
			setupExe,
			fmt.Sprintf(`/DIR="%s"`, absOutDir),
			fmt.Sprintf(`/COMPONENTS="%s"`, components),
			"/VERYSILENT",
			"/SUPPRESSMSGBOXES",
			"/NORESTART",
			"/SP-",
		)
	} else {
		wineBin, err := exec.LookPath("wine")
		if err != nil {
			fmt.Printf("\n%s[!] Error: 'wine' is required to run the decompression pipeline on %s.%s\n", ui.Red, runtime.GOOS, ui.Reset)
			fmt.Println("Please install Wine via your package manager before running extraction.")
			return
		}

		wineTarget := "Z:" + filepath.ToSlash(absOutDir)
		cmd = exec.Command(
			wineBin,
			setupExe,
			"/DIR="+wineTarget,
			"/COMPONENTS="+components,
			"/VERYSILENT",
			"/SUPPRESSMSGBOXES",
			"/NORESTART",
			"/SP-",
		)
		cmd.Env = append(os.Environ(), "WINEDEBUG=-all")
	}

	cmd.Dir = repackDir

	fmt.Printf("\n%sLaunching headless extraction process...%s\n", ui.Cyan, ui.Reset)
	tStart := time.Now()

	if err := cmd.Start(); err != nil {
		fmt.Printf("%s[!] Failed to launch unpacker: %v%s\n", ui.Red, err, ui.Reset)
		return
	}

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinIdx := 0

	fmt.Printf("%sUnpacking is running in background (FreeArc / SREP / LOLZ active)...%s\n", ui.Dim, ui.Reset)

	// Channel to signal process completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
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
			os.Stdout.WriteString(status)
		}
	}

	os.Stdout.WriteString("\r" + strings.Repeat(" ", 80) + "\r")
	totalDuration := time.Since(tStart).Seconds()

	if procErr == nil {
		fmt.Printf("%s[✓] Repack unpacking completed successfully in %s!%s\n", ui.Green, ui.FormatSeconds(totalDuration), ui.Reset)
	} else {
		fmt.Printf("%s[!] Installer process exited with error: %v (%s elapsed).%s\n", ui.Yellow, procErr, ui.FormatSeconds(totalDuration), ui.Reset)
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
			fmt.Printf("\nFound verification reference: %s%s%s\n", ui.Cyan, filepath.Base(foundMD5), ui.Reset)
			verifier.VerifyInstallation(absOutDir, foundMD5)
		} else {
			fmt.Printf("\n%sNote: No 'fitgirl.md5' file detected in output directory.%s\n", ui.Dim, ui.Reset)
		}
	}
}
