package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golift.io/starr/sonarr"
)

var bracketRegex = regexp.MustCompile(`^\[.*?\]\s*`)
var trailingBracketRegex = regexp.MustCompile(`\s*\[.*?\]`)
var versionSuffix = regexp.MustCompile(`v\d+$`)

// ParseFilename extracts the series title and episode number from a filename.
// Expected format (after stripping brackets): "Title - 03v2.mkv"
func ParseFilename(filename string) (title string, episodeNum int, err error) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Strip leading [bracket] groups (e.g. [SubGroup])
	for bracketRegex.MatchString(name) {
		name = bracketRegex.ReplaceAllString(name, "")
	}
	name = strings.TrimSpace(name)

	// Split on " - " to separate title from episode info
	parts := strings.SplitN(name, " - ", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("unable to parse filename: %q (no ' - ' separator found)", filename)
	}

	title = strings.TrimSpace(parts[0])
	epPart := strings.TrimSpace(parts[1])

	// Strip trailing bracket groups (e.g. "03 [1080p][HEVC]" → "03")
	epPart = trailingBracketRegex.ReplaceAllString(epPart, "")
	// Strip version suffix (e.g. "03v2" → "03")
	epPart = versionSuffix.ReplaceAllString(epPart, "")
	epPart = strings.TrimSpace(epPart)

	episodeNum, err = strconv.Atoi(epPart)
	if err != nil {
		return "", 0, fmt.Errorf("unable to parse episode number from %q: %w", filename, err)
	}

	return title, episodeNum, nil
}

// NormalizeTitle lowercases and strips non-alphanumeric characters for fuzzy matching.
func NormalizeTitle(title string) string {
	var b strings.Builder
	lastSpace := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
		} else if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

// MatchSeries finds the Sonarr series whose title or alternate titles match the parsed title.
func MatchSeries(allSeries []*sonarr.Series, title string) *sonarr.Series {
	normalized := NormalizeTitle(title)

	for _, s := range allSeries {
		if NormalizeTitle(s.Title) == normalized {
			return s
		}
		for _, alt := range s.AlternateTitles {
			if NormalizeTitle(alt.Title) == normalized {
				return s
			}
		}
	}
	return nil
}
