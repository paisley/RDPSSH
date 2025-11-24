package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	RemoteHost            string `json:"remote_host"`
	RemoteUser            string `json:"remote_user"`
	LocalPort             string `json:"local_port"`
	P12Path               string `json:"p12_path"`
	MinimizeToTrayWarning bool   `json:"minimize_to_tray_warning"`
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, strings.ToLower(AppName))
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return &Config{LocalPort: "33890", MinimizeToTrayWarning: true}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{LocalPort: "33890", MinimizeToTrayWarning: true}, nil
	}

	var cfg Config
	cfg.MinimizeToTrayWarning = true
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{LocalPort: "33890", MinimizeToTrayWarning: true}, nil
	}

	if cfg.LocalPort == "" {
		cfg.LocalPort = "33890"
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
