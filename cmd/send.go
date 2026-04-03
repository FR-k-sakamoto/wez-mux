package cmd

import (
	"fmt"
	"strings"

	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/k-sakamoto/wez-mux/internal/wezterm"
	"github.com/spf13/cobra"
)

func init() {
	var noEnter bool

	cmd := &cobra.Command{
		Use:   "send label message",
		Short: "Send text to a registered pane",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return err
			}

			pane, err := reg.Resolve(args[0])
			if err != nil {
				return err
			}

			message := strings.Join(args[1:], " ")

			client := wezterm.NewClient("")
			if err := client.SendText(pane.PaneID, message); err != nil {
				return fmt.Errorf("send text to %s: %w", args[0], err)
			}

			if !noEnter {
				if err := client.SendEnter(pane.PaneID); err != nil {
					return fmt.Errorf("send enter to %s: %w", args[0], err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noEnter, "no-enter", false, "Send the message without pressing Enter")

	rootCmd.AddCommand(cmd)
}
