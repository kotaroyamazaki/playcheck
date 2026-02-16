package preflight

import "fmt"

// Severity represents the importance level of a finding.
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Location identifies where a finding was detected.
type Location struct {
	File string
	Line int
	Col  int
}

func (l Location) String() string {
	if l.Line > 0 {
		if l.Col > 0 {
			return fmt.Sprintf("%s:%d:%d", l.File, l.Line, l.Col)
		}
		return fmt.Sprintf("%s:%d", l.File, l.Line)
	}
	return l.File
}

// Finding represents a single compliance issue detected by a check.
type Finding struct {
	CheckID     string
	Title       string
	Description string
	Severity    Severity
	Location    Location
	Suggestion  string
}

func (f Finding) String() string {
	return fmt.Sprintf("[%s] %s: %s (%s)", f.Severity, f.CheckID, f.Title, f.Location)
}

// CheckResult holds the outcome of running a single compliance check.
type CheckResult struct {
	CheckID  string
	Passed   bool
	Findings []Finding
	Err      error
}

// Checker is the interface that all compliance checks must implement.
type Checker interface {
	ID() string
	Name() string
	Description() string
	Run(projectDir string) (*CheckResult, error)
}
