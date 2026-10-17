package verifier

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cliunrepacker/internal/ui"
)

// MD5Entry stores a file checksum entry
type MD5Entry struct {
	ExpectedHash string
	RelPath      string
}

// ParseMD5File parses standard MD5 / QuickSFV file format
func ParseMD5File(md5FilePath string) ([]MD5Entry, error) {
	file, err := os.Open(md5FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []MD5Entry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// QuickSFV / MD5 format: 32 hex characters followed by whitespace and filename
		if len(line) < 33 {
			continue
		}

		expectedHash := strings.ToLower(line[:32])
		relPath := strings.TrimSpace(line[32:])

		// Strip leading asterisk
		relPath = strings.TrimPrefix(relPath, "*")
		relPath = strings.TrimSpace(relPath)

		// Normalize paths
		relPath = strings.ReplaceAll(relPath, `..\`, "")
		relPath = strings.ReplaceAll(relPath, `../`, "")
		relPath = strings.ReplaceAll(relPath, `\`, "/")

		entries = append(entries, MD5Entry{
			ExpectedHash: expectedHash,
			RelPath:      relPath,
		})
	}

	return entries, scanner.Err()
}

// CalculateMD5WithProgress calculates MD5 hash while rendering a real-time progress bar
func CalculateMD5WithProgress(filePath string, chunkSize int64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	totalSize := stat.Size()
	if chunkSize <= 0 {
		chunkSize = 4 * 1024 * 1024 // 4 MB chunk
	}

	hasher := md5.New()
	buf := make([]byte, chunkSize)
	var bytesRead int64
	tStart := time.Now()

	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			hasher.Write(buf[:n])
			bytesRead += int64(n)

			elapsed := time.Since(tStart).Seconds()
			speed := 0.0
			if elapsed > 0 {
				speed = (float64(bytesRead) / (1024 * 1024)) / elapsed
			}

			ui.RenderProgressBar(filepath.Base(filePath), bytesRead, totalSize, speed)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			ui.ClearProgressBar()
			return "", readErr
		}
	}

	ui.ClearProgressBar()
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
