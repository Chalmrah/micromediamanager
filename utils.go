package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func getDestinationFileName(dst string, season int, extension string) (string, error) {

	dstFolder := filepath.Join(dst, "Season "+strconv.Itoa(season))
	episodeNumber, err := getEpisodeNumber(dstFolder)
	if errors.Is(err, os.ErrNotExist) {
		episodeNumber = 1
	} else if err != nil {
		return "", err
	}
	episodename := fmt.Sprintf("%v - %vx%v%v", filepath.Base(dst), season, episodeNumber, extension)
	return episodename, nil
}

func getEpisodeNumber(folder string) (int, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return 0, err
	}
	return len(files) + 1, nil
}

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
		return fmt.Errorf("failed to copy content: %w", err)
	}

	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
func filterFileList(fileList []os.DirEntry, pattern string) []os.DirEntry {
	filteredList := []os.DirEntry{}
	for _, file := range fileList {
		matched, err := filepath.Match(pattern, file.Name())
		if err != nil {
			log.Printf("Error matching pattern %v: %v", file.Name(), err)
			continue
		}
		if matched {
			filteredList = append(filteredList, file)
		}

	}
	return filteredList
}

func readSourceFolderFiles(sourceFolder string) []os.DirEntry {

	dir, err := os.ReadDir(sourceFolder)
	if err != nil {
		log.Fatalf("Unable to read folder %v", sourceFolder)
	}

	return dir
}
