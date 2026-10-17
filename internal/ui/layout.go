package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// TerminalWidth returns the detected terminal width clamped between 40 and 100 columns for readable layouts.
func TerminalWidth() int {
	w := RawTerminalWidth()
	if w < 40 {
		return 40
	}
	if w > 100 {
		return 100
	}
	return w
}

// RawTerminalWidth returns the actual number of columns in the terminal, or 80 if undetermined.
func RawTerminalWidth() int {
	if colStr := os.Getenv("COLUMNS"); colStr != "" {
		if w, err := strconv.Atoi(colStr); err == nil && w > 0 {
			return w
		}
	}
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

// TerminalHeight returns the detected terminal height, or 24 if undetermined.
func TerminalHeight() int {
	if _, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && h > 0 {
		return h
	}
	if rowStr := os.Getenv("LINES"); rowStr != "" {
		if h, err := strconv.Atoi(rowStr); err == nil && h > 0 {
			return h
		}
	}
	return 24
}

// IsTerminal checks if stdout is a character device (TTY).
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// StripANSI removes ANSI escape codes from string for accurate width calculation.
func StripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEscape = false
			}
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// StringWidth calculates the visible display column width of s, excluding ANSI escapes.
// Accurately handles wide Unicode characters (CJK, emojis, and fullwidth symbols).
func StringWidth(s string) int {
	clean := StripANSI(s)
	width := 0
	for _, r := range clean {
		if (r >= 0x1100 && r <= 0x115F) ||
			(r >= 0x2E80 && r <= 0xA4CF) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x1F300 && r <= 0x1F64F) ||
			(r >= 0x1F680 && r <= 0x1F6FF) {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// TruncateString truncates s to fit within maxWidth columns, adding suffix if truncated.
func TruncateString(s string, maxWidth int, suffix string) string {
	if maxWidth <= 0 {
		return ""
	}
	if StringWidth(s) <= maxWidth {
		return s
	}

	suffixW := StringWidth(suffix)
	targetW := maxWidth - suffixW
	if targetW < 0 {
		targetW = 0
		suffix = ""
	}

	clean := StripANSI(s)
	curW := 0
	var out []rune
	for _, r := range clean {
		rW := 1
		if (r >= 0x1100 && r <= 0x115F) || (r >= 0x2E80 && r <= 0xA4CF) ||
			(r >= 0xAC00 && r <= 0xD7A3) || (r >= 0x1F300 && r <= 0x1F6FF) {
			rW = 2
		}
		if curW+rW > targetW {
			break
		}
		out = append(out, r)
		curW += rW
	}
	return string(out) + suffix
}

// PadRight pads s with spaces until it reaches targetWidth columns.
func PadRight(s string, targetWidth int) string {
	w := StringWidth(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// WrapText wraps text cleanly on word boundaries to fit within maxWidth columns.
func WrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if StringWidth(line) <= maxWidth {
			result = append(result, line)
			continue
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		cur := words[0]
		for _, w := range words[1:] {
			if StringWidth(cur)+1+StringWidth(w) <= maxWidth {
				cur += " " + w
			} else {
				result = append(result, cur)
				cur = w
			}
		}
		if cur != "" {
			result = append(result, cur)
		}
	}
	return result
}

// Rule creates a clean, responsive horizontal divider with an optional title.
func Rule(title string) string {
	w := TerminalWidth()
	if title == "" {
		return Dim + strings.Repeat("─", w) + Reset
	}
	titleStr := " " + title + " "
	titleW := StringWidth(titleStr)
	if titleW >= w-4 {
		return Bold + Cyan + title + Reset
	}
	dashesLeft := 3
	dashesRight := w - dashesLeft - titleW
	if dashesRight < 2 {
		dashesRight = 2
	}
	return Dim + strings.Repeat("─", dashesLeft) + Reset +
		Bold + Cyan + titleStr + Reset +
		Dim + strings.Repeat("─", dashesRight) + Reset
}

// PrintHeader prints a styled, responsive section title.
func PrintHeader(title string) {
	fmt.Printf("\n%s\n\n", Rule(title))
}

// ClearLine clears the current terminal line reliably on ANSI terminals.
func ClearLine() {
	if IsTerminal() {
		os.Stdout.WriteString("\r\033[K")
	} else {
		w := RawTerminalWidth()
		os.Stdout.WriteString("\r" + strings.Repeat(" ", w) + "\r")
	}
}
