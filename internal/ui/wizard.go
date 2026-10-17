package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cliunrepacker/internal/system"
)

// WizardHandlers provides the operations dispatched by the wizard pages
type WizardHandlers struct {
	VerifyBins         func(repackDir string)
	InspectRepack      func(repackDir string)
	UnpackRepack       func(repackDir string, outDir string, components string, verifyAfter bool)
	VerifyInstallation func(targetDir string, md5Path string)
}

// WizardState preserves the user's choices across wizard pages
type WizardState struct {
	RepackDir         string
	Meta              RepackMeta
	Language          string // "en" or "ru"
	LimitRAM2GB       bool
	DestinationDir    string
	NoStartMenu       bool // [x] Don't create a Start Menu folder
	CreateDesktopIcon bool // [x] Create a desktop icon
	ComponentMain     bool // [x] Main game files
	ComponentDirectX  bool // [ ] Update DirectX
	ComponentVCx86    bool // [ ] Download and Install C++ Redist 2015-2019 x86
	ComponentVCx64    bool // [ ] Download and Install C++ Redist 2015-2019 x64
	PostVerifyMD5     bool // [x] Verify files integrity (recommended)
	PostRedirect      bool // [ ] Redirect fake FitGirl sites to the real one
	PostLaunch        bool // [ ] Launch game
	Handlers          WizardHandlers
}

// RunInstallerWizard runs the authentic Inno Setup interactive terminal wizard
func RunInstallerWizard(repackDir string, handlers WizardHandlers) {
	absRepackDir, err := filepath.Abs(repackDir)
	if err != nil {
		absRepackDir = repackDir
	}

	meta := LoadRepackMeta(absRepackDir)
	sysInfo := system.RunSystemCheck()

	// Default RAM limit recommendation based on verified host memory (<= 8GB)
	limitRAM := sysInfo.MemTotalGB > 0 && sysInfo.MemTotalGB <= 8.5

	defaultDest := filepath.Join(absRepackDir, "unpacked_game")

	state := WizardState{
		RepackDir:         absRepackDir,
		Meta:              meta,
		Language:          "en",
		LimitRAM2GB:       limitRAM,
		DestinationDir:    defaultDest,
		NoStartMenu:       true, // Default in original installer: don't create start menu folder
		CreateDesktopIcon: true, // Default in original installer: create desktop icon
		ComponentMain:     true, // Mandatory main game files
		PostVerifyMD5:     true, // Default: verify file integrity
		Handlers:          handlers,
	}

	// Step 0: Language Selection Dialog
	langChoice := showLanguageDialog(&state)
	if langChoice == -1 {
		return
	}

	currentStep := 1 // 1: Welcome, 2: Info, 3: Destination, 4: Components, 5: Installing, 6: Finished

	for {
		switch currentStep {
		case 1:
			next := showWelcomePage(&state, sysInfo)
			if next == -1 {
				return // User cancelled
			}
			currentStep = next

		case 2:
			next := showInfoPage(&state)
			if next == -1 {
				return
			}
			currentStep = next

		case 3:
			next := showSelectDirPage(&state)
			if next == -1 {
				return
			}
			currentStep = next

		case 4:
			next := showComponentsPage(&state)
			if next == -1 {
				return
			}
			currentStep = next

		case 5:
			next := runInstallingPage(&state)
			currentStep = next

		case 6:
			showFinishedPage(&state)
			return
		}
	}
}

// Step 0: "Select Setup Language" Modal
func showLanguageDialog(state *WizardState) int {
	RenderLanguageDialog()

	for {
		choice := strings.ToLower(PromptInput("Select language [1: English | 2: Русский | C: Cancel]", "1"))
		switch choice {
		case "1", "en", "english", "":
			state.Language = "en"
			return 1
		case "2", "ru", "russian", "русский":
			state.Language = "ru"
			return 1
		case "c", "cancel", "q":
			if confirmExit() {
				return -1
			}
			RenderLanguageDialog()
		default:
			fmt.Printf("%s[!] Invalid choice. Enter 1 for English, 2 for Russian.%s\n", Yellow, Reset)
		}
	}
}

