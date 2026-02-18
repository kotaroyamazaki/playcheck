---
name: playcheck
description: >
  Scan Android projects for Google Play Store policy compliance issues before submission.
  Use when a user wants to check their Android app for Play Store violations, dangerous permissions,
  privacy/data safety gaps, security vulnerabilities, SDK compliance issues, or pre-submission audits.
  Triggers on: Play Store compliance, policy violations, app rejection, dangerous permissions,
  data safety, privacy policy, targetSdkVersion, exported components, pre-submission check.
---

# playcheck

Static analysis tool that scans Android projects for Google Play Store policy violations.
Checks 31+ rules across 8 categories: dangerous permissions, privacy/data safety, SDK compliance,
account management, manifest validation, security, code scanning, and monetization.

Works with any Android framework: native Android (Kotlin/Java), Flutter, React Native,
Cordova/Ionic, Xamarin/.NET MAUI, Unity, NativeScript, and KMM.

## When to Use

- User wants to check an Android app for Play Store compliance before submission
- User is debugging a Play Store rejection or policy warning
- User asks about dangerous permissions, data safety, or privacy policy requirements
- User wants a pre-submission compliance audit
- User needs to find security issues (cleartext traffic, hardcoded secrets, exported components)

## When NOT to Use

- iOS App Store checks (Android only)
- General code quality or linting (this is compliance-specific)
- Runtime testing, UI testing, or performance profiling

## Installation

1. Check if `playcheck` is already installed:

```bash
playcheck --version 2>/dev/null && echo "playcheck is installed" || echo "playcheck not found"
```

2. If not installed, check for Go and install accordingly:

```bash
# Option A: Install via Go (preferred)
go install github.com/kotaroyamazaki/playcheck/cmd/playcheck@latest

# Option B: Install pre-built binary (no Go required)
# Run the install script bundled with this skill
bash "$(dirname "$0")/scripts/install-playcheck.sh"
```

3. Verify installation:

```bash
playcheck --version
```

## Identify the Correct Scan Path

Different frameworks place Android project files in different directories.
Identify the framework and use the appropriate scan path.

Read `references/framework-paths.md` for the full reference table.

Quick decision tree:

- **Native Android (Kotlin/Java)**: scan the project root (`.`)
- **Flutter**: scan `./android`
- **React Native**: scan `./android`
- **Cordova/Ionic**: scan `./platforms/android` (build first)
- **Xamarin/.NET MAUI**: scan `./Platforms/Android`
- **Unity**: scan `./Temp/gradleOut` (build first)
- **KMM**: scan `./androidApp`
- **NativeScript**: scan `./platforms/android`

To confirm the correct path, locate the AndroidManifest.xml:

```bash
find <project-path> -name "AndroidManifest.xml" -maxdepth 5 2>/dev/null
```

## Run the Scan

Basic scan with colored terminal output:

```bash
playcheck scan <project-path>
```

JSON output for machine-readable results:

```bash
playcheck scan <project-path> --format json
```

Save report to a file:

```bash
playcheck scan <project-path> --format json --output report.json
```

Filter by minimum severity:

```bash
# Only show critical/error-level issues
playcheck scan <project-path> --severity critical

# Show warnings and above
playcheck scan <project-path> --severity warn
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `terminal` | Output format: `terminal` or `json` |
| `--severity` | `-s` | `all` | Minimum severity: `all`, `critical`, `warn`, `info` |
| `--output` | `-o` | (stdout) | Write report to file |

### Exit Codes

- `0` — No critical or error-level issues found
- `1` — Critical issues detected (non-zero exit from the command)

## Interpret Results

### Severity Levels

- **CRITICAL**: Must fix before Play Store submission. App will be rejected.
  Examples: SMS permissions without being default handler, outdated targetSdkVersion, debuggable builds, hardcoded secrets, missing account deletion
- **ERROR**: Likely to cause rejection or policy warning.
  Examples: cleartext traffic, exported components without protection, weak cryptography, non-Play billing for digital goods
- **WARNING**: Should review and potentially fix.
  Examples: unused permissions, missing Data Safety disclosures, WebView JavaScript enabled, missing backup rules
- **INFO**: Informational, no action required.
  Examples: deep link intent filter suggestions

### JSON Output Structure

When using `--format json`, the output contains:

```json
{
  "metadata": {
    "project_path": "...",
    "start_time": "...",
    "duration": "..."
  },
  "summary": {
    "total": 5,
    "critical": 2,
    "error": 1,
    "warning": 2,
    "info": 0
  },
  "findings": [
    {
      "check_id": "DP001",
      "severity": "CRITICAL",
      "title": "SMS Permission Usage",
      "description": "...",
      "location": "AndroidManifest.xml",
      "suggestion": "..."
    }
  ]
}
```

## Help Fix Issues

For each finding, provide remediation guidance based on the category.
Read `references/policy-rules.md` for the complete rule-by-rule reference with fix instructions and policy links.

### Category Overview

1. **Dangerous Permissions (DP001-DP010)**: Remove unused permissions. Submit Permissions Declaration Form for justified use. Replace restricted permissions with scoped alternatives.
2. **Privacy & Data Safety (PDS001-PDS004)**: Add privacy policy URL. Implement data deletion. Update Data Safety form in Play Console. Add prominent disclosure before data collection.
3. **SDK Compliance (SDK001-SDK004)**: Update targetSdkVersion to 35+. Update Play Core library to 1.10.0+. Integrate consent management for ads.
4. **Account Management (AD001-AD002)**: Implement account deletion flow. Disclose auth data in Data Safety section.
5. **Manifest Validation (MV001-MV005)**: Add app icon. Remove debuggable flag. Add versionCode. Configure backup rules. Add BROWSABLE category to deep links.
6. **Security (MS001-MS004)**: Use HTTPS everywhere. Move secrets to environment variables. Protect exported components with permissions. Avoid addJavascriptInterface with untrusted content.
7. **Monetization (MP002)**: Use Google Play Billing for digital goods and subscriptions.
8. **Content Policy (MC001)**: Filter WebView content. Ensure content rating accuracy.

## Validation

After fixing issues, re-run the scan to verify:

```bash
# Re-scan and confirm no critical issues
playcheck scan <project-path> --severity critical

# Compare before/after with JSON output
playcheck scan <project-path> --format json --output report-after.json
```

A successful fix results in exit code `0` and zero critical findings in the summary.
