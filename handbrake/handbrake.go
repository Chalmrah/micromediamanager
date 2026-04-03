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
		"--encoder-preset=veryslow",
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
		return false, err
	}

	return true, nil
}

func verifyHandbrakeInstalled() bool {
	_, err := exec.LookPath("handbrakecli")
	return err == nil
}
