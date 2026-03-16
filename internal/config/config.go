package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config represents the top-level YAML structure
type Config struct {
	Server ServerConfig  `yaml:"server"`
	Routes []RouteConfig `yaml:"routes"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type RouteConfig struct {
	Path     string   `yaml:"path"`
	Backends []string `yaml:"backends"` // Changed from Backend to Backends array
}

// LoadConfig reads the YAML file and unmarshals it into our Config struct
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}