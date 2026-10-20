package system

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunSystemCheck(t *testing.T) {
	info := RunSystemCheck()
	if info.OS != runtime.GOOS {
		t.Errorf("Expected OS %s, got %s", runtime.GOOS, info.OS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Expected Arch %s, got %s", runtime.GOARCH, info.Arch)
	}
	if info.MemTotalGB <= 0 {
		t.Logf("MemTotalGB reported: %f", info.MemTotalGB)
	}
}

func TestTermuxDetection(t *testing.T) {
	// Test default host environment
	initialTermux := IsTermux()

	// Emulate Termux via environment variable
	t.Setenv("TERMUX_VERSION", "0.118.0")
	if !IsTermux() {
		t.Errorf("Expected IsTermux to return true when TERMUX_VERSION is set")
	}

	info := RunSystemCheck()
	if !info.IsTermux {
		t.Errorf("Expected SysInfo.IsTermux to be true")
	}
	if info.Environment != "Termux (Android)" {
		t.Errorf("Expected Environment 'Termux (Android)', got %q", info.Environment)
	}

	// Restore / clear
	t.Setenv("TERMUX_VERSION", "")
	if IsTermux() != initialTermux {
		t.Errorf("Expected IsTermux to return %v after clearing TERMUX_VERSION", initialTermux)
	}
}

func TestGetFreeDiskSpaceNonExistentPath(t *testing.T) {
	tmpDir := t.TempDir()
	nestedNonExistent := filepath.Join(tmpDir, "does", "not", "exist", "yet")

	free, err := GetFreeDiskSpace(nestedNonExistent)
	if err != nil {
		t.Fatalf("GetFreeDiskSpace failed on non-existent path: %v", err)
	}
	if free == 0 {
		t.Errorf("Expected non-zero free disk space, got %d", free)
	}
}
