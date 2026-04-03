package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wez-mux",
	Short: "Manage multi-agent pane orchestration in WezTerm",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

type paneRow struct {
	Label   string
	PaneID  int
	Model   string
	Runtime string
	Status  string
}

func printPaneTable(rows []paneRow) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "LABEL\tPANE_ID\tMODEL\tRUNTIME\tSTATUS")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", row.Label, row.PaneID, row.Model, row.Runtime, row.Status)
	}
	_ = w.Flush()
}
