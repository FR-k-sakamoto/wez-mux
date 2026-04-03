package cmd

import (
	"sort"

	"github.com/k-sakamoto/wez-mux/internal/registry"
	"github.com/k-sakamoto/wez-mux/internal/wezterm"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show live status for registered panes",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load()
			if err != nil {
				return err
			}

			client := wezterm.NewClient("")
			panes, err := client.ListPanes()
			if err != nil {
				return err
			}

			active := make(map[int]struct{}, len(panes))
			for _, pane := range panes {
				active[pane.PaneID] = struct{}{}
			}

			labels := reg.Labels()
			sort.Strings(labels)

			rows := make([]paneRow, 0, len(labels))
			for _, label := range labels {
				pane := reg.Panes[label]
				status := "dead"
				if _, ok := active[pane.PaneID]; ok {
					status = "active"
				}

				rows = append(rows, paneRow{
					Label:   label,
					PaneID:  pane.PaneID,
					Model:   pane.Model,
					Runtime: pane.Runtime,
					Status:  status,
				})
			}

			printPaneTable(rows)
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}
