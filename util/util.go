package util

import (
	"bytes"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("sawyer.util")

var supportedFormats []string

func IsBackground(backgroundDescriptors ...interface{}) string {
	var doRetries = true
	for index, backgroundDescriptor := range backgroundDescriptors {
		logger.Tracef("%v# try to find out if background using %v", index, backgroundDescriptor)
		retry := true
		for retry {
			retry = false
			switch bdt := backgroundDescriptor.(type) {
			case string: // Try by file extension, and else, open file and hand over to format decoder
				bd := backgroundDescriptor.(string)
				format := strings.ToLower(bdt)
				for _, supportedFormat := range supportedFormats {
					if strings.HasSuffix(format, supportedFormat) {
						logger.Tracef("File %v ends in %v so it's supported format", format, supportedFormat)
						return format
					}
				}
				if info, err := os.Stat(bd); err != nil {
					logger.Tracef("Failed to stat in %v so skipping failthrough. %v", bd, err)
					continue
				} else if info.IsDir() {
					logger.Tracef("File is a directory, skipping %v", bd)
					return ""
				}
				if doRetries {
					var err error
					fd, err := os.Open(bd)
					if err != nil {
						logger.Errorf("Failed to open file %v, %v", bd, err)
						continue
					}
					defer fd.Close()
					backgroundDescriptor = fd
					logger.Tracef("File not recognished by extension, handing over to decoders")
					retry = true
				} else {
					logger.Tracef("File not recognished by extension")
				}
			case io.Reader: // Try to decode the image (slow...)
				bd := backgroundDescriptor.(io.Reader)
				_, format, err := image.Decode(bd)
				if err != nil {
					logger.Tracef("Format (%v) not recognished. %v", format, err)
					return ""
				}
				for _, supportedFormat := range supportedFormats {
					if supportedFormat == format {
						logger.Tracef("Format %v supported", format)
						return format
					}
				}
				logger.Tracef("Format %v is not supported", format)
				return ""
			case []byte: // Convert byte slice to reader for continuing with decoding (slow...)
				bd := backgroundDescriptor.([]byte)
				backgroundDescriptor = bytes.NewReader(bd)
				retry = true
			case bool: // Disable handover between filename and reader
				bd := backgroundDescriptor.(bool)
				doRetries = bd
			default:
				logger.Warningf("No idea how to use %T to determine if valid background", bdt)
			}
		}
	}
	logger.Tracef("File not a background by default. %v", backgroundDescriptors)
	return ""
}

func RegisterSupportedFormat(format string) {
	supportedFormats = append(supportedFormats, format)
}

func GetPhotosForPath(path string) []string {
	fileList := make([]string, 0)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			logger.Debugf("Photos can only be files, skipping dir %v", path)
			return err
		}
		file, err := filepath.Abs(path)
		if err != nil {
			logger.Errorf("Path %v could not be converted to absolute path: %v", path, err)
			return err
		}
		fileReader, err := os.Open(file)
		if err != nil {
			logger.Warningf("Couldn't open %v for supported format filtering. %v", path, err)
			return err
		}
		defer fileReader.Close()

		if "" == IsBackground(file) {
			logger.Debugf("File %v not a background", file)
			return err
		}
		logger.Debugf("Image found, path %v", file)
		fileList = append(fileList, file)
		return err
	})
	return fileList
}
