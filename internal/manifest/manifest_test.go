package manifest

import (
	"os"
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

const sampleManifest = `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.testapp"
    android:versionCode="1"
    android:versionName="1.0">

    <uses-sdk
        android:minSdkVersion="24"
        android:targetSdkVersion="35" />

    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
    <uses-permission android:name="android.permission.CAMERA" />

    <application
        android:usesCleartextTraffic="false">

        <activity
            android:name=".MainActivity"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>

        <activity
            android:name=".SecondActivity"
            android:exported="false">
        </activity>

        <service
            android:name=".MyService">
            <intent-filter>
                <action android:name="com.example.ACTION" />
            </intent-filter>
        </service>

        <receiver
            android:name=".MyReceiver"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.BOOT_COMPLETED" />
            </intent-filter>
        </receiver>

    </application>
</manifest>
`

func TestParseBasic(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if m.Package != "com.example.testapp" {
		t.Errorf("Package = %q, want %q", m.Package, "com.example.testapp")
	}
	if m.VersionCode != 1 {
		t.Errorf("VersionCode = %d, want 1", m.VersionCode)
	}
	if m.VersionName != "1.0" {
		t.Errorf("VersionName = %q, want %q", m.VersionName, "1.0")
	}
	if m.MinSdkVersion != 24 {
		t.Errorf("MinSdkVersion = %d, want 24", m.MinSdkVersion)
	}
	if m.TargetSdkVersion != 35 {
		t.Errorf("TargetSdkVersion = %d, want 35", m.TargetSdkVersion)
	}
}

func TestParsePermissions(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(m.Permissions) != 3 {
		t.Fatalf("got %d permissions, want 3", len(m.Permissions))
	}

	want := []string{
		"android.permission.INTERNET",
		"android.permission.ACCESS_FINE_LOCATION",
		"android.permission.CAMERA",
	}
	for i, p := range m.Permissions {
		if p.Name != want[i] {
			t.Errorf("Permission[%d] = %q, want %q", i, p.Name, want[i])
		}
		if p.Line <= 0 {
			t.Errorf("Permission[%d] line should be > 0, got %d", i, p.Line)
		}
	}
}

func TestParseComponents(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(m.Activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(m.Activities))
	}

	// MainActivity should be exported with a launcher intent filter
	main := m.Activities[0]
	if main.Name != ".MainActivity" {
		t.Errorf("Activity[0].Name = %q, want %q", main.Name, ".MainActivity")
	}
	if main.Exported == nil || !*main.Exported {
		t.Errorf("Activity[0].Exported should be true")
	}
	if len(main.IntentFilters) != 1 {
		t.Fatalf("Activity[0] has %d intent filters, want 1", len(main.IntentFilters))
	}

	// Service should have no android:exported set
	if len(m.Services) != 1 {
		t.Fatalf("got %d services, want 1", len(m.Services))
	}
	svc := m.Services[0]
	if svc.Exported != nil {
		t.Errorf("Service.Exported should be nil (not set), got %v", *svc.Exported)
	}
	if len(svc.IntentFilters) != 1 {
		t.Errorf("Service has %d intent filters, want 1", len(svc.IntentFilters))
	}

	// Receiver
	if len(m.Receivers) != 1 {
		t.Fatalf("got %d receivers, want 1", len(m.Receivers))
	}
}

func TestHasLauncherActivity(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !m.HasLauncherActivity() {
		t.Error("HasLauncherActivity should return true")
	}
}

func TestCleartextTraffic(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !m.HasCleartext {
		t.Error("HasCleartext should be true (explicitly set)")
	}
	if m.UsesCleartext {
		t.Error("UsesCleartext should be false")
	}
}

