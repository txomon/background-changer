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

var obc de.OSBackgroundChanger

func configure() error {
	logger.SetLogLevel(loggo.INFO)

	// Configuration origins
	viper.SetConfigType("json")
	viper.SetConfigFile("config.json")
	viper.AddConfigPath("/etc/sawyer")
	viper.AddConfigPath("~/.cache/sawyer")
	viper.AddConfigPath(".")

	// Variables
	viper.SetDefault(util.ConfigurationChangeInterval, 10)
	viper.SetDefault(util.ConfigurationCacheDir, "./cache")
	viper.SetDefault(util.ConfigurationProviders, make([]interface{}, 0))

	// Load config
	err := viper.ReadInConfig()
	if err != nil {
		logger.Infof("Configuration file not found, not retrying")
		return err
	}

	//loggo.GetLogger("sawyer.util").SetLogLevel(loggo.INFO)
	return nil
}

func DaemonMain() {
	pictureStream := make(chan string)

	err := configure()
	for err != nil {
		time.Sleep(time.Duration(10000000000))
		err = configure()
	}

	switch runtime.GOOS {
	case "linux":
		obc = de.LinuxBackgroundChanger{}
	default:
		logger.Errorf("OS %v is not supported by background changer", runtime.GOOS)
		return
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
