package main

import (
	"testing"

	"golift.io/starr/sonarr"
)

func TestBuildDestinationPath(t *testing.T) {
	config := Config{SonarrBasePath: "/data", LocalBasePath: "/mnt"}

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "plain title",
			title: "My Anime",
			want:  "/mnt/anime/My Anime/Season 1/My Anime - S01E02.mkv",
		},
		{
			name:  "slash in title sanitized",
			title: "Fate/stay night",
			want:  "/mnt/anime/My Anime/Season 1/Fate-stay night - S01E02.mkv",
		},
		{
			name:  "colon in title sanitized",
			title: "Re:Zero",
			want:  "/mnt/anime/My Anime/Season 1/ReZero - S01E02.mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			series := &sonarr.Series{Title: tt.title, Path: "/data/anime/My Anime"}
			episode := &Episode{Episode: sonarr.Episode{SeasonNumber: 1, EpisodeNumber: 2}}
			got := buildDestinationPath(config, series, episode, ".mkv")
			if got != tt.want {
				t.Errorf("buildDestinationPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
