package main

import (
	"fmt"
	"os"
	"os/exec"
)

// remuxToMKV copies every stream from inputFile into a Matroska container at
// outputFile without re-encoding (ffmpeg -map 0 -c copy). Stream dispositions
// such as the forced flag are preserved.
func remuxToMKV(inputFile, outputFile string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in $PATH; please install ffmpeg")
	}

	cmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", inputFile,
		"-map", "0",
		"-c", "copy",
		"-y",
		outputFile,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Don't leave a partial file behind for Sonarr to import.
		os.Remove(outputFile)
		return err
	}
	return nil
}
