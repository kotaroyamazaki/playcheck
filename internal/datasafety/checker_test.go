package datasafety

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

func setupTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestChecker_ID(t *testing.T) {
	c := &Checker{}
	if c.ID() != "DATA_SAFETY" {
		t.Errorf("expected ID DATA_SAFETY, got %s", c.ID())
	}
}

func TestChecker_Run_ViolatingApp(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"app/src/main/AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.test">
    <uses-permission android:name="android.permission.SEND_SMS" />
    <uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
    <uses-permission android:name="android.permission.ACCESS_BACKGROUND_LOCATION" />
    <uses-permission android:name="android.permission.CAMERA" />
    <application />
</manifest>`,
		"app/src/main/java/com/example/test/Main.java": `package com.example.test;
public class Main {
    public void createAccount(String email) {
        UserRepo.createUser(email, "password");
    }
    public void collectLocation() {
        getLastKnownLocation(provider);
    }
}`,
		"app/src/main/res/values/strings.xml": `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">Test App</string>
</resources>`,
	})

	c := &Checker{}
	result, err := c.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result.Passed {
		t.Error("expected Passed to be false for violating app")
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings for violating app")
	}

	// Check that privacy policy missing is detected
	hasPrivacyPolicyFinding := false
	for _, f := range result.Findings {
		if f.CheckID == "PDS001" {
			hasPrivacyPolicyFinding = true
		}
	}
	if !hasPrivacyPolicyFinding {
		t.Error("expected PDS001 (privacy policy missing) finding")
	}

	// Check that account deletion finding is present
	hasAccountDeletion := false
	for _, f := range result.Findings {
		if f.CheckID == "AD001" {
			hasAccountDeletion = true
		}
	}
	if !hasAccountDeletion {
		t.Error("expected AD001 (account deletion) finding")
	}
}

func TestChecker_Run_CleanApp(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"app/src/main/AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.clean">
    <uses-permission android:name="android.permission.INTERNET" />
    <application>
        <meta-data android:name="privacy_policy_url" android:value="https://example.com/privacy" />
    </application>
</manifest>`,
		"app/src/main/java/com/example/clean/Main.java": `package com.example.clean;
public class Main {
    private static final String PRIVACY_POLICY = "https://example.com/privacy-policy";
    public void fetchData() {
        // clean HTTPS-only code
    }
}`,
		"app/src/main/res/values/strings.xml": `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">Clean App</string>
    <string name="privacy_policy_url">https://example.com/privacy</string>
</resources>`,
	})

	c := &Checker{}
	result, err := c.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Clean app should have no privacy policy finding
	for _, f := range result.Findings {
		if f.CheckID == "PDS001" {
			t.Error("did not expect PDS001 finding for clean app with privacy policy")
		}
	}
}

func TestCheckPrivacyPolicy_InManifest(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.test">
    <application>
        <meta-data android:name="privacy_policy_url" android:value="https://example.com/privacy" />
    </application>
</manifest>`,
	})

	manifests := []string{filepath.Join(dir, "AndroidManifest.xml")}
	findings := checkPrivacyPolicy(dir, manifests)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings when privacy policy in manifest, got %d", len(findings))
	}
}

func TestCheckPrivacyPolicy_InStrings(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.test">
    <application />
</manifest>`,
		"res/values/strings.xml": `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="privacy_policy_url">https://example.com/privacy</string>
</resources>`,
	})

	manifests := []string{filepath.Join(dir, "AndroidManifest.xml")}
	findings := checkPrivacyPolicy(dir, manifests)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings when privacy policy in strings.xml, got %d", len(findings))
	}
}

