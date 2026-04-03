package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/k-sakamoto/wez-mux/internal/config"
	"github.com/k-sakamoto/wez-mux/internal/layout"
	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/k-sakamoto/wez-mux/internal/wezterm"
	"github.com/spf13/cobra"
)

// execClaude replaces the current process with the claude command.
func execClaude(command string) error {
	parts := strings.Fields(command)
	binary, err := exec.LookPath(parts[0])
	if err != nil {
		return fmt.Errorf("find %s: %w", parts[0], err)
	}
	return syscall.Exec(binary, parts, os.Environ())
}

func init() {
	var configPath string
	var cwd string
	var noStart bool
	var paneID int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Split current tab into 5 panes and start agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := wezterm.NewClient("")

			var orchestratorID int
			var err error

			if cmd.Flags().Changed("pane-id") {
				orchestratorID = paneID
			} else {
				orchestratorID, err = wezterm.CurrentPaneID(client)
				if err != nil {
					return err
				}
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			if cwd == "" {
				cwd, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("get cwd: %w", err)
				}
			}

			plan, err := layout.BuildPlan(cfg)
			if err != nil {
				return err
			}

			// Split strategy:
			// 1. Split orchestrator bottom 50% → bottom pane (coder)
			// 2. Split orchestrator right 66%  → analyzer
			// 3. Split analyzer right 50%      → designer
			// 4. Split coder right 50%         → tester
			//
			// ┌──────────┬──────────┬──────────┐
			// │orchestr. │ analyzer │ designer │
			// ├──────────┴────┬─────┴──────────┤
			// │    coder      │     tester     │
			// └───────────────┴────────────────┘

			coderID, err := client.SplitPane(orchestratorID, wezterm.SplitPaneOptions{
				Direction: "bottom",
				Percent:   50,
				CWD:       cwd,
			})
			if err != nil {
				return fmt.Errorf("create coder pane: %w", err)
			}

			analyzerID, err := client.SplitPane(orchestratorID, wezterm.SplitPaneOptions{
				Direction: "right",
				Percent:   66,
				CWD:       cwd,
			})
			if err != nil {
				return fmt.Errorf("create analyzer pane: %w", err)
			}

			designerID, err := client.SplitPane(analyzerID, wezterm.SplitPaneOptions{
				Direction: "right",
				Percent:   50,
				CWD:       cwd,
			})
			if err != nil {
				return fmt.Errorf("create designer pane: %w", err)
			}

			testerID, err := client.SplitPane(coderID, wezterm.SplitPaneOptions{
				Direction: "right",
				Percent:   50,
				CWD:       cwd,
			})
			if err != nil {
				return fmt.Errorf("create tester pane: %w", err)
			}

			reg := registry.Registry{
				Workspace: plan.Workspace,
				CreatedAt: time.Now().UTC(),
				CWD:       cwd,
				Panes: map[string]registry.Pane{
					"orchestrator": {
						PaneID:  orchestratorID,
						Model:   "opus",
						Runtime: "claude-code",
						Skill:   "agent-orchestrator",
					},
					"analyzer": registry.FromSpec(analyzerID, plan.Specs["analyzer"]),
					"designer": registry.FromSpec(designerID, plan.Specs["designer"]),
					"coder":    registry.FromSpec(coderID, plan.Specs["coder"]),
					"tester":   registry.FromSpec(testerID, plan.Specs["tester"]),
				},
			}

			if err := registry.Save(reg); err != nil {
				return err
			}

			if !noStart {
				// Start agents in the 4 sub-panes
				agentLabels := []string{"analyzer", "designer", "coder", "tester"}
				paneIDs := map[string]int{
					"analyzer": analyzerID,
					"designer": designerID,
					"coder":    coderID,
					"tester":   testerID,
				}
				for _, label := range agentLabels {
					spec := plan.Specs[label]
					if err := client.SendText(paneIDs[label], spec.StartCommand); err != nil {
						return fmt.Errorf("start %s agent: %w", label, err)
					}
					if err := client.SendEnter(paneIDs[label]); err != nil {
						return fmt.Errorf("send enter to %s: %w", label, err)
					}
				}
			}

			fmt.Printf("Initialized workspace %q (5 panes) — registry at %s\n", reg.Workspace, registry.MustPath())

			if !noStart {
				// Replace this process with orchestrator claude
				// exec replaces the current shell, so this must be last
				orchestratorSpec := plan.Specs["orchestrator"]
				fmt.Printf("Starting orchestrator: %s\n", orchestratorSpec.StartCommand)
				return execClaude(orchestratorSpec.StartCommand)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to YAML config")
	cmd.Flags().StringVar(&cwd, "cwd", "", "Working directory for new panes")
	cmd.Flags().BoolVar(&noStart, "no-start", false, "Create panes without starting agents")
	cmd.Flags().IntVar(&paneID, "pane-id", 0, "Specify the orchestrator pane ID (auto-detected if omitted)")

	rootCmd.AddCommand(cmd)
}
