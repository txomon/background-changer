package provider

import (
	"path/filepath"

	"os"

	"fmt"

	"crypto/sha1"

	"github.com/txomon/sawyer/util"
)

type PhotoLinker struct {
	backend  PhotoProvider
	cacheDir string
}

func (pl PhotoLinker) run() {
	pl.backend.run()
}

func (pl PhotoLinker) getBackendName() string {
	return fmt.Sprintf("%vLinker", pl.backend.getBackendName())
}

func (pl PhotoLinker) getPhotos() ([]string, error) {
	var photos = make([]string, 0)

	backendPhotos, err := pl.backend.getPhotos()
	photoDir := filepath.Join(pl.cacheDir, pl.backend.getBackendName())
	if _, err = os.Stat(photoDir); err != nil {
		logger.Debugf("Cache dir for %v backend doesn't exist", pl.getBackendName())
		if err = os.MkdirAll(photoDir, 0755); err != nil {
			logger.Errorf("Failed to create cache dir for %v", pl.getBackendName())
			return nil, err
		}
	}

	for _, backendPhotoPath := range backendPhotos {
		if info, err := os.Stat(backendPhotoPath); err != nil {
			logger.Infof("Could not stat file %v. %v", backendPhotoPath, err)
			continue
		} else if info.IsDir() {
			logger.Debugf("File %v is directory, skipping", backendPhotoPath)
		}

		backendPhotoFile, err := os.Open(backendPhotoPath)
		if err != nil {
			logger.Infof("Failed to open file %v", backendPhotoPath)
			continue
		}
		defer backendPhotoFile.Close()

		var backendPhotoContent []byte
		if _, err = backendPhotoFile.Read(backendPhotoContent); err != nil {
			logger.Infof("Failed to read contents from file %v", backendPhotoPath)
			continue
		}

		sum := sha1.Sum(backendPhotoContent)
		format := util.IsBackground(backendPhotoContent)
		backendPhotoFileName := fmt.Sprintf("%v.%v", sum, format)

		photoPath := filepath.Join(photoDir, backendPhotoFileName)

		if info, err := os.Stat(photoPath); err == nil {
			if info.IsDir() {
				logger.Warningf("The to-be-linked exists and is a directory! %v", photoPath)
			}
			continue
		}

		if err := os.Link(backendPhotoPath, photoPath); err != nil {
			logger.Debugf("Failed to link file, copying it. %v", err)
			if photoFile, err := os.Create(photoPath); err != nil {
				logger.Warningf("Failed to create file in cache dir %v", photoPath)
				continue

			} else {
				go func() {
					photoFile.Write(backendPhotoContent)
					photoFile.Close()
				}()
			}
		} else {
			logger.Tracef("Linked file %v to %v", backendPhotoPath, photoPath)
		}
		photos = append(photos, photoPath)
	}
	return photos, nil
}
