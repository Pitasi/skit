package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "deck",
		Short: "A CLI for building presentations from Markdown",
	}

	root.AddCommand(
		newInitCmd(),
		newBuildCmd(),
		newServeCmd(),
		newPDFCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
