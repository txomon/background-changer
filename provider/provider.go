package provider

import (
	"github.com/juju/loggo"
	"github.com/txomon/sawyer/util"
)

var logger = loggo.GetLogger("sawyer.provider")

type PhotoProvider interface {
	getPhotos() ([]string, error)
	getName() string
	run(*PhotoProvider)
	setStorageLocation(string)
}

var registeredProviders = make(map[string]func(map[string]interface{}) PhotoProvider)

func RegisterProvider(providerType string, constructor func(map[string]interface{}) PhotoProvider) {
	logger.Debugf("Registering provider %v", constructor)
	registeredProviders[providerType] = constructor
}

func GetProvider(config map[string]interface{}) PhotoProvider {
	providerType, ok := config["type"]
	if !ok {
		logger.Errorf("Configuration doesn't have type. %v", config)
		return nil
	}

	constructor, ok := registeredProviders[providerType.(string)]
	if !ok {
		logger.Errorf("Provider '%v' doesn't exist", providerType)
		return nil
	}
	provider := constructor(config)
	logger.Tracef("Using provider %v for %v", provider, config)
	return provider
}

func RunProviders(cacheDirectory string, configs []interface{}) {
	logger.Debugf("Providers registered %v", registeredProviders)
	logger.Debugf("Config %v", configs)
	for _, interfaceConfig := range configs {
		config := interfaceConfig.(map[string]interface{})
		provider := GetProvider(config)
		if provider == nil {
			logger.Warningf("No provider found for %v", config)
		} else {
			storageDir := util.CreateStorageDir(cacheDirectory, provider.getName())
			logger.Debugf("Setting %v(%p) storage dir %v", provider, &provider, storageDir)
			provider.setStorageLocation(storageDir)
			go provider.run(nil)
		}
	}
}
