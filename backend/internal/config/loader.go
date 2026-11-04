package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TaskConfig struct {
	ID      int      `yaml:"id"`
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

type AppConfig struct {
	Version   string       `yaml:"version"`
	Workflows []TaskConfig `yaml:"workflows"`
}

func LoadConfig(filename string) (*AppConfig, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
