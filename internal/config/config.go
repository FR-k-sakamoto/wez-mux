package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Workspace string       `yaml:"workspace"`
	Layout    LayoutConfig `yaml:"layout"`
}

type LayoutConfig struct {
	Rows     []RowConfig `yaml:"rows"`
	RowRatio []int       `yaml:"row_ratio"`
}

type RowConfig struct {
	Panes []PaneConfig `yaml:"panes"`
}

type PaneConfig struct {
	Label   string `yaml:"label"`
	Model   string `yaml:"model"`
	Skill   string `yaml:"skill"`
	Percent int    `yaml:"percent"`
	Codex   bool   `yaml:"codex,omitempty"`
}

// findDefaultConfig looks for default.yaml in standard locations.
func findDefaultConfig() string {
	// 1. ~/.config/wez-mux/default.yaml
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".config", "wez-mux", "default.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// 2. configs/default.yaml relative to executable
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), "configs", "default.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// 3. configs/default.yaml relative to cwd
	if cwd, err := os.Getwd(); err == nil {
		path := filepath.Join(cwd, "configs", "default.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func Load(path string) (Config, error) {
	if path == "" {
		path = findDefaultConfig()
	}

	if path == "" {
		return Config{}, fmt.Errorf("no config file found. Use --config to specify one")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Workspace == "" {
		cfg.Workspace = "wez-mux"
	}

	return cfg, nil
}