func TestValidateTargetSDK(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		wantFind bool
	}{
		{
			name: "target sdk 35 passes",
			xml: `<manifest package="test">
				<uses-sdk android:targetSdkVersion="35" />
			</manifest>`,
			wantFind: false,
		},
		{
			name: "target sdk 34 fails",
			xml: `<manifest package="test">
				<uses-sdk android:targetSdkVersion="34" />
			</manifest>`,
			wantFind: true,
		},
		{
			name: "missing target sdk fails",
			xml:      `<manifest package="test"></manifest>`,
			wantFind: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := Parse([]byte(tc.xml))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			v := NewValidator(m)
			findings := v.CheckTargetSDK()
			if tc.wantFind && len(findings) == 0 {
				t.Error("expected findings, got none")
			}
			if !tc.wantFind && len(findings) > 0 {
				t.Errorf("expected no findings, got %d", len(findings))
			}
		})
	}
}

func TestValidateDangerousPermissions(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	v := NewValidator(m)
	findings := v.CheckDangerousPermissions()

	// INTERNET is not dangerous; FINE_LOCATION and CAMERA are
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2", len(findings))
	}

	for _, f := range findings {
		if f.Location.Line <= 0 {
			t.Errorf("finding %q should have line > 0", f.Title)
		}
	}
}

func TestValidateExportedComponents(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	// Service has intent-filter but no android:exported -> Error
	// MainActivity has intent-filter and exported=true -> Info
	// Receiver has intent-filter and exported=true -> Info
	var errors, infos int
	for _, f := range findings {
		switch f.Severity {
		case preflight.SeverityError:
			errors++
		case preflight.SeverityInfo:
			infos++
		}
	}

	if errors != 1 {
		t.Errorf("got %d error findings, want 1 (service missing exported)", errors)
	}
	if infos != 2 {
		t.Errorf("got %d info findings, want 2 (main activity + receiver exported)", infos)
	}
}

func TestValidateLauncherActivity(t *testing.T) {
	noLauncher := `<manifest package="test">
		<application>
			<activity android:name=".NoLauncher" android:exported="true">
				<intent-filter>
					<action android:name="android.intent.action.VIEW" />
				</intent-filter>
			</activity>
		</application>
	</manifest>`

	m, err := Parse([]byte(noLauncher))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	v := NewValidator(m)
	findings := v.CheckLauncherActivity()
	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1", len(findings))
	}
}

func TestValidateCleartextTraffic(t *testing.T) {
	cleartext := `<manifest package="test">
		<uses-sdk android:targetSdkVersion="35" />
		<application android:usesCleartextTraffic="true">
		</application>
	</manifest>`

	m, err := Parse([]byte(cleartext))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	v := NewValidator(m)
	findings := v.CheckCleartextTraffic()
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1", len(findings))
	}
	if findings[0].Severity != preflight.SeverityError {
		t.Errorf("severity = %v, want ERROR", findings[0].Severity)
	}
}

func TestValidateAll_WithParsedManifest(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	v := NewValidator(m)
	findings := v.ValidateAll()

	// Should have findings for:
	// - dangerous permissions (FINE_LOCATION, CAMERA) = 2
	// - exported component without android:exported (service) = 1
	// - exported component security (main activity, receiver) = 2
	// No findings for: target SDK (35), launcher (present), cleartext (false)
	if len(findings) < 3 {
		t.Errorf("got %d findings, want at least 3", len(findings))
	}
}

func TestParseMalformedXML(t *testing.T) {
	malformed := `<manifest package="test">`
	// Should not panic, may or may not error depending on XML parser behavior
	_, _ = Parse([]byte(malformed))
}

func TestLineNumbers(t *testing.T) {
	m, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// All permissions should have distinct, positive line numbers
	seen := make(map[int]bool)
	for _, p := range m.Permissions {
		if p.Line <= 0 {
			t.Errorf("permission %q has line %d, want > 0", p.Name, p.Line)
		}
		if seen[p.Line] {
			t.Errorf("permission %q shares line %d with another permission", p.Name, p.Line)
		}
		seen[p.Line] = true
	}

	// Activities should have distinct line numbers
	for _, a := range m.Activities {
		if a.Line <= 0 {
			t.Errorf("activity %q has line %d, want > 0", a.Name, a.Line)
		}
	}
}

func TestManifestScanner(t *testing.T) {
	scanner := NewScanner()
	if scanner.ID() != "manifest" {
		t.Errorf("ID = %q, want %q", scanner.ID(), "manifest")
	}
	if scanner.Name() == "" {
		t.Error("Name should not be empty")
	}
	if scanner.Description() == "" {
		t.Error("Description should not be empty")
	}
}