func TestCheckPrivacyPolicy_Missing(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.test">
    <application />
</manifest>`,
		"res/values/strings.xml": `<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_name">Test</string>
</resources>`,
	})

	manifests := []string{filepath.Join(dir, "AndroidManifest.xml")}
	findings := checkPrivacyPolicy(dir, manifests)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for missing privacy policy, got %d", len(findings))
	}
	if findings[0].CheckID != "PDS001" {
		t.Errorf("expected check ID PDS001, got %s", findings[0].CheckID)
	}
}

func TestCheckAccountDeletion_MissingDeletion(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void createUser(String email, String pw) {}
}`,
	})

	findings := checkAccountDeletion(dir)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for missing account deletion, got %d", len(findings))
	}
	if findings[0].CheckID != "AD001" {
		t.Errorf("expected check ID AD001, got %s", findings[0].CheckID)
	}
	if findings[0].Severity != preflight.SeverityError {
		t.Errorf("expected severity ERROR, got %s", findings[0].Severity)
	}
}

func TestCheckAccountDeletion_HasDeletion(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void createUser(String email, String pw) {}
    public void deleteUser(String id) {}
}`,
	})

	findings := checkAccountDeletion(dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when deletion exists, got %d", len(findings))
	}
}

func TestCheckAccountDeletion_NoAccountCode(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void fetchData() {}
}`,
	})

	findings := checkAccountDeletion(dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when no account code, got %d", len(findings))
	}
}

