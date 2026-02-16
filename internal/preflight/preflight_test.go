package preflight

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// mockScanner implements Checker for testing the Runner.
type mockScanner struct {
	id       string
	findings []Finding
	err      error
}

func (m *mockScanner) ID() string          { return m.id }
func (m *mockScanner) Name() string        { return "Mock " + m.id }
func (m *mockScanner) Description() string { return "Mock scanner for testing" }

func (m *mockScanner) Run(projectDir string) (*CheckResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CheckResult{
		CheckID:  m.id,
		Passed:   len(m.findings) == 0,
		Findings: m.findings,
	}, nil
}

func TestRunner_NoScanners(t *testing.T) {
	r := &Runner{}
	result := r.Run("/tmp", nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}

func TestRunner_SingleScanner(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{
		id: "test-scanner",
		findings: []Finding{
			{CheckID: "T001", Title: "Test finding", Severity: SeverityWarning},
		},
	})

	result := r.Run("/tmp", nil)
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	if result.TotalFailed != 1 {
		t.Errorf("expected 1 failed, got %d", result.TotalFailed)
	}
}

func TestRunner_MultipleScanners(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{
		id: "scanner-a",
		findings: []Finding{
			{CheckID: "A001", Title: "Finding A", Severity: SeverityCritical},
		},
	})
	r.RegisterScanner(&mockScanner{
		id:       "scanner-b",
		findings: nil, // passes
	})
	r.RegisterScanner(&mockScanner{
		id: "scanner-c",
		findings: []Finding{
			{CheckID: "C001", Title: "Finding C", Severity: SeverityWarning},
		},
	})

	result := r.Run("/tmp", nil)
	if len(result.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(result.Findings))
	}
	if result.TotalPassed != 1 {
		t.Errorf("expected 1 passed, got %d", result.TotalPassed)
	}
	if result.TotalFailed != 2 {
		t.Errorf("expected 2 failed, got %d", result.TotalFailed)
	}
	// Findings should be sorted by severity (critical first)
	if len(result.Findings) >= 2 && result.Findings[0].Severity < result.Findings[1].Severity {
		t.Error("expected findings sorted by severity descending")
	}
}

func TestRunner_ProgressCallback(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{id: "s1"})
	r.RegisterScanner(&mockScanner{id: "s2"})
	r.RegisterScanner(&mockScanner{id: "s3"})

	callCount := &atomic.Int32{}
	r.Run("/tmp", func() {
		callCount.Add(1)
	})
	if callCount.Load() != 3 {
		t.Errorf("expected progress called 3 times, got %d", callCount.Load())
	}
}

func TestRunner_ScannerError(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{
		id:  "error-scanner",
		err: fmt.Errorf("scanner failed"),
	})

	result := r.Run("/tmp", nil)
	cr := result.ByScanner["error-scanner"]
	if cr == nil {
		t.Fatal("expected check result for error-scanner")
	}
	if cr.Err == nil {
		t.Error("expected error in check result")
	}
}

func TestRunner_Deduplication(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{
		id: "scanner-1",
		findings: []Finding{
			{CheckID: "DUP", Title: "Dup finding", Severity: SeverityWarning, Location: Location{File: "a.java", Line: 10}},
		},
	})
	r.RegisterScanner(&mockScanner{
		id: "scanner-2",
		findings: []Finding{
			{CheckID: "DUP", Title: "Dup finding", Severity: SeverityWarning, Location: Location{File: "a.java", Line: 10}},
		},
	})

	result := r.Run("/tmp", nil)
	dupCount := 0
	for _, f := range result.Findings {
		if f.CheckID == "DUP" {
			dupCount++
		}
	}
	if dupCount != 1 {
		t.Errorf("expected 1 deduplicated finding, got %d", dupCount)
	}
}

func TestRunner_Metadata(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{id: "m1"})
	r.RegisterScanner(&mockScanner{id: "m2"})

	result := r.Run("/some/path", nil)
	if result.ScanMeta.ProjectPath != "/some/path" {
		t.Errorf("expected project path /some/path, got %s", result.ScanMeta.ProjectPath)
	}
	if result.ScanMeta.Duration < 0 {
		t.Error("expected non-negative duration")
	}
	if len(result.ScanMeta.ScannerIDs) != 2 {
		t.Errorf("expected 2 scanner IDs, got %d", len(result.ScanMeta.ScannerIDs))
	}
}

func TestRunner_Checkers(t *testing.T) {
	r := &Runner{}
	r.RegisterScanner(&mockScanner{id: "c1"})
	r.RegisterScanner(&mockScanner{id: "c2"})
	checkers := r.Checkers()
	if len(checkers) != 2 {
		t.Errorf("expected 2 checkers, got %d", len(checkers))
	}
}

