package main

import (
	"fmt"
	"os"

	"github.com/yourusername/playcheck/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
