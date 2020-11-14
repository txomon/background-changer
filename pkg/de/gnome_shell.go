package de

import (
	"os/exec"

	"runtime"
)

type GnomeShellBackgroundChanger struct{}

func (lbc *GnomeShellBackgroundChanger) Set(pictureStream chan string) {
	for {
		picture := <-pictureStream
		command := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", picture)
		err := command.Run()
		if err != nil {
			logger.Errorf("%v", err)
		}
	}
}
func (lbc *GnomeShellBackgroundChanger) GetSupportedFormats() []string {
	return []string{"jpeg", "png", "jpg"}
}

func GnomeShellDetect() DEBackgroundChanger {
	if runtime.GOOS != "linux" {
		return nil
	}
	return &GnomeShellBackgroundChanger{}
}

func init() {
	RegisterDE("gnome-shell", GnomeShellDetect)
}
