package main

import (
	"bytes"
	"image"
	"io"
	"os"
	"os/exec"
	"runtime"

	"time"

	_ "image/jpeg"
	_ "image/png"
	"path/filepath"

	"math"

	"net/http"

	"strings"

	"crypto/sha1"

	"fmt"

	"github.com/juju/loggo"
	"github.com/spf13/viper"
)

var logger loggo.Logger = loggo.GetLogger("main")

type OSBackgroundChanger interface {
	ChangeBackground(chan string)
	getSupportedFormats() []string
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
func (lbc LinuxBackgroundChanger) getSupportedFormats() []string {
	return []string{"jpeg", "png", "jpg"}
}

const (
	ConfigurationPicturePath    = "picture_path"
	ConfigurationChangeInterval = "change_interval"
	ConfigurationCacheDir       = "cache_dir"
)

func configure() {
	// Configuration origins
	viper.SetConfigType("json")
	viper.SetConfigFile("config.json")
	viper.AddConfigPath(".")

	// Variables
	viper.SetDefault(ConfigurationPicturePath, ".")
	viper.SetDefault(ConfigurationChangeInterval, time.Duration(math.Pow(10, 10))) // 10 seconds
	viper.SetDefault(ConfigurationCacheDir, "./cache")

	// Read and monitor configuration
	viper.ReadInConfig()
	viper.WatchConfig()

	logger.SetLogLevel(loggo.DEBUG)
}

func isBackground(backgroundDescriptors ...interface{}) string {
	for index, backgroundDescriptor := range backgroundDescriptors {
		logger.Tracef("%v# try to find out if background using %v", index, backgroundDescriptor)
		retry := true
		for retry {
			retry = false
			switch bdt := backgroundDescriptor.(type) {
			case string:
				bd := backgroundDescriptor.(string)
				format := strings.ToLower(bdt)
				for _, supportedFormat := range obc.getSupportedFormats() {
					if strings.HasSuffix(format, supportedFormat) {
						logger.Tracef("File %v ends in %v so it's supported format", bd, format)
						return format
					}
				}
				info, err := os.Stat(bd)
				if err != nil {
					logger.Tracef("Failed to stat in %v so skipping failthrough", bd)
					continue
				}
				if info.IsDir() {
					logger.Tracef("File is a directory, skipping %v", bd)
					return ""
				}
				backgroundDescriptor, err = os.Open(bd)
				if err != nil {
					continue
				}
				logger.Tracef("File not recognished by extension, handing over to decoders")
				retry = true
			case io.Reader:
				bd := backgroundDescriptor.(io.Reader)
				_, format, err := image.Decode(bd)
				if err != nil {
					logger.Tracef("Format (%v) not recognished. %v", format, err)
					return ""
				}
				for _, supportedFormat := range obc.getSupportedFormats() {
					if supportedFormat == format {
						logger.Tracef("Format %v supported", format)
						return format
					}
				}
				logger.Tracef("Format %v is not supported", format)
				return ""
			case []byte:
				bd := backgroundDescriptor.([]byte)
				backgroundDescriptor = bytes.NewReader(bd)
				retry = true
			default:
				logger.Warningf("No idea how to use %T to determine if valid background", bdt)
			}
		}
	}
	logger.Tracef("File not a background by default. %v", backgroundDescriptors)
	return ""
}

func getPhotosForPath(path string) []string {
	fileList := make([]string, 0)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			logger.Debugf("Photos can only be files, skipping dir %v", path)
			return err
		}
		file, err := filepath.Abs(path)
		if err != nil {
			logger.Errorf("Path %v could not be converted to absolute path: %v", path, err)
			return err
		}
		fileReader, err := os.Open(file)
		if err != nil {
			logger.Warningf("Couldn't open %v for supported format filtering. %v", path, err)
			return err
		}
		if "" == isBackground(file) {
			logger.Debugf("File %v not a background", file)
			return err
		}
		fileReader.Close()
		logger.Debugf("Image found, path %v", file)
		fileList = append(fileList, file)
		return err
	})
	return fileList
}

