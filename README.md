# MicroMediaManager

A CLI tool that automates importing anime and series media files into [Sonarr](https://sonarr.tv/). It parses filenames from a source folder, matches them against your Sonarr library, transcodes H.264 files to HEVC using HandBrake, and copies the results to Sonarr-managed paths.

## Features

- **Sonarr integration** -- Fetches your series library and episode data via the Sonarr API. Automatically triggers a rescan after importing files.
- **Anime-aware filename parsing** -- Parses `[SubGroup] Title - 03.mkv` style filenames, handling leading bracket groups, version suffixes (`v2`), and trailing metadata tags (`[1080p][HEVC]`).
- **Fuzzy title matching** -- Normalizes titles (case, punctuation, whitespace) and matches against Sonarr series titles and alternate titles.
- **Automatic transcoding** -- Detects video codec via ffprobe. H.264 files are transcoded to HEVC (NVENC) with HandBrake; HEVC files are copied directly.
- **Episode deduplication** -- Skips episodes that already have a file in Sonarr.
- **Colorized output** -- Shows processing progress and a summary of processed/unmatched files.

## Requirements

- Go 1.24+
- [ffprobe](https://ffmpeg.org/) (part of FFmpeg)
- [HandBrake CLI](https://handbrake.fr/) (`handbrakecli`) -- required only for H.264 transcoding
- A running Sonarr instance with API access

## Installation

```sh
go build -o micromediamanager .
```

Build with version metadata using ldflags:

```sh
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildCommit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%d)" -o micromediamanager .
```

## Configuration

Create a JSON config file:

```json
{
  "sonarrUrl": "http://localhost:8989",
  "sonarrApiKey": "your-api-key-here",
  "ignoreCertificate": false,
  "handbrakeQuality": 24
}
```

| Field | Description |
|---|---|
| `sonarrUrl` | Base URL of your Sonarr instance |
| `sonarrApiKey` | Sonarr API key (Settings > General) |
| `ignoreCertificate` | Skip TLS certificate verification (for self-signed certs) |
| `handbrakeQuality` | HandBrake constant quality (CQ) value. Lower = higher quality. Default: `24`. See below. |

### HandBrake Quality Guide

The `handbrakeQuality` value controls the constant quality (CQ/CRF) setting passed to HandBrake's `-q` flag. Lower values produce higher quality and larger files:

| Value | Quality | Use case |
|---|---|---|
| 18-20 | Visually lossless | Archival, high-bitrate sources |
| 22-24 | High quality | Good default for anime and TV (default: 24) |
| 26-28 | Medium quality | Smaller files, minor quality loss |
| 30+ | Low quality | Maximum compression, noticeable degradation |

## Usage

```sh
micromediamanager --configFile config.json --sourceFolder /path/to/media
```

| Flag | Short | Description |
|---|---|---|
| `--configFile` | `-c` | Path to the JSON config file (required) |
| `--sourceFolder` | `-s` | Path to the folder containing source media files (required) |
| `--version` | `-v` | Print version information and exit |

## How It Works

1. Reads all files from the source folder.
2. Parses each filename to extract the series title and episode number.
3. Matches the parsed title against your Sonarr library (including alternate titles).
4. Skips episodes that already have a file in Sonarr.
5. Detects the video codec with ffprobe:
   - **HEVC** -- Copies the file directly.
   - **Any other codec** -- Transcodes to HEVC using HandBrake (NVENC encoder).
6. Places the output file in the Sonarr series path: `{Series Path}/Season {N}/{Title} - S{ss}E{ee}.mkv`
7. Triggers a Sonarr rescan for each affected series.

## Supported Filename Formats

```
[SubGroup] Series Title - 03.mkv
[SubGroup] Series Title - 03v2.mkv
[SubGroup] Series Title - 03 [1080p][HEVC].mkv
[SubGroup][720p] Series Title - 12.mkv
```

The parser expects a ` - ` separator between the title and episode number.

## Testing

```sh
go test ./...
```
