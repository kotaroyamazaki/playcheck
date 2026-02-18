# Policy Rules Reference

Complete reference for all playcheck rules. Use this when helping users fix specific findings.

## Dangerous Permissions

### DP001: SMS Permission Usage
- Severity: CRITICAL
- Detects: READ_SMS, SEND_SMS, RECEIVE_SMS, WRITE_SMS permissions
- Fix: Remove SMS permissions unless the app is the default SMS handler. Submit Permissions Declaration Form if required.
- Policy: https://support.google.com/googleplay/android-developer/answer/9047303

### DP002: Call Log Permission Usage
- Severity: CRITICAL
- Detects: READ_CALL_LOG, WRITE_CALL_LOG, PROCESS_OUTGOING_CALLS permissions
- Fix: Remove call log permissions unless the app is the default phone handler. Submit Permissions Declaration Form if required.
- Policy: https://support.google.com/googleplay/android-developer/answer/9047303

### DP003: Location in Background Permission
- Severity: CRITICAL
- Detects: ACCESS_BACKGROUND_LOCATION permission
- Fix: Use foreground location instead. If background location is essential, submit a Permissions Declaration Form explaining the user benefit.
- Policy: https://support.google.com/googleplay/android-developer/answer/9799150

### DP004: Camera Permission Without Usage
- Severity: WARNING
- Detects: CAMERA permission declared but no camera API usage in code
- Fix: Remove CAMERA permission if camera functionality is not implemented.
- Policy: https://support.google.com/googleplay/android-developer/answer/9888170

### DP005: Storage Permission (Broad Access)
- Severity: ERROR
- Detects: MANAGE_EXTERNAL_STORAGE permission
- Fix: Use scoped storage APIs (MediaStore, Storage Access Framework) instead. Submit declaration form if broad access is essential.
- Policy: https://support.google.com/googleplay/android-developer/answer/9956427

### DP006: Exact Alarm Permission
- Severity: WARNING
- Detects: SCHEDULE_EXACT_ALARM permission
- Fix: Use inexact alarms (setAndAllowWhileIdle, setWindow) unless the app is an alarm clock, timer, or calendar.
- Policy: https://support.google.com/googleplay/android-developer/answer/12253906

### DP007: Query All Packages Permission
- Severity: WARNING
- Detects: QUERY_ALL_PACKAGES permission
- Fix: Use targeted package visibility with `<queries>` element instead.
- Policy: https://support.google.com/googleplay/android-developer/answer/10158779

### DP008: Accessibility Service Permission
- Severity: CRITICAL
- Detects: BIND_ACCESSIBILITY_SERVICE or AccessibilityService usage
- Fix: Only use AccessibilityService for genuine accessibility purposes. Provide clear documentation for review.
- Policy: https://support.google.com/googleplay/android-developer/answer/10964491

### DP009: VPN Service Permission
- Severity: ERROR
- Detects: BIND_VPN_SERVICE or VpnService usage
- Fix: Ensure VPN is the app's core purpose. Disclose all data handling in the listing and privacy policy.
- Policy: https://support.google.com/googleplay/android-developer/answer/9888170

### DP010: Foreground Service Type Missing
- Severity: ERROR
- Detects: Foreground service without foregroundServiceType attribute (required for Android 14+)
- Fix: Add `android:foregroundServiceType` to all foreground `<service>` elements. Valid types: camera, connectedDevice, dataSync, health, location, mediaPlayback, mediaProjection, microphone, phoneCall, remoteMessaging, shortService, specialUse, systemExempted.
- Policy: https://developer.android.com/about/versions/14/changes/foreground-service-types

## Privacy & Data Safety

### PDS001: Missing Privacy Policy
- Severity: CRITICAL
- Detects: No privacy policy URL in code or manifest metadata
- Fix: Add a privacy policy accessible from within the app and from the Play Store listing. Disclose data collection, usage, and sharing.
- Policy: https://support.google.com/googleplay/android-developer/answer/9859455

### PDS002: Data Collection Without Disclosure
- Severity: ERROR
- Detects: Device ID, IMEI, subscriber ID, advertising ID, contacts access without disclosure
- Fix: Add a prominent in-app disclosure before collecting user data explaining what, why, and how.
- Policy: https://support.google.com/googleplay/android-developer/answer/10144311

### PDS003: Data Safety Section Mismatch
- Severity: ERROR
- Detects: Location tracking, Firebase Analytics, or stored personal data that may not be declared in Data Safety
- Fix: Update the Data Safety section in Play Console to accurately reflect all collected/shared/processed data.
- Policy: https://support.google.com/googleplay/android-developer/answer/10787469

### PDS004: Missing Data Deletion Mechanism
- Severity: WARNING
- Detects: No account/data deletion functionality found in code
- Fix: Implement in-app and web-based data deletion. Provide the URL in the Data Safety section.
- Policy: https://support.google.com/googleplay/android-developer/answer/13327111

## SDK Compliance

### SDK001: Outdated Target SDK Version
- Severity: CRITICAL
- Detects: targetSdkVersion below 35
- Fix: Update targetSdkVersion in build.gradle to at least API level 35 (Android 15). Test compatibility.
- Policy: https://support.google.com/googleplay/android-developer/answer/11926878