// Page 1: Welcome to the [AppName] Setup Wizard
func showWelcomePage(state *WizardState, sysInfo system.SysInfo) int {
	RenderWelcomeHeader(state.Meta.AppName)

	ramMark := " "
	if state.LimitRAM2GB {
		ramMark = "X"
	}

	PrintBoxLine(fmt.Sprintf("%s┌──────────────────────┐%s  %s%sWelcome to the %s%s",
		Cyan, Reset, Bold, White, state.Meta.AppName, Reset))
	PrintBoxLine(fmt.Sprintf("%s│ THE LEGEND OF ZELDA  │%s  %s%sSetup Wizard%s",
		Cyan, Reset, Bold, White, Reset))
	PrintBoxLine(fmt.Sprintf("%s│   Link's Awakening   │%s",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s│         ▲            │%s  This will install %s on your",
		Cyan, Reset, state.Meta.AppName))
	PrintBoxLine(fmt.Sprintf("%s│        ▲ ▲           │%s  computer.",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s│                      │%s",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s│                      │%s  It is recommended that you close all other",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s│                      │%s  applications before continuing.",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s│   FitGirl Repack     │%s",
		Cyan, Reset))
	PrintBoxLine(fmt.Sprintf("%s└──────────────────────┘%s  Click Next to continue, or Cancel to exit Setup.",
		Cyan, Reset))

	PrintEmptyBoxLine()

	// Checkbox: Limit installer to 2 GB of RAM usage
	PrintBoxLine(fmt.Sprintf(" %s[%s] Limit installer to 2 GB of RAM usage%s   (Toggle: [R])",
		Bold, ramMark, Reset))
	PrintEmptyBoxLine()

	// Exact red warning text reverse-engineered from setup.exe
	PrintBoxLine(fmt.Sprintf(" %sIf you have a low-memory machine (8 GB of RAM or less),%s", Red, Reset))
	PrintBoxLine(fmt.Sprintf(" %sYOU MUST CHECK THIS CHECKBOX.%s", Red, Reset))
	PrintEmptyBoxLine()
	PrintBoxLine(fmt.Sprintf(" %sThis won't allow the installer to use more than 2 GB of RAM.%s", Red, Reset))
	PrintBoxLine(fmt.Sprintf(" %sInstallation will be slower, but stable.%s", Red, Reset))
	PrintEmptyBoxLine()

	// Host memory diagnostic status
	if sysInfo.MemTotalGB > 0 {
		memStatus := fmt.Sprintf("[%sSystem RAM: %.1f GB detected (%.1f GB free)%s]", Dim, sysInfo.MemTotalGB, sysInfo.MemAvailGB, Reset)
		PrintBoxLine(" " + memStatus)
	}

	RenderWizardFooter(false, "Next > (Enter)", "[R] RAM | [V] Verify | [I] Info | [D] Diag", IsMusicPlaying())

	for {
		choice := strings.ToLower(PromptInput("Enter choice [Next: Enter | R: RAM | M: Music | V: Verify | I: Info | D: Diag | C: Cancel]", ""))

		switch choice {
		case "", "next", "n":
			return 2
		case "r":
			state.LimitRAM2GB = !state.LimitRAM2GB
			return 1
		case "m":
			ToggleMusic()
			return 1
		case "v":
			if state.Handlers.VerifyBins != nil {
				state.Handlers.VerifyBins(state.RepackDir)
			}
			Pause()
			return 1
		case "i":
			if state.Handlers.InspectRepack != nil {
				state.Handlers.InspectRepack(state.RepackDir)
			}
			Pause()
			return 1
		case "d":
			PrintHeader("System Diagnostics & Environment")
			info := system.RunSystemCheck()
			fmt.Printf("%sOperating System:%s  %s (%s)\n", Bold, Reset, info.OS, info.Arch)
			fmt.Printf("%sinnoextract:%s       %s\n", Bold, Reset, info.Innoextract)
			fmt.Printf("%sWine runtime:%s      %s\n", Bold, Reset, info.Wine)
			fmt.Printf("%sPhysical RAM:%s      %.1f GB total (%.1f GB available)\n", Bold, Reset, info.MemTotalGB, info.MemAvailGB)
			Pause()
			return 1
		case "p":
			newDirStr := PromptInput("Enter path to repack folder", state.RepackDir)
			if newDirStr != "" {
				if p, err := filepath.Abs(newDirStr); err == nil {
					if stat, err := os.Stat(p); err == nil && stat.IsDir() {
						state.RepackDir = p
						state.Meta = LoadRepackMeta(p)
						fmt.Printf("%s[✓] Working directory changed to: %s%s\n", Green, p, Reset)
					} else {
						fmt.Printf("%s[!] Directory '%s' does not exist.%s\n", Red, p, Reset)
					}
				}
			}
			return 1
		case "c", "q", "0", "exit", "cancel":
			if confirmExit() {
				return -1
			}
			return 1
		default:
			fmt.Printf("%s[!] Invalid choice. Press Enter for Next, R to toggle RAM, C to Cancel.%s\n", Yellow, Reset)
		}
	}
}

