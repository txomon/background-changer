package provider

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/txomon/sawyer/util"
)

type LocalPhotoProvider struct {
	path     string
	interval int
}

func (lpp *LocalPhotoProvider) String() string {
	return lpp.getName()
}
func (lpp *LocalPhotoProvider) setStorageLocation(cacheDirectory string) {
}

func (lpp *LocalPhotoProvider) run(photoProvider *PhotoProvider) {
	var pp PhotoProvider = lpp
	if photoProvider == nil {
		photoProvider = &pp
	}
	logger.Debugf("Running %v with %v", lpp, photoProvider)
	for {
		if photos, err := (*photoProvider).getPhotos(); err == nil {
			logger.Debugf("Got %v photos", len(photos))
		} else {
			logger.Infof("Failed to use photos from %v. %v", lpp.path, err)
		}
		time.Sleep(time.Duration(lpp.interval))
	}
}
func (lpp *LocalPhotoProvider) getName() string {
	name := strings.Map(func(char rune) rune {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			return char
		}
		return -1
	}, lpp.path)
	return fmt.Sprintf("local-%v", name)
}

func (lpp *LocalPhotoProvider) getPhotos() ([]string, error) {
	photos := util.GetPhotosForPath(lpp.path)
	return photos, nil
}

func GetLocalPhotoProvider(config map[string]interface{}) PhotoProvider {
	path, ok := config["path"].(string)
	if !ok {
		logger.Errorf("path config parameter is not a string as expected")
		return nil
	}

	interval, ok := config["poll_interval"].(int)
	if !ok {
		interval = 10
	}
	interval *= 1000000000
	var lpp PhotoProvider = &LocalPhotoProvider{path: path, interval: interval}
	var pl PhotoProvider = &PhotoLinker{backend: &lpp}
	logger.Tracef("Creating %v[%p](backend:%v[%p])", pl, &pl, lpp, &lpp)
	return pl
}

func init() {
	RegisterProvider("local", GetLocalPhotoProvider)
}
