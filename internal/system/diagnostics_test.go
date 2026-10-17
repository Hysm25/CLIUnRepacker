package system

import (
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
