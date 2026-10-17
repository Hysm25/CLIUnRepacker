//go:build darwin

package system

import (
	"encoding/binary"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// GetMemoryInfo retrieves total physical RAM on macOS via sysctl
func GetMemoryInfo() (totalBytes uint64, availBytes uint64, err error) {
	// Attempt syscall.Sysctl
	s, err := syscall.Sysctl("hw.memsize")
	if err == nil {
		if len(s) == 8 {
			totalBytes = binary.LittleEndian.Uint64([]byte(s))
		} else {
			totalBytes, _ = strconv.ParseUint(strings.TrimSpace(s), 10, 64)
		}
	}

	// Fallback to sysctl command if needed
	if totalBytes == 0 {
		out, cmdErr := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if cmdErr == nil {
			totalBytes, _ = strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		}
	}

	return totalBytes, totalBytes / 2, nil
}
