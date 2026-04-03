package cmd

import (
	"fmt"

	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/k-sakamoto/wez-mux/internal/wezterm"
	"github.com/spf13/cobra"
)

func init() {
	var killAll bool

	cmd := &cobra.Command{
		Use:   "kill label",
		Short: "Kill one registered pane or all non-orchestrator panes",
		Args: func(cmd *cobra.Command, args []string) error {
			if killAll {
				if len(args) != 0 {
					return fmt.Errorf("--all does not take a label")
				}
				return nil
			}

			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return err
			}

			client := wezterm.NewClient("")

			if killAll {
				for _, label := range reg.Labels() {
					if label == "orchestrator" {
						continue
					}

					pane := reg.Panes[label]
					if err := client.KillPane(pane.PaneID); err != nil {
						return fmt.Errorf("kill %s: %w", label, err)
					}
					delete(reg.Panes, label)
				}

				return registry.Save(reg)
			}

			label := args[0]
			pane, err := reg.Resolve(label)
			if err != nil {
				return err
			}

			if err := client.KillPane(pane.PaneID); err != nil {
				return fmt.Errorf("kill %s: %w", label, err)
			}

			delete(reg.Panes, label)
			return registry.Save(reg)
		},
	}

	cmd.Flags().BoolVar(&killAll, "all", false, "Kill all registered panes except orchestrator")

	rootCmd.AddCommand(cmd)
}
