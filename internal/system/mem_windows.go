//go:build windows

package system

import (
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

var (
	kernel32Dll          = syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx = kernel32Dll.NewProc("GlobalMemoryStatusEx")
)

// GetMemoryInfo retrieves total and available physical RAM on Windows
func GetMemoryInfo() (totalBytes uint64, availBytes uint64, err error) {
	var status memoryStatusEx
	status.Length = uint32(unsafe.Sizeof(status))

	r1, _, callErr := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if r1 == 0 {
		return 0, 0, callErr
	}
	return status.TotalPhys, status.AvailPhys, nil
}
