package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command for the playcheck CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "playcheck",
		Short: "Google Play Store compliance scanner",
		Long:  "Scans Android projects for Google Play Store policy compliance issues before submission.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(NewScanCmd())

	return rootCmd
}
