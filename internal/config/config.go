package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultConfigYAML = `workspace: wez-mux
layout:
  rows:
    - panes:
        - label: analyzer
          model: sonnet
          skill: agent-analyzer
          percent: 66
        - label: designer
          model: opus
          skill: agent-designer
          percent: 50
    - panes:
        - label: coder
          model: sonnet
          skill: agent-coder
          percent: 50
        - label: tester
          model: haiku
          skill: agent-tester
          percent: 50
  row_ratio: [50, 50]
`

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

func Load(path string) (Config, error) {
	var data []byte
	var err error

	if path == "" {
		data = []byte(defaultConfigYAML)
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config %s: %w", path, err)
		}
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
