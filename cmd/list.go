package cmd

import (
	"sort"

	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered panes",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return err
			}

			labels := reg.Labels()
			sort.Strings(labels)

			rows := make([]paneRow, 0, len(labels))
			for _, label := range labels {
				pane := reg.Panes[label]
				rows = append(rows, paneRow{
					Label:   label,
					PaneID:  pane.PaneID,
					Model:   pane.Model,
					Runtime: pane.Runtime,
					Status:  "registered",
				})
			}

			printPaneTable(rows)
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}