func TestCheckUserConsent_WithoutConsent(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Tracker.java": `package com.example;
public class Tracker {
    public void track() {
        getLastKnownLocation(provider);
    }
}`,
	})

	findings := checkUserConsent(dir)
	if len(findings) == 0 {
		t.Error("expected findings for data collection without consent")
	}
	found := false
	for _, f := range findings {
		if f.CheckID == "PDS003" {
			found = true
		}
	}
	if !found {
		t.Error("expected PDS003 finding for data collection without consent")
	}
}

func TestCheckUserConsent_WithConsent(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Tracker.java": `package com.example;
public class Tracker {
    // Gets user consent before collecting data
    public void track() {
        if (userConsent) {
            getLastKnownLocation(provider);
        }
    }
}`,
	})

	findings := checkUserConsent(dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when consent present, got %d", len(findings))
	}
}

func TestFindLineNumber(t *testing.T) {
	content := "line1\nline2\nline3\nfoobar\nline5"
	line := findLineNumber(content, "foobar")
	if line != 4 {
		t.Errorf("expected line 4, got %d", line)
	}

	line = findLineNumber(content, "notfound")
	if line != 0 {
		t.Errorf("expected line 0 for not found, got %d", line)
	}
}

func TestCheckBackgroundLocation(t *testing.T) {
	m := manifestInfo{
		FilePath: "/test/AndroidManifest.xml",
		Permissions: []string{
			"android.permission.ACCESS_BACKGROUND_LOCATION",
			"android.permission.ACCESS_FINE_LOCATION",
		},
	}

	findings := checkBackgroundLocation(m, "AndroidManifest.xml", "/test")
	// Should have at least one finding about background location
	hasBackgroundFinding := false
	for _, f := range findings {
		if f.CheckID == "DP006" {
			hasBackgroundFinding = true
		}
	}
	if !hasBackgroundFinding {
		t.Error("expected DP006 finding for background location")
	}
}

func TestCheckBackgroundLocation_WithoutForeground(t *testing.T) {
	m := manifestInfo{
		FilePath: "/test/AndroidManifest.xml",
		Permissions: []string{
			"android.permission.ACCESS_BACKGROUND_LOCATION",
		},
	}

	findings := checkBackgroundLocation(m, "AndroidManifest.xml", "/test")
	// Should have 2 findings: background location + missing foreground
	if len(findings) < 2 {
		t.Errorf("expected at least 2 findings for background without foreground, got %d", len(findings))
	}
}

func TestCheckPermissionDisclosures(t *testing.T) {
	manifests := []manifestInfo{
		{
			FilePath: "/test/AndroidManifest.xml",
			Permissions: []string{
				"android.permission.READ_SMS",
				"android.permission.CAMERA",
				"android.permission.INTERNET",
			},
			HasMeta: map[string]bool{},
		},
	}

	findings := checkPermissionDisclosures(manifests, "/test")
	// Should find disclosures for READ_SMS and CAMERA (INTERNET is not dangerous)
	hasSMSDisclosure := false
	hasCameraDisclosure := false
	for _, f := range findings {
		if f.CheckID == "PDS002" {
			if strings.Contains(f.Description, "Text messages") {
				hasSMSDisclosure = true
			}
			if strings.Contains(f.Description, "Photos/Videos") {
				hasCameraDisclosure = true
			}
		}
	}
	if !hasSMSDisclosure {
		t.Error("expected disclosure finding for SMS permission")
	}
	if !hasCameraDisclosure {
		t.Error("expected disclosure finding for CAMERA permission")
	}
}

// --- Tests for checkSDKDisclosures ---

func TestCheckSDKDisclosures_FirebaseAnalytics(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"app/build.gradle": `plugins {
    id 'com.android.application'
}
dependencies {
    implementation 'com.google.firebase:firebase-analytics:21.5.0'
    implementation 'com.google.firebase:firebase-crashlytics:18.6.0'
}`,
	})

	findings := checkSDKDisclosures(dir)
	if len(findings) == 0 {
		t.Fatal("expected findings for Firebase SDK dependencies")
	}

	hasAnalytics := false
	hasCrashlytics := false
	for _, f := range findings {
		if f.CheckID == "SDK001" {
			if strings.Contains(f.Description, "Firebase Analytics") {
				hasAnalytics = true
			}
			if strings.Contains(f.Description, "Firebase Crashlytics") {
				hasCrashlytics = true
			}
		}
	}
	if !hasAnalytics {
		t.Error("expected finding for Firebase Analytics SDK")
	}
	if !hasCrashlytics {
		t.Error("expected finding for Firebase Crashlytics SDK")
	}
}

func TestCheckSDKDisclosures_NoGradleFiles(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `class Main {}`,
	})

	findings := checkSDKDisclosures(dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when no gradle files, got %d", len(findings))
	}
}

func TestCheckSDKDisclosures_CleanGradle(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"app/build.gradle": `plugins {
    id 'com.android.application'
}
dependencies {
    implementation 'androidx.core:core-ktx:1.12.0'
    implementation 'androidx.appcompat:appcompat:1.6.1'
}`,
	})

	findings := checkSDKDisclosures(dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for clean gradle, got %d", len(findings))
	}
}

func TestCheckSDKDisclosures_MultipleSDKs(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"app/build.gradle.kts": `dependencies {
    implementation("com.facebook.android:facebook-android-sdk:16.0.0")
    implementation("com.adjust.sdk:adjust-android:4.38.0")
    implementation("com.stripe:stripe-android:20.30.0")
}`,
	})

	findings := checkSDKDisclosures(dir)
	if len(findings) < 3 {
		t.Errorf("expected at least 3 findings for multiple SDKs, got %d", len(findings))
	}
}

// --- Tests for crossReferencePermissionsWithCode ---

func TestCrossReferencePermissions_UsedInCode(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
import android.hardware.camera2.CameraManager;
public class Main {
    public void open() {
        CameraManager cm = getSystemService(CameraManager.class);
    }
}`,
	})

	manifests := []manifestInfo{
		{
			FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
			Permissions: []string{"android.permission.CAMERA"},
			HasMeta:     map[string]bool{},
		},
	}

	findings := crossReferencePermissionsWithCode(manifests, dir)
	for _, f := range findings {
		if f.CheckID == "SDK004" && strings.Contains(f.Description, "CAMERA") {
			t.Error("did not expect unused CAMERA finding when CameraManager is in code")
		}
	}
}

