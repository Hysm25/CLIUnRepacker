package ui

import (
	"fmt"
	"math"
	"os"
)

var (
	colorEnabled = isColorSupported()

	Reset     = code("\033[0m")
	Bold      = code("\033[1m")
	Dim       = code("\033[2m")
	Italic    = code("\033[3m")
	Underline = code("\033[4m")

	Red     = code("\033[91m")
	Green   = code("\033[92m")
	Yellow  = code("\033[93m")
	Blue    = code("\033[94m")
	Magenta = code("\033[95m")
	Cyan    = code("\033[96m")
	White   = code("\033[97m")
	Gray    = code("\033[90m")
)

func code(c string) string {
	if colorEnabled {
		return c
	}
	return ""
}

// EnableColors sets whether ANSI colors should be used in output.
func EnableColors(enable bool) {
	colorEnabled = enable
	Reset = code("\033[0m")
	Bold = code("\033[1m")
	Dim = code("\033[2m")
	Italic = code("\033[3m")
	Underline = code("\033[4m")
	Red = code("\033[91m")
	Green = code("\033[92m")
	Yellow = code("\033[93m")
	Blue = code("\033[94m")
	Magenta = code("\033[95m")
	Cyan = code("\033[96m")
	White = code("\033[97m")
	Gray = code("\033[90m")
}

func isColorSupported() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	InitTerminal()
	return true
}

// FormatBytes formats byte sizes into human-readable strings (e.g. 1.25 GB, 45.2 MB)
func FormatBytes(numBytes int64) string {
	if numBytes < 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	val := float64(numBytes)
	for _, unit := range units {
		if math.Abs(val) < 1024.0 || unit == "PB" {
			if unit == "B" {
				return fmt.Sprintf("%d B", int64(val))
			}
			return fmt.Sprintf("%.2f %s", val, unit)
		}
		val /= 1024.0
	}
	return fmt.Sprintf("%.2f PB", val)
}

// FormatSeconds formats seconds into clean MM:SS or HH:MM:SS
func FormatSeconds(seconds float64) string {
	totalSecs := int64(seconds)
	if totalSecs < 0 {
		totalSecs = 0
	}
	secs := totalSecs % 60
	mins := (totalSecs / 60) % 60
	hours := totalSecs / 3600

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
