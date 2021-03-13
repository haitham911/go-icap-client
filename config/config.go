package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig represents the app configuration
type AppConfig struct {
	Scheme            string
	Host              string
	Port              int
	ICAP_Resource     string
	CheckFile         bool
	Timeout           int
	ProcessExtensions []string
}

var appCfg AppConfig

// Init initializes the configuration
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	appCfg = AppConfig{
		Scheme:            viper.GetString("app.scheme"),
		Host:              viper.GetString("app.host"),
		Port:              viper.GetInt("app.port"),
		ICAP_Resource:     viper.GetString("app.icap_resource"),
		CheckFile:         viper.GetBool("app.checkfile"),
		Timeout:           viper.GetInt("app.timeout"),
		ProcessExtensions: viper.GetStringSlice("app.process_extensions"),
	}

}

// InitTestConfig initializes the app with the test config file (for integration test)
func InitTestConfig() {
	viper.SetConfigName("config.test")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	appCfg = AppConfig{
		Port: viper.GetInt("app.port"),
	}
}

// App returns the the app configuration instance
func App() *AppConfig {
	return &appCfg
}