// Page 2: Information Screen
func showInfoPage(state *WizardState) int {
	RenderWizardHeader(state.Meta.AppName, "Information", "Please read the following important information before continuing.")

	PrintBoxLine(fmt.Sprintf("%sWhen you are ready to continue with Setup, click Next.%s", Dim, Reset))
	PrintEmptyBoxLine()

	// Authentic yellow warning box with blue border (verified from setup.exe)
	PrintBoxLine(fmt.Sprintf("%s╔══════════════════════════════════════════════════════════════════════╗%s", Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %s%sfitgirl-repacks.site%s%s is THE ONLY official site of my repacks.        %s║%s",
		Blue, Underline, Yellow, Reset, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %sAll other \"mirrors\" (fitgirlrepacks.co, fitgirl-repacks.cc,         %s║%s",
		Blue, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %srepack-fitgirl.com, fitgirl-repack.com, etc.) are ALL fakes,        %s║%s",
		Blue, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %smade to infect you with malware, show you tons of ads and get your   %s║%s",
		Blue, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %smoney as donations. Make sure to enable redirection to the real     %s║%s",
		Blue, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s║  %ssite at the end of the installation.                                %s║%s",
		Blue, Yellow, Blue, Reset))
	PrintBoxLine(fmt.Sprintf("%s╚══════════════════════════════════════════════════════════════════════╝%s", Blue, Reset))
	PrintEmptyBoxLine()

	// Verified Game Metadata
	PrintBoxLine(fmt.Sprintf("%s%s%s", Bold, Blue, state.Meta.AppName))
	PrintBoxLine(fmt.Sprintf("%sRelease Date:%s %s", Bold, Reset, state.Meta.ReleaseDate))
	if state.Meta.GenresTags != "" {
		PrintBoxLine(fmt.Sprintf("%sGenres/Tags:%s  %s", Bold, Reset, state.Meta.GenresTags))
	}
	PrintBoxLine(fmt.Sprintf("%sDeveloper:%s    %s | %sPublisher:%s %s", Bold, Reset, state.Meta.Developer, Bold, Reset, state.Meta.Publisher))
	if state.Meta.Engine != "" {
		PrintBoxLine(fmt.Sprintf("%sEngine:%s       %s", Bold, Reset, state.Meta.Engine))
	}
	PrintEmptyBoxLine()
	PrintBoxLine(fmt.Sprintf("%sRepack Features:%s", Bold, Reset))
	for _, feat := range state.Meta.RepackFeatures {
		PrintBoxLine(fmt.Sprintf("  • %s", feat))
	}

	RenderWizardFooter(true, "Next > (Enter)", "", IsMusicPlaying())

	for {
		choice := strings.ToLower(PromptInput("Enter choice [Next: Enter | B: Back | C: Cancel]", ""))
		switch choice {
		case "", "next", "n":
			return 3
		case "b", "back":
			return 1
		case "c", "q", "cancel":
			if confirmExit() {
				return -1
			}
			return 2
		}
	}
}

