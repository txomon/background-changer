package provider

import (
	"fmt"
	"time"

	"github.com/txomon/sawyer/util"
)

type LocalPhotoProvider struct {
	path     string
	interval int
}

func (lpp LocalPhotoProvider) run() {
	for {
		lpp.getPhotos()
		time.Sleep(time.Duration(lpp.interval))
	}
}
func (lpp LocalPhotoProvider) getBackendName() string {
	return fmt.Sprintf("Local(%v)", lpp.path)
}

func (lpp LocalPhotoProvider) getPhotos() ([]string, error) {
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
	return PhotoLinker{backend: LocalPhotoProvider{path: path, interval: interval}}
}

func init() {
	RegisterProvider("local", GetLocalPhotoProvider)
}
