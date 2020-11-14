package sawyer

import (
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/txomon/sawyer/pkg/util"
)

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
		cachePath := viper.GetString(util.ConfigurationCacheDir)
		_, err := os.Stat(cachePath)
		if err != nil {
			os.MkdirAll(cachePath, 0755)
		}
		nextFileList = util.GetPhotosForPath(cachePath)

		nextFile = getNextInList(lastFile, lastFileList, nextFileList)

		if _, err := os.Stat(nextFile); err == nil {
			logger.Infof("Next background %v", nextFile)
			pictureStream <- nextFile
		} else {
			logger.Infof("There is no background file available")
		}

		// We wait for Duration before changing again
		configuredDuration := viper.GetDuration(util.ConfigurationChangeInterval)
		time.Sleep(configuredDuration * 1000000000)
		lastFile, nextFile = nextFile, ""
		lastFileList, nextFileList = nextFileList, nil
	}
}
