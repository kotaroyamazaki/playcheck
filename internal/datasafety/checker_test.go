package datasafety

import (
	"os"
	"path/filepath"
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
			if containsString(f.Description, "Text messages") {
				hasSMSDisclosure = true
			}
			if containsString(f.Description, "Photos/Videos") {
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

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
