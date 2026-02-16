package manifest

import (
	"fmt"
	"strings"

	"github.com/yourusername/playcheck/internal/preflight"
)

// ManifestScanner implements preflight.Checker for manifest validation.
type ManifestScanner struct{}

func (s *ManifestScanner) ID() string          { return "manifest" }
func (s *ManifestScanner) Name() string        { return "AndroidManifest Validator" }
func (s *ManifestScanner) Description() string { return "Validates AndroidManifest.xml for Play Store compliance" }

func (s *ManifestScanner) Run(projectDir string) (*preflight.CheckResult, error) {
	m, err := FindAndParse(projectDir)
	if err != nil {
		return &preflight.CheckResult{
			CheckID: s.ID(),
			Passed:  false,
			Err:     err,
		}, err
	}

	v := NewValidator(m)
	findings := v.ValidateAll()

	return &preflight.CheckResult{
		CheckID:  s.ID(),
		Passed:   len(findings) == 0,
		Findings: findings,
	}, nil
}

// NewScanner creates a new ManifestScanner for use with the preflight runner.
func NewScanner() *ManifestScanner {
	return &ManifestScanner{}
}

// Validator runs compliance checks against a parsed AndroidManifest.
type Validator struct {
	manifest *AndroidManifest
}

// NewValidator creates a new manifest validator.
func NewValidator(m *AndroidManifest) *Validator {
	return &Validator{manifest: m}
}

// ValidateAll runs all manifest validation checks and returns findings.
func (v *Validator) ValidateAll() []preflight.Finding {
	var findings []preflight.Finding
	findings = append(findings, v.CheckTargetSDK()...)
	findings = append(findings, v.CheckDangerousPermissions()...)
	findings = append(findings, v.CheckExportedComponents()...)
	findings = append(findings, v.CheckLauncherActivity()...)
	findings = append(findings, v.CheckCleartextTraffic()...)
	return findings
}

// CheckTargetSDK validates that targetSdkVersion meets Play Store requirements.
func (v *Validator) CheckTargetSDK() []preflight.Finding {
	m := v.manifest
	if m.TargetSdkVersion == 0 {
		return []preflight.Finding{{
			CheckID:     RuleTargetSDK,
			Title:       "Missing targetSdkVersion",
			Description: "targetSdkVersion is not set in the manifest. Play Store requires targetSdkVersion >= 35.",
			Severity:    preflight.SeverityCritical,
			Location:    preflight.Location{File: m.filePath},
			Suggestion:  "Set targetSdkVersion to 35 or higher in your build.gradle or AndroidManifest.xml.",
		}}
	}

	if m.TargetSdkVersion < MinTargetSDKVersion {
		return []preflight.Finding{{
			CheckID:     RuleTargetSDK,
			Title:       fmt.Sprintf("targetSdkVersion %d is below required minimum", m.TargetSdkVersion),
			Description: fmt.Sprintf("targetSdkVersion is %d but Play Store requires >= %d for new apps and updates.", m.TargetSdkVersion, MinTargetSDKVersion),
			Severity:    preflight.SeverityCritical,
			Location:    preflight.Location{File: m.filePath},
			Suggestion:  fmt.Sprintf("Update targetSdkVersion to %d or higher.", MinTargetSDKVersion),
		}}
	}

	return nil
}

// CheckDangerousPermissions flags dangerous permissions that require disclosure.
func (v *Validator) CheckDangerousPermissions() []preflight.Finding {
	var findings []preflight.Finding
	for _, perm := range v.manifest.Permissions {
		info, isDangerous := dangerousPermissions[perm.Name]
		if !isDangerous {
			continue
		}
		findings = append(findings, preflight.Finding{
			CheckID:     info.RuleID,
			Title:       fmt.Sprintf("Dangerous permission: %s", shortPermName(perm.Name)),
			Description: info.Description,
			Severity:    severityForPermission(perm.Name),
			Location: preflight.Location{
				File: v.manifest.filePath,
				Line: perm.Line,
			},
			Suggestion: fmt.Sprintf("Ensure %s permission usage complies with Play Store policies. Add prominent disclosure if required.", info.Category),
		})
	}
	return findings
}