// --- Tests for ParseFile ---

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/AndroidManifest.xml"
	content := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.parsefile">
    <uses-sdk android:targetSdkVersion="35" />
    <uses-permission android:name="android.permission.INTERNET" />
    <application />
</manifest>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if m.Package != "com.example.parsefile" {
		t.Errorf("Package = %q, want %q", m.Package, "com.example.parsefile")
	}
	if m.FilePath() != path {
		t.Errorf("FilePath = %q, want %q", m.FilePath(), path)
	}
	if m.TargetSdkVersion != 35 {
		t.Errorf("TargetSdkVersion = %d, want 35", m.TargetSdkVersion)
	}
	if len(m.Permissions) != 1 {
		t.Errorf("got %d permissions, want 1", len(m.Permissions))
	}
}

func TestParseFile_NonexistentPath(t *testing.T) {
	_, err := ParseFile("/nonexistent/AndroidManifest.xml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseFile_MalformedXML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/AndroidManifest.xml"
	if err := os.WriteFile(path, []byte("<manifest><broken"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseFile(path)
	if err == nil {
		t.Error("expected error for malformed XML")
	}
}

// --- Tests for FindAndParse ---

func TestFindAndParse_AppSrcMain(t *testing.T) {
	dir := t.TempDir()
	manifestDir := dir + "/app/src/main"
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.findandparse">
    <uses-sdk android:targetSdkVersion="35" />
    <application />
</manifest>`
	if err := os.WriteFile(manifestDir+"/AndroidManifest.xml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := FindAndParse(dir)
	if err != nil {
		t.Fatalf("FindAndParse error: %v", err)
	}
	if m.Package != "com.example.findandparse" {
		t.Errorf("Package = %q, want %q", m.Package, "com.example.findandparse")
	}
}

func TestFindAndParse_RootManifest(t *testing.T) {
	dir := t.TempDir()
	content := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.root">
    <uses-sdk android:targetSdkVersion="35" />
    <application />
</manifest>`
	if err := os.WriteFile(dir+"/AndroidManifest.xml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := FindAndParse(dir)
	if err != nil {
		t.Fatalf("FindAndParse error: %v", err)
	}
	if m.Package != "com.example.root" {
		t.Errorf("Package = %q, want %q", m.Package, "com.example.root")
	}
}

func TestFindAndParse_SrcMainManifest(t *testing.T) {
	dir := t.TempDir()
	manifestDir := dir + "/src/main"
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.srcmain">
    <uses-sdk android:targetSdkVersion="35" />
    <application />
</manifest>`
	if err := os.WriteFile(manifestDir+"/AndroidManifest.xml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := FindAndParse(dir)
	if err != nil {
		t.Fatalf("FindAndParse error: %v", err)
	}
	if m.Package != "com.example.srcmain" {
		t.Errorf("Package = %q, want %q", m.Package, "com.example.srcmain")
	}
}

func TestFindAndParse_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindAndParse(dir)
	if err == nil {
		t.Error("expected error when no manifest found")
	}
}

func TestFindAndParse_PriorityOrder(t *testing.T) {
	// When both app/src/main/ and root exist, app/src/main/ should be preferred
	dir := t.TempDir()
	manifestDir := dir + "/app/src/main"
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		t.Fatal(err)
	}

	appManifest := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.app">
    <uses-sdk android:targetSdkVersion="35" />
    <application />
</manifest>`
	rootManifest := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.root">
    <uses-sdk android:targetSdkVersion="35" />
    <application />
</manifest>`

	if err := os.WriteFile(manifestDir+"/AndroidManifest.xml", []byte(appManifest), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/AndroidManifest.xml", []byte(rootManifest), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := FindAndParse(dir)
	if err != nil {
		t.Fatalf("FindAndParse error: %v", err)
	}
	// app/src/main/ is checked first, so should return that one
	if m.Package != "com.example.app" {
		t.Errorf("Package = %q, want %q (app/src/main should take priority)", m.Package, "com.example.app")
	}
}