func TestCrossReferencePermissions_NotUsedInCode(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void doNothing() {}
}`,
	})

	manifests := []manifestInfo{
		{
			FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
			Permissions: []string{"android.permission.CAMERA"},
			HasMeta:     map[string]bool{},
		},
	}

	findings := crossReferencePermissionsWithCode(manifests, dir)
	found := false
	for _, f := range findings {
		if f.CheckID == "SDK004" && strings.Contains(f.Description, "CAMERA") {
			found = true
		}
	}
	if !found {
		t.Error("expected SDK004 finding for unused CAMERA permission")
	}
}

func TestCrossReferencePermissions_NonDangerousPermission(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `class Main {}`,
	})

	manifests := []manifestInfo{
		{
			FilePath:    "/test/AndroidManifest.xml",
			Permissions: []string{"android.permission.INTERNET"},
			HasMeta:     map[string]bool{},
		},
	}

	findings := crossReferencePermissionsWithCode(manifests, dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-dangerous permission, got %d", len(findings))
	}
}

// --- Tests for checkRuntimePermissions ---

func TestCheckRuntimePermissions_WithRequest(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void askPerms() {
        ActivityCompat.requestPermissions(this, new String[]{Manifest.permission.CAMERA}, 100);
    }
}`,
	})

	m := manifestInfo{
		FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
		Permissions: []string{"android.permission.CAMERA"},
		HasMeta:     map[string]bool{},
	}

	findings := checkRuntimePermissions(m, dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when runtime permission request present, got %d", len(findings))
	}
}

func TestCheckRuntimePermissions_WithCheckSelfPermission(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void check() {
        ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA);
    }
}`,
	})

	m := manifestInfo{
		FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
		Permissions: []string{"android.permission.CAMERA"},
		HasMeta:     map[string]bool{},
	}

	findings := checkRuntimePermissions(m, dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when checkSelfPermission present, got %d", len(findings))
	}
}

func TestCheckRuntimePermissions_Missing(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    public void doNothing() {}
}`,
	})

	m := manifestInfo{
		FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
		Permissions: []string{"android.permission.CAMERA"},
		HasMeta:     map[string]bool{},
	}

	findings := checkRuntimePermissions(m, dir)
	found := false
	for _, f := range findings {
		if f.CheckID == "PDS004" {
			found = true
		}
	}
	if !found {
		t.Error("expected PDS004 finding for missing runtime permission request")
	}
}

func TestCheckRuntimePermissions_NoDangerousPerms(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"Main.java": `class Main {}`,
	})

	m := manifestInfo{
		FilePath:    filepath.Join(dir, "AndroidManifest.xml"),
		Permissions: []string{"android.permission.INTERNET"},
		HasMeta:     map[string]bool{},
	}

	findings := checkRuntimePermissions(m, dir)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when no dangerous permissions, got %d", len(findings))
	}
}

// --- Tests for parseManifests ---

func TestParseManifests(t *testing.T) {
	dir := setupTestProject(t, map[string]string{
		"AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example">
    <uses-permission android:name="android.permission.CAMERA" />
    <uses-permission android:name="android.permission.INTERNET" />
    <application>
        <meta-data android:name="privacy_policy_url" android:value="https://example.com/privacy" />
    </application>
</manifest>`,
	})

	paths := []string{filepath.Join(dir, "AndroidManifest.xml")}
	result := parseManifests(paths)

	if len(result) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(result))
	}
	if len(result[0].Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(result[0].Permissions))
	}
	if !result[0].HasMeta["privacy_policy_url"] {
		t.Error("expected privacy_policy_url in HasMeta")
	}
}

func TestParseManifests_NonexistentFile(t *testing.T) {
	result := parseManifests([]string{"/nonexistent/AndroidManifest.xml"})
	if len(result) != 0 {
		t.Errorf("expected 0 results for nonexistent file, got %d", len(result))
	}
}

// --- Tests for Checker.Name and Checker.Description ---

func TestChecker_Name(t *testing.T) {
	c := &Checker{}
	if c.Name() == "" {
		t.Error("Name should not be empty")
	}
}

func TestChecker_Description(t *testing.T) {
	c := &Checker{}
	if c.Description() == "" {
		t.Error("Description should not be empty")
	}
}

