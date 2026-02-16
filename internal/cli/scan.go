package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kotaroyamazaki/playcheck/internal/codescan"
	"github.com/kotaroyamazaki/playcheck/internal/datasafety"
	"github.com/kotaroyamazaki/playcheck/internal/manifest"
	"github.com/kotaroyamazaki/playcheck/internal/preflight"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

type scanOptions struct {
	format   string
	severity string
	output   string
}

// NewScanCmd creates the scan subcommand.
func NewScanCmd() *cobra.Command {
	opts := &scanOptions{}

	cmd := &cobra.Command{
		Use:   "scan [project-path]",
		Short: "Scan an Android project for Play Store compliance issues",
		Long:  "Analyzes an Android project directory and reports any Google Play Store policy violations or compliance issues.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "terminal", "Output format: terminal, json")
	cmd.Flags().StringVarP(&opts.severity, "severity", "s", "all", "Minimum severity to display: all, critical, warn, info")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Write report to file instead of stdout")

	return cmd
}

func runScan(projectPath string, opts *scanOptions) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access project path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", absPath)
	}

	minSeverity, err := parseSeverityFilter(opts.severity)
	if err != nil {
		return err
	}

	runner := preflight.NewDefaultRunner(func(r *preflight.Runner) {
		r.RegisterScanner(manifest.NewScanner())
		r.RegisterScanner(codescan.NewScanner())
		r.RegisterScanner(&datasafety.Checker{})
	})
	checkers := runner.Checkers()

	bar := progressbar.NewOptions(len(checkers),
		progressbar.OptionSetDescription("Scanning..."),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(50*time.Millisecond),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetPredictTime(false),
	)

	scanResult := runner.Run(absPath, func() {
		bar.Add(1)
	})

	bar.Finish()
	fmt.Fprint(os.Stderr, "\r\033[K") // clear progress bar line

	report := preflight.NewReport(scanResult, minSeverity)

	var outputData []byte

	switch opts.format {
	case "json":
		outputData, err = json.MarshalIndent(report.ToJSON(), "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		outputData = append(outputData, '\n')
	case "terminal":
		outputData = []byte(report.RenderTerminal())
	default:
		return fmt.Errorf("unknown format: %s (use 'terminal' or 'json')", opts.format)
	}

	if opts.output != "" {
		if err := os.WriteFile(opts.output, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", opts.output)
	} else {
		fmt.Print(string(outputData))
	}

	if report.HasCritical() {
		return fmt.Errorf("critical issues detected")
	}
	return nil
}

func parseSeverityFilter(s string) (preflight.Severity, error) {
	switch s {
	case "all":
		return preflight.SeverityInfo, nil
	case "info":
		return preflight.SeverityInfo, nil
	case "warn", "warning":
		return preflight.SeverityWarning, nil
	case "critical", "error":
		return preflight.SeverityCritical, nil
	default:
		return 0, fmt.Errorf("unknown severity filter: %s (use all, critical, warn, or info)", s)
	}
}
