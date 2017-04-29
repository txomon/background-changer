package sawyer

import (
	"math"
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

const (
	ConfigurationChangeInterval = "change_interval"
	ConfigurationCacheDir       = "cache_dir"
	ConfigurationProviders      = "providers"
)

func configure() {
	// Configuration origins
	viper.SetConfigType("json")
	viper.SetConfigFile("config.json")
	viper.AddConfigPath(".")

	// Variables
	viper.SetDefault(ConfigurationChangeInterval, time.Duration(math.Pow(10, 10))) // 10 seconds
	viper.SetDefault(ConfigurationCacheDir, "./cache")
	viper.SetDefault(ConfigurationProviders, make([]map[string]interface{}, 0))

	// Load config
	viper.ReadInConfig()

	logger.SetLogLevel(loggo.TRACE)
}

func DaemonMain() {
	pictureStream := make(chan string)

	configure()

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
	providerConfigs := viper.Get(ConfigurationProviders)
	logger.Infof("Read config for providers %v", providerConfigs)
	provider.RunProviders(providerConfigs.([]map[string]interface{}))
	go pictureMonitor(pictureStream)
	obc.ChangeBackground(pictureStream)
}
