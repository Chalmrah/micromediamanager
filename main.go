package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"micromediamanager/handbrake"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	pflag.BoolVarP(&dryRun, "dry-run", "d", false, "Preview actions without transcoding, copying, or triggering rescans")
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
	if dryRun {
		log.Printf("Fetched %d series from Sonarr", len(allSeries))
	}

	fileList := readSourceFolderFiles(sourceFolder)

	// Build HTTP client for raw API calls (reuses TLS config from Sonarr client)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	if config.IgnoreCertificate {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}

	episodeCache := make(map[int64][]*Episode)
	affectedSeries := make(map[int64]*sonarr.Series)
	var unmatchedFiles []string
	processedCount := 0
	skippedCount := 0

	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		title, seasonNum, episodeNum, explicitSeason, err := ParseFilename(file.Name())
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
			episodes, err = getSeriesEpisodes(config, httpClient, series.ID)
			if err != nil {
				log.Printf("Failed to fetch episodes for %s: %v", series.Title, err)
				continue
			}
			episodeCache[series.ID] = episodes
		}

		matchedEpisode := MatchEpisode(episodes, seasonNum, episodeNum, series.SeriesType == "anime", explicitSeason)

		if matchedEpisode == nil {
			log.Printf("%s No episode %d found for %s in Sonarr", yellow("WARN"), episodeNum, series.Title)
			unmatchedFiles = append(unmatchedFiles, file.Name())
			continue
		}

		if matchedEpisode.HasFile {
			skippedCount++
			continue
		}

		srcExt := filepath.Ext(file.Name())
		srcPath := filepath.Join(sourceFolder, file.Name())
		// Everything is normalised to MKV so the container is consistent and
		// the forced-flag fix below always has an MKV to edit.
		destinationPath := buildDestinationPath(config, series, matchedEpisode, ".mkv")

		fmt.Printf("%s %s %s\n", green(file.Name()), red("-->"), filepath.Base(destinationPath))

		encoding, err := getVideoCodec(srcPath)
		if err != nil {
			log.Printf("Get video codec %s error: %v", file.Name(), err)
			continue
		}

		if dryRun {
			switch {
			case encoding == "hevc" && strings.EqualFold(srcExt, ".mkv"):
				log.Printf("Action: copy (already HEVC MKV)")
			case encoding == "hevc":
				log.Printf("Action: remux %s to MKV (already HEVC)", srcExt)
			default:
				log.Printf("Action: transcode %s to HEVC MKV (quality %d)", encoding, config.HandbrakeQuality)
			}
			if tracks, err := englishForcedTracksToClear(srcPath); err == nil && len(tracks) > 0 {
				log.Printf("Action: clear forced flag on subtitle track(s) %v", tracks)
			}
		}

		if !dryRun {
			// Ensure destination directory exists
			folderName := filepath.Dir(destinationPath)
			if _, err := os.Stat(folderName); os.IsNotExist(err) {
				if err := os.MkdirAll(folderName, os.ModePerm); err != nil {
					log.Printf("Unable to create folder %q: %v", folderName, err)
					continue
				}
			}

			if encoding == "hevc" {
				if strings.EqualFold(srcExt, ".mkv") {
					if err := copyFile(srcPath, destinationPath); err != nil {
						log.Printf("Unable to copy file %s: %v", file.Name(), err)
						continue
					}
				} else {
					if err := remuxToMKV(srcPath, destinationPath); err != nil {
						log.Printf("Unable to remux file %s: %v", file.Name(), err)
						continue
					}
				}
			} else {
				if _, err := handbrake.Run(srcPath, destinationPath, config.HandbrakeQuality); err != nil {
					log.Printf("Handbrake error for %s: %v", file.Name(), err)
					continue
				}
			}

			// Clear a bogus forced flag on the English subtitle track when the
			// only English track is forced (header edit, no re-encode).
			if tracks, err := englishForcedTracksToClear(destinationPath); err != nil {
				log.Printf("%s subtitle probe failed for %s: %v", yellow("WARN"), filepath.Base(destinationPath), err)
			} else if len(tracks) > 0 {
				if err := clearForcedFlags(destinationPath, tracks); err != nil {
					log.Printf("%s mkvpropedit failed for %s: %v", yellow("WARN"), filepath.Base(destinationPath), err)
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
	if skippedCount > 0 {
		fmt.Printf("%s %d file(s) skipped (episode already has a file in Sonarr)\n", cyan("SKIP"), skippedCount)
	}

	if len(unmatchedFiles) > 0 {
		fmt.Printf("\n%s %d unmatched file(s):\n", yellow("WARNING"), len(unmatchedFiles))
		for _, f := range unmatchedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
}
