package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// SysInfo represents the system diagnostics result
type SysInfo struct {
	OS          string  `json:"os"`
	Arch        string  `json:"arch"`
	IsTermux    bool    `json:"is_termux,omitempty"`
	IsAndroid   bool    `json:"is_android,omitempty"`
	Environment string  `json:"environment,omitempty"`
	Innoextract string  `json:"innoextract,omitempty"`
	Wine        string  `json:"wine,omitempty"`
	REA         string  `json:"rea,omitempty"`
	Prlimit     string  `json:"prlimit,omitempty"`
	MemTotalGB  float64 `json:"mem_total_gb"`
	MemAvailGB  float64 `json:"mem_avail_gb"`
}

// IsTermux detects whether the process is executing inside Termux on Android
func IsTermux() bool {
	if os.Getenv("TERMUX_VERSION") != "" || os.Getenv("TERMUX_MAIN_PACKAGE_NAME") != "" {
		return true
	}
	prefix := os.Getenv("PREFIX")
	if strings.Contains(prefix, "com.termux") {
		return true
	}
	if _, err := os.Stat("/data/data/com.termux/files/usr"); err == nil {
		if _, err := os.Stat("/system/bin/app_process"); err == nil {
			return true
		}
	}
	return false
}

// IsAndroid detects if running on the Android platform (natively or under Termux)
func IsAndroid() bool {
	if runtime.GOOS == "android" {
		return true
	}
	if os.Getenv("ANDROID_ROOT") != "" || os.Getenv("ANDROID_DATA") != "" {
		return true
	}
	if _, err := os.Stat("/system/bin/sh"); err == nil {
		return true
	}
	return IsTermux()
}

// RunSystemCheck checks dependencies, OS, arch, environment, and system memory
func RunSystemCheck() SysInfo {
	info := SysInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		IsTermux:  IsTermux(),
		IsAndroid: IsAndroid(),
	}

	if info.IsTermux {
		info.Environment = "Termux (Android)"
	} else if info.IsAndroid {
		info.Environment = "Android"
	}

	termuxPrefix := os.Getenv("PREFIX")
	if termuxPrefix == "" && info.IsTermux {
		termuxPrefix = "/data/data/com.termux/files/usr"
	}

	// Find innoextract
	if p, err := exec.LookPath("innoextract"); err == nil {
		info.Innoextract = p
	} else {
		// Check Termux prefix
		if termuxPrefix != "" {
			tInno := filepath.Join(termuxPrefix, "bin", "innoextract")
			if _, err := os.Stat(tInno); err == nil {
				info.Innoextract = tInno
			}
		}
		if info.Innoextract == "" {
			// Check ~/.local/bin/innoextract
			home, err := os.UserHomeDir()
			if err == nil {
				localInno := filepath.Join(home, ".local", "bin", "innoextract")
				if _, err := os.Stat(localInno); err == nil {
					info.Innoextract = localInno
				}
			}
		}
	}

	// Find REA (Reverse Engineering Assistant)
	if p, err := exec.LookPath("rea"); err == nil {
		info.REA = p
	} else {
		if termuxPrefix != "" {
			tREA := filepath.Join(termuxPrefix, "bin", "rea")
			if _, err := os.Stat(tREA); err == nil {
				info.REA = tREA
			}
		}
		if info.REA == "" {
			home, err := os.UserHomeDir()
			if err == nil {
				localREA := filepath.Join(home, ".local", "bin", "rea")
				if _, err := os.Stat(localREA); err == nil {
					info.REA = localREA
				}
			}
		}
	}

	// Find prlimit for memory limit enforcement on Linux
	if p, err := exec.LookPath("prlimit"); err == nil {
		info.Prlimit = p
	} else if termuxPrefix != "" {
		tPrlimit := filepath.Join(termuxPrefix, "bin", "prlimit")
		if _, err := os.Stat(tPrlimit); err == nil {
			info.Prlimit = tPrlimit
		}
	}

	// Find wine if not Windows
	if runtime.GOOS != "windows" {
		if p, err := exec.LookPath("wine"); err == nil {
			info.Wine = p
		} else if termuxPrefix != "" {
			tWine := filepath.Join(termuxPrefix, "bin", "wine")
			if _, err := os.Stat(tWine); err == nil {
				info.Wine = tWine
			}
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
