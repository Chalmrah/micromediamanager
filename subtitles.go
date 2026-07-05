package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	ffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

// subtitleTrack is the minimal view of a subtitle stream needed to decide
// whether its forced flag is bogus. The slice order matches the file's
// subtitle track order, so a track's 1-based index maps to mkvpropedit's
// track:sN selector.
type subtitleTrack struct {
	Language string
	Forced   bool
}

// isEnglish reports whether an ffprobe language tag denotes English.
func isEnglish(lang string) bool {
	switch strings.ToLower(lang) {
	case "eng", "en":
		return true
	}
	return false
}

// forcedTracksToClear returns the 1-based subtitle track numbers whose forced
// flag should be cleared. It only acts when every English subtitle track is
// forced (i.e. there is no normal/full English track); a genuine full+forced
// English pair is left untouched.
func forcedTracksToClear(subs []subtitleTrack) []int {
	var forced []int
	engNormal := false
	for i, s := range subs {
		if !isEnglish(s.Language) {
			continue
		}
		if s.Forced {
			forced = append(forced, i+1)
		} else {
			engNormal = true
		}
	}
	if engNormal {
		return nil
	}
	return forced
}

// englishForcedTracksToClear probes filePath and returns the subtitle track
// numbers (1-based, for mkvpropedit track:sN) whose forced flag should be
// cleared, applying the full+forced-pair guard.
func englishForcedTracksToClear(filePath string) ([]int, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	probeData, err := ffprobe.ProbeURL(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("error probing file: %w", err)
	}

	var subs []subtitleTrack
	for _, stream := range probeData.Streams {
		if stream.CodecType != string(ffprobe.StreamSubtitle) {
			continue
		}
		lang, _ := stream.TagList.GetString("language")
		subs = append(subs, subtitleTrack{
			Language: lang,
			Forced:   stream.Disposition.Forced == 1,
		})
	}

	return forcedTracksToClear(subs), nil
}

// clearForcedFlags clears (sets to 0) the forced flag on the given 1-based
// subtitle tracks of an MKV using mkvpropedit. Only that header property is
// touched; no stream data is re-encoded.
func clearForcedFlags(filePath string, tracks []int) error {
	if _, err := exec.LookPath("mkvpropedit"); err != nil {
		return fmt.Errorf("mkvpropedit not found in $PATH; please install mkvtoolnix")
	}

	args := []string{filePath}
	for _, n := range tracks {
		args = append(args, "--edit", fmt.Sprintf("track:s%d", n), "--set", "flag-forced=0")
	}

	cmd := exec.Command("mkvpropedit", args...)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
