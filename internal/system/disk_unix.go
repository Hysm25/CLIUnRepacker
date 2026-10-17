//go:build !windows

package system

import "syscall"

// GetFreeDiskSpace returns available disk space in bytes for POSIX filesystems
func GetFreeDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}
