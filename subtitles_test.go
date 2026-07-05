package main

import (
	"reflect"
	"testing"
)

func TestForcedTracksToClear(t *testing.T) {
	tests := []struct {
		name string
		subs []subtitleTrack
		want []int
	}{
		{
			name: "only English track is forced -> clear it",
			subs: []subtitleTrack{{Language: "eng", Forced: true}},
			want: []int{1},
		},
		{
			name: "full + forced English pair -> leave alone",
			subs: []subtitleTrack{
				{Language: "eng", Forced: false},
				{Language: "eng", Forced: true},
			},
			want: nil,
		},
		{
			name: "full English only -> nothing to do",
			subs: []subtitleTrack{{Language: "eng", Forced: false}},
			want: nil,
		},
		{
			name: "no English track -> nothing to do",
			subs: []subtitleTrack{
				{Language: "jpn", Forced: true},
				{Language: "spa", Forced: false},
			},
			want: nil,
		},
		{
			name: "forced English after non-English tracks -> correct ordinal",
			subs: []subtitleTrack{
				{Language: "jpn", Forced: false},
				{Language: "spa", Forced: false},
				{Language: "eng", Forced: true},
			},
			want: []int{3},
		},
		{
			name: "multiple forced-only English tracks -> clear all",
			subs: []subtitleTrack{
				{Language: "en", Forced: true},
				{Language: "eng", Forced: true},
			},
			want: []int{1, 2},
		},
		{
			name: "en variant counts as English full -> leave forced alone",
			subs: []subtitleTrack{
				{Language: "en", Forced: false},
				{Language: "eng", Forced: true},
			},
			want: nil,
		},
		{
			name: "no subtitles at all",
			subs: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := forcedTracksToClear(tt.subs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("forcedTracksToClear(%v) = %v, want %v", tt.subs, got, tt.want)
			}
		})
	}
}
