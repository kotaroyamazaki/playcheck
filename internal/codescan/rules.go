package codescan

import "github.com/yourusername/playcheck/internal/preflight"

// Rule IDs for code scanning checks.
const (
	RuleHTTPUsage         = "CS001"
	RulePrivacyPolicy     = "CS002"
	RuleFirebaseAnalytics = "CS003"
	RuleAdMob             = "CS004"
	RuleAdIDUsage         = "CS005"
	RuleAccountCreation   = "CS006"
	RuleAccountDeletion   = "CS007"
	RuleSMSUsage          = "CS008"
	RuleLocationUsage     = "CS009"
	RuleCameraUsage       = "CS010"
	RuleCryptoUsage       = "CS011"
	RuleWebViewJS         = "CS012"
	RuleFacebookSDK       = "CS013"
	RuleThirdPartyTracker = "CS014"
)

// codeRule describes a single code scanning rule with its detection pattern.
type codeRule struct {
	ID          string
	Title       string
	Description string
	Severity    preflight.Severity
	Suggestion  string
	Patterns    []string // regex patterns
}

// codeRules is the list of all code scanning rules.
var codeRules = []codeRule{
	{
		ID:          RuleHTTPUsage,
		Title:       "Unencrypted HTTP URL detected",
		Description: "Code contains a hardcoded HTTP URL. Unencrypted data transmission can expose user data and violate Play Store security policies.",
		Severity:    preflight.SeverityError,
		Suggestion:  "Replace http:// URLs with https:// to ensure encrypted data transmission.",
		Patterns: []string{
			`"http://[^"]+?"`,
			`'http://[^']+?'`,
			`\bHttpURLConnection\b`,
		},
	},
	{
		ID:          RulePrivacyPolicy,
		Title:       "Privacy policy URL found",
		Description: "A privacy policy URL was detected in the code. Verify it is accessible and up to date.",
		Severity:    preflight.SeverityInfo,
		Suggestion:  "Ensure the privacy policy URL is accessible, up to date, and covers all data collection disclosed in the Data Safety section.",
		Patterns: []string{
			`(?i)privacy[_\-\s]?policy`,
			`(?i)privacypolicy`,
		},
	},
	{
		ID:          RuleFirebaseAnalytics,
		Title:       "Firebase Analytics SDK usage detected",
		Description: "Firebase Analytics collects data that must be disclosed in the Data Safety section, including app interactions, device identifiers, and diagnostics.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Disclose Firebase Analytics data collection in your Data Safety form. Include: App interactions, Device or other IDs, and Diagnostics.",
		Patterns: []string{
			`com\.google\.firebase\.analytics`,
			`FirebaseAnalytics`,
			`logEvent\s*\(`,
			`setAnalyticsCollectionEnabled`,
		},
	},
	{
		ID:          RuleAdMob,
		Title:       "AdMob SDK usage detected",
		Description: "Google AdMob collects advertising data that must be disclosed in the Data Safety section.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Disclose AdMob data collection in your Data Safety form. Include: Advertising ID, approximate location (if enabled), and device info.",
		Patterns: []string{
			`com\.google\.android\.gms\.ads`,
			`AdRequest`,
			`AdView`,
			`InterstitialAd`,
			`RewardedAd`,
			`MobileAds\.initialize`,
		},
	},
	{
		ID:          RuleAdIDUsage,
		Title:       "Advertising ID usage detected",
		Description: "The app accesses the advertising ID, which must be disclosed in the Data Safety section. Advertising ID policies require a privacy policy.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Disclose advertising ID collection in your Data Safety form under 'Device or other IDs'. Ensure you have a privacy policy.",
		Patterns: []string{
			`AdvertisingIdClient`,
			`getAdvertisingIdInfo`,
			`advertisingId`,
			`ADVERTISING_ID`,
		},
	},
	{
		ID:          RuleAccountCreation,
		Title:       "Account creation pattern detected",
		Description: "Account creation flows must comply with Play Store account deletion requirements. Apps that support account creation must also provide account deletion.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Ensure your app provides an in-app account deletion option as required by Play Store policy. The deletion flow must be easy to find and complete.",
		Patterns: []string{
			`(?i)createAccount`,
			`(?i)signUp\s*\(`,
			`(?i)registerUser`,
			`(?i)createUser`,
			`FirebaseAuth\.getInstance\(\)\.createUser`,
		},
	},
	{
		ID:          RuleAccountDeletion,
		Title:       "Account deletion pattern detected",
		Description: "Account deletion support was detected. Verify the deletion flow meets Play Store requirements for completeness and accessibility.",
		Severity:    preflight.SeverityInfo,
		Suggestion:  "Ensure account deletion deletes all user data or clearly discloses data retention policies. The deletion option must be accessible in-app.",
		Patterns: []string{
			`(?i)deleteAccount`,
			`(?i)removeAccount`,
			`(?i)deleteUser`,
			`(?i)accountDeletion`,
		},
	},
	{
		ID:          RuleSMSUsage,
		Title:       "SMS API usage detected in code",
		Description: "Direct SMS API usage is restricted by Play Store. Only apps with default handler status or approved exceptions may use SMS APIs.",
		Severity:    preflight.SeverityCritical,
		Suggestion:  "Remove direct SMS API usage unless your app is a default SMS handler. Use alternative verification methods like Firebase Auth Phone verification or SMS Retriever API.",
		Patterns: []string{
			`SmsManager`,
			`sendTextMessage\s*\(`,
			`sendMultipartTextMessage`,
		},
	},
	{
		ID:          RuleLocationUsage,
		Title:       "Location API usage detected in code",
		Description: "Location access in code must be disclosed in the Data Safety section and requires runtime permission with prominent disclosure.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Ensure location usage is disclosed in your Data Safety form. Provide prominent disclosure before requesting location permission at runtime.",
		Patterns: []string{
			`FusedLocationProviderClient`,
			`LocationManager`,
			`requestLocationUpdates`,
			`getLastKnownLocation`,
			`getLastLocation`,
			`LocationRequest`,
		},
	},
	{
		ID:          RuleCameraUsage,
		Title:       "Camera API usage detected in code",
		Description: "Camera access in code must be disclosed in the Data Safety section and requires runtime permission.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Ensure camera usage is disclosed in your Data Safety form. Request camera permission at runtime with clear context.",
		Patterns: []string{
			`CameraManager`,
			`CameraX`,
			`Camera2`,
			`camera\.open`,
			`CameraDevice`,
			`MediaStore\.ACTION_IMAGE_CAPTURE`,
		},
	},
	{
		ID:          RuleCryptoUsage,
		Title:       "Weak cryptography usage detected",
		Description: "Usage of weak or deprecated cryptographic algorithms was detected. Play Store security policies recommend strong encryption.",
		Severity:    preflight.SeverityError,
		Suggestion:  "Use strong encryption algorithms (AES-256, RSA-2048+). Replace DES, MD5, and SHA-1 for security-critical operations.",
		Patterns: []string{
			`DES/`,
			`"DES"`,
			`Cipher\.getInstance\(\s*"DES`,
			`MessageDigest\.getInstance\(\s*"MD5"`,
			`MessageDigest\.getInstance\(\s*"SHA-1"`,
		},
	},
	{
		ID:          RuleWebViewJS,
		Title:       "WebView JavaScript enabled",
		Description: "WebView with JavaScript enabled can be a security risk. Ensure proper URL validation and content security policies are in place.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Validate all URLs loaded in WebView. Implement proper content security and consider using SafeBrowsing API.",
		Patterns: []string{
			`setJavaScriptEnabled\s*\(\s*true\s*\)`,
			`addJavascriptInterface`,
		},
	},
	{
		ID:          RuleFacebookSDK,
		Title:       "Facebook SDK usage detected",
		Description: "Facebook SDK collects data that must be disclosed in the Data Safety section, including device info and app events.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Disclose Facebook SDK data collection in your Data Safety form. Review Facebook's data collection documentation for full disclosure requirements.",
		Patterns: []string{
			`com\.facebook\.`,
			`FacebookSdk`,
			`AppEventsLogger`,
			`LoginManager`,
		},
	},
	{
		ID:          RuleThirdPartyTracker,
		Title:       "Third-party tracking SDK detected",
		Description: "A third-party tracking or analytics SDK was detected. Data collection must be disclosed in the Data Safety section.",
		Severity:    preflight.SeverityWarning,
		Suggestion:  "Disclose all third-party SDK data collection in your Data Safety form. Review each SDK's documentation for specific data types collected.",
		Patterns: []string{
			`com\.adjust\.sdk`,
			`com\.appsflyer`,
			`com\.amplitude`,
			`com\.mixpanel`,
			`com\.segment\.analytics`,
			`com\.braze`,
			`com\.crashlytics`,
		},
	},
}
