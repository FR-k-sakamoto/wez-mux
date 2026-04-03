package main

import (
	"os"

	"github.com/k-sakamoto/wez-mux/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
