package ui

import (
	"fmt"
	"strings"
)

// WindowWidth defines the standard column width of the terminal installer window
const WindowWidth = 78

// stringWidth calculates the visual display column width of a string (ignoring ANSI escape codes)
func stringWidth(s string) int {
	inEscape := false
	width := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		// Wide runes (emojis / double-width characters)
		if r == '🥄' || r == '📁' || r == '🖵' || r == '🔊' || r == '🔇' {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// RenderTitleBar prints the authentic Windows / Inno Setup title bar with window controls
func RenderTitleBar(appName string) {
	leftPart := fmt.Sprintf("┌─ Setup - %s ", appName)
	controls := " [ _ ] [ □ ] [ ✕ ] ─┐"
	middleDashes := WindowWidth - stringWidth(leftPart) - stringWidth(controls)
	if middleDashes < 1 {
		middleDashes = 1
	}

	fmt.Printf("\n%s%s%s%s%s%s%s\n",
		Bold, Blue, leftPart,
		strings.Repeat("─", middleDashes),
		White, controls, Reset,
	)
}

// RenderWizardHeader prints the authentic Inno Setup step header with Amélie avatar
func RenderWizardHeader(appName string, stepTitle string, stepSubtitle string) {
	RenderTitleBar(appName)

	// Header layout:
	// Left: "│  " (3 columns)
	// Text Area: 59 columns
	// Spacing: "  " (2 columns)
	// Avatar Area: 11 columns
	// Right: " │" (3 columns) -> 3 + 59 + 2 + 11 + 3 = 78 columns
	textWidth := 59

	var subLines []string
	words := strings.Fields(stepSubtitle)
	curLine := ""
	for _, w := range words {
		if stringWidth(curLine)+stringWidth(w)+1 > textWidth {
			subLines = append(subLines, curLine)
			curLine = w
		} else {
			if curLine == "" {
				curLine = w
			} else {
				curLine += " " + w
			}
		}
	}
	if curLine != "" {
		subLines = append(subLines, curLine)
	}
	if len(subLines) == 0 {
		subLines = []string{""}
	}

	// Line 1: Title + Avatar top (╭─────╮)
	titlePad := textWidth - stringWidth(stepTitle)
	if titlePad < 0 {
		titlePad = 0
	}
	titleStr := Bold + White + stepTitle + Reset + strings.Repeat(" ", titlePad)
	avatar1 := Magenta + "╭─────╮" + Reset
	fmt.Printf("%s│%s  %s  %s      %s│%s\n", Blue, Reset, titleStr, avatar1, Blue, Reset)

	// Line 2: Subtitle L1 + Avatar middle with spoon: │(◕‿◕)│🥄
	sub1 := subLines[0]
	subPad1 := textWidth - stringWidth(sub1)
	if subPad1 < 0 {
		subPad1 = 0
	}
	sub1Str := Dim + sub1 + Reset + strings.Repeat(" ", subPad1)
	avatar2 := Magenta + "│" + White + "(◕‿◕)" + Magenta + "│" + Yellow + "🥄" + Reset
	fmt.Printf("%s│%s  %s  %s   %s│%s\n", Blue, Reset, sub1Str, avatar2, Blue, Reset)

	// Line 3: Subtitle L2 + Avatar bottom (╰──┬──╯)
	sub2 := ""
	if len(subLines) > 1 {
		sub2 = subLines[1]
	}
	subPad2 := textWidth - stringWidth(sub2)
	if subPad2 < 0 {
		subPad2 = 0
	}
	sub2Str := Dim + sub2 + Reset + strings.Repeat(" ", subPad2)
	avatar3 := Magenta + "╰──┬──╯" + Reset
	fmt.Printf("%s│%s  %s  %s      %s│%s\n", Blue, Reset, sub2Str, avatar3, Blue, Reset)

	// Divider below header
	fmt.Printf("%s├%s┤%s\n", Blue, strings.Repeat("─", WindowWidth-2), Reset)
}

// RenderWelcomeHeader prints the title bar for the Welcome screen (no header box)
func RenderWelcomeHeader(appName string) {
	RenderTitleBar(appName)
}

// PrintBoxLine prints a line enclosed within the window's left and right borders
func PrintBoxLine(content string) {
	vis := stringWidth(content)
	pad := WindowWidth - 4 - vis
	if pad < 0 {
		pad = 0
	}
	fmt.Printf("%s│%s %s%s %s│%s\n", Blue, Reset, content, strings.Repeat(" ", pad), Blue, Reset)
}

// PrintEmptyBoxLine prints a blank content line enclosed by window borders
func PrintEmptyBoxLine() {
	fmt.Printf("%s│%s%s%s│%s\n", Blue, Reset, strings.Repeat(" ", WindowWidth-2), Blue, Reset)
}

// RenderWizardFooter prints the authentic Inno Setup bottom bar with music toggle and navigation
func RenderWizardFooter(canBack bool, nextLabel string, extraAction string, isMusicPlaying bool) {
	fmt.Printf("%s├%s┤%s\n", Blue, strings.Repeat("─", WindowWidth-2), Reset)

	// Music indicator icon
	speakerIcon := fmt.Sprintf("%s[🔇]%s", Dim, Reset)
	if isMusicPlaying {
		speakerIcon = fmt.Sprintf("%s[🔊]%s", Green, Reset)
	}

	brandingLink := fmt.Sprintf("%sFitGirl Repacks%s", Cyan, Reset)
	gameInfoLink := fmt.Sprintf("%sGame Info%s", Cyan, Reset)

	backStr := "[ < Back (B) ]"
	if !canBack {
		backStr = "              "
	}

	if nextLabel == "" {
		nextLabel = "Next > (Enter)"
	}
	nextStr := fmt.Sprintf("[%s]", nextLabel)
	cancelStr := "[ Cancel (C) ]"

	// Line 1: Speaker, Links, and Buttons
	leftSide := fmt.Sprintf("  %s %s  %s", speakerIcon, brandingLink, gameInfoLink)
	rightSide := fmt.Sprintf("%s  %s  %s  ", backStr, nextStr, cancelStr)

	visibleLeft := stringWidth(leftSide)
	visibleRight := stringWidth(rightSide)
	middleGap := WindowWidth - 2 - visibleLeft - visibleRight
	if middleGap < 1 {
		middleGap = 1
	}

	fmt.Printf("%s│%s%s%s%s%s│%s\n",
		Blue, Reset,
		leftSide,
		strings.Repeat(" ", middleGap),
		rightSide,
		Blue, Reset,
	)

	// Line 2: Subtitle links / Official URL & Extra actions
	siteURL := fmt.Sprintf("%sofficial site%s", Dim, Reset)
	visibleLeft2 := stringWidth("         ") + stringWidth(siteURL)
	extraText := ""
	if extraAction != "" {
		extraText = fmt.Sprintf("%s%s%s  ", Dim, extraAction, Reset)
	}
	visibleRight2 := stringWidth(extraText)
	gap2 := WindowWidth - 2 - visibleLeft2 - visibleRight2
	if gap2 < 1 {
		gap2 = 1
	}

	fmt.Printf("%s│%s         %s%s%s%s│%s\n",
		Blue, Reset,
		siteURL,
		strings.Repeat(" ", gap2),
		extraText,
		Blue, Reset,
	)

	fmt.Printf("%s└%s┘%s\n\n", Blue, strings.Repeat("─", WindowWidth-2), Reset)
}

// RenderExitDialog renders the authentic "Exit Setup" message box modal
func RenderExitDialog() {
	boxWidth := 66
	title := "┌─ Exit Setup "
	dashes := boxWidth - stringWidth(title) - 1
	if dashes < 2 {
		dashes = 2
	}

	fmt.Printf("\n%s%s%s%s┐%s\n", Bold, Yellow, title, strings.Repeat("─", dashes), Reset)
	fmt.Printf("%s│%s  %s(?)%s %sSetup is not complete. If you exit now, the program%s       %s│%s\n",
		Yellow, Reset, Bold, Reset, Bold, Reset, Yellow, Reset)
	fmt.Printf("%s│%s      will not be installed.                                     %s│%s\n",
		Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s                                                                 %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s      You may run Setup again at another time to complete        %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s      the installation.                                          %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s                                                                 %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s      Exit Setup?                                                %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s                                                                 %s│%s\n", Yellow, Reset, Yellow, Reset)
	fmt.Printf("%s│%s                    %s[ Yes (Y) ]%s     %s[ No (Enter/N) ]%s             %s│%s\n",
		Yellow, Reset, Red, Reset, Green, Reset, Yellow, Reset)
	fmt.Printf("%s└%s┘%s\n\n", Yellow, strings.Repeat("─", boxWidth-2), Reset)
}

// RenderLanguageDialog renders the authentic "Select Setup Language" modal
func RenderLanguageDialog() {
	boxWidth := 66
	title := "┌─ Select Setup Language "
	dashes := boxWidth - stringWidth(title) - 1
	if dashes < 2 {
		dashes = 2
	}

	fmt.Printf("\n%s%s%s%s┐%s\n", Bold, Cyan, title, strings.Repeat("─", dashes), Reset)
	fmt.Printf("%s│%s  %s[🖵]%s Select the language to use during the installation:        %s│%s\n",
		Cyan, Reset, Bold, Reset, Cyan, Reset)
	fmt.Printf("%s│%s                                                                 %s│%s\n", Cyan, Reset, Cyan, Reset)
	fmt.Printf("%s│%s      ┌──────────────────────────────────────────────────┐       %s│%s\n", Cyan, Reset, Cyan, Reset)
	fmt.Printf("%s│%s      │ %s[1] English (default)%s                          │       %s│%s\n",
		Cyan, Reset, Bold, Reset, Cyan, Reset)
	fmt.Printf("%s│%s      │ %s[2] Русский (Russian)%s                          │       %s│%s\n",
		Cyan, Reset, Bold, Reset, Cyan, Reset)
	fmt.Printf("%s│%s      └──────────────────────────────────────────────────┘       %s│%s\n", Cyan, Reset, Cyan, Reset)
	fmt.Printf("%s│%s                                                                 %s│%s\n", Cyan, Reset, Cyan, Reset)
	fmt.Printf("%s│%s                    %s[ OK (Enter) ]%s    %s[ Cancel (C) ]%s             %s│%s\n",
		Cyan, Reset, Green, Reset, Dim, Reset, Cyan, Reset)
	fmt.Printf("%s└%s┘%s\n\n", Cyan, strings.Repeat("─", boxWidth-2), Reset)
}
