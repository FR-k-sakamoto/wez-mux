package layout

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k-sakamoto/wez-mux/internal/config"
)

type PaneSpec struct {
	Label        string
	Model        string
	Skill        string
	Runtime      string
	StartCommand string
}

type Plan struct {
	Workspace string
	Specs     map[string]PaneSpec
}

var requiredLabels = []string{"analyzer", "designer", "coder", "tester"}

// skillFilePath returns the absolute path to a skill's SKILL.md.
// Search order:
//  1. .agent/skills/<skill>/SKILL.md (relative to wez-mux binary or cwd)
//  2. ~/.claude/skills/<skill>/SKILL.md
func skillFilePath(skill string) string {
	// 1. Check .agent/skills/ relative to executable
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), ".agent", "skills", skill, "SKILL.md")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// Also check cwd
	if cwd, err := os.Getwd(); err == nil {
		path := filepath.Join(cwd, ".agent", "skills", skill, "SKILL.md")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 2. Fallback to ~/.claude/skills/
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".claude", "skills", skill, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func buildStartCommand(model, label, skill, runtime string) string {
	skillPath := skillFilePath(skill)

	if runtime == "codex" {
		cmd := "codex --full-auto"
		if model != "" {
			cmd += fmt.Sprintf(" --model %s", model)
		}
		// Pass SKILL.md content as initial prompt via shell expansion
		if skillPath != "" {
			cmd += fmt.Sprintf(` "$(cat %s)"`, skillPath)
		}
		return cmd
	}

	// Claude Code
	cmd := fmt.Sprintf("claude --model %s --name %s --permission-mode auto", model, label)
	if skillPath != "" {
		cmd += fmt.Sprintf(" --append-system-prompt-file %s", skillPath)
	}
	return cmd
}

func BuildPlan(cfg config.Config) (Plan, error) {
	specs := map[string]PaneSpec{
		"orchestrator": {
			Label:        "orchestrator",
			Model:        "opus",
			Skill:        "agent-orchestrator",
			Runtime:      "claude-code",
			StartCommand: buildStartCommand("opus", "orchestrator", "agent-orchestrator", "claude-code"),
		},
	}

	panesByLabel := make(map[string]config.PaneConfig)
	for _, row := range cfg.Layout.Rows {
		for _, pane := range row.Panes {
			panesByLabel[pane.Label] = pane
		}
	}

	for _, label := range requiredLabels {
		pane, ok := panesByLabel[label]
		if !ok {
			return Plan{}, fmt.Errorf("config is missing pane %q", label)
		}

		runtime := "claude-code"
		if pane.Codex {
			runtime = "codex"
		}

		specs[label] = PaneSpec{
			Label:        label,
			Model:        pane.Model,
			Skill:        pane.Skill,
			Runtime:      runtime,
			StartCommand: buildStartCommand(pane.Model, label, pane.Skill, runtime),
		}
	}

	return Plan{
		Workspace: cfg.Workspace,
		Specs:     specs,
	}, nil
}
