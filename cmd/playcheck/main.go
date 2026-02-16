package main

import (
	"fmt"
	"os"

	"github.com/kotaroyamazaki/playcheck/internal/cli"
)

// Version information set via ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cli.NewRootCmd()
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