// CheckExportedComponents validates android:exported on components with intent filters.
// Since Android 12 (API 31), components with intent-filters must explicitly set android:exported.
func (v *Validator) CheckExportedComponents() []preflight.Finding {
	var findings []preflight.Finding

	checkComponent := func(name, kind string, exported *bool, filters []IntentFilter, line int) {
		if len(filters) == 0 {
			return
		}
		if exported == nil {
			findings = append(findings, preflight.Finding{
				CheckID:     RuleExportedComponent,
				Title:       fmt.Sprintf("%s missing android:exported", kind),
				Description: fmt.Sprintf("Component %q has intent-filters but does not set android:exported. This is required since Android 12 (API 31) and will cause installation failures.", name),
				Severity:    preflight.SeverityError,
				Location: preflight.Location{
					File: v.manifest.filePath,
					Line: line,
				},
				Suggestion: fmt.Sprintf("Add android:exported=\"true\" or android:exported=\"false\" to the <%s> element.", strings.ToLower(kind)),
			})
		} else if *exported {
			// Warn about explicitly exported components for security review.
			findings = append(findings, preflight.Finding{
				CheckID:     RuleComponentSecurity,
				Title:       fmt.Sprintf("Exported %s: %s", kind, shortComponentName(name)),
				Description: fmt.Sprintf("Component %q is exported and accessible to other apps. Ensure this is intentional and properly secured.", name),
				Severity:    preflight.SeverityInfo,
				Location: preflight.Location{
					File: v.manifest.filePath,
					Line: line,
				},
				Suggestion: "Review exported components to ensure they don't expose sensitive functionality.",
			})
		}
	}

	for _, a := range v.manifest.Activities {
		checkComponent(a.Name, "Activity", a.Exported, a.IntentFilters, a.Line)
	}
	for _, s := range v.manifest.Services {
		checkComponent(s.Name, "Service", s.Exported, s.IntentFilters, s.Line)
	}
	for _, r := range v.manifest.Receivers {
		checkComponent(r.Name, "Receiver", r.Exported, r.IntentFilters, r.Line)
	}
	for _, p := range v.manifest.Providers {
		checkComponent(p.Name, "Provider", p.Exported, p.IntentFilters, p.Line)
	}

	return findings
}

// CheckLauncherActivity checks that the manifest has a launcher activity.
func (v *Validator) CheckLauncherActivity() []preflight.Finding {
	if v.manifest.HasLauncherActivity() {
		return nil
	}
	return []preflight.Finding{{
		CheckID:     RuleLauncherActivity,
		Title:       "No launcher activity found",
		Description: "The manifest does not define a launcher activity (an activity with ACTION_MAIN and CATEGORY_LAUNCHER). The app may not appear in the launcher.",
		Severity:    preflight.SeverityWarning,
		Location:    preflight.Location{File: v.manifest.filePath},
		Suggestion:  "Add an intent-filter with action MAIN and category LAUNCHER to your main activity.",
	}}
}

// CheckCleartextTraffic checks the android:usesCleartextTraffic setting.
func (v *Validator) CheckCleartextTraffic() []preflight.Finding {
	if !v.manifest.HasCleartext {
		// Not explicitly set; default depends on target SDK.
		// For targetSdkVersion >= 28, default is false (secure).
		if v.manifest.TargetSdkVersion > 0 && v.manifest.TargetSdkVersion < 28 {
			return []preflight.Finding{{
				CheckID:     RuleCleartextTraffic,
				Title:       "Cleartext traffic may be enabled by default",
				Description: "With targetSdkVersion < 28, cleartext traffic is allowed by default. Consider setting android:usesCleartextTraffic=\"false\".",
				Severity:    preflight.SeverityWarning,
				Location:    preflight.Location{File: v.manifest.filePath},
				Suggestion:  "Set android:usesCleartextTraffic=\"false\" in the <application> element or use a Network Security Config.",
			}}
		}
		return nil
	}

	if v.manifest.UsesCleartext {
		return []preflight.Finding{{
			CheckID:     RuleCleartextTraffic,
			Title:       "Cleartext traffic explicitly enabled",
			Description: "android:usesCleartextTraffic is set to true, allowing unencrypted HTTP connections. This is a security risk and may trigger Play Store warnings.",
			Severity:    preflight.SeverityError,
			Location:    preflight.Location{File: v.manifest.filePath},
			Suggestion:  "Set android:usesCleartextTraffic=\"false\" and use HTTPS for all network communication.",
		}}
	}

	return nil
}

// shortPermName returns a human-friendly short permission name.
func shortPermName(fullName string) string {
	parts := splitLast(fullName, ".")
	if parts != "" {
		return parts
	}
	return fullName
}

// shortComponentName returns a human-friendly short component name.
func shortComponentName(fullName string) string {
	parts := splitLast(fullName, ".")
	if parts != "" {
		return parts
	}
	return fullName
}

func splitLast(s, sep string) string {
	idx := lastIndex(s, sep)
	if idx < 0 {
		return ""
	}
	return s[idx+len(sep):]
}

func lastIndex(s, sub string) int {
	for i := len(s) - len(sub); i >= 0; i-- {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
