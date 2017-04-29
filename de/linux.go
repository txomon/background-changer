package de

import (
	"os/exec"

	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("sawyer.de")

type OSBackgroundChanger interface {
	ChangeBackground(chan string)
	GetSupportedFormats() []string
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
func (lbc LinuxBackgroundChanger) GetSupportedFormats() []string {
	return []string{"jpeg", "png", "jpg"}
}
