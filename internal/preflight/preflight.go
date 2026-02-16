package preflight

import (
	"sort"
	"sync"
	"time"
)

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

// Runner orchestrates compliance checkers and aggregates results.
type Runner struct {
	checkers []Checker
}

// RegisterScanner adds a checker to the runner.
func (r *Runner) RegisterScanner(s Checker) {
	r.checkers = append(r.checkers, s)
}

// Checkers returns the list of registered checkers.
func (r *Runner) Checkers() []Checker {
	return r.checkers
}

// Run executes all registered checkers against the project directory.
// The onComplete callback is invoked after each checker finishes, which
// is used by the CLI to advance the progress bar.
// Checkers run concurrently for better performance.
func (r *Runner) Run(projectDir string, onComplete func()) *ScanResult {
	startTime := time.Now()

	result := &ScanResult{
		ByScanner: make(map[string]*CheckResult, len(r.checkers)),
		ScanMeta: ScanMetadata{
			ProjectPath: projectDir,
			StartTime:   startTime,
		},
	}

	for _, c := range r.checkers {
		result.ScanMeta.ScannerIDs = append(result.ScanMeta.ScannerIDs, c.ID())
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, c := range r.checkers {
		wg.Add(1)
		go func(checker Checker) {
			defer wg.Done()

			cr, err := checker.Run(projectDir)
			if cr == nil {
				cr = &CheckResult{
					CheckID: checker.ID(),
				}
			}
			if err != nil {
				cr.Err = err
			}

			mu.Lock()
			result.ByScanner[checker.ID()] = cr
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
		}(c)
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
