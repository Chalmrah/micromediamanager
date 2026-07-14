package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("failed to copy content: %w", err)
	}

	if err = dstFile.Sync(); err != nil {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// On network filesystems write errors can surface at close, so check it.
	if err = dstFile.Close(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	return nil
}

func readSourceFolderFiles(sourceFolder string) []os.DirEntry {
	dir, err := os.ReadDir(sourceFolder)
	if err != nil {
		log.Fatalf("Unable to read folder %v: %v", sourceFolder, err)
	}

	return dir
}

// filenameSanitizer replaces characters that are illegal in file names on
// Linux or common SMB/NTFS mounts, so titles like "Fate/stay night" don't
// create bogus subdirectories.
var filenameSanitizer = strings.NewReplacer(
	"/", "-", "\\", "-",
	":", "", "*", "", "?", "", `"`, "", "<", "", ">", "", "|", "",
)

// buildDestinationPath constructs the full output path using Sonarr series/episode metadata.
// Format: {series.Path}/Season {N}/{Title} - S{ss}E{ee}{ext}
// The series path is remapped from Sonarr's path to the local path using config.
func buildDestinationPath(config Config, series *sonarr.Series, episode *Episode, ext string) string {
	seasonDir := fmt.Sprintf("Season %d", episode.SeasonNumber)
	title := filenameSanitizer.Replace(series.Title)
	filename := fmt.Sprintf("%s - S%02dE%02d%s", title, episode.SeasonNumber, episode.EpisodeNumber, ext)
	return filepath.Join(config.RemapPath(series.Path), seasonDir, filename)
}
