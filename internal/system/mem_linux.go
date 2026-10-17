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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			if fields[0] == "MemTotal:" {
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					totalBytes = kb * 1024
				}
			} else if fields[0] == "MemAvailable:" {
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					availBytes = kb * 1024
				}
			}
		}
	}
	return totalBytes, availBytes, scanner.Err()
}
