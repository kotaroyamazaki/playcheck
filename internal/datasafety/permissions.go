package datasafety

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yourusername/playcheck/internal/preflight"
	"github.com/yourusername/playcheck/pkg/utils"
)

// permissionDisclosure maps dangerous Android permissions to required data safety disclosures.
type permissionDisclosure struct {
	Permission    string
	DataType      string
	DisclosureMsg string
	CheckID       string
}

var dangerousPermissionDisclosures = []permissionDisclosure{
	{
		Permission:    "android.permission.READ_SMS",
		DataType:      "Text messages",
		DisclosureMsg: "READ_SMS permission requires disclosure of text message data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.RECEIVE_SMS",
		DataType:      "Text messages",
		DisclosureMsg: "RECEIVE_SMS permission requires disclosure of text message data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.READ_CALL_LOG",
		DataType:      "Call logs",
		DisclosureMsg: "READ_CALL_LOG permission requires disclosure of call log data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.READ_CONTACTS",
		DataType:      "Contacts",
		DisclosureMsg: "READ_CONTACTS permission requires disclosure of contacts data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.ACCESS_FINE_LOCATION",
		DataType:      "Precise location",
		DisclosureMsg: "ACCESS_FINE_LOCATION permission requires disclosure of precise location data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.ACCESS_COARSE_LOCATION",
		DataType:      "Approximate location",
		DisclosureMsg: "ACCESS_COARSE_LOCATION permission requires disclosure of approximate location data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.CAMERA",
		DataType:      "Photos/Videos",
		DisclosureMsg: "CAMERA permission requires disclosure of photo/video data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.RECORD_AUDIO",
		DataType:      "Audio recordings",
		DisclosureMsg: "RECORD_AUDIO permission requires disclosure of audio data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.READ_EXTERNAL_STORAGE",
		DataType:      "Files and documents",
		DisclosureMsg: "READ_EXTERNAL_STORAGE permission requires disclosure of file/document data access",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.READ_CALENDAR",
		DataType:      "Calendar events",
		DisclosureMsg: "READ_CALENDAR permission requires disclosure of calendar data collection",
		CheckID:       "PDS002",
	},
	{
		Permission:    "android.permission.BODY_SENSORS",
		DataType:      "Health data",
		DisclosureMsg: "BODY_SENSORS permission requires disclosure of health/fitness data collection",
		CheckID:       "PDS002",
	},
}

// checkPermissionDisclosures validates that manifest permissions have corresponding data safety disclosures.
func checkPermissionDisclosures(manifests []manifestInfo, projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	for _, m := range manifests {
		relPath, _ := filepath.Rel(projectDir, m.FilePath)

		for _, perm := range m.Permissions {
			for _, disc := range dangerousPermissionDisclosures {
				if perm == disc.Permission {
					findings = append(findings, preflight.Finding{
						CheckID:     disc.CheckID,
						Title:       "Permission requires data safety disclosure",
						Description: disc.DisclosureMsg + ". Data type: " + disc.DataType,
						Severity:    preflight.SeverityWarning,
						Location:    preflight.Location{File: relPath},
						Suggestion:  "Declare '" + disc.DataType + "' data collection in your Play Console Data Safety form.",
					})
				}
			}
		}

		// Check background location access.
		findings = append(findings, checkBackgroundLocation(m, relPath, projectDir)...)

		// Check runtime permission requests in code.
		findings = append(findings, checkRuntimePermissions(m, projectDir)...)
	}

	return findings
}

// checkBackgroundLocation validates background location permission usage.
func checkBackgroundLocation(m manifestInfo, relPath, projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	hasBackgroundLocation := false
	hasForegroundLocation := false
	for _, p := range m.Permissions {
		switch p {
		case "android.permission.ACCESS_BACKGROUND_LOCATION":
			hasBackgroundLocation = true
		case "android.permission.ACCESS_FINE_LOCATION", "android.permission.ACCESS_COARSE_LOCATION":
			hasForegroundLocation = true
		}
	}

	if hasBackgroundLocation {
		findings = append(findings, preflight.Finding{
			CheckID:     "DP006",
			Title:       "Background location access declared",
			Description: "ACCESS_BACKGROUND_LOCATION requires prominent disclosure and Play Store policy review. Apps must demonstrate the need for background location.",
			Severity:    preflight.SeverityError,
			Location:    preflight.Location{File: relPath},
			Suggestion:  "Ensure your app has a prominent in-app disclosure explaining why background location is needed. Submit the permission declaration form in Play Console.",
		})

		if !hasForegroundLocation {
			findings = append(findings, preflight.Finding{
				CheckID:     "DP006",
				Title:       "Background location without foreground location",
				Description: "ACCESS_BACKGROUND_LOCATION is declared but no foreground location permission (ACCESS_FINE_LOCATION or ACCESS_COARSE_LOCATION) is present. Background location requires a foreground location permission.",
				Severity:    preflight.SeverityError,
				Location:    preflight.Location{File: relPath},
				Suggestion:  "Add ACCESS_FINE_LOCATION or ACCESS_COARSE_LOCATION permission alongside ACCESS_BACKGROUND_LOCATION.",
			})
		}
	}

	return findings
}

