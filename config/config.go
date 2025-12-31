package config

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigPath = "config.yaml"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type R2Config struct {
	Enabled         bool   `yaml:"enabled"`
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	AccessKeyID     string `yaml:"access-key-id"`
	SecretAccessKey string `yaml:"secret-access-key"`
}

type TwitcastingConfig struct {
	Cookie string `yaml:"cookie"`
}

type Config struct {
	Streamers []*struct {
		ScreenId     string  `yaml:"screen-id" validate:"required"`
		Schedule     string  `yaml:"schedule" validate:"required"`
		EncodeOption *string `yaml:"encode-option"`
	} `yaml:"streamers" validate:"dive"`
	R2          *R2Config          `yaml:"r2"`
	Twitcasting *TwitcastingConfig `yaml:"twitcasting"`
}

func GetDefaultConfig() *Config {
	config, err := parseConfig(defaultConfigPath)
	if err != nil {
		log.Fatal("Error parsing config file: \n", err)
	}
	return config
}

func parseConfig(configPath string) (*Config, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("Paniced parsing user config: ", r)
		}
	}()

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, err
	}

	return config, validate.Struct(config)
}
