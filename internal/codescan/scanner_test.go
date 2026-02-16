package codescan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kotaroyamazaki/playcheck/internal/preflight"
)

func setupTestDir(t *testing.T, files map[string]string) string {
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

func TestScanner_ID(t *testing.T) {
	s := NewScanner()
	if s.ID() != "code-scan" {
		t.Errorf("expected ID code-scan, got %s", s.ID())
	}
}

func TestScanner_Name(t *testing.T) {
	s := NewScanner()
	if s.Name() == "" {
		t.Error("Name should not be empty")
	}
}

func TestScanner_Description(t *testing.T) {
	s := NewScanner()
	if s.Description() == "" {
		t.Error("Description should not be empty")
	}
}

func TestScanner_Run_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !result.Passed {
		t.Error("expected Passed=true for empty project")
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}

func TestScanner_Run_HTTPDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Main.java": `package com.example;
import java.net.URL;
public class Main {
    String url = "http://insecure.example.com/api";
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleHTTPUsage {
			found = true
			if f.Severity != preflight.SeverityError {
				t.Errorf("HTTP finding severity should be ERROR, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected CS001 (HTTP usage) finding")
	}
}

func TestScanner_Run_SMSDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"SmsSender.java": `package com.example;
import android.telephony.SmsManager;
public class SmsSender {
    public void send() {
        SmsManager sms = SmsManager.getDefault();
        sms.sendTextMessage("+1234567890", null, "Hello", null, null);
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleSMSUsage {
			found = true
			if f.Severity != preflight.SeverityCritical {
				t.Errorf("SMS finding severity should be CRITICAL, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected CS008 (SMS usage) finding")
	}
}

func TestScanner_Run_LocationDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Tracker.kt": `package com.example
import android.location.LocationManager
class Tracker(val ctx: android.content.Context) {
    fun track() {
        val lm = ctx.getSystemService("location") as LocationManager
        lm.requestLocationUpdates("gps", 1000L, 10f, listener)
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleLocationUsage {
			found = true
		}
	}
	if !found {
		t.Error("expected CS009 (location usage) finding")
	}
}

func TestScanner_Run_CameraDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"CameraActivity.java": `package com.example;
import android.hardware.camera2.CameraManager;
public class CameraActivity {
    public void open() {
        CameraManager manager = getSystemService(CameraManager.class);
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleCameraUsage {
			found = true
		}
	}
	if !found {
		t.Error("expected CS010 (camera usage) finding")
	}
}

func TestScanner_Run_AdvertisingIDDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"AdHelper.java": `package com.example;
public class AdHelper {
    public void init() {
        AdvertisingIdClient.getAdvertisingIdInfo(context);
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleAdIDUsage {
			found = true
		}
	}
	if !found {
		t.Error("expected CS005 (advertising ID) finding")
	}
}

func TestScanner_Run_AccountCreationDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"AuthService.kt": `package com.example
class AuthService {
    fun createAccount(email: String, password: String) {
        repo.createUser(email, password)
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleAccountCreation {
			found = true
		}
	}
	if !found {
		t.Error("expected CS006 (account creation) finding")
	}
}

func TestScanner_Run_WebViewJSDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"WebActivity.java": `package com.example;
public class WebActivity {
    public void setup(android.webkit.WebView wv) {
        wv.getSettings().setJavaScriptEnabled(true);
        wv.addJavascriptInterface(new JsInterface(), "android");
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleWebViewJS {
			found = true
		}
	}
	if !found {
		t.Error("expected CS012 (WebView JS) finding")
	}
}

func TestScanner_Run_CryptoDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Crypto.java": `package com.example;
import javax.crypto.Cipher;
public class Crypto {
    public void weak() {
        Cipher c = Cipher.getInstance("DES/ECB/PKCS5Padding");
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleCryptoUsage {
			found = true
			if f.Severity != preflight.SeverityError {
				t.Errorf("crypto finding severity should be ERROR, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected CS011 (weak crypto) finding")
	}
}

func TestScanner_Run_CleanProject(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Main.java": `package com.example;
import javax.net.ssl.HttpsURLConnection;
public class Main {
    private static final String URL = "https://secure.example.com/api";
    public void fetch() {
        java.net.URL url = new java.net.URL(URL);
        HttpsURLConnection conn = (HttpsURLConnection) url.openConnection();
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Clean file should have no critical/error findings
	for _, f := range result.Findings {
		if f.Severity >= preflight.SeverityError {
			t.Errorf("unexpected error+ finding in clean project: %s - %s", f.CheckID, f.Title)
		}
	}
}

func TestScanner_Run_SkipsComments(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Main.java": `package com.example;
public class Main {
    // SmsManager is not used here
    // sendTextMessage is just a comment
    /* CameraManager is documented */
    * this is a continuation of a block comment
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, f := range result.Findings {
		if f.CheckID == RuleSMSUsage || f.CheckID == RuleCameraUsage {
			t.Errorf("should not match in comments: %s at line %d", f.CheckID, f.Location.Line)
		}
	}
}

func TestScanner_Run_MaxMatchesPerRule(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Many.java": `package com.example;
public class Many {
    String u1 = "http://example.com/1";
    String u2 = "http://example.com/2";
    String u3 = "http://example.com/3";
    String u4 = "http://example.com/4";
    String u5 = "http://example.com/5";
    String u6 = "http://example.com/6";
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	httpCount := 0
	for _, f := range result.Findings {
		if f.CheckID == RuleHTTPUsage {
			httpCount++
		}
	}
	// Should be capped at maxMatchesPerRule (3) per file
	if httpCount > 3 {
		t.Errorf("expected at most 3 HTTP findings per file, got %d", httpCount)
	}
}

func TestScanner_Run_LineNumbers(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"Lines.java": `package com.example;
public class Lines {
    String a = "safe";
    String b = "http://example.com";
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, f := range result.Findings {
		if f.CheckID == RuleHTTPUsage {
			if f.Location.Line != 4 {
				t.Errorf("expected HTTP finding at line 4, got %d", f.Location.Line)
			}
			if f.Location.File != "Lines.java" {
				t.Errorf("expected file Lines.java, got %s", f.Location.File)
			}
		}
	}
}

func TestScanner_Run_FacebookSDKDetection(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"SocialLogin.kt": `package com.example
class SocialLogin {
    fun init() {
        FacebookSdk.sdkInitialize(context)
        AppEventsLogger.activateApp(application)
    }
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleFacebookSDK {
			found = true
		}
	}
	if !found {
		t.Error("expected CS013 (Facebook SDK) finding")
	}
}

func TestScanner_Run_KotlinFiles(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"App.kt": `package com.example
fun main() {
    val url = "http://insecure.example.com"
}`,
	})

	s := NewScanner()
	result, err := s.Run(dir)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.CheckID == RuleHTTPUsage {
			found = true
		}
	}
	if !found {
		t.Error("expected HTTP finding in Kotlin file")
	}
}

func TestNewScanner_CompilesAllRules(t *testing.T) {
	s := NewScanner()
	if len(s.compiled) == 0 {
		t.Fatal("expected compiled rules, got 0")
	}
	if len(s.compiled) < len(codeRules) {
		t.Errorf("expected at least %d compiled rules, got %d", len(codeRules), len(s.compiled))
	}
}

func TestCompileRules(t *testing.T) {
	rules := []codeRule{
		{
			ID:       "TEST001",
			Patterns: []string{`validRegex`},
		},
		{
			ID:       "TEST002",
			Patterns: []string{`[invalid`}, // invalid regex
		},
	}

	compiled := compileRules(rules)
	// TEST001 should compile, TEST002 should be skipped
	if len(compiled) != 1 {
		t.Errorf("expected 1 compiled rule (invalid skipped), got %d", len(compiled))
	}
}

func TestCompilePattern_Caching(t *testing.T) {
	pattern := `test\d+`
	re1, err := compilePattern(pattern)
	if err != nil {
		t.Fatal(err)
	}
	re2, err := compilePattern(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if re1 != re2 {
		t.Error("expected same regex object from cache")
	}
}