// runtimePermissionRe matches calls to ActivityCompat.requestPermissions or requestPermissions.
var runtimePermissionRe = regexp.MustCompile(`requestPermissions?\s*\(`)
var checkSelfPermissionRe = regexp.MustCompile(`checkSelfPermission\s*\(`)

// checkRuntimePermissions verifies that dangerous permissions are requested at runtime.
func checkRuntimePermissions(m manifestInfo, projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	// Only check if the manifest has dangerous permissions that require runtime request.
	hasDangerousPerm := false
	for _, p := range m.Permissions {
		for _, d := range dangerousPermissionDisclosures {
			if p == d.Permission {
				hasDangerousPerm = true
				break
			}
		}
		if hasDangerousPerm {
			break
		}
	}
	if !hasDangerousPerm {
		return findings
	}

	codeFiles, err := utils.WalkFiles(projectDir, utils.WithExtensions(".kt", ".java"))
	if err != nil {
		return findings
	}

	hasRuntimeRequest := false
	for _, cf := range codeFiles {
		data, err := os.ReadFile(cf)
		if err != nil {
			continue
		}
		content := string(data)
		if runtimePermissionRe.MatchString(content) || checkSelfPermissionRe.MatchString(content) {
			hasRuntimeRequest = true
			break
		}
	}

	if !hasRuntimeRequest {
		relPath, _ := filepath.Rel(projectDir, m.FilePath)
		findings = append(findings, preflight.Finding{
			CheckID:     "PDS004",
			Title:       "No runtime permission request detected",
			Description: "Dangerous permissions are declared in manifest but no runtime permission requests (requestPermissions/checkSelfPermission) were found in code. Android 6.0+ requires runtime permission requests for dangerous permissions.",
			Severity:    preflight.SeverityError,
			Location:    preflight.Location{File: relPath},
			Suggestion:  "Implement runtime permission requests using ActivityCompat.requestPermissions() or the Activity Result API.",
		})
	}

	return findings
}

// sdkInfo describes a third-party SDK that requires data safety disclosure.
type sdkInfo struct {
	Name           string
	Dependencies   []string
	DisclosureNote string
}

// thirdPartySDKs lists common SDKs that require data safety form disclosures.
var thirdPartySDKs = []sdkInfo{
	{
		Name:           "Firebase Analytics",
		Dependencies:   []string{"com.google.firebase:firebase-analytics", "firebase-analytics-ktx"},
		DisclosureNote: "Collects app interactions, device identifiers, and crash data. Disclose 'App interactions', 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Firebase Crashlytics",
		Dependencies:   []string{"com.google.firebase:firebase-crashlytics", "firebase-crashlytics-ktx"},
		DisclosureNote: "Collects crash logs and device state. Disclose 'Crash logs', 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Google AdMob",
		Dependencies:   []string{"com.google.android.gms:play-services-ads", "com.google.ads:"},
		DisclosureNote: "Collects advertising ID, device info, and interaction data. Disclose 'Device or other IDs', 'Ads data' in Data Safety.",
	},
	{
		Name:           "Facebook SDK",
		Dependencies:   []string{"com.facebook.android:facebook-", "implementation 'com.facebook.android"},
		DisclosureNote: "Collects device info, app events, and advertising data. Disclose 'Device or other IDs', 'App interactions' in Data Safety.",
	},
	{
		Name:           "Adjust SDK",
		Dependencies:   []string{"com.adjust.sdk:adjust-android"},
		DisclosureNote: "Collects device identifiers and attribution data. Disclose 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "AppsFlyer SDK",
		Dependencies:   []string{"com.appsflyer:af-android-sdk"},
		DisclosureNote: "Collects device identifiers, install referrer, and attribution data. Disclose 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Sentry SDK",
		Dependencies:   []string{"io.sentry:sentry-android"},
		DisclosureNote: "Collects crash logs and device state. Disclose 'Crash logs', 'Diagnostics' in Data Safety.",
	},
	{
		Name:           "Google Maps SDK",
		Dependencies:   []string{"com.google.android.gms:play-services-maps", "com.google.android.gms:play-services-location"},
		DisclosureNote: "May collect location data. Disclose 'Approximate location' or 'Precise location' in Data Safety if location is used.",
	},
	{
		Name:           "Mixpanel SDK",
		Dependencies:   []string{"com.mixpanel.android:mixpanel-android"},
		DisclosureNote: "Collects app interactions, device identifiers. Disclose 'App interactions', 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Amplitude SDK",
		Dependencies:   []string{"com.amplitude:android-sdk", "com.amplitude:analytics-android"},
		DisclosureNote: "Collects app interactions and device identifiers. Disclose 'App interactions', 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Braze SDK",
		Dependencies:   []string{"com.braze:android-sdk"},
		DisclosureNote: "Collects device info, push tokens, and user interactions. Disclose 'Device or other IDs', 'App interactions' in Data Safety.",
	},
	{
		Name:           "OneSignal SDK",
		Dependencies:   []string{"com.onesignal:OneSignal"},
		DisclosureNote: "Collects push notification tokens and device identifiers. Disclose 'Device or other IDs' in Data Safety.",
	},
	{
		Name:           "Stripe SDK",
		Dependencies:   []string{"com.stripe:stripe-android"},
		DisclosureNote: "Processes payment information. Disclose 'Financial info' and 'Purchase history' in Data Safety.",
	},
}

