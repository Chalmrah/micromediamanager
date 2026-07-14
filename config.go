package main

import (
	"encoding/json"
	"os"
	"strings"
)

type Config struct {
	SonarrURL         string `json:"sonarrUrl"`
	SonarrAPIKey      string `json:"sonarrApiKey"`
	IgnoreCertificate bool   `json:"ignoreCertificate"`
	HandbrakeQuality  int    `json:"handbrakeQuality"`
	SonarrBasePath    string `json:"sonarrBasePath"`
	LocalBasePath     string `json:"localBasePath"`
}

// RemapPath replaces the Sonarr base path prefix with the local base path.
func (c Config) RemapPath(sonarrPath string) string {
	if c.SonarrBasePath != "" && c.LocalBasePath != "" {
		return strings.Replace(sonarrPath, c.SonarrBasePath, c.LocalBasePath, 1)
	}
	return sonarrPath
}

const defaultHandbrakeQuality = 24

func ReadConfig(fileName string) (Config, error) {
	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		return Config{}, err
	}

	if config.HandbrakeQuality == 0 {
		config.HandbrakeQuality = defaultHandbrakeQuality
	}

	// Avoid double slashes when building raw API URLs.
	config.SonarrURL = strings.TrimRight(config.SonarrURL, "/")

	return config, nil
}
