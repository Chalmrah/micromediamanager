package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"log"
	"mkvcompressor/handbrake"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	configFile   string
	sourceFolder string
	version      bool
	goVersion    = runtime.Version()
	buildDate    = "unknown"
	buildCommit  = "dev"
	buildVersion = "unknown"
	changedFiles = false
)

func main() {

	pflag.StringVarP(&configFile, "configFile", "c", "", "Location of config json")
	pflag.StringVarP(&sourceFolder, "sourceFolder", "s", "", "Location of source media folder")
	pflag.BoolVarP(&version, "version", "v", false, "Displays version information")
	pflag.Parse()

	if version {
		fmt.Printf("MicroMediaManager  %s\n- commit hssh: %s\n- build date: %s\n- go version: %s\n", buildVersion, buildCommit, buildDate, goVersion)
		os.Exit(0)
	}

	green := color.New(color.FgHiGreen).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
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
		filteredFileList := filterFileList(fileList, v.FileName+"*")

		if len(showList) == 0 {
			//log.Printf("No matching files for %v found", v.FileName)
			continue
		}

		for _, file := range filteredFileList {
			fmt.Printf("%s", green(file.Name()))

			encoding, err := getVideoCodec(filepath.Join(sourceFolder, file.Name()))
			if err != nil {
				log.Fatalf("Get video codec %s error:%v", file.Name(), err)
			}

			episodeName, err := getDestinationFileName(
				v.MappingFolder,
				v.Season,
				filepath.Ext(file.Name()),
			)
			if err != nil {
				log.Fatalf("Unable to get destination file name:%v ", file.Name())
			}
			destinationPath := filepath.Join(v.MappingFolder, "Season "+strconv.Itoa(v.Season), episodeName)

			folderName := filepath.Dir(destinationPath)

			if _, err := os.Stat(folderName); os.IsNotExist(err) {
				err := os.MkdirAll(folderName, os.ModePerm)
				if err != nil {
					log.Printf("Unable to create folder:%v ", folderName)
				}
			}

			fmt.Printf(" %s %s\n", red("-->"), filepath.Base(episodeName))
			changedFiles = true

			switch encoding {
			case "h264":
				_, err := handbrake.Run(filepath.Join(sourceFolder, file.Name()), destinationPath, 24)
				if err != nil {
					return
				}
			case "hevc":
				err := copyFile(filepath.Join(sourceFolder, file.Name()), episodeName)
				if err != nil {
					log.Printf("Unable to copy file to %s, error:%v", file.Name(), err)
					return
				}
			}
		}
	}
	if !changedFiles {
		fmt.Printf("No files detected")
	}
}
