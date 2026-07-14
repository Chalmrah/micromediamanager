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
var versionSuffix = regexp.MustCompile(`(?i)v\d+$`)
// seasonPatterns match a trailing season indicator in a title, capturing the
// season number. Tried in order: "Title S2", "Title 2nd Season", "Title Season 2".
var seasonPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\s+S(\d+)$`),
	regexp.MustCompile(`(?i)\s+(\d+)(?:st|nd|rd|th)\s+Season$`),
	regexp.MustCompile(`(?i)\s+Season\s+(\d+)$`),
}
// romanSeasonSuffix matches a trailing roman numeral used as a season indicator
// (e.g. "Title II" → season 2). Only applied for values >= 2 so a stray trailing
// "I" is never misread as a season.
var romanSeasonSuffix = regexp.MustCompile(`(?i)\s+([ivx]+)$`)
var sceneSeasonEpisode = regexp.MustCompile(`(?i)[\.\s_]S(\d+)E(\d+)(?:[\.\s_]|$)`)

// romanToInt converts a roman numeral (case-insensitive) to its integer value,
// returning 0 if the string is not a well-formed roman numeral.
func romanToInt(s string) int {
	values := map[rune]int{'i': 1, 'v': 5, 'x': 10, 'l': 50, 'c': 100, 'd': 500, 'm': 1000}
	runes := []rune(strings.ToLower(s))
	total := 0
	for i, r := range runes {
		v, ok := values[r]
		if !ok {
			return 0
		}
		if i+1 < len(runes) && values[runes[i+1]] > v {
			total -= v
		} else {
			total += v
		}
	}
	return total
}

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

	// Split on " - " to separate title from episode info. A title may itself
	// contain " - " (e.g. "Show - The Second Act - 03"), so try each
	// occurrence left to right and use the first split whose right-hand side
	// parses as an episode number.
	var epErr error
	for searchFrom := 0; ; {
		i := strings.Index(name[searchFrom:], " - ")
		if i < 0 {
			break
		}
		i += searchFrom
		searchFrom = i + 3

		epPart := strings.TrimSpace(name[i+3:])

		// Strip trailing tag groups: [1080p], (1080p), [HEVC], [hash], etc.
		epPart = trailingTagRegex.ReplaceAllString(epPart, "")
		// Strip version suffix (e.g. "03v2" → "03")
		epPart = versionSuffix.ReplaceAllString(epPart, "")
		epPart = strings.TrimSpace(epPart)

		episodeNum, err = strconv.Atoi(epPart)
		if err != nil {
			epErr = err
			continue
		}

		title = strings.TrimSpace(name[:i])

		// Extract and strip season suffix from title (e.g. "Title S2" or
		// "Title 2nd Season" → "Title", season 2)
		seasonNum = 1
		for _, re := range seasonPatterns {
			if m := re.FindStringSubmatch(title); m != nil {
				seasonNum, _ = strconv.Atoi(m[1])
				title = strings.TrimSpace(re.ReplaceAllString(title, ""))
				explicitSeason = true
				break
			}
		}

		// Roman-numeral season suffix (e.g. "Title II" → "Title", season 2).
		// Only applied for values >= 2 so a stray trailing "I" is not misread.
		if !explicitSeason {
			if m := romanSeasonSuffix.FindStringSubmatch(title); m != nil {
				if n := romanToInt(m[1]); n >= 2 {
					seasonNum = n
					title = strings.TrimSpace(romanSeasonSuffix.ReplaceAllString(title, ""))
					explicitSeason = true
				}
			}
		}

		return title, seasonNum, episodeNum, explicitSeason, nil
	}

	// Fallback: scene release format (e.g. "Title.Name.S01E02.stuff.mkv")
	if m := sceneSeasonEpisode.FindStringSubmatchIndex(name); m != nil {
		// Title is everything before the SxxExx match, with dot/underscore
		// separators replaced by spaces
		title = strings.TrimSpace(name[:m[0]])
		title = strings.ReplaceAll(title, ".", " ")
		title = strings.ReplaceAll(title, "_", " ")
		seasonNum, _ = strconv.Atoi(name[m[2]:m[3]])
		episodeNum, _ = strconv.Atoi(name[m[4]:m[5]])
		return title, seasonNum, episodeNum, true, nil
	}

	if epErr != nil {
		return "", 0, 0, false, fmt.Errorf("unable to parse episode number from %q: %w", filename, epErr)
	}
	return "", 0, 0, false, fmt.Errorf("unable to parse filename: %q (no ' - ' separator or SxxExx pattern found)", filename)
}

// NormalizeTitle lowercases and strips non-alphanumeric characters for fuzzy matching.
func NormalizeTitle(title string) string {
	// Spell out ampersands so "Tom & Jerry" matches "Tom and Jerry".
	title = strings.ReplaceAll(title, "&", " and ")
	var b strings.Builder
	lastSpace := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
		} else if r == '\'' || r == '’' {
			// Drop apostrophes rather than treating them as separators, so
			// "Aren't" normalizes to "arent" and matches scene releases that
			// omit the apostrophe entirely ("Arent").
			continue
		} else if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

// MatchEpisode finds the matching episode from a list using the following strategy:
//   - If the series is anime and the filename has no explicit season, match by absolute episode number
//   - Otherwise match by season + episode number, falling back to scene numbering
//     (handles anime where fansub seasons differ from TVDB structure)
func MatchEpisode(episodes []*Episode, seasonNum, episodeNum int, isAnime, explicitSeason bool) *Episode {
	useAbsolute := isAnime && !explicitSeason
	for _, ep := range episodes {
		if useAbsolute && ep.AbsoluteEpisodeNumber == episodeNum {
			return ep
		}
		if !useAbsolute && ep.EpisodeNumber == episodeNum && ep.SeasonNumber == seasonNum {
			return ep
		}
		if !useAbsolute && ep.SceneEpisodeNumber == episodeNum && ep.SceneSeasonNumber == seasonNum {
			return ep
		}
	}
	return nil
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
