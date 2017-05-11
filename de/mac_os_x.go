package de

import (
	"os/exec"

	"fmt"
	"runtime"
)

type MacOsXBackgroundChanger struct{}

func (lbc *MacOsXBackgroundChanger) ChangeBackground(pictureStream chan string) {
	for {
		picture := <-pictureStream
		query := fmt.Sprintf("update data set value = '%v';", picture)
		command := exec.Command("sqlite3", "~/Library/Application Support/Dock/desktoppicture.db", query)
		err := command.Run()
		if err != nil {
			logger.Errorf("%v", err)
		}
		command = exec.Command("killall", "Dock")
		err = command.Run()
		if err != nil {
			logger.Errorf("%v", err)
		}
	}
}
func (lbc *MacOsXBackgroundChanger) GetSupportedFormats() []string {
	return []string{"jpeg", "png", "jpg"}
}

func MacOsXDetect() DEBackgroundChanger {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return &MacOsXBackgroundChanger{}
}

func init() {
	RegisterDE("mac-os-x", MacOsXDetect)
}