func getNextInList(lastItem string, previousList []string, nextList []string) string {
	found := false
	for _, nextItem := range previousList {
		if lastItem == nextItem {
			found = true
			continue
		}
		if found {
			for _, newNextItem := range nextList {
				if newNextItem == nextItem {
					return newNextItem
				}
			}
		}
	}

	if len(nextList) > 0 {
		return nextList[0]
	}

	return ""
}

func pictureMonitor(pictureStream chan string) {
	var lastFile, nextFile string
	var lastFileList, nextFileList []string

	for {
		nextFileList = getPhotosForPath(viper.GetString(ConfigurationPicturePath))

		nextFile = getNextInList(lastFile, lastFileList, nextFileList)

		if _, err := os.Stat(nextFile); err == nil {
			logger.Infof("Next background %v", nextFile)
			pictureStream <- nextFile
		} else {
			logger.Infof("There is no background file available")
		}

		// We wait for Duration before changing again
		time.Sleep(viper.GetDuration(ConfigurationChangeInterval))
		lastFile, nextFile = nextFile, ""
		lastFileList, nextFileList = nextFileList, nil
	}
}

type PhotoBackend interface {
	getPhotos() ([]string, error)
	getBackendName() string
}

type PhotoDownloader struct {
	backend        PhotoBackend
	cacheDirectory string
}

func (photoDownloader PhotoDownloader) getBackendName() string {
	return fmt.Sprintf("%vDownloaded", photoDownloader.backend.getBackendName())
}

func (photoDownloader PhotoDownloader) getPhotos() ([]string, error) {
	var photos []string
	backendPhotos, err := photoDownloader.backend.getPhotos()
	if err != nil {
		logger.Errorf("PhotoDownloader encountered an error from backend %v getPhotos", photoDownloader.backend.getBackendName())
		return nil, err
	}
	backendCacheDirectory := filepath.Join(photoDownloader.cacheDirectory, photoDownloader.backend.getBackendName())
	err = os.MkdirAll(backendCacheDirectory, 0755)
	if err != nil {
		logger.Errorf("Failed creating %v cache dir. %v", backendCacheDirectory, err)
		return nil, err
	}
	for _, photo := range backendPhotos {
		res, err := http.Get(photo)
		if err != nil {
			logger.Warningf("Failed to GET photo %v. %v", photo, err)
		}
		var photoContent []byte
		_, err = res.Body.Read(photoContent)
		if err != nil {
			logger.Infof("Failed to read all the body from %v", photo)
		}
		format := isBackground(photoContent)
		if "" == format {
			logger.Infof("Photo %v format is not supported by backend", photo)
			continue
		}

		photoName := sha1.Sum(photoContent)
		photoExtension := format
		photoFilePath := fmt.Sprintf("%v.%v", photoName, photoExtension)
		photoPath := filepath.Join(backendCacheDirectory, photoFilePath)
		stat, err := os.Stat(photoPath)
		if err == nil {
			if stat.IsDir() {
				logger.Warningf("For some reason %v is a directory...", stat)
			}
			continue
		}

		photoFile, err := os.Create(photoPath)
		if err != nil {
			logger.Warningf("Creating file %v failed. %v", photoPath, err)
			continue
		}

		size, err := photoFile.Write(photoContent)
		if err != nil {
			logger.Warningf("Writing in file %v failed. %v", photoPath, err)
			continue
		}
		logger.Debugf("Written %v bytes to %v", size, photoPath)

		err = photoFile.Close()
		if err != nil {
			logger.Warningf("Closing the file failed. %v", err)
			continue
		}

		photos = append(photos, photoPath)
	}
	return photos, err
}

var obc OSBackgroundChanger

func main() {
	pictureStream := make(chan string)

	configure()

	switch runtime.GOOS {
	case "linux":
		obc = LinuxBackgroundChanger{}
	default:
		logger.Errorf("OS %v is not supported by background changer", runtime.GOOS)
		return
	}
	go pictureMonitor(pictureStream)
	obc.ChangeBackground(pictureStream)
}
