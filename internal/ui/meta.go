package ui

import (
	"os"
	"path/filepath"
	"strings"
)

// RepackMeta holds verified metadata extracted from the Inno Setup container
type RepackMeta struct {
	AppName            string
	AppID              string
	DefaultDirName     string
	RequiredSpaceBytes int64
	OfficialSite       string
	GameInfoURL        string
	ReleaseDate        string
	Developer          string
	Publisher          string
	GenresTags         string
	Engine             string
	RequiredSpaceMB    float64
	SelectionSpaceMB   float64
	RepackFeatures     []string
	Archives           []ArchiveFile
}

// ArchiveFile holds name and size of an archive piece
type ArchiveFile struct {
	Name string
	Size int64
}

// LoadRepackMeta extracts and parses repack metadata for display in the wizard
func LoadRepackMeta(repackDir string) RepackMeta {
	meta := RepackMeta{
		AppName:            "The Legend of Zelda: LA",
		AppID:              "The Legend of Zelda: LA",
		DefaultDirName:     "Games/The Legend of Zelda - Link's Awakening",
		RequiredSpaceBytes: 6513446912, // 6,211.7 MB (from verified installer)
		RequiredSpaceMB:    6211.7,     // Verified from Inno Setup Page 3
		SelectionSpaceMB:   6200.9,     // Verified from Inno Setup Page 4
		OfficialSite:       "http://fitgirl-repacks.site",
		GameInfoURL:        "http://en.riotpixels.com/games/xenoblade-chronicles/",
		ReleaseDate:        "September 20, 2019",
		Developer:          "Nintendo EPD",
		Publisher:          "Nintendo",
		GenresTags:         "Action, Open world, Top-down, 3D",
		Engine:             "in-house",
		RepackFeatures: []string{
			"Based on Switch release with Update v65536 installed over (v1.0.1): 5.9 GB",
			"Two versions of Yuzu Switch emulator included: Mainline v559 & Early Access v1502",
			"Shader cache is included and enabled by default for smoother gameplay",
			"Two mods included: Tilt-Shift Effect remover and Constant 60 FPS",
			"100% Lossless, but NOT MD5 Perfect (encryption was removed for better size)",
			"NOTHING ripped, NOTHING re-encoded",
			"Significantly smaller archive size: compressed from 6.1 GB to 1.4 GB",
			"After-install integrity check included",
			"HDD space after installation: 6.1 GB",
		},
	}

	// Detect actual fg-*.bin archives
	bins, _ := filepath.Glob(filepath.Join(repackDir, "fg-*.bin"))
	for _, b := range bins {
		if s, err := os.Stat(b); err == nil {
			meta.Archives = append(meta.Archives, ArchiveFile{
				Name: filepath.Base(b),
				Size: s.Size(),
			})
		}
	}

	// If install_script.iss is present, parse overrides
	issPath := filepath.Join(repackDir, "install_script.iss")
	if data, err := os.ReadFile(issPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "AppName=") {
				meta.AppName = strings.TrimPrefix(l, "AppName=")
			} else if strings.HasPrefix(l, "DefaultDirName=") {
				raw := strings.TrimPrefix(l, "DefaultDirName=")
				raw = strings.ReplaceAll(raw, "{code:NoSD}\\", "")
				raw = strings.ReplaceAll(raw, `\`, "/")
				meta.DefaultDirName = raw
			}
		}
	}

	return meta
}
