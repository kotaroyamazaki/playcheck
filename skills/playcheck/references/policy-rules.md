# Policy Rules Reference

Complete reference for all playcheck rules, organized by check_id as they appear in scan output.

## Manifest Scanner Rules

### SDK001: Target SDK Version
- Severity: CRITICAL
- Detects: Missing or outdated targetSdkVersion (minimum required: 35)
- Fix: Set targetSdkVersion to 35 or higher in build.gradle or AndroidManifest.xml.
- Note: This ID is also emitted by the Data Safety checker for third-party SDK disclosure (see below).

### DP001: Restricted Dangerous Permission
- Severity: CRITICAL (SMS, RECORD_AUDIO) or WARNING (others)
- Detects: RECORD_AUDIO, READ_SMS, SEND_SMS, RECEIVE_SMS, BODY_SENSORS permissions in manifest
- Fix: Remove unused restricted permissions. Submit Permissions Declaration Form if required.

### DP002: Location Permission
- Severity: CRITICAL (ACCESS_BACKGROUND_LOCATION) or WARNING (fine/coarse location)
- Detects: ACCESS_FINE_LOCATION, ACCESS_COARSE_LOCATION, ACCESS_BACKGROUND_LOCATION
- Fix: Add prominent disclosure for location usage. Use foreground-only location if possible. Submit declaration form for background location.

### DP003: Camera Permission
- Severity: WARNING
- Detects: CAMERA permission in manifest
- Fix: Ensure camera usage is disclosed in Data Safety form. Request runtime permission.

### DP004: Contacts Permission
- Severity: WARNING
- Detects: READ_CONTACTS, WRITE_CONTACTS permissions in manifest
- Fix: Ensure contacts access is disclosed in Data Safety form. Add prominent disclosure.

### DP005: Storage Permission
- Severity: CRITICAL (MANAGE_EXTERNAL_STORAGE) or WARNING (READ/WRITE_EXTERNAL_STORAGE)
- Detects: READ_EXTERNAL_STORAGE, WRITE_EXTERNAL_STORAGE, MANAGE_EXTERNAL_STORAGE
- Fix: Use scoped storage APIs (MediaStore, Storage Access Framework) instead. Submit declaration form for broad access.

### DP006: Phone Permission
- Severity: CRITICAL (READ_CALL_LOG) or WARNING (READ_PHONE_STATE, CALL_PHONE)
- Detects: READ_PHONE_STATE, CALL_PHONE, READ_CALL_LOG permissions in manifest
- Fix: Remove unused phone permissions. Submit Permissions Declaration Form for call log access.
- Note: This ID is also emitted by the Data Safety checker for background location (see below).

### DP007: Calendar Permission
- Severity: WARNING
- Detects: READ_CALENDAR, WRITE_CALENDAR permissions in manifest
- Fix: Ensure calendar data access is disclosed in Data Safety form.

### MV001: Missing android:exported Attribute
- Severity: ERROR
- Detects: Activities, services, receivers, or providers with intent-filters but no android:exported attribute (required since Android 12 / API 31)
- Fix: Add `android:exported="true"` or `android:exported="false"` to components with intent-filters.

### MV002: No Launcher Activity
- Severity: WARNING
- Detects: No activity with ACTION_MAIN + CATEGORY_LAUNCHER intent-filter
- Fix: Add an intent-filter with action MAIN and category LAUNCHER to the main activity.

### MV004: Cleartext Traffic Enabled
- Severity: ERROR (explicitly enabled) or WARNING (enabled by default on targetSdk < 28)
- Detects: `android:usesCleartextTraffic="true"` or implicit cleartext on older SDK targets
- Fix: Set `android:usesCleartextTraffic="false"`. Use HTTPS for all network requests. Use Network Security Config for exceptions.

### MC001: Exported Component Security
- Severity: INFO
- Detects: Components with `android:exported="true"` that are accessible to other apps
- Fix: Review exported components to ensure they don't expose sensitive functionality. Add permission protection if needed.

## Code Scanner Rules

### CS001: HTTP URL Detected
- Severity: ERROR
- Detects: Hardcoded `http://` URLs and HttpURLConnection usage in Kotlin/Java code
- Fix: Replace `http://` URLs with `https://` to ensure encrypted data transmission.

### CS002: Privacy Policy URL Found
- Severity: INFO
- Detects: Privacy policy references in code (informational, confirms policy exists)
- Fix: Verify the privacy policy URL is accessible, up to date, and covers all Data Safety disclosures.

### CS003: Firebase Analytics Usage
- Severity: WARNING
- Detects: FirebaseAnalytics, logEvent, setAnalyticsCollectionEnabled
- Fix: Disclose Firebase Analytics data collection in Data Safety form: App interactions, Device or other IDs, Diagnostics.

### CS004: AdMob SDK Usage
- Severity: WARNING
- Detects: AdRequest, AdView, InterstitialAd, RewardedAd, MobileAds.initialize
- Fix: Disclose AdMob data collection in Data Safety form: Advertising ID, approximate location, device info.

