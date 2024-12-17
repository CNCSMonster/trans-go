package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

type TomlConfig struct {
	BaseUrl     string  `toml:"OPENAI_API_BASE_URL"`
	ApiKey      string  `toml:"OPENAI_API_KEY"`
	Model       string  `toml:"MODEL"`
	Temperature float64 `toml:"TEMPERATURE"`
}

type Config struct {
	TomlConfig
	ConfigPath string
}

func (c Config) String() string {
	return fmt.Sprintf("BaseUrl: %s\nApiKey: %s\nModel: %s\nTemperature: %f\nConfigPath:%s", c.BaseUrl, c.ApiKey, c.Model, c.Temperature, c.ConfigPath)
}

func DefaultConfig() (config *Config) {
	// from XDG_COFNIG_HOME/trans-go/config.toml
	// or from HOME/.trans-go/config.toml
	// or with nil
	var configPath string
	if c, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		configPath = c + "/trans-go/config.toml"
	} else if c, ok := os.LookupEnv("HOME"); ok {
		configPath = c + "/.trans-go/config.toml"
	}

	if configPath == "" {
		return nil
	}
	// parse toml
	var tomlConfig TomlConfig
	if _, err := toml.DecodeFile(configPath, &tomlConfig); err != nil {
		log.Fatal(err)
	}

	config = &Config{
		TomlConfig: TomlConfig{
			BaseUrl:     tomlConfig.BaseUrl,
			ApiKey:      tomlConfig.ApiKey,
			Model:       tomlConfig.Model,
			Temperature: tomlConfig.Temperature,
		},
		ConfigPath: configPath,
	}
	return
}

func NewConfig() Config {
	var config *Config
	if c := DefaultConfig(); c != nil {
		config = c
	} else {
		config = &Config{}
	}

	if url, ok := os.LookupEnv("OPENAI_API_BASE_URL"); ok {
		config.BaseUrl = url
	}
	if ak, ok := os.LookupEnv("OPENAI_API_KEY"); ok {
		config.ApiKey = ak
	}
	if m, ok := os.LookupEnv("MODEL"); ok {
		config.Model = m
	}
	if ts, ok := os.LookupEnv("TEMPERATURE"); ok {
		if t, err := strconv.ParseFloat(ts, 64); err == nil {
			config.Temperature = t
		} else {
			panic(fmt.Sprintf("fail to parse \"%s\" as float64", ts))
		}
	}

	return *config
}
