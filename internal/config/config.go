// Package config loads the confluence-mcp configuration file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Confluence ConfluenceConfig `yaml:"confluence"`
}

type ConfluenceConfig struct {
	URL      string `yaml:"url"`
	PAT      string `yaml:"pat"`
	ReadOnly bool   `yaml:"read_only"`
}

// Path returns the location of the config file (~/.config/confluence-mcp/config.yaml).
func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "confluence-mcp", "config.yaml"), nil
}

// Load reads and validates the config file.
func Load() (*Config, error) {
	configPath, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read config %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	if cfg.Confluence.URL == "" {
		return nil, fmt.Errorf("confluence.url is required")
	}

	if cfg.Confluence.PAT == "" {
		return nil, fmt.Errorf("confluence.pat is required")
	}

	return &cfg, nil
}
