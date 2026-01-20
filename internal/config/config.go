package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/goccy/go-yaml"
)

type Config struct {
	// LLM API keys
	AkGrok   string `yaml:"akGrok"`
	AkOpenAI string `yaml:"akOpenAI"`
	AkClaude string `yaml:"akClaude"`

	// Simulation parameters
	EventFile        string  `yaml:"eventFile"`
	NumHouseholds    int     `yaml:"numHouseholds"`
	PopPerHousehold  int     `yaml:"popPerHousehold"`
	NumStartingFirms int     `yaml:"numStartingFirms"`
	NumTicks         int     `yaml:"numTicks"`

	// Economic parameters
	InitialWage   float64 `yaml:"initialWage"`
	WageFloor     float64 `yaml:"wageFloor"`
	MaxWageChange float64 `yaml:"maxWageChange"`
	SavingsRate   float64 `yaml:"savingsRate"`
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

	// Apply defaults for simulation parameters
	if config.EventFile == "" {
		config.EventFile = "simrun/events.ndjson"
	}
	if config.NumHouseholds == 0 {
		config.NumHouseholds = 100
	}
	if config.PopPerHousehold == 0 {
		config.PopPerHousehold = 30
	}
	if config.NumStartingFirms == 0 {
		config.NumStartingFirms = 5
	}
	if config.NumTicks == 0 {
		config.NumTicks = 500
	}
	if config.InitialWage == 0 {
		config.InitialWage = 10.0
	}
	if config.WageFloor == 0 {
		config.WageFloor = 5.0
	}
	if config.MaxWageChange == 0 {
		config.MaxWageChange = 0.05
	}
	if config.SavingsRate == 0 {
		config.SavingsRate = 0.2
	}

	return &config, nil
}
