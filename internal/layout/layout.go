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
// Checks ~/.claude/skills/<skill>/SKILL.md first.
func skillFilePath(skill string) string {
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

func buildStartCommand(model, label, skill string) string {
	cmd := fmt.Sprintf("claude --model %s --name %s --permission-mode auto", model, label)
	if path := skillFilePath(skill); path != "" {
		cmd += fmt.Sprintf(" --append-system-prompt-file %s", path)
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
			StartCommand: buildStartCommand("opus", "orchestrator", "agent-orchestrator"),
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
			runtime = "claude-code+codex"
		}

		specs[label] = PaneSpec{
			Label:        label,
			Model:        pane.Model,
			Skill:        pane.Skill,
			Runtime:      runtime,
			StartCommand: buildStartCommand(pane.Model, label, pane.Skill),
		}
	}

	return Plan{
		Workspace: cfg.Workspace,
		Specs:     specs,
	}, nil
}
