package main

import (
	"testing"

	"golift.io/starr/sonarr"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		wantTitle      string
		wantSeason     int
		wantEpNum      int
		wantExplicit   bool
		wantErr        bool
	}{
		{
			name:       "standard format",
			filename:   "[SubGroup] My Anime - 03.mkv",
			wantTitle:  "My Anime",
			wantSeason: 1,
			wantEpNum:  3,
		},
		{
			name:       "version suffix",
			filename:   "[SubGroup] My Anime - 03v2.mkv",
			wantTitle:  "My Anime",
			wantSeason: 1,
			wantEpNum:  3,
		},
		{
			name:       "trailing bracket tags",
			filename:   "[SubGroup] My Anime - 03 [1080p][HEVC].mkv",
			wantTitle:  "My Anime",
			wantSeason: 1,
			wantEpNum:  3,
		},
		{
			name:       "trailing tags with version",
			filename:   "[SubGroup] My Anime - 03v2 [1080p][HEVC].mkv",
			wantTitle:  "My Anime",
			wantSeason: 1,
			wantEpNum:  3,
		},
		{
			name:       "multiple leading bracket groups",
			filename:   "[SubGroup][720p] My Anime - 12.mkv",
			wantTitle:  "My Anime",
			wantSeason: 1,
			wantEpNum:  12,
		},
		{
			name:         "parenthesized tags",
			filename:     "[SubsPlease] Dorohedoro S2 - 01 (1080p) [B0159228].mkv",
			wantTitle:    "Dorohedoro",
			wantSeason:   2,
			wantEpNum:    1,
			wantExplicit: true,
		},
		{
			name:         "season suffix stripped from title",
			filename:     "Tensei Shitara Slime Datta Ken S4 - 01 (1080p) [182FF0C9].mkv",
			wantTitle:    "Tensei Shitara Slime Datta Ken",
			wantSeason:   4,
			wantEpNum:    1,
			wantExplicit: true,
		},
		{
			name:         "version with parenthesized and bracket tags",
			filename:     "Otonari no Tenshi-sama ni Itsunomanika Dame Ningen ni Sareteita Ken S2 - 01v2 (1080p) [6504C9CC].mkv",
			wantTitle:    "Otonari no Tenshi-sama ni Itsunomanika Dame Ningen ni Sareteita Ken",
			wantSeason:   2,
			wantEpNum:    1,
			wantExplicit: true,
		},
		{
			name:         "empty leading bracket",
			filename:     "[] Sousou no Frieren S2 - 04 (1080p) [698A157A].mkv",
			wantTitle:    "Sousou no Frieren",
			wantSeason:   2,
			wantEpNum:    4,
			wantExplicit: true,
		},
		{
			name:       "no season defaults to 1",
			filename:   "[Sub] Title - 05.mkv",
			wantTitle:  "Title",
			wantSeason: 1,
			wantEpNum:  5,
		},
		{
			name:    "no separator",
			filename: "[SubGroup] My Anime 03.mkv",
			wantErr: true,
		},
		{
			name:    "non-numeric episode",
			filename: "[SubGroup] My Anime - SP1.mkv",
			wantErr: true,
		},
		{
			name:         "scene release dot-separated",
			filename:     "Farming.Life.in.Another.World.S02E01.1080p.ADN.WEB-DL.JPN.AAC2.0.H.264.MSubs-ToonsHub.mkv",
			wantTitle:    "Farming Life in Another World",
			wantSeason:   2,
			wantEpNum:    1,
			wantExplicit: true,
		},
		{
			name:         "scene release with higher episode",
			filename:     "Some.Show.S02E13.Episode.Title.720p.WEB-DL.mkv",
			wantTitle:    "Some Show",
			wantSeason:   2,
			wantEpNum:    13,
			wantExplicit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, seasonNum, epNum, explicitSeason, err := ParseFilename(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got title=%q season=%d epNum=%d", title, seasonNum, epNum)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if seasonNum != tt.wantSeason {
				t.Errorf("seasonNum = %d, want %d", seasonNum, tt.wantSeason)
			}
			if epNum != tt.wantEpNum {
				t.Errorf("epNum = %d, want %d", epNum, tt.wantEpNum)
			}
			if explicitSeason != tt.wantExplicit {
				t.Errorf("explicitSeason = %v, want %v", explicitSeason, tt.wantExplicit)
			}
		})
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "lowercase", input: "My Anime", want: "my anime"},
		{name: "punctuation stripped", input: "Re:Zero!", want: "re zero"},
		{name: "mixed case and symbols", input: "Sword Art Online: Alicization", want: "sword art online alicization"},
		{name: "CJK characters preserved", input: "進撃の巨人", want: "進撃の巨人"},
		{name: "multiple spaces collapsed", input: "Title   With   Spaces", want: "title with spaces"},
		{name: "leading trailing spaces", input: "  Hello  ", want: "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTitle(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMatchSeries(t *testing.T) {
	series := []*sonarr.Series{
		{
			Title: "Attack on Titan",
			AlternateTitles: []*sonarr.AlternateTitle{
				{Title: "Shingeki no Kyojin"},
			},
		},
		{
			Title: "My Hero Academia",
		},
	}

	tests := []struct {
		name      string
		title     string
		wantMatch string
		wantNil   bool
	}{
		{name: "exact match", title: "Attack on Titan", wantMatch: "Attack on Titan"},
		{name: "case insensitive", title: "attack on titan", wantMatch: "Attack on Titan"},
		{name: "alternate title", title: "Shingeki no Kyojin", wantMatch: "Attack on Titan"},
		{name: "no match", title: "Naruto", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchSeries(series, tt.title)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %q", got.Title)
				}
				return
			}
			if got == nil {
				t.Fatal("expected match, got nil")
			}
			if got.Title != tt.wantMatch {
				t.Errorf("matched %q, want %q", got.Title, tt.wantMatch)
			}
		})
	}
}
