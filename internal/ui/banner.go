package ui

import (
	"fmt"
	"strings"
)

// PrintBanner prints a clean, modern, minimalist header
func PrintBanner() {
	w := TerminalWidth()
	fmt.Printf("\n%s%sCLIUnRepacker%s  %sv1.0.0%s\n", Bold, Cyan, Reset, Dim, Reset)
	fmt.Printf("%sCross-Platform FitGirl Repack Inspector & Unpacker%s\n", Dim, Reset)
	fmt.Println(Dim + strings.Repeat("─", w) + Reset)
}

// PrintRepackSummary displays an aligned summary of the detected repack in repackDir
func PrintRepackSummary(meta RepackMeta, repackDir string) {
	w := TerminalWidth()
	appName := meta.AppName
	if appName == "" {
		appName = "FitGirl Repack"
	}

	fmt.Printf("  %sRepack:%s    %s%s%s\n", Dim, Reset, Bold, appName, Reset)
	fmt.Printf("  %sDirectory:%s %s\n", Dim, Reset, repackDir)

	if len(meta.Archives) > 0 {
		var totalSize int64
		var names []string
		for _, a := range meta.Archives {
			totalSize += a.Size
			names = append(names, fmt.Sprintf("%s (%s)", a.Name, FormatBytes(a.Size)))
		}
		fmt.Printf("  %sArchives:%s  %d files, %s total [%s]\n",
			Dim, Reset, len(meta.Archives), FormatBytes(totalSize), strings.Join(names, ", "))
	} else {
		fmt.Printf("  %sArchives:%s  %sNo fg-*.bin archives detected in directory%s\n", Dim, Reset, Yellow, Reset)
	}

	fmt.Println(Dim + strings.Repeat("─", w) + Reset)
}
