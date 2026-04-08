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
var trailingTagRegex = regexp.MustCompile(`\s*[\[\(].*?[\]\)]`)
var versionSuffix = regexp.MustCompile(`v\d+$`)
var seasonSuffix = regexp.MustCompile(`(?i)\s+S(\d+)$`)
var sceneSeasonEpisode = regexp.MustCompile(`(?i)[\.\s]S(\d+)E(\d+)[\.\s]`)

// ParseFilename extracts the series title, season number, and episode number from a filename.
// Expected format (after stripping brackets): "Title S2 - 03v2.mkv"
// Also supports scene release format: "Title.Name.S01E02.stuff.mkv"
// If no season suffix is present, seasonNum defaults to 1.
// explicitSeason is true when the filename contains an explicit season indicator (S2, S01E01, etc.).
func ParseFilename(filename string) (title string, seasonNum int, episodeNum int, explicitSeason bool, err error) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Strip leading [bracket] groups (e.g. [SubGroup])
	for bracketRegex.MatchString(name) {
		name = bracketRegex.ReplaceAllString(name, "")
	}
	name = strings.TrimSpace(name)

	// Split on " - " to separate title from episode info
	parts := strings.SplitN(name, " - ", 2)
	if len(parts) == 2 {
		title = strings.TrimSpace(parts[0])

		// Extract and strip season suffix from title (e.g. "Title S2" → "Title", season 2)
		seasonNum = 1
		if m := seasonSuffix.FindStringSubmatch(title); m != nil {
			seasonNum, _ = strconv.Atoi(m[1])
			title = seasonSuffix.ReplaceAllString(title, "")
			explicitSeason = true
		}

		epPart := strings.TrimSpace(parts[1])

		// Strip trailing tag groups: [1080p], (1080p), [HEVC], [hash], etc.
		epPart = trailingTagRegex.ReplaceAllString(epPart, "")
		// Strip version suffix (e.g. "03v2" → "03")
		epPart = versionSuffix.ReplaceAllString(epPart, "")
		epPart = strings.TrimSpace(epPart)

		episodeNum, err = strconv.Atoi(epPart)
		if err != nil {
			return "", 0, 0, false, fmt.Errorf("unable to parse episode number from %q: %w", filename, err)
		}

		return title, seasonNum, episodeNum, explicitSeason, nil
	}

	// Fallback: scene release format (e.g. "Title.Name.S01E02.stuff.mkv")
	if m := sceneSeasonEpisode.FindStringSubmatchIndex(name); m != nil {
		// Title is everything before the SxxExx match, with dots replaced by spaces
		title = strings.ReplaceAll(strings.TrimSpace(name[:m[0]]), ".", " ")
		seasonNum, _ = strconv.Atoi(name[m[2]:m[3]])
		episodeNum, _ = strconv.Atoi(name[m[4]:m[5]])
		return title, seasonNum, episodeNum, true, nil
	}

	return "", 0, 0, false, fmt.Errorf("unable to parse filename: %q (no ' - ' separator or SxxExx pattern found)", filename)
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
