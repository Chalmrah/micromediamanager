# MicroMediaManager

## Overview
MicroMediaManager is a command-line tool designed to help organize and manage media files efficiently. It allows users to specify a configuration file and source media folder for structured media organization.

## Usage
Run the program with the following flags:

```
micromediamanager [flags]
```

### Flags
- `-c, --configFile` *(string)*: Location of the config JSON file.
- `-s, --sourceFolder` *(string)*: Location of the source media folder.
- `-v, --version` *(bool)*: Displays version information.

## Example
```
micromediamanager -c config.json -s /path/to/media
```

## Configuration File Format
The configuration file is a JSON array where each object represents a media file mapping.

### Example `config.json`
```json
[
  {
    "FileName": "TV show file name",
    "MappingFolder": "M:/Media/TV/TV Show",
    "Season": 2
  },
  {
    "FileName": "TV Show File name ",
    "MappingFolder": "M:/Media/TV/My custom tv show folder name"
  }
]
```

## Notes
- The `FileName` field should match the media file name.
- The `MappingFolder` field specifies where the media should be moved or categorized.
- The optional `Season` field can be used to indicate season numbers.

## License
This project is open-source and available for use under an appropriate license.

