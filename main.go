package main

import (
	"log"
	"os/exec"
	"runtime"
)

type OSBackgroundChanger interface {
	ChangeBackground(string)
}

type LinuxBackgroundChanger struct{}

func (lbc LinuxBackgroundChanger) ChangeBackground(picture string) {
	command := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", picture)
	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var background_changer OSBackgroundChanger
	switch runtime.GOOS {
	case "linux":
		background_changer = LinuxBackgroundChanger{}
	}
	background_changer.ChangeBackground("/home/javier/Documents/profile.jpg")
}
