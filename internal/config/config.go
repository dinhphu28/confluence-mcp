// Package config loads, migrates and persists the confluence-mcp config file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CurrentConfigVersion is the schema version this binary writes and understands.
// Bump it whenever the config layout changes, and add a migration step in
// Migrate for the new version.
const CurrentConfigVersion = 1

type Config struct {
	// Version is the config schema version. Existing files written before
	// versioning was introduced have 0 and are migrated on next setup.
	Version    int              `yaml:"config_version"`
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

// Read parses the config file without validating required fields. found is false
// when the file does not exist (a non-error condition for first-time setup).
func Read() (cfg *Config, found bool, err error) {
	configPath, err := Path()
	if err != nil {
		return nil, false, err
	}

	data, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	cfg = &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, true, fmt.Errorf("cannot parse config %s: %w", configPath, err)
	}

	return cfg, true, nil
}

// Save writes cfg to the config file, creating the directory if needed.
func Save(cfg *Config) error {
	configPath, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0o600)
}

// Migrate upgrades cfg in place to CurrentConfigVersion, applying each step in
// order. It returns true if anything changed.
func Migrate(cfg *Config) bool {
	start := cfg.Version

	// v0 -> v1: introduce config_version. No field changes; older files are
	// already shape-compatible.
	if cfg.Version < 1 {
		cfg.Version = 1
	}

	// Future migrations go here, e.g.:
	// if cfg.Version < 2 { ...transform...; cfg.Version = 2 }

	return cfg.Version != start
}

// Load reads and validates the config for use by the server. It rejects configs
// from a newer schema than this binary supports.
func Load() (*Config, error) {
	cfg, found, err := Read()
	if err != nil {
		return nil, err
	}
	if !found {
		path, _ := Path()
		return nil, fmt.Errorf("config %s not found (run: confluence-mcp setup)", path)
	}

	if cfg.Version > CurrentConfigVersion {
		return nil, fmt.Errorf(
			"config version %d is newer than this binary supports (%d); please upgrade confluence-mcp",
			cfg.Version, CurrentConfigVersion)
	}

	if cfg.Confluence.URL == "" {
		return nil, fmt.Errorf("confluence.url is required")
	}
	if cfg.Confluence.PAT == "" {
		return nil, fmt.Errorf("confluence.pat is required")
	}

	return cfg, nil
}
