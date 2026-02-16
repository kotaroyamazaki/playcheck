# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-02-16

### Added
- AndroidManifest.xml validation (SDK version, permissions, components)
- Kotlin/Java code scanning (14 detection rules)
- Data safety compliance checking
- 31+ policy rules covering Google Play Store requirements
- Colored terminal output with severity indicators
- JSON output format for CI/CD integration
- Severity filtering (critical, warning, info)
- Progress bar during scanning
- 109 comprehensive tests with 85-95% coverage
- Sample test applications (violating and clean)
- Concurrent file scanning with goroutines (8 parallel workers)
- Regex pattern caching for performance
- Embedded policy database with go:embed
- Proper exit codes (0 = pass, 1 = critical issues)

### Detection Capabilities
- **Dangerous Permissions**: SMS, call log, location (foreground/background), camera, contacts, storage, calendar, microphone, body sensors, accessibility, VPN, exact alarms, query all packages
- **Privacy & Data Safety**: Privacy policy detection, user consent validation, account deletion requirements, permission disclosure mapping
- **SDK Compliance**: Target SDK version 35 enforcement, deprecated API detection
- **Security Issues**: Cleartext traffic, hardcoded secrets, exported components, weak cryptography
- **Third-party SDKs**: Firebase Analytics, Firebase Crashlytics, AdMob, Facebook SDK, tracking SDKs (Adjust, AppsFlyer, Amplitude, Mixpanel, etc.)

### Documentation
- Comprehensive README with usage examples
- Japanese README (README.ja.md)
- Framework support guide (FRAMEWORKS.md)
- Contributing guidelines (CONTRIBUTING.md)
- MIT License
- GitHub issue templates (bug report, feature request)
- GitHub Actions workflows (CI, Release)

---

[Unreleased]: https://github.com/kotaroyamazaki/playcheck/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/kotaroyamazaki/playcheck/releases/tag/v1.0.0
