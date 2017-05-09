package provider

import (
	"fmt"
	"net/http"
	"time"
)

type ImgurProvider struct {
	album    string
	interval int
	client   http.Client
}

func buildImgurURL(endpoint string) string {
	return fmt.Sprintf("https://api.imgur.com%v", endpoint)
}

func imgurGet(endpoint string) (interface{}, error) {
	return nil, nil
}

func (ip *ImgurProvider) getPhotos() ([]string, error) {
	photos := make([]string, 0)
	// TODO
	return photos, nil
}

func (ip *ImgurProvider) getName() string {
	return fmt.Sprintf("imgur-%v", ip.album)
}

func (ip *ImgurProvider) setStorageLocation(location string) {
}

func (ip *ImgurProvider) run(photoProvider *PhotoProvider) {
	var pp PhotoProvider = ip

	if photoProvider == nil {
		photoProvider = &pp
	}

	for {
		if photos, err := (*photoProvider).getPhotos(); err == nil {
			logger.Debugf("Got %v photos", len(photos))
		} else {
			logger.Infof("Failed to get photos from %v. %v", ip.album, err)
		}
		time.Sleep(time.Duration(ip.interval))
	}
}

func GetImgurPhotoProvider(config map[string]interface{}) PhotoProvider {
	album, ok := config["album"].(string)
	if !ok {
		logger.Errorf("path config parameter is not a string as expected")
		return nil
	}

	interval, ok := config["poll_interval"].(int)
	if !ok {
		interval = 1000
	}
	interval *= 1000000000

	var pl PhotoProvider = &PhotoDownloader{
		backend: &ImgurProvider{album: album, interval: interval},
	}

	return pl
}
