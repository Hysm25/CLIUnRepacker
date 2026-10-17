//go:build !linux && !windows && !darwin

package system

// GetMemoryInfo fallback for unsupported operating systems
func GetMemoryInfo() (totalBytes uint64, availBytes uint64, err error) {
	return 0, 0, nil
}
