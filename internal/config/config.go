package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	BaseURL      string `json:"base_url"`
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

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://apis.usps.com"
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client_id and client_secret are required in %s", configPath())
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir(), 0755); err != nil {
		return err
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://apis.usps.com"
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
		fmt.Fprintf(os.Stderr, "  1. Register at https://developers.usps.com\n")
		fmt.Fprintf(os.Stderr, "  2. Create an app and get your Consumer Key + Secret\n")
		fmt.Fprintf(os.Stderr, "  3. Create %s with:\n", configPath())
		fmt.Fprintf(os.Stderr, `     {
       "client_id": "YOUR_CONSUMER_KEY",
       "client_secret": "YOUR_CONSUMER_SECRET"
     }
`)
		os.Exit(1)
	}
	return cfg
}
