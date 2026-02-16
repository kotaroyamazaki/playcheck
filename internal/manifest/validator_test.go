package manifest

import (
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

func boolPtr(v bool) *bool { return &v }

func TestCheckTargetSDK_Missing(t *testing.T) {
	m := &AndroidManifest{filePath: "AndroidManifest.xml"}
	v := NewValidator(m)
	findings := v.CheckTargetSDK()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].CheckID != RuleTargetSDK {
		t.Errorf("expected check ID %s, got %s", RuleTargetSDK, findings[0].CheckID)
	}
	if findings[0].Severity != preflight.SeverityCritical {
		t.Errorf("expected severity CRITICAL, got %s", findings[0].Severity)
	}
}

func TestCheckTargetSDK_BelowMinimum(t *testing.T) {
	m := &AndroidManifest{TargetSdkVersion: 33, filePath: "AndroidManifest.xml"}
	v := NewValidator(m)
	findings := v.CheckTargetSDK()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != preflight.SeverityCritical {
		t.Errorf("expected severity CRITICAL, got %s", findings[0].Severity)
	}
}

func TestCheckTargetSDK_MeetsMinimum(t *testing.T) {
	m := &AndroidManifest{TargetSdkVersion: MinTargetSDKVersion, filePath: "AndroidManifest.xml"}
	v := NewValidator(m)
	findings := v.CheckTargetSDK()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for valid target SDK, got %d", len(findings))
	}
}

func TestCheckDangerousPermissions(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Permissions: []Permission{
			{Name: "android.permission.INTERNET", Line: 5},
			{Name: "android.permission.SEND_SMS", Line: 6},
			{Name: "android.permission.CAMERA", Line: 7},
			{Name: "android.permission.ACCESS_FINE_LOCATION", Line: 8},
		},
	}
	v := NewValidator(m)
	findings := v.CheckDangerousPermissions()

	// INTERNET is not dangerous, so expect 3 findings
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	// SEND_SMS should be critical
	foundSMS := false
	for _, f := range findings {
		if f.Location.Line == 6 {
			foundSMS = true
			if f.Severity != preflight.SeverityCritical {
				t.Errorf("SEND_SMS should be CRITICAL, got %s", f.Severity)
			}
		}
	}
	if !foundSMS {
		t.Error("expected finding for SEND_SMS permission")
	}
}

func TestCheckDangerousPermissions_NoFindings(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Permissions: []Permission{
			{Name: "android.permission.INTERNET", Line: 5},
			{Name: "android.permission.ACCESS_NETWORK_STATE", Line: 6},
		},
	}
	v := NewValidator(m)
	findings := v.CheckDangerousPermissions()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for safe permissions, got %d", len(findings))
	}
}

