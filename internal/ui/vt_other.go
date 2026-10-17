//go:build !windows

package ui

func InitTerminal() {
	// No-op on POSIX/Unix systems
}
