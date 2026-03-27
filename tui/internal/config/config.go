// ABOUTME: Loads and saves user configuration from ~/.texty.yaml.
// ABOUTME: Stores username and default server URL.

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds persisted user settings.
type Config struct {
	Username  string `yaml:"username"`
	ServerURL string `yaml:"server_url,omitempty"`
}

// Path returns the config file location.
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".texty.yaml")
}

// Load reads the config file. Returns zero Config if it doesn't exist.
func Load() Config {
	data, err := os.ReadFile(Path())
	if err != nil {
		return Config{}
	}
	var cfg Config
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

// Save writes the config file.
func Save(cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0644)
}
