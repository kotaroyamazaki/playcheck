package integration_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/codescan"
	"github.com/kotaroyamazaki/playcheck/internal/datasafety"
	"github.com/kotaroyamazaki/playcheck/internal/manifest"
	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func newFullRunner() *preflight.Runner {
	return preflight.NewDefaultRunner(func(r *preflight.Runner) {
		r.RegisterScanner(manifest.NewScanner())
		r.RegisterScanner(codescan.NewScanner())
		r.RegisterScanner(&datasafety.Checker{})
	})
}

func TestIntegration_ViolatingApp(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "violating-app")

	runner := newFullRunner()
	result := runner.Run(appDir, nil)

	if result == nil {
		t.Fatal("expected non-nil scan result")
	}

	// The violating app should have multiple findings
	if len(result.Findings) == 0 {
		t.Fatal("expected findings for violating app, got none")
	}

	// Check that we got findings from multiple scanners
	scannersSeen := make(map[string]bool)
	for _, f := range result.Findings {
		scannersSeen[f.CheckID] = true
	}

	t.Logf("Total findings: %d", len(result.Findings))
	for _, f := range result.Findings {
		t.Logf("  [%s] %s: %s (%s)", f.Severity, f.CheckID, f.Title, f.Location)
	}

	// Verify specific expected findings
	checkIDs := make(map[string]bool)
	for _, f := range result.Findings {
		checkIDs[f.CheckID] = true
	}

	// Manifest findings: target SDK too low (SDK001)
	if !checkIDs["SDK001"] {
		t.Error("expected SDK001 (target SDK too low) finding")
	}

	// Manifest findings: dangerous permissions
	hasDangerousPerm := checkIDs["DP001"] || checkIDs["DP002"] || checkIDs["DP003"]
	if !hasDangerousPerm {
		t.Error("expected at least one dangerous permission finding")
	}

	// Manifest findings: cleartext traffic (MV004)
	if !checkIDs["MV004"] {
		t.Error("expected MV004 (cleartext traffic) finding")
	}

	// Code scan findings: HTTP usage (CS001)
	if !checkIDs["CS001"] {
		t.Error("expected CS001 (HTTP usage) finding")
	}

	// Code scan findings: SMS usage (CS008)
	if !checkIDs["CS008"] {
		t.Error("expected CS008 (SMS usage) finding")
	}

	// Data safety: privacy policy missing (PDS001)
	if !checkIDs["PDS001"] {
		t.Error("expected PDS001 (privacy policy missing) finding")
	}

	// Data safety: account deletion missing (AD001)
	if !checkIDs["AD001"] {
		t.Error("expected AD001 (account deletion missing) finding")
	}

	// Report should indicate failure
	report := preflight.NewReport(result, preflight.SeverityInfo)
	if !report.HasCritical() {
		t.Error("expected HasCritical=true for violating app")
	}
	if report.CriticalCount == 0 {
		t.Error("expected critical count > 0")
	}
}

func TestIntegration_CleanApp(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "clean-app")

	runner := newFullRunner()
	result := runner.Run(appDir, nil)

	if result == nil {
		t.Fatal("expected non-nil scan result")
	}

	t.Logf("Total findings: %d", len(result.Findings))
	for _, f := range result.Findings {
		t.Logf("  [%s] %s: %s (%s)", f.Severity, f.CheckID, f.Title, f.Location)
	}

	// The clean app should have no critical findings
	report := preflight.NewReport(result, preflight.SeverityInfo)
	if report.CriticalCount > 0 {
		t.Errorf("expected 0 critical findings for clean app, got %d", report.CriticalCount)
		for _, f := range result.Findings {
			if f.Severity >= preflight.SeverityError {
				t.Logf("  unexpected critical: [%s] %s: %s", f.Severity, f.CheckID, f.Title)
			}
		}
	}

	// The clean app should not flag missing privacy policy
	for _, f := range result.Findings {
		if f.CheckID == "PDS001" {
			t.Error("did not expect PDS001 (privacy policy missing) for clean app")
		}
	}
}

func TestIntegration_JSONReport(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "violating-app")

	runner := newFullRunner()
	result := runner.Run(appDir, nil)

	report := preflight.NewReport(result, preflight.SeverityInfo)
	jsonReport := report.ToJSON()

	if jsonReport.ProjectPath == "" {
		t.Error("expected non-empty project path in JSON report")
	}
	if len(jsonReport.Findings) == 0 {
		t.Error("expected findings in JSON report")
	}
	if jsonReport.Timestamp == "" {
		t.Error("expected timestamp in JSON report")
	}

	// Verify all findings have required fields
	for _, f := range jsonReport.Findings {
		if f.CheckID == "" {
			t.Error("expected non-empty check_id in JSON finding")
		}
		if f.Severity == "" {
			t.Error("expected non-empty severity in JSON finding")
		}
		if f.Title == "" {
			t.Error("expected non-empty title in JSON finding")
		}
	}
}

func TestIntegration_ScanMetadata(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "clean-app")

	runner := newFullRunner()
	result := runner.Run(appDir, nil)

	if result.ScanMeta.ProjectPath != appDir {
		t.Errorf("expected project path %s, got %s", appDir, result.ScanMeta.ProjectPath)
	}
	if result.ScanMeta.Duration <= 0 {
		t.Error("expected positive scan duration")
	}
	if len(result.ScanMeta.ScannerIDs) != 3 {
		t.Errorf("expected 3 scanner IDs (manifest, codescan, datasafety), got %d", len(result.ScanMeta.ScannerIDs))
	}
}

func TestIntegration_SeverityFilter(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "violating-app")

	runner := newFullRunner()
	result := runner.Run(appDir, nil)

	reportAll := preflight.NewReport(result, preflight.SeverityInfo)
	reportCritical := preflight.NewReport(result, preflight.SeverityCritical)

	if len(reportCritical.Findings) >= len(reportAll.Findings) {
		t.Error("critical filter should show fewer findings than all")
	}
	if len(reportCritical.Findings) == 0 {
		t.Error("expected some critical findings in violating app")
	}
}

func TestIntegration_ProgressCallback(t *testing.T) {
	root := projectRoot()
	appDir := filepath.Join(root, "testdata", "sample-apps", "clean-app")

	runner := newFullRunner()
	callCount := 0
	runner.Run(appDir, func() {
		callCount++
	})

	// 3 scanners = 3 progress callbacks
	if callCount != 3 {
		t.Errorf("expected 3 progress callbacks, got %d", callCount)
	}
}
