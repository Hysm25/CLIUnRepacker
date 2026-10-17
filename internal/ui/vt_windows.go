//go:build windows

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

func InitTerminal() {
	var (
		kernel32       = syscall.NewLazyDLL("kernel32.dll")
		getConsoleMode = kernel32.NewProc("GetConsoleMode")
		setConsoleMode = kernel32.NewProc("SetConsoleMode")
	)
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	r1, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r1 != 0 {
		const enableVirtualTerminalProcessing = 0x0004
		setConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
	}
}
