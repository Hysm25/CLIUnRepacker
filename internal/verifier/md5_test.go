package verifier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMD5File(t *testing.T) {
	content := `; QuickSFV test file
6bd2d927e1deb25a4d094e3a44abd174 *..\fg-01.bin
07afd147d334d7ec9e4b51c151cb9b8d *fg-02.bin
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md5")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	entries, err := ParseMD5File(filePath)
	if err != nil {
		t.Fatalf("ParseMD5File failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].ExpectedHash != "6bd2d927e1deb25a4d094e3a44abd174" || entries[0].RelPath != "fg-01.bin" {
		t.Errorf("Unexpected entry 0: %+v", entries[0])
	}
	if entries[1].ExpectedHash != "07afd147d334d7ec9e4b51c151cb9b8d" || entries[1].RelPath != "fg-02.bin" {
		t.Errorf("Unexpected entry 1: %+v", entries[1])
	}
}

func TestCalculateMD5WithProgress(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	// MD5 of "Hello World\n" is ed076287532e86365e841e92bfc50d8c
	content := []byte("Hello World\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write sample file: %v", err)
	}

	hash, err := CalculateMD5WithProgress(testFile, 4)
	if err != nil {
		t.Fatalf("CalculateMD5WithProgress failed: %v", err)
	}

	expected := "e59ff97941044f85df5297e1c302d260"
	if hash != expected {
		t.Errorf("Hash mismatch: got %s, expected %s", hash, expected)
	}
}