// Page 3: Select Destination Location (wpSelectDir)
// REA Verified: Contains both Destination Folder AND Start Menu / Desktop Checkboxes!
func showSelectDirPage(state *WizardState) int {
	RenderWizardHeader(state.Meta.AppName, "Select Destination Location", "Where should "+state.Meta.AppName+" be installed?")

	PrintBoxLine(fmt.Sprintf("%s[📁] Setup will install %s into the following folder.%s", Bold, state.Meta.AppName, Reset))
	PrintEmptyBoxLine()
	PrintBoxLine("To continue, click Next. If you would like to select a different folder,")
	PrintBoxLine("enter [P] to browse/edit the path.")
	PrintEmptyBoxLine()

	PrintBoxLine(fmt.Sprintf("%sDestination folder:%s", Bold, Reset))
	PrintBoxLine(fmt.Sprintf("  %s%s%s   %s[P: Browse / Edit]%s", Cyan, state.DestinationDir, Reset, Dim, Reset))
	PrintEmptyBoxLine()

	// Checkboxes verified directly on this page from setup.exe screenshot
	noSmMark := " "
	if state.NoStartMenu {
		noSmMark = "X"
	}
	deskMark := " "
	if state.CreateDesktopIcon {
		deskMark = "X"
	}

	PrintBoxLine(fmt.Sprintf("%s[1] [%s] Don't create a Start Menu folder%s  (Toggle: [1])", Bold, noSmMark, Reset))
	PrintBoxLine(fmt.Sprintf("%s[2] [%s] Create a desktop icon%s             (Toggle: [2])", Bold, deskMark, Reset))
	PrintEmptyBoxLine()

	// Required space line verified from setup.exe screenshot: 6,211.7 MB
	PrintBoxLine(fmt.Sprintf("At least %s%.1f MB%s of free disk space is required.", Bold, state.Meta.RequiredSpaceMB, Reset))

	if free, err := system.GetFreeDiskSpace(state.DestinationDir); err == nil {
		freeMB := float64(free) / (1024 * 1024)
		statusColor := Green
		if freeMB < state.Meta.RequiredSpaceMB {
			statusColor = Red
		}
		PrintBoxLine(fmt.Sprintf("Available disk space on destination drive: %s%.1f MB%s", statusColor, freeMB, Reset))
	}

	RenderWizardFooter(true, "Next > (Enter)", "[P] Path | [1-2] Toggle", IsMusicPlaying())

	for {
		choice := strings.ToLower(PromptInput("Enter choice [Next: Enter | P: Path | 1/2: Toggle | B: Back | C: Cancel]", ""))
		switch choice {
		case "", "next", "n":
			return 4
		case "p", "path":
			newPath := PromptInput("Enter new destination directory", state.DestinationDir)
			if newPath != "" {
				if abs, err := filepath.Abs(newPath); err == nil {
					state.DestinationDir = abs
				} else {
					state.DestinationDir = newPath
				}
			}
			return 3
		case "1":
			state.NoStartMenu = !state.NoStartMenu
			return 3
		case "2":
			state.CreateDesktopIcon = !state.CreateDesktopIcon
			return 3
		case "b", "back":
			return 2
		case "c", "q", "cancel":
			if confirmExit() {
				return -1
			}
			return 3
		}
	}
}

// Page 4: Select Components (wpSelectComponents)
// REA Verified: Features "Default Installation" dropdown, component checklist, and [Install] button
func showComponentsPage(state *WizardState) int {
	RenderWizardHeader(state.Meta.AppName, "Select Components", "Which components should be installed?")

	PrintBoxLine("Select the components you want to install; clear the components you do not")
	PrintBoxLine("want to install. Click Install when you are ready to continue.")
	PrintEmptyBoxLine()

	PrintBoxLine(fmt.Sprintf("%sSetup Type:%s [ Default Installation ]", Bold, Reset))
	PrintEmptyBoxLine()

	dxMark := " "
	if state.ComponentDirectX {
		dxMark = "X"
	}
	vc86Mark := " "
	if state.ComponentVCx86 {
		vc86Mark = "X"
	}
	vc64Mark := " "
	if state.ComponentVCx64 {
		vc64Mark = "X"
	}

	// Verified checklist items matching setup.exe screenshot
	PrintBoxLine(fmt.Sprintf("%s[1] [X] Main game files%s                                   [Required]", Bold, Reset))
	PrintBoxLine(fmt.Sprintf("%s[2] [%s] Update DirectX%s                                    [Optional]", Bold, dxMark, Reset))
	PrintBoxLine(fmt.Sprintf("%s[3] [%s] Download and Install C++ Redist 2015-2019 x86%s     [Optional]", Bold, vc86Mark, Reset))
	PrintBoxLine(fmt.Sprintf("%s[4] [%s] Download and Install C++ Redist 2015-2019 x64%s     [Optional]", Bold, vc64Mark, Reset))
	PrintEmptyBoxLine()

	// Verified space line from setup.exe screenshot: 6,200.9 MB
	PrintBoxLine(fmt.Sprintf("Current selection requires at least %s%.1f MB%s of disk space.", Bold, state.Meta.SelectionSpaceMB, Reset))
	PrintBoxLine(fmt.Sprintf("%sTip: Leaving optional redists unchecked avoids network timeouts on Wine.%s", Dim, Reset))

	// In the real installer, the button here is [Install], NOT [Next >]!
	RenderWizardFooter(true, "Install (Enter)", "[2-4] Toggle", IsMusicPlaying())

	for {
		choice := strings.ToLower(PromptInput("Enter choice [Install: Enter | 2/3/4: Toggle | B: Back | C: Cancel]", ""))
		switch choice {
		case "", "install", "i", "next":
			return 5 // Transition directly to Installing
		case "1":
			fmt.Printf("  %s[!] Main game files are mandatory and cannot be deselected.%s\n", Yellow, Reset)
		case "2":
			state.ComponentDirectX = !state.ComponentDirectX
			return 4
		case "3":
			state.ComponentVCx86 = !state.ComponentVCx86
			return 4
		case "4":
			state.ComponentVCx64 = !state.ComponentVCx64
			return 4
		case "b", "back":
			return 3
		case "c", "q", "cancel":
			if confirmExit() {
				return -1
			}
			return 4
		}
	}
}

