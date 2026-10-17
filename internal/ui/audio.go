package ui

import (
	"os/exec"
)

var (
	musicCmd      *exec.Cmd
	musicPlaying  bool
	musicFilePath = "/tmp/fg_audio/tmp/music.aac"
)

// IsMusicPlaying returns the current background music state
func IsMusicPlaying() bool {
	return musicPlaying
}

// ToggleMusic toggles background repack music playback
func ToggleMusic() bool {
	if musicPlaying {
		StopMusic()
		return false
	}
	return StartMusic()
}

// StartMusic launches background music if an audio player is available
func StartMusic() bool {
	if musicPlaying {
		return true
	}

	player, err := exec.LookPath("mpv")
	var args []string
	if err == nil {
		args = []string{"--no-video", "--loop", musicFilePath}
	} else {
		player, err = exec.LookPath("ffplay")
		if err == nil {
			args = []string{"-nodisp", "-loop", "0", musicFilePath}
		}
	}

	if player == "" {
		return false
	}

	musicCmd = exec.Command(player, args...)
	if err := musicCmd.Start(); err == nil {
		musicPlaying = true
		return true
	}
	return false
}

// StopMusic stops background music playback
func StopMusic() {
	if musicCmd != nil && musicCmd.Process != nil {
		_ = musicCmd.Process.Kill()
		_ = musicCmd.Wait()
		musicCmd = nil
	}
	musicPlaying = false
}
