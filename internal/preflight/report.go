package preflight

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Report holds the analyzed scan results and provides rendering methods.
type Report struct {
	ProjectPath   string
	ScanResult    *ScanResult
	MinSeverity   Severity
	CriticalCount int
	WarningCount  int
	InfoCount     int
	Findings      []Finding
}

// JSONReport is the JSON-serializable representation of a scan report.
type JSONReport struct {
	Timestamp   string        `json:"timestamp"`
	ProjectPath string        `json:"project_path"`
	Summary     JSONSummary   `json:"summary"`
	Findings    []JSONFinding `json:"findings"`
}

// JSONSummary holds aggregate counts for JSON output.
type JSONSummary struct {
	TotalChecks   int    `json:"total_checks"`
	Passed        int    `json:"passed"`
	Failed        int    `json:"failed"`
	CriticalCount int    `json:"critical"`
	WarningCount  int    `json:"warning"`
	InfoCount     int    `json:"info"`
	Duration      string `json:"duration"`
}

// JSONFinding is a single finding in JSON format.
type JSONFinding struct {
	CheckID     string `json:"check_id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// NewReport creates a Report from a ScanResult, filtering findings by minimum severity.
func NewReport(result *ScanResult, minSeverity Severity) *Report {
	r := &Report{
		ProjectPath: result.ScanMeta.ProjectPath,
		ScanResult:  result,
		MinSeverity: minSeverity,
	}

	for _, f := range result.Findings {
		if f.Severity < minSeverity {
			continue
		}
		r.Findings = append(r.Findings, f)
		switch f.Severity {
		case SeverityCritical, SeverityError:
			r.CriticalCount++
		case SeverityWarning:
			r.WarningCount++
		case SeverityInfo:
			r.InfoCount++
		}
	}

	return r
}

// HasCritical returns true if any critical-level findings exist (unfiltered).
func (r *Report) HasCritical() bool {
	for _, f := range r.ScanResult.Findings {
		if f.Severity == SeverityCritical || f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// ToJSON returns a JSON-serializable report structure.
func (r *Report) ToJSON() JSONReport {
	findings := make([]JSONFinding, 0, len(r.Findings))
	for _, f := range r.Findings {
		findings = append(findings, JSONFinding{
			CheckID:     f.CheckID,
			Severity:    f.Severity.String(),
			Title:       f.Title,
			Description: f.Description,
			Location:    f.Location.String(),
			Suggestion:  f.Suggestion,
		})
	}

	totalChecks := r.ScanResult.TotalPassed + r.ScanResult.TotalFailed

	return JSONReport{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ProjectPath: r.ProjectPath,
		Summary: JSONSummary{
			TotalChecks:   totalChecks,
			Passed:        r.ScanResult.TotalPassed,
			Failed:        r.ScanResult.TotalFailed,
			CriticalCount: r.CriticalCount,
			WarningCount:  r.WarningCount,
			InfoCount:     r.InfoCount,
			Duration:      r.ScanResult.ScanMeta.Duration.String(),
		},
		Findings: findings,
	}
}

// RenderTerminal produces colored, human-readable terminal output.
func (r *Report) RenderTerminal() string {
	var b strings.Builder

	headerColor := color.New(color.FgCyan, color.Bold)
	criticalColor := color.New(color.FgRed, color.Bold)
	warningColor := color.New(color.FgYellow)
	infoColor := color.New(color.FgBlue)
	passedColor := color.New(color.FgGreen, color.Bold)
	dimColor := color.New(color.Faint)

	totalChecks := r.ScanResult.TotalPassed + r.ScanResult.TotalFailed

	b.WriteString("\n")
	headerColor.Fprint(&b, "=== Play Store Compliance Report ===")
	b.WriteString("\n")
	dimColor.Fprintf(&b, "Project: %s", r.ProjectPath)
	b.WriteString("\n")
	dimColor.Fprintf(&b, "Duration: %s", r.ScanResult.ScanMeta.Duration)
	b.WriteString("\n\n")

	if len(r.Findings) == 0 {
		passedColor.Fprint(&b, "All checks passed! No issues found.")
		b.WriteString("\n")
	} else {
		var criticals, warnings, infos []Finding
		for _, f := range r.Findings {
			switch f.Severity {
			case SeverityCritical, SeverityError:
				criticals = append(criticals, f)
			case SeverityWarning:
				warnings = append(warnings, f)
			case SeverityInfo:
				infos = append(infos, f)
			}
		}

		if len(criticals) > 0 {
			criticalColor.Fprintf(&b, "CRITICAL (%d)", len(criticals))
			b.WriteString("\n")
			for _, f := range criticals {
				renderFinding(&b, f, criticalColor, dimColor)
			}
			b.WriteString("\n")
		}

		if len(warnings) > 0 {
			warningColor.Fprintf(&b, "WARNING (%d)", len(warnings))
			b.WriteString("\n")
			for _, f := range warnings {
				renderFinding(&b, f, warningColor, dimColor)
			}
			b.WriteString("\n")
		}

		if len(infos) > 0 {
			infoColor.Fprintf(&b, "INFO (%d)", len(infos))
			b.WriteString("\n")
			for _, f := range infos {
				renderFinding(&b, f, infoColor, dimColor)
			}
			b.WriteString("\n")
		}
	}

	// Summary bar
	b.WriteString(strings.Repeat("-", 50))
	b.WriteString("\n")
	fmt.Fprintf(&b, "Checks run: %d | Passed: ", totalChecks)
	passedColor.Fprintf(&b, "%d", r.ScanResult.TotalPassed)
	b.WriteString(" | Critical: ")
	if r.CriticalCount > 0 {
		criticalColor.Fprintf(&b, "%d", r.CriticalCount)
	} else {
		fmt.Fprintf(&b, "%d", r.CriticalCount)
	}
	b.WriteString(" | Warnings: ")
	if r.WarningCount > 0 {
		warningColor.Fprintf(&b, "%d", r.WarningCount)
	} else {
		fmt.Fprintf(&b, "%d", r.WarningCount)
	}
	b.WriteString(" | Info: ")
	fmt.Fprintf(&b, "%d", r.InfoCount)
	b.WriteString("\n")

	if r.CriticalCount > 0 {
		b.WriteString("\n")
		criticalColor.Fprint(&b, "RESULT: FAIL")
		b.WriteString(" - Critical issues must be resolved before submission.\n")
	} else {
		b.WriteString("\n")
		passedColor.Fprint(&b, "RESULT: PASS")
		b.WriteString(" - No critical issues found.\n")
	}

	return b.String()
}

func renderFinding(b *strings.Builder, f Finding, severityColor *color.Color, dimColor *color.Color) {
	severityColor.Fprintf(b, "  [%s]", f.Severity)
	fmt.Fprintf(b, " %s", f.Title)
	b.WriteString("\n")
	if f.Location.File != "" {
		dimColor.Fprintf(b, "         %s", f.Location)
		b.WriteString("\n")
	}
	if f.Description != "" {
		fmt.Fprintf(b, "         %s", f.Description)
		b.WriteString("\n")
	}
	if f.Suggestion != "" {
		dimColor.Fprintf(b, "         Suggestion: %s", f.Suggestion)
		b.WriteString("\n")
	}
}
