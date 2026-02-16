package codescan

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
	"github.com/kotaroyamazaki/playcheck/pkg/utils"
)

// Scanner scans Kotlin and Java source files for Play Store compliance issues.
type Scanner struct {
	compiled []compiledRule
}

// NewScanner creates a Scanner with the default rule set pre-compiled.
func NewScanner() *Scanner {
	return &Scanner{
		compiled: compileRules(codeRules),
	}
}

// ID implements preflight.Checker.
func (s *Scanner) ID() string { return "code-scan" }

// Name implements preflight.Checker.
func (s *Scanner) Name() string { return "Code Scanner" }

// Description implements preflight.Checker.
func (s *Scanner) Description() string {
	return "Scans Kotlin and Java source files for Play Store compliance issues"
}

// maxSnippetLen is the maximum length of a code snippet included in findings.
const maxSnippetLen = 120

// maxConcurrency limits the number of files scanned concurrently.
const maxConcurrency = 8

// Run implements preflight.Checker. It walks the project directory for .kt and
// .java files, scans them concurrently, and returns aggregated findings.
func (s *Scanner) Run(projectDir string) (*preflight.CheckResult, error) {
	files, err := utils.WalkFiles(projectDir,
		utils.WithExtensions(".kt", ".java"),
	)
	if err != nil {
		return nil, err
	}

	result := &preflight.CheckResult{
		CheckID: s.ID(),
		Passed:  true,
	}

	if len(files) == 0 {
		return result, nil
	}

	// Scan files concurrently with a semaphore to limit parallelism.
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, maxConcurrency)
		findings []preflight.Finding
	)

	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }() // release

			ff := s.scanFile(path, projectDir)
			if len(ff) > 0 {
				mu.Lock()
				findings = append(findings, ff...)
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()

	result.Findings = findings
	result.Passed = len(findings) == 0

	return result, nil
}

// scanFile scans a single file against all compiled rules and returns findings.
func (s *Scanner) scanFile(filePath, projectDir string) []preflight.Finding {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	relPath, err := filepath.Rel(projectDir, filePath)
	if err != nil {
		relPath = filePath
	}

	var findings []preflight.Finding

	// Track which rule IDs have already matched in this file to avoid
	// excessive duplicate findings from the same rule.
	matched := make(map[string]int) // rule ID -> count
	const maxMatchesPerRule = 3

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comment-only lines to reduce false positives.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		for i := range s.compiled {
			cr := &s.compiled[i]

			if matched[cr.rule.ID] >= maxMatchesPerRule {
				continue
			}

			for _, re := range cr.patterns {
				if re.MatchString(line) {
					matched[cr.rule.ID]++

					snippet := strings.TrimSpace(line)
					if len(snippet) > maxSnippetLen {
						snippet = snippet[:maxSnippetLen] + "..."
					}

					findings = append(findings, preflight.Finding{
						CheckID:     cr.rule.ID,
						Title:       cr.rule.Title,
						Description: cr.rule.Description + "\n  Code: " + snippet,
						Severity:    cr.rule.Severity,
						Location: preflight.Location{
							File: relPath,
							Line: lineNum,
						},
						Suggestion: cr.rule.Suggestion,
					})
					break // one match per rule per line is enough
				}
			}
		}
	}

	return findings
}
