package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey string `json:"api_key"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "track-")
}

func ConfigDir() string {
	return configDir()
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found — run 'track- setup' or create %s", configPath())
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required in %s", configPath())
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0600)
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err)
		fmt.Fprintf(os.Stderr, "To get started:\n")
		fmt.Fprintf(os.Stderr, "  1. Sign up free at https://www.ship24.com\n")
		fmt.Fprintf(os.Stderr, "  2. Copy your API key from the dashboard\n")
		fmt.Fprintf(os.Stderr, "  3. Run 'track- setup' or create %s with:\n", configPath())
		fmt.Fprintf(os.Stderr, `     {
       "api_key": "YOUR_SHIP24_API_KEY"
     }
`)
		os.Exit(1)
	}
	return cfg
}
