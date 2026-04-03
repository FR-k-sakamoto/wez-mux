package cmd

import (
	"fmt"
	"strconv"

	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/k-sakamoto/wez-mux/internal/wezterm"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "read label [lines]",
		Short: "Read text from a registered pane (includes scrollback)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lineCount := 50
			if len(args) == 2 {
				n, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid line count %q: %w", args[1], err)
				}
				if n < 1 {
					return fmt.Errorf("line count must be positive")
				}
				lineCount = n
			}

			reg, err := registry.Load()
			if err != nil {
				return err
			}

			pane, err := reg.Resolve(args[0])
			if err != nil {
				return err
			}

			client := wezterm.NewClient("")

			// Read from deep scrollback to capture full output.
			// Use a large negative start-line to get scrollback history.
			startLine := -10000
			text, err := client.GetText(pane.PaneID, startLine)
			if err != nil {
				// Fallback to viewport only if scrollback fails
				text, err = client.GetText(pane.PaneID, 0)
				if err != nil {
					return fmt.Errorf("read %s: %w", args[0], err)
				}
			}

			fmt.Print(wezterm.LastLines(text, lineCount))
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}
