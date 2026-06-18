package main

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
	URL string `yaml:"url"`
	PAT string `yaml:"pat"`
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "confluence-mcp", "config.yaml")

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
