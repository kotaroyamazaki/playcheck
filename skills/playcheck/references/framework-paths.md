# Framework Scan Paths

Use this table to determine the correct directory to pass to `playcheck scan`.

| Framework | Scan Path | Notes |
|-----------|-----------|-------|
| Native Android (Kotlin/Java) | `.` (project root) | Full support: manifest, code, build config |
| Flutter | `./android` | Dart code not scanned |
| React Native | `./android` | JS/TS code not scanned |
| Cordova / Ionic | `./platforms/android` | Run `cordova build android` first |
| Xamarin / .NET MAUI | `./Platforms/Android` | C# code not scanned; manifest checks only |
| Unity | `./Temp/gradleOut` | Run Unity build first to generate Android project |
| NativeScript | `./platforms/android` | Full support for native plugins |
| Kotlin Multiplatform (KMM) | `./androidApp` | Scans the Android target module |

## How to Confirm the Right Path

Locate the AndroidManifest.xml to verify:

```bash
find <project-root> -name "AndroidManifest.xml" -maxdepth 5 2>/dev/null
```

The scan path should be the parent Android project directory containing the manifest.

## What Gets Scanned

Regardless of framework, playcheck analyzes:
- `AndroidManifest.xml` — permissions, components, attributes
- Kotlin (`.kt`) and Java (`.java`) files — code patterns
- `build.gradle` / `build.gradle.kts` — SDK versions, dependencies

Framework-specific code (Dart, JS/TS, C#, etc.) is **not** scanned because
Play Store policy violations are enforced at the Android platform level.
