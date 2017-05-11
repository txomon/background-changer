package sawyer

import (
	"runtime"

	"time"

	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/txomon/sawyer/de"
	"github.com/txomon/sawyer/provider"
	"github.com/txomon/sawyer/util"
)

var logger = loggo.GetLogger("sawyer")

func configure() error {
	logger.SetLogLevel(loggo.INFO)

	// Configuration origins
	viper.SetConfigType("json")
	viper.SetConfigFile("config.json")

	// Variables
	viper.SetDefault(util.ConfigurationChangeInterval, 10)
	viper.SetDefault(util.ConfigurationCacheDir, "cache")
	viper.SetDefault(util.ConfigurationProviders, make([]interface{}, 0))

	// Load config
	err := viper.ReadInConfig()
	if err != nil {
		logger.Infof("Configuration file read failed. %v", err)
		return err
	}

	//loggo.GetLogger("sawyer.util").SetLogLevel(loggo.INFO)
	return nil
}

func DaemonMain() {
	pictureStream := make(chan string)

	switch runtime.GOOS {
	case "linux":
		viper.AddConfigPath(".")
		viper.AddConfigPath("~/.cache/sawyer")
		viper.AddConfigPath("~/.config/sawyer")
		viper.AddConfigPath("/etc/sawyer")
	case "darwin":
		viper.AddConfigPath(".")
		viper.AddConfigPath("~/Library/Caches/sawyer")
		viper.AddConfigPath("~/Library/Preferences/sawyer")
	default:
		logger.Errorf("OS %v is not supported by background changer", runtime.GOOS)
		return
	}
	obc := de.GetDEBackgroundChanger("")

	err := configure()
	for err != nil {
		logger.Infof("Retrying in 10 seconds")
		time.Sleep(time.Duration(10000000000))
		err = configure()
	}

	for _, supportedFormat := range obc.GetSupportedFormats() {
		util.RegisterSupportedFormat(supportedFormat)
	}
	providerConfigs := viper.Get(util.ConfigurationProviders).([]interface{})
	logger.Infof("Read config for providers %v", providerConfigs)
	provider.RunProviders(viper.GetString(util.ConfigurationCacheDir), providerConfigs)
	go pictureMonitor(pictureStream)
	obc.ChangeBackground(pictureStream)
}
