package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// PromptInput prompts the user for text input with an optional default value
func PromptInput(promptText string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s%s%s [%s%s%s]: ", Bold, promptText, Reset, Cyan, defaultVal, Reset)
	} else {
		fmt.Printf("%s%s%s: ", Bold, promptText, Reset)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultVal
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

// Pause waits for the user to press Enter before continuing
func Pause() {
	fmt.Printf("\n%sPress Enter to return to main menu...%s", Dim, Reset)
	_, _ = reader.ReadString('\n')
}
