package ui

import (
	"path/filepath"
)

// MenuHandlers defines callbacks for menu actions to prevent circular package imports
type MenuHandlers struct {
	InspectRepack      func(repackDir string)
	VerifyBins         func(repackDir string)
	UnpackRepack       func(repackDir string, outDir string, components string, verifyAfter bool)
	VerifyInstallation func(targetDir string, md5Path string)
}

// InteractiveMenu launches the authentic terminal-based Inno Setup installer wizard
func InteractiveMenu(initialRepackDir string, handlers MenuHandlers) {
	currentRepackDir, err := filepath.Abs(initialRepackDir)
	if err != nil {
		currentRepackDir = initialRepackDir
	}

	RunInstallerWizard(currentRepackDir, WizardHandlers{
		InspectRepack:      handlers.InspectRepack,
		VerifyBins:         handlers.VerifyBins,
		UnpackRepack:       handlers.UnpackRepack,
		VerifyInstallation: handlers.VerifyInstallation,
	})
}
