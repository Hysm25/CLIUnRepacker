//go:build !windows

package system

import (
	"os"
	"path/filepath"
	"syscall"
)

// GetFreeDiskSpace returns available disk space in bytes for POSIX filesystems
func GetFreeDiskSpace(path string) (uint64, error) {
	// If path doesn't exist yet, traverse up to the nearest existing directory
	checkPath := path
	for checkPath != "" && checkPath != "." && checkPath != "/" {
		if _, err := os.Stat(checkPath); err == nil {
			break
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			break
		}
		checkPath = parent
	}
	if checkPath == "" {
		checkPath = "."
	}

	var stat syscall.Statfs_t
	err := syscall.Statfs(checkPath, &stat)
	if err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}