// crossReferencePermissionsWithCode checks that permissions declared in manifest
// are actually used in code, and flags unused dangerous permissions.
func crossReferencePermissionsWithCode(manifests []manifestInfo, projectDir string) []preflight.Finding {
	var findings []preflight.Finding

	codeFiles, err := utils.WalkFiles(projectDir, utils.WithExtensions(".kt", ".java"))
	if err != nil {
		return findings
	}

	// Build a set of all code content for searching.
	var allCode strings.Builder
	for _, cf := range codeFiles {
		data, err := os.ReadFile(cf)
		if err != nil {
			continue
		}
		allCode.Write(data)
		allCode.WriteByte('\n')
	}
	codeContent := allCode.String()

	// Map of permission -> common API usage patterns.
	permissionAPIs := map[string][]*regexp.Regexp{
		"android.permission.CAMERA": {
			regexp.MustCompile(`Camera|CameraManager|CameraDevice|CameraX|camera2`),
		},
		"android.permission.RECORD_AUDIO": {
			regexp.MustCompile(`MediaRecorder|AudioRecord|SpeechRecognizer`),
		},
		"android.permission.READ_CONTACTS": {
			regexp.MustCompile(`ContactsContract|ContactsProvider|READ_CONTACTS`),
		},
		"android.permission.ACCESS_FINE_LOCATION": {
			regexp.MustCompile(`LocationManager|FusedLocationProvider|LocationRequest|getLastKnownLocation|requestLocationUpdates`),
		},
		"android.permission.ACCESS_COARSE_LOCATION": {
			regexp.MustCompile(`LocationManager|FusedLocationProvider|LocationRequest|getLastKnownLocation|requestLocationUpdates`),
		},
		"android.permission.READ_SMS": {
			regexp.MustCompile(`SmsManager|Telephony\.Sms|SmsMessage`),
		},
		"android.permission.READ_CALL_LOG": {
			regexp.MustCompile(`CallLog|CallLog\.Calls`),
		},
		"android.permission.READ_CALENDAR": {
			regexp.MustCompile(`CalendarContract|CalendarProvider`),
		},
		"android.permission.BODY_SENSORS": {
			regexp.MustCompile(`SensorManager|Sensor\.TYPE_HEART|HealthServicesClient`),
		},
	}

	for _, m := range manifests {
		relPath, _ := filepath.Rel(projectDir, m.FilePath)
		for _, perm := range m.Permissions {
			apis, exists := permissionAPIs[perm]
			if !exists {
				continue
			}
			usedInCode := false
			for _, api := range apis {
				if api.MatchString(codeContent) {
					usedInCode = true
					break
				}
			}
			if !usedInCode {
				shortPerm := perm
				if idx := strings.LastIndex(perm, "."); idx >= 0 {
					shortPerm = perm[idx+1:]
				}
				findings = append(findings, preflight.Finding{
					CheckID:     "SDK004",
					Title:       "Declared permission not used in code",
					Description: shortPerm + " is declared in manifest but no corresponding API usage was detected in code. Unused dangerous permissions may cause rejection.",
					Severity:    preflight.SeverityWarning,
					Location:    preflight.Location{File: relPath},
					Suggestion:  "Remove the " + shortPerm + " permission from your manifest if it is not needed, or verify it is used by a library.",
				})
			}
		}
	}

	return findings
}