func TestNewDefaultRunner(t *testing.T) {
	r := NewDefaultRunner(func(r *Runner) {
		r.RegisterScanner(&mockScanner{id: "custom"})
	})
	if len(r.Checkers()) != 1 {
		t.Errorf("expected 1 checker, got %d", len(r.Checkers()))
	}
}

func TestDeduplicateFindings_Empty(t *testing.T) {
	result := deduplicateFindings(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestDeduplicateFindings_NoDuplicates(t *testing.T) {
	findings := []Finding{
		{CheckID: "A", Location: Location{File: "a.java", Line: 1}},
		{CheckID: "B", Location: Location{File: "b.java", Line: 2}},
	}
	result := deduplicateFindings(findings)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestReport_NewReport(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "C1", Severity: SeverityCritical},
			{CheckID: "W1", Severity: SeverityWarning},
			{CheckID: "I1", Severity: SeverityInfo},
		},
		ScanMeta: ScanMetadata{ProjectPath: "/test"},
	}

	report := NewReport(sr, SeverityInfo)
	if report.CriticalCount != 1 {
		t.Errorf("expected 1 critical, got %d", report.CriticalCount)
	}
	if report.WarningCount != 1 {
		t.Errorf("expected 1 warning, got %d", report.WarningCount)
	}
	if report.InfoCount != 1 {
		t.Errorf("expected 1 info, got %d", report.InfoCount)
	}
	if len(report.Findings) != 3 {
		t.Errorf("expected 3 findings, got %d", len(report.Findings))
	}
}

func TestReport_SeverityFilter(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "C1", Severity: SeverityCritical},
			{CheckID: "W1", Severity: SeverityWarning},
			{CheckID: "I1", Severity: SeverityInfo},
		},
		ScanMeta: ScanMetadata{ProjectPath: "/test"},
	}

	report := NewReport(sr, SeverityWarning)
	if len(report.Findings) != 2 {
		t.Errorf("expected 2 findings (warning+), got %d", len(report.Findings))
	}
}

func TestReport_HasCritical(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "C1", Severity: SeverityCritical},
		},
		ScanMeta: ScanMetadata{ProjectPath: "/test"},
	}
	report := NewReport(sr, SeverityInfo)
	if !report.HasCritical() {
		t.Error("expected HasCritical to be true")
	}
}

func TestReport_HasCritical_OnlyWarnings(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "W1", Severity: SeverityWarning},
		},
		ScanMeta: ScanMetadata{ProjectPath: "/test"},
	}
	report := NewReport(sr, SeverityInfo)
	if report.HasCritical() {
		t.Error("expected HasCritical to be false for warnings only")
	}
}

func TestReport_ToJSON(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "C1", Severity: SeverityCritical, Title: "Test", Location: Location{File: "a.java", Line: 1}},
		},
		TotalPassed: 1,
		TotalFailed: 1,
		ScanMeta:    ScanMetadata{ProjectPath: "/test"},
	}
	report := NewReport(sr, SeverityInfo)
	jsonReport := report.ToJSON()

	if jsonReport.ProjectPath != "/test" {
		t.Errorf("expected project path /test, got %s", jsonReport.ProjectPath)
	}
	if len(jsonReport.Findings) != 1 {
		t.Errorf("expected 1 finding in JSON, got %d", len(jsonReport.Findings))
	}
	if jsonReport.Summary.TotalChecks != 2 {
		t.Errorf("expected 2 total checks, got %d", jsonReport.Summary.TotalChecks)
	}
}

func TestReport_RenderTerminal(t *testing.T) {
	sr := &ScanResult{
		Findings: []Finding{
			{CheckID: "C1", Severity: SeverityCritical, Title: "Critical Issue"},
			{CheckID: "W1", Severity: SeverityWarning, Title: "Warning Issue"},
		},
		TotalPassed: 1,
		TotalFailed: 1,
		ScanMeta:    ScanMetadata{ProjectPath: "/test"},
	}
	report := NewReport(sr, SeverityInfo)
	output := report.RenderTerminal()

	if output == "" {
		t.Error("expected non-empty terminal output")
	}
	if !strings.Contains(output, "Play Store Compliance Report") {
		t.Error("expected report header in output")
	}
	if !strings.Contains(output, "FAIL") {
		t.Error("expected FAIL result when critical findings present")
	}
}

func TestReport_RenderTerminal_AllPassed(t *testing.T) {
	sr := &ScanResult{
		Findings:    nil,
		TotalPassed: 3,
		ScanMeta:    ScanMetadata{ProjectPath: "/test"},
	}
	report := NewReport(sr, SeverityInfo)
	output := report.RenderTerminal()

	if !strings.Contains(output, "PASS") {
		t.Error("expected PASS result when no findings")
	}
}

