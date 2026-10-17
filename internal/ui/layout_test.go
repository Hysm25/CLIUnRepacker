package ui

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	colored := "\033[92mHello \033[1mWorld\033[0m!"
	expected := "Hello World!"
	actual := StripANSI(colored)
	if actual != expected {
		t.Errorf("StripANSI failed: got %q, expected %q", actual, expected)
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Hello World", 11},
		{"\033[91mRed Text\033[0m", 8},
		{"日本語", 6}, // 3 wide characters * 2 = 6
		{"🚀 Rocket", 9}, // 2 + 1 + 6 = 9
		{"", 0},
	}

	for _, tt := range tests {
		actual := StringWidth(tt.input)
		if actual != tt.expected {
			t.Errorf("StringWidth(%q): got %d, expected %d", tt.input, actual, tt.expected)
		}
	}
}

func TestWrapText(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	wrapped := WrapText(text, 15)
	if len(wrapped) < 3 {
		t.Errorf("Expected at least 3 wrapped lines, got %d: %v", len(wrapped), wrapped)
	}
	for i, line := range wrapped {
		if StringWidth(line) > 15 {
			t.Errorf("Line %d exceeds maxWidth 15: %q (width %d)", i, line, StringWidth(line))
		}
	}
}

func TestTruncateString(t *testing.T) {
	s := "VeryLongFileNameThatExceedsTheBudget.bin"
	truncated := TruncateString(s, 20, "...")
	if StringWidth(truncated) > 20 {
		t.Errorf("Truncated string exceeds maxWidth 20: %q", truncated)
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Errorf("Expected suffix '...', got %q", truncated)
	}
}

func TestRule(t *testing.T) {
	rule := Rule("Section Title")
	clean := StripANSI(rule)
	if !strings.Contains(clean, "Section Title") {
		t.Errorf("Rule does not contain title: %q", clean)
	}
	if len(clean) < 40 {
		t.Errorf("Rule too short: %d", len(clean))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1024 * 1024 * 5, "5.00 MB"},
		{1410635437, "1.31 GB"},
	}

	for _, tt := range tests {
		actual := FormatBytes(tt.bytes)
		if actual != tt.expected {
			t.Errorf("FormatBytes(%d): got %s, expected %s", tt.bytes, actual, tt.expected)
		}
	}
}

func TestFormatSeconds(t *testing.T) {
	if s := FormatSeconds(65); s != "01:05" {
		t.Errorf("FormatSeconds(65): got %s, expected 01:05", s)
	}
	if s := FormatSeconds(3665); s != "01:01:05" {
		t.Errorf("FormatSeconds(3665): got %s, expected 01:01:05", s)
	}
}
