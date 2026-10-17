package ui

import (
	"fmt"
	"os"
	"strings"
)

// RenderProgressBar renders an adaptive, responsive progress bar on a single line.
// It automatically scales according to the current terminal width without overflowing or wrapping.
func RenderProgressBar(label string, current int64, total int64, speedMBps float64) {
	termW := RawTerminalWidth()
	if termW < 30 {
		termW = 30
	}

	pct := 0.0
	if total > 0 {
		pct = float64(current) / float64(total) * 100.0
		if pct > 100.0 {
			pct = 100.0
		}
	}

	// Prepare stats text on the right
	var stats string
	if speedMBps > 0 {
		stats = fmt.Sprintf("%5.1f%% (%s / %s) %6.1f MB/s",
			pct, FormatBytes(current), FormatBytes(total), speedMBps)
	} else {
		stats = fmt.Sprintf("%5.1f%% (%s / %s)",
			pct, FormatBytes(current), FormatBytes(total))
	}

	// Truncate label if too long for terminal
	maxLabelW := termW / 3
	if maxLabelW > 25 {
		maxLabelW = 25
	}
	displayLabel := TruncateString(label, maxLabelW, "...")
	if displayLabel != "" {
		displayLabel = displayLabel + " "
	}

	// Calculate space for bar
	overhead := StringWidth(displayLabel) + StringWidth(stats) + 6 // "  " + displayLabel + "[" + bar + "] " + stats
	barW := termW - overhead

	var line string
	if barW >= 8 {
		filled := 0
		if total > 0 {
			filled = int(int64(barW) * current / total)
		}
		if filled > barW {
			filled = barW
		}
		if filled < 0 {
			filled = 0
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
		line = fmt.Sprintf("\r  %s%s%s[%s]%s %s%s%s",
			Cyan, displayLabel, Reset,
			bar,
			Reset, Yellow, stats, Reset,
		)
	} else {
		// Terminal too narrow for visual bar: show compact text
		line = fmt.Sprintf("\r  %s%s%s %s%s%s",
			Cyan, displayLabel, Reset,
			Yellow, stats, Reset,
		)
	}

	ClearLine()
	os.Stdout.WriteString(line)
}

// ClearProgressBar erases the active progress line.
func ClearProgressBar() {
	ClearLine()
}