// Page 5: Installing Screen (wpInstalling)
func runInstallingPage(state *WizardState) int {
	StopMusic() // Silence music during heavy extraction

	RenderWizardHeader(state.Meta.AppName, "Installing", "Please wait while Setup installs "+state.Meta.AppName+" on your computer.")

	PrintBoxLine(fmt.Sprintf("%sUnpacking...%s", Cyan, Reset))
	PrintEmptyBoxLine()

	// Build component argument matching Inno Setup selections
	compList := []string{"text"}
	if state.ComponentDirectX {
		compList = append(compList, "directx")
	}
	if state.ComponentVCx86 {
		compList = append(compList, "vc_2019_x86")
	}
	if state.ComponentVCx64 {
		compList = append(compList, "vc_2019_x64")
	}
	componentsArg := strings.Join(compList, ",")

	if state.Handlers.UnpackRepack != nil {
		state.Handlers.UnpackRepack(state.RepackDir, state.DestinationDir, componentsArg, false)
	}

	return 6
}

// Page 6: Completing the Setup Wizard (wpFinished)
func showFinishedPage(state *WizardState) {
	RenderWizardHeader(state.Meta.AppName, "Completing the "+state.Meta.AppName+" Setup Wizard", "Setup has finished installing "+state.Meta.AppName+" on your computer.")

	PrintBoxLine(fmt.Sprintf("%sSetup has finished installing %s on your computer.%s", Bold, state.Meta.AppName, Reset))
	PrintBoxLine("Click Finish to exit Setup.")
	PrintEmptyBoxLine()

	verifyMark := " "
	if state.PostVerifyMD5 {
		verifyMark = "X"
	}
	redirectMark := " "
	if state.PostRedirect {
		redirectMark = "X"
	}
	launchMark := " "
	if state.PostLaunch {
		launchMark = "X"
	}

	PrintBoxLine(fmt.Sprintf("%sPost-Installation Actions:%s", Bold, Reset))
	PrintBoxLine(fmt.Sprintf(" %s[1] [%s] Verify files integrity (recommended)%s", Bold, verifyMark, Reset))
	PrintBoxLine(fmt.Sprintf(" %s[2] [%s] Redirect fake FitGirl sites to the real one%s", Bold, redirectMark, Reset))
	PrintBoxLine(fmt.Sprintf(" %s[3] [%s] Launch %s%s", Bold, launchMark, state.Meta.AppName, Reset))
	PrintEmptyBoxLine()

	PrintBoxLine(fmt.Sprintf("%sOfficial Site:%s %s", Dim, Reset, state.Meta.OfficialSite))

	RenderWizardFooter(false, "Finish (Enter)", "[1-3] Toggle", false)

	for {
		choice := strings.ToLower(PromptInput("Enter choice [Finish: Enter | 1/2/3: Toggle]", ""))
		switch choice {
		case "", "finish", "f", "next":
			// If post-verification requested, run MD5 check
			if state.PostVerifyMD5 && state.Handlers.VerifyInstallation != nil {
				state.Handlers.VerifyInstallation(state.DestinationDir, "")
			}
			fmt.Printf("\n%s[✓] Setup Wizard completed successfully. Enjoy your game!%s\n\n", Green, Reset)
			return
		case "1":
			state.PostVerifyMD5 = !state.PostVerifyMD5
			showFinishedPage(state)
			return
		case "2":
			state.PostRedirect = !state.PostRedirect
			showFinishedPage(state)
			return
		case "3":
			state.PostLaunch = !state.PostLaunch
			showFinishedPage(state)
			return
		}
	}
}

// confirmExit renders the authentic "Exit Setup" modal and asks for confirmation
func confirmExit() bool {
	RenderExitDialog()
	ans := strings.ToLower(PromptInput("Exit Setup? [y/N]", "N"))
	return ans == "y" || ans == "yes"
}
