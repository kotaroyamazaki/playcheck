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
Runs three scanners — manifest validator, code scanner, and data safety checker — covering
dangerous permissions, privacy/data safety, SDK compliance, account management, manifest
validation, code patterns (HTTP, crypto, SDKs), and component security.

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

2. If not installed, install via one of these methods:

```bash
# Option A: Install via Go (preferred)
go install github.com/kotaroyamazaki/playcheck/cmd/playcheck@latest
```

If Go is not available, run the install script bundled with this skill.
Locate the skill directory (where this SKILL.md is installed) and run:

```bash
bash <skill-directory>/scripts/install-playcheck.sh
```

The skill directory is typically `~/.claude/skills/playcheck/`,
`.claude/skills/playcheck/`, or `.agents/skills/playcheck/`
depending on the agent and installation method.

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

playcheck uses four severity levels. In the JSON summary, `critical` is the combined count
of CRITICAL and ERROR findings. The `--severity critical` flag shows both CRITICAL and ERROR.

- **CRITICAL**: Must fix before Play Store submission. App will be rejected.
  Examples: outdated targetSdkVersion, SMS permission without default handler, SMS API in code
- **ERROR**: Likely to cause rejection or policy warning.
  Examples: cleartext traffic, missing android:exported, HTTP URLs, weak cryptography, missing privacy policy, no runtime permission requests
- **WARNING**: Should review and potentially fix.
  Examples: unused permissions, missing Data Safety disclosures, SDK data collection, WebView JavaScript, no launcher activity
- **INFO**: Informational, no action required.
  Examples: exported component review, privacy policy URL found, account deletion detected

### JSON Output Structure

When using `--format json`, the output contains:

```json
{
  "timestamp": "2026-01-15T10:30:00Z",
  "project_path": "/path/to/android/project",
  "summary": {
    "total_checks": 31,
    "passed": 26,
    "failed": 5,
    "critical": 2,
    "warning": 2,
    "info": 1,
    "duration": "1.234s"
  },
  "findings": [
    {
      "check_id": "DP001",
      "severity": "CRITICAL",
      "title": "SMS Permission Usage",
      "description": "App requests SMS permission ...",
      "location": "AndroidManifest.xml:15",
      "suggestion": "Remove SMS permissions unless ..."
    }
  ]
}
```

Note: `location` and `suggestion` fields are omitted when empty.

## Help Fix Issues

For each finding, provide remediation guidance based on the category.
Read `references/policy-rules.md` for the complete rule-by-rule reference with fix instructions and policy links.

### By Scanner

**Manifest Scanner** (SDK001, DP001-DP007, MV001/MV002/MV004, MC001):
- Validates targetSdkVersion >= 35, flags dangerous permissions, checks android:exported, verifies launcher activity, detects cleartext traffic.

**Code Scanner** (CS001-CS014):
- Scans Kotlin/Java files for HTTP URLs, weak crypto, SMS API usage, SDK data collection patterns (Firebase, AdMob, Facebook, trackers), WebView JavaScript, account creation/deletion, location/camera API usage.

**Data Safety Checker** (PDS001-PDS004, AD001, SDK001, SDK004, DP006):
- Verifies privacy policy presence, checks permission-to-Data-Safety-form alignment, detects data collection without consent, validates runtime permission requests, checks account deletion requirement, flags unused manifest permissions, audits third-party SDK disclosures.

## Validation

After fixing issues, re-run the scan to verify:

```bash
# Re-scan and confirm no critical issues
playcheck scan <project-path> --severity critical

# Compare before/after with JSON output
playcheck scan <project-path> --format json --output report-after.json
```

A successful fix results in exit code `0` and zero critical findings in the summary.
