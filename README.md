# playcheck

A CLI tool that scans **any Android project** for Google Play Store compliance issues before submission. Works with native Android (Kotlin/Java), Flutter, React Native, and other frameworks. Catches policy violations, dangerous permissions, security issues, and data safety gaps early in development.

> 💡 **Inspired by [Greenlight](https://github.com/RevylAI/greenlight)** - bringing the same compliance scanning capabilities from iOS to Android developers.

English | [日本語](README.ja.md)

## Features

- **Manifest validation** - SDK version checks, dangerous permissions, exported components, cleartext traffic
- **Code scanning** - Detects HTTP URLs, SMS API usage, advertising IDs, weak cryptography, third-party SDK data collection
- **Data safety compliance** - Privacy policy detection, account deletion requirements, permission disclosure checks, user consent validation
- **31+ policy rules** - Covers dangerous permissions, privacy, SDK compliance, account management, security, and more
- **Multiple output formats** - Colored terminal output and JSON for CI/CD integration
- **Severity filtering** - Filter findings by severity level (critical, warning, info)

## Installation

```bash
go install github.com/kotaroyamazaki/playcheck/cmd/playcheck@latest
```

Or build from source:

```bash
git clone https://github.com/kotaroyamazaki/playcheck.git
cd playcheck
go build -o playcheck ./cmd/playcheck
```

## Usage

### Basic scan

```bash
playcheck scan /path/to/android/project
```

### Output formats

```bash
# Colored terminal output (default)
playcheck scan ./my-app

# JSON output for CI/CD pipelines
playcheck scan ./my-app --format json

# Write report to file
playcheck scan ./my-app --format json --output report.json
```

### Severity filtering

```bash
# Show all findings (default)
playcheck scan ./my-app --severity all

# Show only critical and error findings
playcheck scan ./my-app --severity critical

# Show warnings and above
playcheck scan ./my-app --severity warn
```

### Exit codes

- `0` - No critical or error-level issues found
- `1` - Critical or error-level issues detected that must be resolved before Play Store submission

## Supported Rules

### Dangerous Permissions (DP001-DP010)

| ID | Rule | Severity |
|----|------|----------|
| DP001 | SMS Permission Usage | CRITICAL |
| DP002 | Call Log Permission Usage | CRITICAL |
| DP003 | Location in Background | CRITICAL |
| DP004 | Camera Permission Without Usage | WARNING |
| DP005 | Storage Permission (Broad Access) | ERROR |
| DP006 | Exact Alarm Permission | WARNING |
| DP007 | Query All Packages | WARNING |
| DP008 | Accessibility Service Permission | CRITICAL |
| DP009 | VPN Service Permission | ERROR |
| DP010 | Foreground Service Type Missing | ERROR |

### Privacy & Data Safety (PDS001-PDS004)

| ID | Rule | Severity |
|----|------|----------|
| PDS001 | Missing Privacy Policy | CRITICAL |
| PDS002 | Data Collection Without Disclosure | ERROR |
| PDS003 | Data Safety Section Mismatch | ERROR |
| PDS004 | Missing Data Deletion Mechanism | WARNING |

### SDK Compliance (SDK001-SDK004)

| ID | Rule | Severity |
|----|------|----------|
| SDK001 | Outdated Target SDK Version | CRITICAL |
| SDK002 | Missing Play Core Library Update | WARNING |
| SDK003 | Missing Ads SDK Consent Integration | ERROR |
| SDK004 | Deprecated API Usage | WARNING |

### Account Management (AD001-AD002)

| ID | Rule | Severity |
|----|------|----------|
| AD001 | Missing Account Deletion Option | CRITICAL |
| AD002 | Login Without Data Safety Disclosure | WARNING |

### Manifest Validation (MV001-MV005)

| ID | Rule | Severity |
|----|------|----------|
| MV001 | Missing App Icon | ERROR |
| MV002 | Debuggable Build | CRITICAL |
| MV003 | Missing Version Code | ERROR |
| MV004 | Backup Rules Missing | WARNING |
| MV005 | Intent Filter Without BROWSABLE | INFO |

### Security (MS001-MS004)

| ID | Rule | Severity |
|----|------|----------|
| MS001 | Insecure Network Communication | ERROR |
| MS002 | Hardcoded Secrets or API Keys | CRITICAL |
| MS003 | Exported Components Without Protection | ERROR |
| MS004 | WebView JavaScript Interface Vulnerability | ERROR |

### Code Scanning (CS001-CS014)

| ID | Rule | Severity |
|----|------|----------|
| CS001 | Unencrypted HTTP URL | ERROR |
| CS002 | Privacy Policy URL Found | INFO |
| CS003 | Firebase Analytics SDK Usage | WARNING |
| CS004 | AdMob SDK Usage | WARNING |
| CS005 | Advertising ID Usage | WARNING |
| CS006 | Account Creation Pattern | WARNING |
| CS007 | Account Deletion Pattern | INFO |
| CS008 | SMS API Usage | CRITICAL |
| CS009 | Location API Usage | WARNING |
| CS010 | Camera API Usage | WARNING |
| CS011 | Weak Cryptography Usage | ERROR |
| CS012 | WebView JavaScript Enabled | WARNING |
| CS013 | Facebook SDK Usage | WARNING |
| CS014 | Third-Party Tracking SDK | WARNING |

### Monetization (MP002)

| ID | Rule | Severity |
|----|------|----------|
| MP002 | Non-Play Billing for Digital Goods | CRITICAL |

### Content Policy (MC001)

| ID | Rule | Severity |
|----|------|----------|
| MC001 | Content Rating Missing | WARNING |

## Output Examples

### Terminal output

```
=== Play Store Compliance Report ===
Project: /path/to/android/project
Duration: 45ms

CRITICAL (3)
  [CRITICAL] targetSdkVersion 33 is below required minimum
         AndroidManifest.xml
         Suggestion: Update targetSdkVersion to 35 or higher.

  [CRITICAL] Dangerous permission: SEND_SMS
         AndroidManifest.xml:6
         Suggestion: Ensure SMS permission usage complies with Play Store policies.

  [CRITICAL] SMS API usage detected in code
         app/src/main/java/com/example/Main.java:15
         Suggestion: Remove direct SMS API usage unless your app is a default SMS handler.

WARNING (5)
  [WARNING] Dangerous permission: CAMERA
         AndroidManifest.xml:10
         Suggestion: Ensure Camera permission usage complies with Play Store policies.

  [WARNING] Firebase Analytics SDK usage detected
         app/src/main/java/com/example/Main.java:22
         Suggestion: Disclose Firebase Analytics data collection in your Data Safety form.

--------------------------------------------------
Checks run: 3 | Passed: 0 | Critical: 3 | Warnings: 5 | Info: 2

RESULT: FAIL - Critical issues must be resolved before submission.
```

### JSON output

```json
{
  "timestamp": "2026-02-16T00:00:00Z",
  "project_path": "/path/to/android/project",
  "summary": {
    "total_checks": 3,
    "passed": 0,
    "failed": 3,
    "critical": 3,
    "warning": 5,
    "info": 2,
    "duration": "45ms"
  },
  "findings": [
    {
      "check_id": "SDK001",
      "severity": "CRITICAL",
      "title": "targetSdkVersion 33 is below required minimum",
      "description": "targetSdkVersion is 33 but Play Store requires >= 35.",
      "location": "AndroidManifest.xml",
      "suggestion": "Update targetSdkVersion to 35 or higher."
    }
  ]
}
```

## Project Structure

```
cmd/playcheck/          CLI entry point
internal/
  cli/                  Cobra command definitions
  codescan/             Kotlin/Java source code scanner
  datasafety/           Data safety and privacy compliance checker
  manifest/             AndroidManifest.xml parser and validator
  policies/             Embedded policy rule database (31+ rules)
  preflight/            Core types, runner, and report formatting
pkg/utils/              File walking utilities
testdata/
  sample-apps/
    violating-app/      Sample app with multiple policy violations
    clean-app/          Sample compliant app for testing false positives
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run integration tests only
go test -run TestIntegration ./...

# Run tests for a specific package
go test ./internal/manifest/...
go test ./internal/codescan/...
go test ./internal/datasafety/...
go test ./internal/policies/...
go test ./internal/preflight/...
```

## Acknowledgments

This project was inspired by [Greenlight](https://github.com/RevylAI/greenlight), an excellent App Store compliance scanner for iOS apps. Just as Greenlight helps iOS developers catch App Store policy violations early, playcheck aims to provide the same value for Android developers targeting Google Play Store.

This project uses the following open-source libraries:

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework
- [Color](https://github.com/fatih/color) - Terminal color output
- [Progressbar](https://github.com/schollz/progressbar) - Progress bar display

## License

MIT - See [LICENSE](LICENSE) for details.
