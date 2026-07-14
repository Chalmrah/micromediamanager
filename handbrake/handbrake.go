package handbrake

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func Run(inputFile, outputFile string, quality int) (bool, error) {
	if !verifyHandbrakeInstalled() {
		return false, fmt.Errorf("handbrakecli not found in $PATH; please install HandBrake CLI")
	}

	cmdArgs := []string{
		"--encoder", "nvenc_h265",
		// NVENC presets are fastest..slowest (not x265 names like "veryslow",
		// which HandBrake silently ignores). Medium is the speed/quality
		// balance this tool has always effectively encoded at.
		"--encoder-preset=medium",
		"--encoder-profile=main",
		"--all-audio",
		"--all-subtitles",
		"-q", strconv.Itoa(quality),
		"-f", "av_mkv",
		"--input", inputFile,
		"--output", outputFile,
	}

	cmd := exec.Command("handbrakecli", cmdArgs...)

	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		// Don't leave a partial file behind for Sonarr to import.
		os.Remove(outputFile)
		return false, err
	}

	return true, nil
}

func verifyHandbrakeInstalled() bool {
	_, err := exec.LookPath("handbrakecli")
	return err == nil
}
