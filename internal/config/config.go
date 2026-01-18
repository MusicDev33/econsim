package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/goccy/go-yaml"
)

type Config struct {
	AkGrok   string `yaml:"akGrok"`
	AkOpenAI string `yaml:"akOpenAI"`
	AkClaude string `yaml:"akClaude"`
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		var err error
		instance, err = loadConfig("config.yaml")
		if err != nil {
			panic(fmt.Sprintf("Failed to load config: %v", err))
		}
	})
	return instance
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.AkGrok == "" && config.AkOpenAI == "" && config.AkClaude == "" {
		return nil, fmt.Errorf("at least one API key must be set in config (akGrok, akOpenAI, or akClaude)")
	}

	return &config, nil
}