### CS005: Advertising ID Usage
- Severity: WARNING
- Detects: AdvertisingIdClient, getAdvertisingIdInfo, advertisingId
- Fix: Disclose advertising ID in Data Safety form under "Device or other IDs". Ensure privacy policy exists.

### CS006: Account Creation Pattern
- Severity: WARNING
- Detects: createAccount, signUp, registerUser, createUser, createUserWithEmailAndPassword
- Fix: Ensure an in-app account deletion option exists (required by Play Store policy).

### CS007: Account Deletion Pattern
- Severity: INFO
- Detects: deleteAccount, removeAccount, deleteUser, accountDeletion (informational, confirms deletion exists)
- Fix: Verify the deletion flow is easy to find and deletes all user data or discloses retention.

### CS008: SMS API Usage in Code
- Severity: CRITICAL
- Detects: SmsManager, sendTextMessage, sendMultipartTextMessage
- Fix: Remove direct SMS API usage. Use Firebase Auth Phone verification or SMS Retriever API instead.

### CS009: Location API Usage
- Severity: WARNING
- Detects: FusedLocationProviderClient, LocationManager, requestLocationUpdates, getLastKnownLocation
- Fix: Disclose location usage in Data Safety form. Add prominent disclosure before requesting location permission.

### CS010: Camera API Usage
- Severity: WARNING
- Detects: CameraManager, CameraX, Camera2, CameraDevice, ACTION_IMAGE_CAPTURE
- Fix: Disclose camera usage in Data Safety form. Request camera permission at runtime with context.

### CS011: Weak Cryptography
- Severity: ERROR
- Detects: DES cipher usage, MD5 for security, SHA-1 for security-critical operations
- Fix: Use AES-256, RSA-2048+. Replace DES, MD5, and SHA-1 for security-critical operations.

### CS012: WebView JavaScript Enabled
- Severity: WARNING
- Detects: setJavaScriptEnabled(true), addJavascriptInterface
- Fix: Validate all URLs loaded in WebView. Implement content security. Consider SafeBrowsing API.

### CS013: Facebook SDK Usage
- Severity: WARNING
- Detects: com.facebook, FacebookSdk, AppEventsLogger, LoginManager
- Fix: Disclose Facebook SDK data collection in Data Safety form. Review Facebook docs for full disclosure requirements.

### CS014: Third-Party Tracking SDK
- Severity: WARNING
- Detects: Adjust, AppsFlyer, Amplitude, Mixpanel, Segment, Braze, Crashlytics SDKs
- Fix: Disclose all third-party SDK data collection in Data Safety form.

## Data Safety Checker Rules

### PDS001: Privacy Policy Not Found
- Severity: ERROR
- Detects: No privacy policy URL in AndroidManifest.xml or string resources
- Fix: Add a privacy policy URL via `<meta-data>` tag in manifest or include it in string resources. Policy must cover all data collection.

### PDS002: Permission Requires Data Safety Disclosure
- Severity: WARNING
- Detects: Dangerous permissions (SMS, call log, contacts, location, camera, audio, storage, calendar, sensors) that require Data Safety form entries
- Fix: Declare the corresponding data type in your Play Console Data Safety form.

### PDS003: Data Collection Without Consent
- Severity: WARNING
- Detects: Data collection APIs (getDeviceId, getAdvertisingIdInfo, ANDROID_ID, getAccounts, location APIs) in files without consent-related code
- Fix: Implement a consent dialog. Obtain user consent before collecting personal data.

### PDS004: No Runtime Permission Request
- Severity: ERROR
- Detects: Dangerous permissions declared in manifest but no requestPermissions/checkSelfPermission calls found in code
- Fix: Implement runtime permission requests using ActivityCompat.requestPermissions() or the Activity Result API.

### AD001: Account Deletion Not Found
- Severity: ERROR
- Detects: Account creation patterns in code without corresponding account deletion functionality
- Fix: Implement in-app account deletion. See https://support.google.com/googleplay/android-developer/answer/13327111

### SDK001 (Data Safety): Third-Party SDK Disclosure
- Severity: WARNING
- Detects: Third-party SDKs in build.gradle (Firebase Analytics, Crashlytics, AdMob, Facebook, Adjust, AppsFlyer, Sentry, Maps, Mixpanel, Amplitude, Braze, OneSignal, Stripe)
- Fix: Declare data collection by the detected SDK in your Play Console Data Safety form.
- Note: Shares ID with the manifest scanner's Target SDK Version check. Distinguish by context: manifest findings reference targetSdkVersion; data safety findings reference specific SDK dependencies.

### SDK004 (Data Safety): Unused Permission in Code
- Severity: WARNING
- Detects: Dangerous permissions declared in manifest but no corresponding API usage in Kotlin/Java code
- Fix: Remove the unused permission from manifest, or verify it is used by a library dependency.

### DP006 (Data Safety): Background Location Access
- Severity: ERROR
- Detects: ACCESS_BACKGROUND_LOCATION in manifest, with additional check for missing foreground location permission
- Fix: Add prominent in-app disclosure for background location. Submit permission declaration form in Play Console. Ensure foreground location permission is also present.
- Note: Shares ID with the manifest scanner's Phone Permission check. Distinguish by context: manifest findings reference phone/call permissions; data safety findings reference background location.
