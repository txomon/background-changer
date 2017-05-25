// +build darwin

package de

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS:  -framework AppKit

#import <AppKit/AppKit.h>
#include <stdlib.h>

void change_background(char *image) {
	NSError *error;
	NSURL *imageURL;
	NSDictionary *options = @{};
	imageURL = [NSURL URLWithString:[NSString stringWithFormat:@"file://localhost%@", [NSString stringWithUTF8String:image]]];

    for (NSScreen *screen in [NSScreen screens]) {
        [[NSWorkspace sharedWorkspace] setDesktopImageURL:imageURL forScreen:screen options:@{} error:&error ];
    }
}

*/
import "C"

import (
	"runtime"
	"unsafe"
)

type MacOsXBackgroundChanger struct{}

func (lbc *MacOsXBackgroundChanger) ChangeBackground(pictureStream chan string) {
	for {
		picture := <-pictureStream
		logger.Infof("Picture would be %v", picture)
		pictureString := C.CString(picture)
		defer C.free((unsafe.Pointer)(pictureString))
		C.change_background(pictureString)
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
