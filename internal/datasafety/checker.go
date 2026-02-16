package datasafety

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
	"github.com/kotaroyamazaki/playcheck/pkg/utils"
)

// Checker validates data safety compliance for Google Play Store requirements.
type Checker struct{}

// NewChecker creates a new data safety Checker.
func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) ID() string          { return "DATA_SAFETY" }
func (c *Checker) Name() string        { return "Data Safety Compliance" }
func (c *Checker) Description() string { return "Checks data safety declarations, privacy policies, and disclosure requirements" }

// Run executes all data safety compliance checks on the given project directory.
func (c *Checker) Run(projectDir string) (*preflight.CheckResult, error) {
	result := &preflight.CheckResult{
		CheckID: c.ID(),
		Passed:  true,
	}

	manifests, err := utils.FindAndroidManifests(projectDir)
	if err != nil {
		result.Err = err
		return result, nil
	}

	// Parse manifest permissions and metadata.
	manifestData := parseManifests(manifests)

	// Check privacy policy presence.
	privacyFindings := checkPrivacyPolicy(projectDir, manifests)
	result.Findings = append(result.Findings, privacyFindings...)

	// Check permission disclosures.
	permFindings := checkPermissionDisclosures(manifestData, projectDir)
	result.Findings = append(result.Findings, permFindings...)

	// Check third-party SDK disclosures.
	sdkFindings := checkSDKDisclosures(projectDir)
	result.Findings = append(result.Findings, sdkFindings...)

	// Check account deletion requirement.
	acctFindings := checkAccountDeletion(projectDir)
	result.Findings = append(result.Findings, acctFindings...)

	// Check user consent patterns.
	consentFindings := checkUserConsent(projectDir)
	result.Findings = append(result.Findings, consentFindings...)

	// Cross-reference manifest permissions with actual code usage.
	crossRefFindings := crossReferencePermissionsWithCode(manifestData, projectDir)
	result.Findings = append(result.Findings, crossRefFindings...)

	for _, f := range result.Findings {
		if f.Severity >= preflight.SeverityError {
			result.Passed = false
			break
		}
	}

	return result, nil
}

// manifestInfo holds parsed data from AndroidManifest.xml files.
type manifestInfo struct {
	FilePath    string
	Permissions []string
	HasMeta     map[string]bool
}

var permissionRe = regexp.MustCompile(`<uses-permission\s+android:name="([^"]+)"`)
var metadataNameRe = regexp.MustCompile(`<meta-data\s+android:name="([^"]+)"`)

// Account creation/deletion detection patterns.
var createAccountPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)createUser|signUp|registerUser|createAccount|registerAccount`),
	regexp.MustCompile(`(?i)FirebaseAuth\s*\.\s*getInstance\(\)\s*\.\s*createUser`),
	regexp.MustCompile(`(?i)\.createUserWithEmailAndPassword\(`),
}

var deleteAccountPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)deleteUser|deleteAccount|removeAccount|deactivateAccount`),
	regexp.MustCompile(`(?i)\.delete\(\)\s*//.*account`),
	regexp.MustCompile(`(?i)FirebaseAuth.*\.currentUser.*\.delete\(`),
	regexp.MustCompile(`(?i)account.?delet`),
}

// Data collection and consent detection patterns.
var dataCollectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`getDeviceId\(`),
	regexp.MustCompile(`getAdvertisingIdInfo\(`),
	regexp.MustCompile(`ANDROID_ID`),
	regexp.MustCompile(`getAccounts\(`),
	regexp.MustCompile(`getLastKnownLocation\(`),
	regexp.MustCompile(`requestLocationUpdates\(`),
}

var consentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)consent`),
	regexp.MustCompile(`(?i)user.?agre`),
	regexp.MustCompile(`(?i)opt.?in`),
	regexp.MustCompile(`(?i)permission.?dialog`),
	regexp.MustCompile(`(?i)privacy.?accept`),
}

func parseManifests(paths []string) []manifestInfo {
	var results []manifestInfo
	for _, p := range paths {
		info := manifestInfo{
			FilePath: p,
			HasMeta:  make(map[string]bool),
		}
		data, err := utils.ReadFileWithLimit(p)
		if err != nil {
			continue
		}
		content := string(data)
		for _, m := range permissionRe.FindAllStringSubmatch(content, -1) {
			info.Permissions = append(info.Permissions, m[1])
		}
		for _, m := range metadataNameRe.FindAllStringSubmatch(content, -1) {
			info.HasMeta[m[1]] = true
		}
		results = append(results, info)
	}
	return results
}

// checkSDKDisclosures scans Gradle files for third-party SDKs that require data safety disclosures.
func checkSDKDisclosures(projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	gradleFiles, err := utils.FindGradleFiles(projectDir)
	if err != nil {
		return findings
	}

	for _, gf := range gradleFiles {
		data, err := utils.ReadFileWithLimit(gf)
		if err != nil {
			continue
		}
		content := string(data)
		relPath, _ := filepath.Rel(projectDir, gf)

		for _, sdk := range thirdPartySDKs {
			for _, dep := range sdk.Dependencies {
				if strings.Contains(content, dep) {
					line := findLineNumber(content, dep)
					findings = append(findings, preflight.Finding{
						CheckID:     "SDK001",
						Title:       "Third-party SDK requires data safety disclosure",
						Description: sdk.Name + " SDK detected (" + dep + "). " + sdk.DisclosureNote,
						Severity:    preflight.SeverityWarning,
						Location:    preflight.Location{File: relPath, Line: line},
						Suggestion:  "Declare data collection by " + sdk.Name + " in your Play Console Data Safety form. " + sdk.DisclosureNote,
					})
				}
			}
		}
	}

	return findings
}

// checkAccountDeletion checks if apps that create accounts also provide account deletion.
func checkAccountDeletion(projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	codeFiles, err := utils.WalkFiles(projectDir, utils.WithExtensions(".kt", ".java"))
	if err != nil {
		return findings
	}

	var hasCreateAccount bool
	var hasDeleteAccount bool
	var createAccountLoc preflight.Location

	for _, cf := range codeFiles {
		data, err := utils.ReadFileWithLimit(cf)
		if err != nil {
			continue
		}
		content := string(data)
		relPath, _ := filepath.Rel(projectDir, cf)

		if !hasCreateAccount {
			for _, p := range createAccountPatterns {
				loc := p.FindStringIndex(content)
				if loc != nil {
					hasCreateAccount = true
					line := findLineNumber(content, content[loc[0]:loc[1]])
					createAccountLoc = preflight.Location{File: relPath, Line: line}
					break
				}
			}
		}

		for _, p := range deleteAccountPatterns {
			if p.MatchString(content) {
				hasDeleteAccount = true
				break
			}
		}

		if hasCreateAccount && hasDeleteAccount {
			break
		}
	}

	if hasCreateAccount && !hasDeleteAccount {
		findings = append(findings, preflight.Finding{
			CheckID:     "AD001",
			Title:       "Account deletion not found",
			Description: "App creates user accounts but no account deletion functionality was detected. Google Play requires apps that allow account creation to also provide account deletion.",
			Severity:    preflight.SeverityError,
			Location:    createAccountLoc,
			Suggestion:  "Implement account deletion functionality. See https://support.google.com/googleplay/android-developer/answer/13327111",
		})
	}

	return findings
}

// findLineNumber returns the 1-based line number of the first occurrence of substr in content.
func findLineNumber(content, substr string) int {
	idx := strings.Index(content, substr)
	if idx < 0 {
		return 0
	}
	return strings.Count(content[:idx], "\n") + 1
}

// checkUserConsent scans code files for data collection without consent patterns.
func checkUserConsent(projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	codeFiles, err := utils.WalkFiles(projectDir, utils.WithExtensions(".kt", ".java"))
	if err != nil {
		return findings
	}

	for _, cf := range codeFiles {
		data, err := utils.ReadFileWithLimit(cf)
		if err != nil {
			continue
		}
		content := string(data)
		relPath, _ := filepath.Rel(projectDir, cf)

		for _, dp := range dataCollectionPatterns {
			loc := dp.FindStringIndex(content)
			if loc == nil {
				continue
			}

			// Check if the same file has consent-related code.
			hasConsent := false
			for _, cp := range consentPatterns {
				if cp.MatchString(content) {
					hasConsent = true
					break
				}
			}

			if !hasConsent {
				line := findLineNumber(content, content[loc[0]:loc[1]])
				findings = append(findings, preflight.Finding{
					CheckID:     "PDS003",
					Title:       "Data collection without apparent consent",
					Description: "Data collection API (" + content[loc[0]:loc[1]] + ") detected without consent-related code in the same file.",
					Severity:    preflight.SeverityWarning,
					Location:    preflight.Location{File: relPath, Line: line},
					Suggestion:  "Ensure user consent is obtained before collecting personal data. Consider implementing a consent dialog.",
				})
				break // One finding per file is enough
			}
		}
	}

	return findings
}

