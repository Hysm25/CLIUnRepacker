package ui

import (
	"fmt"
	"math"
	"strings"
)

// PrintBanner displays the authentic FitGirl Repacks Inno Setup installer banner
func PrintBanner() {
	banner := fmt.Sprintf(`%s%s┌─ Setup - FitGirl Repack Installer ──────────────────── [ _ ] [ □ ] [ ✕ ] ─┐
│                                                                           │
│   %s███████╗██╗████████╗ ██████╗ ██╗██████╗ ██╗%s                             │
│   %s██╔════╝██║╚══██╔══╝██╔════╝ ██║██╔══██╗██║%s      %sREPACKS%s                │
│   %s█████╗  ██║   ██║   ██║  ███╗██║██████╔╝██║%s      %shttp://fitgirl-repacks.site%s
│   %s██╔══╝  ██║   ██║   ██║   ██║██║██╔══██╗██║%s      %sInno Setup Enhanced 5.5%s  │
│   %s██║     ██║   ██║   ╚██████╔╝██║██║  ██║███████╗%s  %s(100%% Lossless Repack)%s   │
│   %s╚═╝     ╚═╝   ╚═╝    ╚═════╝ ╚═╝╚═╝  ╚═╝╚══════╝%s                         │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘%s`,
		Bold, Blue,
		Magenta, Blue,
		Magenta, Blue, Bold, Blue,
		Magenta, Blue, Cyan, Blue,
		Magenta, Blue, White, Blue,
		Magenta, Blue, Dim, Blue,
		Magenta, Blue,
		Reset,
	)
	fmt.Println(banner)
}

// PrintHeader displays an Inno Setup styled section header
func PrintHeader(title string) {
	dashes := WindowWidth - 8 - len(title)
	if dashes < 2 {
		dashes = 2
	}
	fmt.Printf("\n%s%s┌─ [ %s ] %s┐%s\n\n", Bold, Blue, title, strings.Repeat("─", dashes), Reset)
}

// FormatBytes formats byte sizes into human-readable strings (e.g. 1.25 GB)
func FormatBytes(numBytes int64) string {
	if numBytes < 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	val := float64(numBytes)
	for _, unit := range units {
		if math.Abs(val) < 1024.0 || unit == "PB" {
			return fmt.Sprintf("%.2f %s", val, unit)
		}
		val /= 1024.0
	}
	return fmt.Sprintf("%.2f PB", val)
}

// FormatSeconds formats seconds into MM:SS or HH:MM:SS
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
