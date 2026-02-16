package cli

import (
	"os"
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

func TestParseSeverityFilter(t *testing.T) {
	tests := []struct {
		input   string
		want    preflight.Severity
		wantErr bool
	}{
		{"all", preflight.SeverityInfo, false},
		{"info", preflight.SeverityInfo, false},
		{"warn", preflight.SeverityWarning, false},
		{"warning", preflight.SeverityWarning, false},
		{"critical", preflight.SeverityCritical, false},
		{"error", preflight.SeverityCritical, false},
		{"", 0, true},
		{"unknown", 0, true},
		{"DEBUG", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseSeverityFilter(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("parseSeverityFilter(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSeverityFilter(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseSeverityFilter(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestRunScan_NonexistentPath(t *testing.T) {
	opts := &scanOptions{format: "terminal", severity: "all"}
	err := runScan("/nonexistent/path/that/does/not/exist", opts)
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestRunScan_NotADirectory(t *testing.T) {
	f := t.TempDir() + "/file.txt"
	if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := &scanOptions{format: "terminal", severity: "all"}
	err := runScan(f, opts)
	if err == nil {
		t.Error("expected error when path is a file, not a directory")
	}
}

func TestRunScan_InvalidSeverity(t *testing.T) {
	dir := t.TempDir()
	opts := &scanOptions{format: "terminal", severity: "badvalue"}
	err := runScan(dir, opts)
	if err == nil {
		t.Error("expected error for invalid severity filter")
	}
}

func TestRunScan_UnknownFormat(t *testing.T) {
	dir := t.TempDir()
	opts := &scanOptions{format: "yaml", severity: "all"}
	err := runScan(dir, opts)
	if err == nil {
		t.Error("expected error for unknown output format")
	}
}

func TestRunScan_JSONOutputToFile(t *testing.T) {
	dir := t.TempDir()
	outFile := dir + "/report.json"
	opts := &scanOptions{format: "json", severity: "all", output: outFile}
	// Scanning an empty directory -- no manifest will be found, but should not panic.
	_ = runScan(dir, opts)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file to be created: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output file")
	}
}

func TestRunScan_TerminalOutputToFile(t *testing.T) {
	dir := t.TempDir()
	outFile := dir + "/report.txt"
	opts := &scanOptions{format: "terminal", severity: "all", output: outFile}
	_ = runScan(dir, opts)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file to be created: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output file")
	}
}

func TestNewScanCmd(t *testing.T) {
	cmd := NewScanCmd()
	if cmd.Use != "scan [project-path]" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	if f := cmd.Flags().Lookup("format"); f == nil {
		t.Error("expected --format flag")
	}
	if s := cmd.Flags().Lookup("severity"); s == nil {
		t.Error("expected --severity flag")
	}
	if o := cmd.Flags().Lookup("output"); o == nil {
		t.Error("expected --output flag")
	}
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "playcheck" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'scan' subcommand")
	}
}
