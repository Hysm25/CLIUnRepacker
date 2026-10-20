//go:build linux

package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// GetMemoryInfo parses /proc/meminfo on Linux for total and available RAM
func GetMemoryInfo() (totalBytes uint64, availBytes uint64, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var memFree, buffers, cached uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			switch fields[0] {
			case "MemTotal:":
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					totalBytes = kb * 1024
				}
			case "MemAvailable:":
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					availBytes = kb * 1024
				}
			case "MemFree:":
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					memFree = kb * 1024
				}
			case "Buffers:":
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					buffers = kb * 1024
				}
			case "Cached:":
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					cached = kb * 1024
				}
			}
		}
	}

	// Fallback for Android/Linux kernels that lack MemAvailable in /proc/meminfo
	if availBytes == 0 && (memFree > 0 || cached > 0) {
		availBytes = memFree + buffers + cached
	}

	return totalBytes, availBytes, scanner.Err()
}
