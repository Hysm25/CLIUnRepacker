package ui

import (
	"os"
)

var (
	colorEnabled = isColorSupported()

	Reset     = code("\033[0m")
	Bold      = code("\033[1m")
	Dim       = code("\033[2m")
	Underline = code("\033[4m")
	Red       = code("\033[91m")
	Green     = code("\033[92m")
	Yellow    = code("\033[93m")
	Blue      = code("\033[94m")
	Magenta   = code("\033[95m")
	Cyan      = code("\033[96m")
	White     = code("\033[97m")
)

func code(c string) string {
	if colorEnabled {
		return c
	}
	return ""
}

func EnableColors(enable bool) {
	colorEnabled = enable
	Reset = code("\033[0m")
	Bold = code("\033[1m")
	Dim = code("\033[2m")
	Underline = code("\033[4m")
	Red = code("\033[91m")
	Green = code("\033[92m")
	Yellow = code("\033[93m")
	Blue = code("\033[94m")
	Magenta = code("\033[95m")
	Cyan = code("\033[96m")
	White = code("\033[97m")
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
