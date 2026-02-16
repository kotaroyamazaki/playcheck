package datasafety

import (
	"encoding/xml"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
	"github.com/kotaroyamazaki/playcheck/pkg/utils"
)

// checkPrivacyPolicy checks for privacy policy URL presence in both
// AndroidManifest.xml and strings.xml resource files.
func checkPrivacyPolicy(projectDir string, manifests []string) []preflight.Finding {
	var findings []preflight.Finding

	manifestHasPolicy := checkManifestPrivacyPolicy(manifests, projectDir)
	stringsHasPolicy := checkStringsPrivacyPolicy(projectDir)

	if !manifestHasPolicy && !stringsHasPolicy {
		// Determine the best location to report.
		loc := preflight.Location{File: "AndroidManifest.xml"}
		if len(manifests) > 0 {
			relPath, _ := filepath.Rel(projectDir, manifests[0])
			loc.File = relPath
		}
		findings = append(findings, preflight.Finding{
			CheckID:     "PDS001",
			Title:       "Privacy policy URL not found",
			Description: "No privacy policy URL detected in AndroidManifest.xml or string resources. Apps that collect personal data or handle sensitive permissions must include a privacy policy.",
			Severity:    preflight.SeverityError,
			Location:    loc,
			Suggestion:  "Add a privacy policy URL to your AndroidManifest.xml via a <meta-data> tag or include it in your string resources.",
		})
	}

	return findings
}

// privacyURLPatterns matches common privacy policy URL patterns.
var privacyURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)privacy.?policy`),
	regexp.MustCompile(`(?i)https?://[^\s"<>]+/privacy`),
	regexp.MustCompile(`(?i)https?://[^\s"<>]+/legal`),
	regexp.MustCompile(`(?i)privacy_policy_url`),
	regexp.MustCompile(`(?i)privacyPolicyUrl`),
}

// checkManifestPrivacyPolicy checks AndroidManifest.xml files for privacy policy references.
func checkManifestPrivacyPolicy(manifests []string, projectDir string) bool {
	for _, m := range manifests {
		data, err := utils.ReadFileWithLimit(m)
		if err != nil {
			continue
		}
		content := string(data)
		for _, p := range privacyURLPatterns {
			if p.MatchString(content) {
				return true
			}
		}
	}
	return false
}

// stringsXMLResource represents a parsed Android strings.xml <resources> element.
type stringsXMLResource struct {
	XMLName xml.Name          `xml:"resources"`
	Strings []stringsXMLEntry `xml:"string"`
}

type stringsXMLEntry struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// checkStringsPrivacyPolicy scans res/values/strings.xml files for privacy policy URLs.
func checkStringsPrivacyPolicy(projectDir string) bool {
	xmlFiles, err := utils.WalkFiles(projectDir, utils.WithFilenames("strings.xml"))
	if err != nil {
		return false
	}

	for _, xf := range xmlFiles {
		// Only consider files under a "values" directory.
		dir := filepath.Base(filepath.Dir(xf))
		if !strings.HasPrefix(dir, "values") {
			continue
		}

		data, err := utils.ReadFileWithLimit(xf)
		if err != nil {
			continue
		}

		var res stringsXMLResource
		if err := xml.Unmarshal(data, &res); err != nil {
			// Fall back to raw text search.
			content := string(data)
			for _, p := range privacyURLPatterns {
				if p.MatchString(content) {
					return true
				}
			}
			continue
		}

		for _, entry := range res.Strings {
			nameLC := strings.ToLower(entry.Name)
			valueLC := strings.ToLower(entry.Value)
			if strings.Contains(nameLC, "privacy") || strings.Contains(nameLC, "policy") {
				return true
			}
			if strings.Contains(valueLC, "privacy") && strings.Contains(valueLC, "http") {
				return true
			}
		}
	}

	return false
}
