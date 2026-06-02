package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Color string `yaml:"color"`
	Theme string `yaml:"theme"`
	Shade *bool  `yaml:"shade,omitempty"`
}

func (c Config) ShadeEnabled() bool {
	if c.Shade == nil {
		return true
	}
	return *c.Shade
}

func Default() Config {
	return Config{
		Color: "auto",
		Theme: "dracula",
	}
}

func Load() (Config, error) {
	cfg := Default()

	path, err := configFilePath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func Save(cfg Config) (string, error) {
	path, err := configFilePath()
	if err != nil {
		return "", fmt.Errorf("cannot resolve config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("cannot marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("cannot write config: %w", err)
	}
	return path, nil
}

func configFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "oc-color", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "oc-color", "config.yaml"), nil
}
