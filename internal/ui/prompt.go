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

// PromptConfirm prompts the user with a Yes/No question
func PromptConfirm(promptText string, defaultYes bool) bool {
	hint := "Y/n"
	if !defaultYes {
		hint = "y/N"
	}
	fmt.Printf("%s%s%s [%s%s%s]: ", Bold, promptText, Reset, Cyan, hint, Reset)

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes
	}
	if line == "y" || line == "yes" {
		return true
	}
	if line == "n" || line == "no" {
		return false
	}
	return defaultYes
}

// Pause waits for the user to press Enter to continue
func Pause() {
	PauseMessage("Press Enter to continue...")
}

// PauseMessage displays a custom pause message and waits for Enter
func PauseMessage(msg string) {
	fmt.Printf("\n%s%s%s", Dim, msg, Reset)
	_, _ = reader.ReadString('\n')
}
