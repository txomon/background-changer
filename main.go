package main

import (
	"os"
	"os/exec"
	"runtime"

	"time"

	"path/filepath"
	"strings"

	"math"

	"github.com/juju/loggo"
	"github.com/spf13/viper"
)

var logger loggo.Logger = loggo.GetLogger("main")

type OSBackgroundChanger interface {
	ChangeBackground(chan string)
}

type LinuxBackgroundChanger struct{}

func (lbc LinuxBackgroundChanger) ChangeBackground(pictureStream chan string) {
	for {
		picture := <-pictureStream
		command := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", picture)
		err := command.Run()
		if err != nil {
			logger.Errorf("%v", err)
		}
	}
}

const (
	ConfigurationPicturePath    = "picture_path"
	ConfigurationChangeInterval = "change_interval"
)

func configure() {
	// Configuration origins
	viper.SetConfigType("json")
	viper.SetConfigFile("config.json")
	viper.AddConfigPath(".")

	// Variables
	viper.SetDefault(ConfigurationPicturePath, ".")
	viper.SetDefault(ConfigurationChangeInterval, time.Duration(math.Pow(10, 10))) // 10 seconds

	// Read and monitor configuration
	viper.ReadInConfig()
	viper.WatchConfig()

	logger.SetLogLevel(loggo.DEBUG)
}

func isPhoto(path string) bool {
	path = strings.ToLower(path)
	imageSuffixes := []string{"jpg", "jpeg", "png"}
	for _, suffix := range imageSuffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func pictureMonitor(pictureStream chan string) {
	var lastFile, nextFile string

	for {
		fileList := make([]string, 0)
		filepath.Walk(viper.GetString(ConfigurationPicturePath), func(path string, info os.FileInfo, err error) error {
			if !isPhoto(path) {
				return err
			}

			file, err := filepath.Abs(path)
			if err != nil {
				logger.Errorf("Path %v could not be converted to absolute path: %v", path, err)
				return err
			}
			logger.Debugf("Image found, path %v", file)
			fileList = append(fileList, file)
			return err
		})

		// Iterate over all files
		found := false
		for _, file := range fileList {
			if found || lastFile == "" {
				nextFile = file
				break
			}
			if lastFile == file {
				found = true
			}
		}
		// If no next file but there are files we are in the end of the loop, and first image should be used
		if nextFile == "" && len(fileList) > 0 {
			nextFile = fileList[0]
		}

		if _, err := os.Stat(nextFile); err == nil {
			logger.Infof("Next background %v", nextFile)
			pictureStream <- nextFile
		} else {
			logger.Infof("There is no background file available")
		}

		// We wait for Duration before changing again
		time.Sleep(viper.GetDuration(ConfigurationChangeInterval))
		lastFile, nextFile = nextFile, ""
	}
}

func main() {
	var background_changer OSBackgroundChanger
	pictureStream := make(chan string)

	configure()

	switch runtime.GOOS {
	case "linux":
		background_changer = LinuxBackgroundChanger{}
	}
	go pictureMonitor(pictureStream)
	background_changer.ChangeBackground(pictureStream)
}
