package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golift.io/starr/sonarr"
)

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("failed to copy content: %w", err)
	}

	err = dstFile.Sync()
	if err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

func readSourceFolderFiles(sourceFolder string) []os.DirEntry {
	dir, err := os.ReadDir(sourceFolder)
	if err != nil {
		log.Fatalf("Unable to read folder %v", sourceFolder)
	}

	return dir
}

// buildDestinationPath constructs the full output path using Sonarr series/episode metadata.
// Format: {series.Path}/Season {N}/{Title} - S{ss}E{ee}{ext}
// The series path is remapped from Sonarr's path to the local path using config.
func buildDestinationPath(config Config, series *sonarr.Series, episode *Episode, ext string) string {
	seasonDir := fmt.Sprintf("Season %d", episode.SeasonNumber)
	filename := fmt.Sprintf("%s - S%02dE%02d%s", series.Title, episode.SeasonNumber, episode.EpisodeNumber, ext)
	return filepath.Join(config.RemapPath(series.Path), seasonDir, filename)
}
