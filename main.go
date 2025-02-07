package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"log"
	"mkvcompressor/handbrake"
	"path/filepath"
	"strings"
)

var (
	configFile   string
	sourceFolder string
	version      bool
)

func main() {

	pflag.StringVarP(&configFile, "configFile", "c", "", "Location of config json")
	pflag.StringVarP(&sourceFolder, "sourceFolder", "s", "", "Location of source media folder")
	pflag.BoolVarP(&version, "version", "v", false, "Print version and exit")
	pflag.Parse()

	// read json file
	// loop over shows
	//   read source directory
	// 	 evaluate whether there are matching files for each series
	// 	 evaluate each file for encoding to see whether to move or encode
	//   get destination file path/name
	//   encode if not hevc. Copy if hevc.

	showList, err := ReadConfig(configFile)
	if err != nil {
		panic(err)
	}

	fileList := readSourceFolderFiles(sourceFolder)

	for _, v := range showList {
		showList := filterFileList(fileList, v.FileName+"*")

		if len(showList) == 0 {
			//log.Printf("No matching files for %v found", v.FileName)
			continue
		}

		for _, file := range showList {
			encoding, err := getVideoCodec(filepath.Join(sourceFolder, file.Name()))
			if err != nil {
				log.Fatalf("Get video codec %s error:%v", file.Name(), err)
			}

			episodePath, err := getDestinationFileName(
				v.MappingFolder,
				v.Season,
				strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			)
			if err != nil {
				log.Fatalf("Unable to get destination file name:%v", file.Name(), err)
			}

			switch encoding {
			case "h264":
				handbrake.Reencode()
			case "hevc":
				err := copyFile(filepath.Join(sourceFolder, file.Name()), episodePath)
				if err != nil {
					return
				}
			}

			fmt.Printf("%v %v\n", encoding, file.Name())
		}

	}
}
