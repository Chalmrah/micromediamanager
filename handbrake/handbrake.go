package handbrake

import (
	"log"
	"os"
	"os/exec"
)

func Run(inputFile, outputFile string, quality int) (bool, error) {
	if !verifyHandbrakeInstalled() {
		log.Printf("Handbrake not installed. Ensure it is installed and in $PATH")
		return false, nil
	}

	cmdArgs := []string{
		"--encoder", "nvenc_h265",
		"--encoder-preset=veryslow",
		"--encoder-profile=main",
		"--all-audio",
		"--all-subtitles",
		"-q", string(rune(quality)),
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
