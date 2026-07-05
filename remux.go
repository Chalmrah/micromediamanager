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
		"-i", inputFile,
		"-map", "0",
		"-c", "copy",
		"-y",
		outputFile,
	)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
