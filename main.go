package main

import (
	"fmt"
	"log"
	"micromediamanager/handbrake"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"golift.io/starr/sonarr"
)

var (
	configFile   string
	sourceFolder string
	version      bool
	dryRun       bool
	goVersion    = runtime.Version()
	buildDate    = "unknown"
	buildCommit  = "dev"
	buildVersion = "unknown"
)

func main() {
	pflag.StringVarP(&configFile, "configFile", "c", "", "Location of config json")
	pflag.StringVarP(&sourceFolder, "sourceFolder", "s", "", "Location of source media folder")
	pflag.BoolVarP(&version, "version", "v", false, "Displays version information")
	pflag.BoolVar(&dryRun, "dryRun", false, "Preview actions without transcoding, copying, or triggering rescans")
	pflag.Parse()

	if version {
		fmt.Printf("MicroMediaManager  %s\n- commit hash: %s\n- build date: %s\n- go version: %s\n", buildVersion, buildCommit, buildDate, goVersion)
		os.Exit(0)
	}

	if configFile == "" || sourceFolder == "" {
		log.Fatalf("Required flags missing. Usage:\n  micromediamanager --configFile <path> --sourceFolder <path>")
	}

	green := color.New(color.FgHiGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()

	if dryRun {
		fmt.Printf("%s No files will be modified\n\n", cyan("[DRY RUN]"))
	}

	config, err := ReadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	client := newSonarrClient(config)

	allSeries, err := client.GetAllSeries()
	if err != nil {
		log.Fatalf("Failed to fetch series from Sonarr: %v", err)
	}
	log.Printf("Fetched %d series from Sonarr", len(allSeries))

	fileList := readSourceFolderFiles(sourceFolder)

	episodeCache := make(map[int64][]*sonarr.Episode)
	affectedSeries := make(map[int64]*sonarr.Series)
	var unmatchedFiles []string
	processedCount := 0

	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		title, episodeNum, err := ParseFilename(file.Name())
		if err != nil {
			log.Printf("%s Unable to parse: %s (%v)", yellow("WARN"), file.Name(), err)
			unmatchedFiles = append(unmatchedFiles, file.Name())
			continue
		}

		series := MatchSeries(allSeries, title)
		if series == nil {
			log.Printf("%s No Sonarr match for: %s (parsed title: %q)", yellow("WARN"), file.Name(), title)
			unmatchedFiles = append(unmatchedFiles, file.Name())
			continue
		}

		// Fetch episodes for this series (cached)
		episodes, ok := episodeCache[series.ID]
		if !ok {
			episodes, err = client.GetSeriesEpisodes(&sonarr.GetEpisode{SeriesID: series.ID})
			if err != nil {
				log.Printf("Failed to fetch episodes for %s: %v", series.Title, err)
				continue
			}
			episodeCache[series.ID] = episodes
		}

		// Find matching episode by absolute number (anime) or episode number
		var matchedEpisode *sonarr.Episode
		isAnime := series.SeriesType == "anime"
		for _, ep := range episodes {
			if isAnime && ep.AbsoluteEpisodeNumber == episodeNum {
				matchedEpisode = ep
				break
			}
			if !isAnime && ep.EpisodeNumber == episodeNum && ep.SeasonNumber == 1 {
				matchedEpisode = ep
				break
			}
		}

		if matchedEpisode == nil {
			log.Printf("%s No episode %d found for %s in Sonarr", yellow("WARN"), episodeNum, series.Title)
			unmatchedFiles = append(unmatchedFiles, file.Name())
			continue
		}

		if matchedEpisode.HasFile {
			log.Printf("Skipping %s - episode already has a file in Sonarr", file.Name())
			continue
		}

		ext := filepath.Ext(file.Name())
		destinationPath := buildDestinationPath(series, matchedEpisode, ext)

		fmt.Printf("%s %s %s\n", green(file.Name()), red("-->"), filepath.Base(destinationPath))

		encoding, err := getVideoCodec(filepath.Join(sourceFolder, file.Name()))
		if err != nil {
			log.Printf("Get video codec %s error: %v", file.Name(), err)
			continue
		}

		if encoding == "hevc" {
			log.Printf("Action: copy (already HEVC)")
		} else {
			log.Printf("Action: transcode %s to HEVC (quality %d)", encoding, config.HandbrakeQuality)
		}

		if !dryRun {
			// Ensure destination directory exists
			folderName := filepath.Dir(destinationPath)
			if _, err := os.Stat(folderName); os.IsNotExist(err) {
				if err := os.MkdirAll(folderName, os.ModePerm); err != nil {
					log.Printf("Unable to create folder: %v", folderName)
					continue
				}
			}

			srcPath := filepath.Join(sourceFolder, file.Name())
			if encoding == "hevc" {
				err := copyFile(srcPath, destinationPath)
				if err != nil {
					log.Printf("Unable to copy file %s: %v", file.Name(), err)
					continue
				}
			} else {
				_, err := handbrake.Run(srcPath, destinationPath, config.HandbrakeQuality)
				if err != nil {
					log.Printf("Handbrake error for %s: %v", file.Name(), err)
					continue
				}
			}
		}

		processedCount++
		affectedSeries[series.ID] = series
	}

	// Trigger Sonarr rescan for each affected series
	for _, series := range affectedSeries {
		if dryRun {
			log.Printf("Would trigger Sonarr rescan for %s", series.Title)
		} else {
			log.Printf("Triggering Sonarr rescan for %s", series.Title)
			_, err := client.SendCommand(&sonarr.CommandRequest{
				Name:     "RescanSeries",
				SeriesID: series.ID,
			})
			if err != nil {
				log.Printf("Failed to trigger rescan for %s: %v", series.Title, err)
			}
		}
	}

	// Summary
	fmt.Println()
	if processedCount > 0 {
		fmt.Printf("%s Processed %d file(s) across %d series\n", green("DONE"), processedCount, len(affectedSeries))
	} else {
		fmt.Println("No files processed")
	}

	if len(unmatchedFiles) > 0 {
		fmt.Printf("\n%s %d unmatched file(s):\n", yellow("WARNING"), len(unmatchedFiles))
		for _, f := range unmatchedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
}
