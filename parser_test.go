package main

import (
	"testing"

	"golift.io/starr/sonarr"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantTitle  string
		wantEpNum  int
		wantErr    bool
	}{
		{
			name:      "standard format",
			filename:  "[SubGroup] My Anime - 03.mkv",
			wantTitle: "My Anime",
			wantEpNum: 3,
		},
		{
			name:      "version suffix",
			filename:  "[SubGroup] My Anime - 03v2.mkv",
			wantTitle: "My Anime",
			wantEpNum: 3,
		},
		{
			name:      "trailing bracket tags",
			filename:  "[SubGroup] My Anime - 03 [1080p][HEVC].mkv",
			wantTitle: "My Anime",
			wantEpNum: 3,
		},
		{
			name:      "trailing tags with version",
			filename:  "[SubGroup] My Anime - 03v2 [1080p][HEVC].mkv",
			wantTitle: "My Anime",
			wantEpNum: 3,
		},
		{
			name:      "multiple leading bracket groups",
			filename:  "[SubGroup][720p] My Anime - 12.mkv",
			wantTitle: "My Anime",
			wantEpNum: 12,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, epNum, err := ParseFilename(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got title=%q epNum=%d", title, epNum)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if epNum != tt.wantEpNum {
				t.Errorf("epNum = %d, want %d", epNum, tt.wantEpNum)
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
