package provider

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"io/ioutil"

	"github.com/txomon/sawyer/util"
)

type PhotoDownloader struct {
	backend        PhotoProvider
	cacheDirectory string
	memory         MapMemory
}

func (pd *PhotoDownloader) run(photoProvider *PhotoProvider) {
	var pp PhotoProvider = pd
	if photoProvider == nil {
		photoProvider = &pp
	}
	pd.backend.run(photoProvider)
}
func (pd *PhotoDownloader) setStorageLocation(cacheDirectory string) {
	pd.cacheDirectory = cacheDirectory
	pd.memory = NewMemory(cacheDirectory)
}
func (pd *PhotoDownloader) String() string {
	return fmt.Sprint("downloader-", pd.getName())
}

func (pd *PhotoDownloader) getName() string {
	return pd.backend.getName()
}

func (pd *PhotoDownloader) getPhotos() ([]string, error) {
	var photos []string
	backendPhotos, err := pd.backend.getPhotos()
	if err != nil {
		logger.Errorf("PhotoDownloader encountered an error from backend %v getPhotos", pd.backend.getName())
		return nil, err
	}
	for _, photo := range backendPhotos {
		if cachedFile := pd.memory.getMemory(photo); cachedFile != "" {
			if _, err := os.Stat(cachedFile); err == nil {
				photos = append(photos, cachedFile)
				logger.Tracef("Cached file, nothing needs to be done")
				continue
			} else {
				logger.Tracef("Cached file deleted, continuing as if not cached")
			}
		}
		res, err := http.Get(photo)
		if err != nil {
			logger.Warningf("Failed to GET photo %v. %v", photo, err)
		}
		logger.Tracef("Got photo %v %v", photo, res.Status)
		photoContent, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Infof("Failed to read all the body from %v. %v", photo, err)
			continue
		}
		format := util.IsBackground(photo, photoContent)
		if "" == format {
			logger.Infof("Photo %v format is not supported by backend", photo)
			continue
		}
		sum := sha1.Sum(photoContent)
		photoName := hex.EncodeToString(sum[:])
		photoExtension := format
		photoFilePath := fmt.Sprintf("%v.%v", photoName, photoExtension)
		photoPath := filepath.Join(pd.cacheDirectory, photoFilePath)

		stat, err := os.Stat(photoPath)
		if err == nil {
			if stat.IsDir() {
				logger.Warningf("For some reason %v is a directory...", stat)
				continue
			}
		} else {
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
		}

		pd.memory.setMemory(photo, photoPath)
		photos = append(photos, photoPath)
	}
	return photos, err
}
