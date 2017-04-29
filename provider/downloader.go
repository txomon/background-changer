package provider

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/txomon/sawyer/util"
)

type PhotoDownloader struct {
	backend        PhotoProvider
	cacheDirectory string
}

func (pd PhotoDownloader) run(photoProvider PhotoProvider) {
	if photoProvider == nil {
		photoProvider = pd
	}
	pd.backend.run(pd)
}
func (pd PhotoDownloader) String() string {
	return fmt.Sprint("downloader-", pd.getBackendName())
}

func (pd PhotoDownloader) getBackendName() string {
	return pd.backend.getBackendName()
}

func (pd PhotoDownloader) getPhotos() ([]string, error) {
	var photos []string
	backendPhotos, err := pd.backend.getPhotos()
	if err != nil {
		logger.Errorf("PhotoDownloader encountered an error from backend %v getPhotos", pd.backend.getBackendName())
		return nil, err
	}
	backendCacheDirectory := filepath.Join(pd.cacheDirectory, pd.backend.getBackendName())
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
		format := util.IsBackground(photoContent)
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

		photos = append(photos, photoPath)
	}
	return photos, err
}
