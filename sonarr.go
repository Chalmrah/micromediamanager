package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"golift.io/starr"
	"golift.io/starr/sonarr"
)

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
