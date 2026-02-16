package manifest

import "github.com/kotaroyamazaki/playcheck/internal/preflight"

// Rule IDs for manifest validation checks.
const (
	RuleTargetSDK         = "SDK001"
	RuleMinSDK            = "SDK004"
	RuleDangerousPerm     = "DP001"
	RuleLocationPerm      = "DP002"
	RuleCameraPerm        = "DP003"
	RuleContactsPerm      = "DP004"
	RuleStoragePerm       = "DP005"
	RulePhonePerm         = "DP006"
	RuleCalendarPerm      = "DP007"
	RuleExportedComponent = "MV001"
	RuleLauncherActivity  = "MV002"
	RuleCleartextTraffic  = "MV004"
	RuleComponentSecurity = "MC001"
)

// dangerousPermissions maps Android permission names to their rule IDs and descriptions.
var dangerousPermissions = map[string]struct {
	RuleID      string
	Category    string
	Description string
}{
	"android.permission.ACCESS_FINE_LOCATION": {
		RuleID:      RuleLocationPerm,
		Category:    "Location",
		Description: "Fine location access requires prominent disclosure and runtime permission",
	},
	"android.permission.ACCESS_COARSE_LOCATION": {
		RuleID:      RuleLocationPerm,
		Category:    "Location",
		Description: "Coarse location access requires prominent disclosure",
	},
	"android.permission.ACCESS_BACKGROUND_LOCATION": {
		RuleID:      RuleLocationPerm,
		Category:    "Location",
		Description: "Background location requires additional justification for Play Store approval",
	},
	"android.permission.CAMERA": {
		RuleID:      RuleCameraPerm,
		Category:    "Camera",
		Description: "Camera access requires prominent disclosure",
	},
	"android.permission.READ_CONTACTS": {
		RuleID:      RuleContactsPerm,
		Category:    "Contacts",
		Description: "Contacts access requires prominent disclosure and justification",
	},
	"android.permission.WRITE_CONTACTS": {
		RuleID:      RuleContactsPerm,
		Category:    "Contacts",
		Description: "Write contacts access requires prominent disclosure and justification",
	},
	"android.permission.READ_EXTERNAL_STORAGE": {
		RuleID:      RuleStoragePerm,
		Category:    "Storage",
		Description: "Broad storage access; consider using scoped storage APIs instead",
	},
	"android.permission.WRITE_EXTERNAL_STORAGE": {
		RuleID:      RuleStoragePerm,
		Category:    "Storage",
		Description: "Broad storage write access; deprecated in favor of scoped storage",
	},
	"android.permission.MANAGE_EXTERNAL_STORAGE": {
		RuleID:      RuleStoragePerm,
		Category:    "Storage",
		Description: "All-files access requires Play Store policy justification",
	},
	"android.permission.READ_PHONE_STATE": {
		RuleID:      RulePhonePerm,
		Category:    "Phone",
		Description: "Phone state access includes device identifiers; requires justification",
	},
	"android.permission.CALL_PHONE": {
		RuleID:      RulePhonePerm,
		Category:    "Phone",
		Description: "Direct call permission requires prominent disclosure",
	},
	"android.permission.READ_CALL_LOG": {
		RuleID:      RulePhonePerm,
		Category:    "Phone",
		Description: "Call log access is restricted; requires default handler or Play Store exception",
	},
	"android.permission.READ_CALENDAR": {
		RuleID:      RuleCalendarPerm,
		Category:    "Calendar",
		Description: "Calendar read access requires prominent disclosure",
	},
	"android.permission.WRITE_CALENDAR": {
		RuleID:      RuleCalendarPerm,
		Category:    "Calendar",
		Description: "Calendar write access requires prominent disclosure",
	},
	"android.permission.RECORD_AUDIO": {
		RuleID:      RuleDangerousPerm,
		Category:    "Microphone",
		Description: "Microphone access requires prominent disclosure and runtime permission",
	},
	"android.permission.READ_SMS": {
		RuleID:      RuleDangerousPerm,
		Category:    "SMS",
		Description: "SMS read access is restricted; requires default handler or Play Store exception",
	},
	"android.permission.SEND_SMS": {
		RuleID:      RuleDangerousPerm,
		Category:    "SMS",
		Description: "SMS send access is restricted; requires default handler or Play Store exception",
	},
	"android.permission.RECEIVE_SMS": {
		RuleID:      RuleDangerousPerm,
		Category:    "SMS",
		Description: "SMS receive access is restricted; requires default handler or Play Store exception",
	},
	"android.permission.BODY_SENSORS": {
		RuleID:      RuleDangerousPerm,
		Category:    "Sensors",
		Description: "Body sensor access requires health-related disclosure",
	},
}

// MinTargetSDKVersion is the minimum target SDK version required by Play Store.
const MinTargetSDKVersion = 35

// severityForPermission returns the severity for a dangerous permission finding.
func severityForPermission(permName string) preflight.Severity {
	// Restricted permissions (SMS, call log) are critical
	switch permName {
	case "android.permission.READ_SMS",
		"android.permission.SEND_SMS",
		"android.permission.RECEIVE_SMS",
		"android.permission.READ_CALL_LOG",
		"android.permission.MANAGE_EXTERNAL_STORAGE",
		"android.permission.ACCESS_BACKGROUND_LOCATION":
		return preflight.SeverityCritical
	default:
		return preflight.SeverityWarning
	}
}