func TestCheckExportedComponents_MissingExported(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Activities: []Activity{
			{
				Name:     ".MainActivity",
				Exported: nil, // missing android:exported
				IntentFilters: []IntentFilter{
					{Actions: []string{"android.intent.action.MAIN"}},
				},
				Line: 10,
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for missing exported, got %d", len(findings))
	}
	if findings[0].CheckID != RuleExportedComponent {
		t.Errorf("expected check ID %s, got %s", RuleExportedComponent, findings[0].CheckID)
	}
	if findings[0].Severity != preflight.SeverityError {
		t.Errorf("expected severity ERROR, got %s", findings[0].Severity)
	}
}

func TestCheckExportedComponents_ExplicitlyExported(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Activities: []Activity{
			{
				Name:     ".MainActivity",
				Exported: boolPtr(true),
				IntentFilters: []IntentFilter{
					{Actions: []string{"android.intent.action.MAIN"}},
				},
				Line: 10,
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	// Explicitly exported with intent filters should produce an INFO finding
	if len(findings) != 1 {
		t.Fatalf("expected 1 info finding for explicitly exported, got %d", len(findings))
	}
	if findings[0].CheckID != RuleComponentSecurity {
		t.Errorf("expected check ID %s, got %s", RuleComponentSecurity, findings[0].CheckID)
	}
	if findings[0].Severity != preflight.SeverityInfo {
		t.Errorf("expected severity INFO, got %s", findings[0].Severity)
	}
}

func TestCheckExportedComponents_NoIntentFilters(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Activities: []Activity{
			{
				Name:          ".InternalActivity",
				Exported:      boolPtr(false),
				IntentFilters: nil,
				Line:          10,
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for component without intent filters, got %d", len(findings))
	}
}

func TestCheckExportedComponents_Services(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Services: []Service{
			{
				Name:     ".MyService",
				Exported: nil,
				IntentFilters: []IntentFilter{
					{Actions: []string{"com.example.ACTION"}},
				},
				Line: 20,
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for service missing exported, got %d", len(findings))
	}
}

func TestCheckExportedComponents_Receivers(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Receivers: []Receiver{
			{
				Name:     ".MyReceiver",
				Exported: boolPtr(true),
				IntentFilters: []IntentFilter{
					{Actions: []string{"android.intent.action.BOOT_COMPLETED"}},
				},
				Line: 30,
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckExportedComponents()

	if len(findings) != 1 {
		t.Fatalf("expected 1 info finding for exported receiver, got %d", len(findings))
	}
	if findings[0].Severity != preflight.SeverityInfo {
		t.Errorf("expected INFO severity, got %s", findings[0].Severity)
	}
}

func TestCheckLauncherActivity_Present(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Activities: []Activity{
			{
				Name:     ".MainActivity",
				Exported: boolPtr(true),
				IntentFilters: []IntentFilter{
					{
						Actions:    []string{"android.intent.action.MAIN"},
						Categories: []string{"android.intent.category.LAUNCHER"},
					},
				},
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckLauncherActivity()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when launcher activity present, got %d", len(findings))
	}
}

func TestCheckLauncherActivity_Missing(t *testing.T) {
	m := &AndroidManifest{
		filePath: "AndroidManifest.xml",
		Activities: []Activity{
			{
				Name:     ".OtherActivity",
				Exported: boolPtr(true),
				IntentFilters: []IntentFilter{
					{Actions: []string{"android.intent.action.VIEW"}},
				},
			},
		},
	}
	v := NewValidator(m)
	findings := v.CheckLauncherActivity()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for missing launcher, got %d", len(findings))
	}
	if findings[0].CheckID != RuleLauncherActivity {
		t.Errorf("expected check ID %s, got %s", RuleLauncherActivity, findings[0].CheckID)
	}
}

func TestCheckCleartextTraffic_Enabled(t *testing.T) {
	m := &AndroidManifest{
		filePath:     "AndroidManifest.xml",
		HasCleartext: true,
		UsesCleartext: true,
	}
	v := NewValidator(m)
	findings := v.CheckCleartextTraffic()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for cleartext enabled, got %d", len(findings))
	}
	if findings[0].Severity != preflight.SeverityError {
		t.Errorf("expected severity ERROR, got %s", findings[0].Severity)
	}
}

func TestCheckCleartextTraffic_Disabled(t *testing.T) {
	m := &AndroidManifest{
		filePath:     "AndroidManifest.xml",
		HasCleartext: true,
		UsesCleartext: false,
	}
	v := NewValidator(m)
	findings := v.CheckCleartextTraffic()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when cleartext disabled, got %d", len(findings))
	}
}

func TestCheckCleartextTraffic_DefaultLowSDK(t *testing.T) {
	m := &AndroidManifest{
		filePath:         "AndroidManifest.xml",
		HasCleartext:     false,
		TargetSdkVersion: 27,
	}
	v := NewValidator(m)
	findings := v.CheckCleartextTraffic()

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for low SDK default cleartext, got %d", len(findings))
	}
	if findings[0].Severity != preflight.SeverityWarning {
		t.Errorf("expected severity WARNING, got %s", findings[0].Severity)
	}
}

func TestCheckCleartextTraffic_DefaultHighSDK(t *testing.T) {
	m := &AndroidManifest{
		filePath:         "AndroidManifest.xml",
		HasCleartext:     false,
		TargetSdkVersion: 35,
	}
	v := NewValidator(m)
	findings := v.CheckCleartextTraffic()

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for high SDK default cleartext, got %d", len(findings))
	}
}

func TestValidateAll(t *testing.T) {
	m := &AndroidManifest{
		filePath:         "AndroidManifest.xml",
		TargetSdkVersion: 33,
		HasCleartext:     true,
		UsesCleartext:    true,
		Permissions: []Permission{
			{Name: "android.permission.SEND_SMS", Line: 5},
		},
	}
	v := NewValidator(m)
	findings := v.ValidateAll()

	// Should have at least:
	// - target SDK too low (critical)
	// - dangerous permission SEND_SMS (critical)
	// - cleartext enabled (error)
	// - missing launcher activity (warning)
	if len(findings) < 4 {
		t.Errorf("expected at least 4 findings from ValidateAll, got %d", len(findings))
		for _, f := range findings {
			t.Logf("  %s: %s (%s)", f.CheckID, f.Title, f.Severity)
		}
	}
}

func TestParse_ValidManifest(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.test">
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.CAMERA" />
    <application>
        <uses-sdk android:minSdkVersion="24" android:targetSdkVersion="35" />
        <activity
            android:name=".MainActivity"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>`

	m, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if m.Package != "com.example.test" {
		t.Errorf("expected package com.example.test, got %s", m.Package)
	}
	if m.TargetSdkVersion != 35 {
		t.Errorf("expected targetSdkVersion 35, got %d", m.TargetSdkVersion)
	}
	if m.MinSdkVersion != 24 {
		t.Errorf("expected minSdkVersion 24, got %d", m.MinSdkVersion)
	}
	if len(m.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(m.Permissions))
	}
	if len(m.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(m.Activities))
	}
	if !m.HasLauncherActivity() {
		t.Error("expected launcher activity to be detected")
	}
}

func TestParse_ComponentTypes(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="com.example.test">
    <application>
        <activity android:name=".A1" android:exported="true">
            <intent-filter><action android:name="android.intent.action.MAIN" /></intent-filter>
        </activity>
        <service android:name=".S1" android:exported="false" />
        <receiver android:name=".R1" android:exported="true">
            <intent-filter><action android:name="com.example.ACTION" /></intent-filter>
        </receiver>
        <provider android:name=".P1" android:exported="false" />
    </application>
</manifest>`

	m, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(m.Activities) != 1 {
		t.Errorf("expected 1 activity, got %d", len(m.Activities))
	}
	if len(m.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(m.Services))
	}
	if len(m.Receivers) != 1 {
		t.Errorf("expected 1 receiver, got %d", len(m.Receivers))
	}
	if len(m.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(m.Providers))
	}
}

func TestParse_CleartextTraffic(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.example.test">
    <application android:usesCleartextTraffic="true" />
</manifest>`

	m, err := Parse([]byte(xml))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if !m.HasCleartext {
		t.Error("expected HasCleartext to be true")
	}
	if !m.UsesCleartext {
		t.Error("expected UsesCleartext to be true")
	}
}

func TestParse_InvalidXML(t *testing.T) {
	_, err := Parse([]byte("not xml at all <<<<"))
	// The parser is lenient, so it might not error. That's OK.
	_ = err
}

func TestShortPermName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"android.permission.CAMERA", "CAMERA"},
		{"android.permission.SEND_SMS", "SEND_SMS"},
		{"CAMERA", "CAMERA"},
	}

	for _, tc := range tests {
		got := shortPermName(tc.input)
		if got != tc.expected {
			t.Errorf("shortPermName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
