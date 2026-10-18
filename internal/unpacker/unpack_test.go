package unpacker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectRepackMissing(t *testing.T) {
	tmpDir := t.TempDir()
	// Inspecting empty directory without setup.exe should not panic
	InspectRepack(tmpDir)
}

func TestInspectRepackWithMock(t *testing.T) {
	tmpDir := t.TempDir()
	setupFile := filepath.Join(tmpDir, "setup.exe")
	if err := os.WriteFile(setupFile, []byte("MOCK_EXE_DATA"), 0644); err != nil {
		t.Fatalf("Failed to create mock setup.exe: %v", err)
	}

	binFile := filepath.Join(tmpDir, "fg-01.bin")
	if err := os.WriteFile(binFile, []byte("ArC\x01\x00\x00\x06\x07"), 0644); err != nil {
		t.Fatalf("Failed to create mock fg-01.bin: %v", err)
	}

	InspectRepack(tmpDir)
}

func TestUnpackRepackMissingSetup(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output")
	// Should cleanly report error without panic
	UnpackRepack(tmpDir, outDir, "", true, false)
}
