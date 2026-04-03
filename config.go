package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	SonarrURL         string `json:"sonarrUrl"`
	SonarrAPIKey      string `json:"sonarrApiKey"`
	IgnoreCertificate bool   `json:"ignoreCertificate"`
	HandbrakeQuality  int    `json:"handbrakeQuality"`
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

	return config, nil
}
