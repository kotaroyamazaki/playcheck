package preflight

import (
	"sort"
	"sync"
	"time"
)

// Scanner is the interface that compliance scanners must implement.
// Each scanner focuses on a specific area (manifest, code, data safety)
// and implements the Checker interface defined in types.go.
type Scanner interface {
	Checker
}

// ScanResult holds the aggregated results from all scanners.
type ScanResult struct {
	Findings    []Finding
	ScanMeta    ScanMetadata
	ByScanner   map[string]*CheckResult
	TotalPassed int
	TotalFailed int
}

// ScanMetadata contains information about the scan execution.
type ScanMetadata struct {
	ProjectPath string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	ScannerIDs  []string
}

// Runner orchestrates compliance scanners and aggregates results.
type Runner struct {
	scanners []Scanner
}

// NewRunner creates a Runner with all built-in scanners registered.
func NewRunner() *Runner {
	r := &Runner{}
	r.registerBuiltinScanners()
	return r
}

// registerBuiltinScanners is a placeholder for scanner registration.
// Scanners are registered externally via RegisterScanner to avoid import cycles,
// since scanner packages import preflight for types.
func (r *Runner) registerBuiltinScanners() {
	// Scanners are registered by NewDefaultRunner() in the register package.
}

// RegisterScanner adds a scanner to the runner.
func (r *Runner) RegisterScanner(s Scanner) {
	r.scanners = append(r.scanners, s)
}

// Checkers returns the list of registered scanners as Checkers.
func (r *Runner) Checkers() []Checker {
	checkers := make([]Checker, len(r.scanners))
	for i, s := range r.scanners {
		checkers[i] = s
	}
	return checkers
}

// Run executes all registered scanners against the project directory.
// The onComplete callback is invoked after each scanner finishes, which
// is used by the CLI to advance the progress bar.
// Scanners run concurrently for better performance.
func (r *Runner) Run(projectDir string, onComplete func()) *ScanResult {
	startTime := time.Now()

	result := &ScanResult{
		ByScanner: make(map[string]*CheckResult, len(r.scanners)),
		ScanMeta: ScanMetadata{
			ProjectPath: projectDir,
			StartTime:   startTime,
		},
	}

	for _, s := range r.scanners {
		result.ScanMeta.ScannerIDs = append(result.ScanMeta.ScannerIDs, s.ID())
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, s := range r.scanners {
		wg.Add(1)
		go func(scanner Scanner) {
			defer wg.Done()

			cr, err := scanner.Run(projectDir)
			if cr == nil {
				cr = &CheckResult{
					CheckID: scanner.ID(),
				}
			}
			if err != nil {
				cr.Err = err
			}

			mu.Lock()
			result.ByScanner[scanner.ID()] = cr
			result.Findings = append(result.Findings, cr.Findings...)
			if cr.Passed {
				result.TotalPassed++
			} else {
				result.TotalFailed++
			}
			mu.Unlock()

			if onComplete != nil {
				onComplete()
			}
		}(s)
	}

	wg.Wait()

	// Deduplicate findings by CheckID + Location.
	result.Findings = deduplicateFindings(result.Findings)

	// Sort findings: critical first, then by severity descending.
	sort.Slice(result.Findings, func(i, j int) bool {
		if result.Findings[i].Severity != result.Findings[j].Severity {
			return result.Findings[i].Severity > result.Findings[j].Severity
		}
		if result.Findings[i].CheckID != result.Findings[j].CheckID {
			return result.Findings[i].CheckID < result.Findings[j].CheckID
		}
		return result.Findings[i].Location.String() < result.Findings[j].Location.String()
	})

	result.ScanMeta.EndTime = time.Now()
	result.ScanMeta.Duration = result.ScanMeta.EndTime.Sub(result.ScanMeta.StartTime)

	return result
}

// deduplicateFindings removes duplicate findings based on CheckID and Location.
func deduplicateFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return findings
	}
	type key struct {
		checkID string
		loc     string
	}
	seen := make(map[key]bool, len(findings))
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		k := key{checkID: f.CheckID, loc: f.Location.String()}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, f)
	}
	return out
}
