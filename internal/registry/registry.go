package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/k-sakamoto/wez-mux/internal/layout"
)

type Pane struct {
	PaneID  int    `json:"pane_id"`
	Model   string `json:"model"`
	Runtime string `json:"runtime"`
	Skill   string `json:"skill"`
}

type Registry struct {
	Workspace string          `json:"workspace"`
	CreatedAt time.Time       `json:"created_at"`
	CWD       string          `json:"cwd"`
	Panes     map[string]Pane `json:"panes"`
}

func FromSpec(paneID int, spec layout.PaneSpec) Pane {
	return Pane{
		PaneID:  paneID,
		Model:   spec.Model,
		Runtime: spec.Runtime,
		Skill:   spec.Skill,
	}
}

func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "wez-mux", "registry.json"), nil
}

func MustPath() string {
	path, err := Path()
	if err != nil {
		return "~/.config/wez-mux/registry.json"
	}

	return path
}

func Load() (Registry, error) {
	path, err := Path()
	if err != nil {
		return Registry{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Registry{}, fmt.Errorf("registry not found at %s", path)
		}
		return Registry{}, fmt.Errorf("read registry %s: %w", path, err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return Registry{}, fmt.Errorf("parse registry %s: %w", path, err)
	}
	if reg.Panes == nil {
		reg.Panes = make(map[string]Pane)
	}

	return reg, nil
}

func Save(reg Registry) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if reg.Panes == nil {
		reg.Panes = make(map[string]Pane)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write registry %s: %w", path, err)
	}

	return nil
}

func (r Registry) Resolve(label string) (Pane, error) {
	pane, ok := r.Panes[label]
	if !ok {
		return Pane{}, fmt.Errorf("unknown pane label %q", label)
	}
	return pane, nil
}

func (r Registry) Labels() []string {
	labels := make([]string, 0, len(r.Panes))
	for label := range r.Panes {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}