### SDK002: Missing Play Core Library Update
- Severity: WARNING
- Detects: Play Core library version below 1.10.0
- Fix: Update Play Core library to 1.10.0+ in build.gradle dependencies.
- Policy: https://developer.android.com/reference/com/google/android/play/core/release-notes

### SDK003: Missing Ads SDK Consent Integration
- Severity: ERROR
- Detects: AdMob, Facebook Ads, Unity Ads, or AppLovin without consent management
- Fix: Integrate Google UMP SDK or equivalent consent management. Initialize consent before loading ads.
- Policy: https://support.google.com/googleplay/android-developer/answer/11112578

### SDK004: Deprecated API Usage
- Severity: WARNING
- Detects: getRunningTasks, getRunningAppProcesses, GET_SIGNATURES, setJavaScriptEnabled
- Fix: Replace deprecated API calls with modern equivalents per Android API reference.
- Policy: https://developer.android.com/distribute/best-practices/develop/target-sdk

## Account Management

### AD001: Missing Account Deletion Option
- Severity: CRITICAL
- Detects: Account creation code (createUser, signUp, register) without matching deletion flow
- Fix: Implement in-app account deletion. Provide web-based deletion option as well.
- Policy: https://support.google.com/googleplay/android-developer/answer/13327111

### AD002: Login Without Data Safety Disclosure
- Severity: WARNING
- Detects: Login/authentication code without Data Safety disclosure
- Fix: Declare authentication-related data (email, name, user IDs) in Play Console Data Safety section.
- Policy: https://support.google.com/googleplay/android-developer/answer/10787469

## Manifest Validation

### MV001: Missing App Icon
- Severity: ERROR
- Detects: `<application>` element missing `android:icon` attribute
- Fix: Add `android:icon` to the `<application>` element pointing to launcher icon resource.
- Policy: https://developer.android.com/guide/topics/manifest/application-element

### MV002: Debuggable Build
- Severity: CRITICAL
- Detects: `android:debuggable="true"` in manifest
- Fix: Remove `android:debuggable="true"` or ensure it is false for release builds. Use Gradle build types.
- Policy: https://developer.android.com/guide/topics/manifest/application-element#debug

### MV003: Missing Version Code
- Severity: ERROR
- Detects: Missing `android:versionCode` in manifest
- Fix: Add `android:versionCode` to `<manifest>` or set it in build.gradle.
- Policy: https://developer.android.com/guide/topics/manifest/manifest-element

### MV004: Backup Rules Missing
- Severity: WARNING
- Detects: Missing `android:dataExtractionRules` or `android:fullBackupContent`
- Fix: Add `android:dataExtractionRules` (Android 12+) and `android:fullBackupContent` (older) to `<application>`.
- Policy: https://developer.android.com/guide/topics/data/autobackup

### MV005: Intent Filter Without BROWSABLE Category
- Severity: INFO
- Detects: Deep link intent filter with VIEW action but no BROWSABLE category
- Fix: Add `<category android:name="android.intent.category.BROWSABLE"/>` to deep link intent filters.
- Policy: https://developer.android.com/training/app-links/deep-linking

## Security

### MS001: Insecure Network Communication
- Severity: ERROR
- Detects: `usesCleartextTraffic="true"`, HTTP URLs in code, cleartext traffic in network security config
- Fix: Set `android:usesCleartextTraffic="false"`. Use HTTPS for all requests. Configure network security config for exceptions.
- Policy: https://developer.android.com/privacy-and-security/security-config

### MS002: Hardcoded Secrets or API Keys
- Severity: CRITICAL
- Detects: API keys, secret keys, passwords, private keys, Firebase credentials in source code
- Fix: Move secrets to environment variables, secrets manager, or Android encrypted SharedPreferences. Never commit secrets.
- Policy: https://support.google.com/googleplay/android-developer/answer/9848633

### MS003: Exported Components Without Protection
- Severity: ERROR
- Detects: Activities, services, or receivers with `exported="true"` but no `android:permission`
- Fix: Add `android:permission` to exported components, or set `android:exported="false"` if external access is not needed.
- Policy: https://developer.android.com/topic/security/best-practices

### MS004: WebView JavaScript Interface Vulnerability
- Severity: ERROR
- Detects: `addJavascriptInterface()`, `setAllowFileAccessFromFileURLs(true)`, `setAllowUniversalAccessFromFileURLs(true)`
- Fix: Avoid addJavascriptInterface with untrusted content. Use WebMessagePort. Disable file access from URLs.
- Policy: https://developer.android.com/privacy-and-security/risks/webview-javascript

## Monetization

### MP002: Non-Play Billing for Digital Goods
- Severity: CRITICAL
- Detects: Stripe, Braintree, PayPal, Razorpay SDK usage for potential digital goods purchases
- Fix: Use Google Play Billing Library for digital goods and subscriptions. Third-party payment is only allowed for physical goods.
- Policy: https://support.google.com/googleplay/android-developer/answer/9858738

## Content Policy

### MC001: Missing Content Rating
- Severity: WARNING
- Detects: WebView loading external URLs, JavaScript interfaces that could serve unrated content
- Fix: Filter WebView content appropriately. Complete content rating questionnaire in Play Console accurately.
- Policy: https://support.google.com/googleplay/android-developer/answer/9859455
