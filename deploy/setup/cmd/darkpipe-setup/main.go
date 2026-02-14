package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "darkpipe-setup",
		Short: "DarkPipe interactive setup with live validation",
		Long:  "Interactive setup tool for DarkPipe mail server deployment",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("darkpipe-setup version %s\n", version)
		},
	}

	rootCmd.AddCommand(versionCmd)
	// Setup command will be added in Task 2

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
