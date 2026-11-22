package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type HueConfig struct {
	IpBridge string `yaml:"ipBridge"`
	ApiKey   string `yaml:"apiKey"`
	Room     string `yaml:"room"`
}

type HrmConfig struct {
	HrmDeviceId string `yaml:"hrmDeviceId"`
	MaxHr       int    `yaml:"maxHr"`
}

type AppConfig struct {
	HueConfig HueConfig `yaml:"hueConfig"`
	HrmConfig HrmConfig `yaml:"hrmConfig"`
}

func ReadConfig() AppConfig {
	return ReadConfigPath("config/config.yaml")
}

func ReadConfigPath(path string) AppConfig {
	appConfig := AppConfig{}
	configFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configFile, &appConfig)
	if err != nil {
		log.Fatal(err)
	}
	return appConfig
}
