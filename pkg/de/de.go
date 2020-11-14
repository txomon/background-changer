package de

import "github.com/juju/loggo"

var logger = loggo.GetLogger("sawyer.de")

type DEBackgroundChanger interface {
	Set(chan string)
	GetSupportedFormats() []string
}

func GetDEBackgroundChanger(de string) DEBackgroundChanger {
	if de != "" {
		constructor, ok := registeredDEs[de]
		if !ok {
			logger.Infof("Background changer %v not found", de)
			return nil
		}
		return constructor()
	}
	for id, constructor := range registeredDEs {
		de := constructor()
		if de != nil {
			logger.Infof("Desktop environment %v matched", id)
			return de
		}
	}
	return nil
}

var registeredDEs = make(map[string]func() DEBackgroundChanger)

func RegisterDE(de string, constructor func() DEBackgroundChanger) {
	logger.Tracef("Registering desktop environment %v", de)
	registeredDEs[de] = constructor
}
