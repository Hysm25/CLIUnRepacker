package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SysInfo represents the system diagnostics result
type SysInfo struct {
	OS          string  `json:"os"`
	Arch        string  `json:"arch"`
	Innoextract string  `json:"innoextract"`
	Wine        string  `json:"wine"`
	MemTotalGB  float64 `json:"mem_total_gb"`
	MemAvailGB  float64 `json:"mem_avail_gb"`
}

// RunSystemCheck checks dependencies, OS, arch, and system memory
func RunSystemCheck() SysInfo {
	info := SysInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Find innoextract
	if p, err := exec.LookPath("innoextract"); err == nil {
		info.Innoextract = p
	} else {
		// Check ~/.local/bin/innoextract
		home, err := os.UserHomeDir()
		if err == nil {
			localInno := filepath.Join(home, ".local", "bin", "innoextract")
			if _, err := os.Stat(localInno); err == nil {
				info.Innoextract = localInno
			}
		}
	}

	// Find wine if not Windows
	if runtime.GOOS != "windows" {
		if p, err := exec.LookPath("wine"); err == nil {
			info.Wine = p
		}
	}

	// Memory detection
	total, avail, err := GetMemoryInfo()
	if err == nil && total > 0 {
		info.MemTotalGB = float64(total) / (1024 * 1024 * 1024)
		info.MemAvailGB = float64(avail) / (1024 * 1024 * 1024)
	}

	return info
}
