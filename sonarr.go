package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golift.io/starr"
	"golift.io/starr/sonarr"
)

// Episode extends the starr Episode type with scene numbering fields
// that the upstream library doesn't include.
type Episode struct {
	sonarr.Episode
	SceneSeasonNumber   int `json:"sceneSeasonNumber"`
	SceneEpisodeNumber  int `json:"sceneEpisodeNumber"`
	SceneAbsoluteNumber int `json:"sceneAbsoluteEpisodeNumber"`
}

func newSonarrClient(config Config) *sonarr.Sonarr {
	starrConfig := starr.New(config.SonarrAPIKey, config.SonarrURL, 30*time.Second)

	if config.IgnoreCertificate {
		starrConfig.Client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}

	return sonarr.New(starrConfig)
}

// getSeriesEpisodes fetches episodes for a series using a raw API call
// to preserve scene numbering fields that the starr library drops.
func getSeriesEpisodes(config Config, client *http.Client, seriesID int64) ([]*Episode, error) {
	url := fmt.Sprintf("%s/api/v3/episode?seriesId=%d", config.SonarrURL, seriesID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Key", config.SonarrAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var episodes []*Episode
	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return nil, fmt.Errorf("failed to decode episodes: %w", err)
	}

	return episodes, nil
}
