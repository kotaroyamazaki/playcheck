package policies

import (
	"testing"
)

func TestLoad(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if db.Version == "" {
		t.Error("expected non-empty version")
	}
	if len(db.AllRules()) < 28 {
		t.Errorf("expected at least 28 rules, got %d", len(db.AllRules()))
	}
}

func TestGetRule(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	r := db.GetRule("DP001")
	if r == nil {
		t.Fatal("GetRule(DP001) returned nil")
	}
	if r.Name != "SMS Permission Usage" {
		t.Errorf("expected SMS Permission Usage, got %s", r.Name)
	}
	if r.Severity != SeverityCritical {
		t.Errorf("expected CRITICAL severity, got %s", r.Severity)
	}
}

func TestGetRulesByCategory(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	dp := db.GetRulesByCategory(CategoryDangerousPermissions)
	if len(dp) == 0 {
		t.Error("expected dangerous_permissions rules, got none")
	}
	for _, r := range dp {
		if r.Category != CategoryDangerousPermissions {
			t.Errorf("rule %s has category %s, expected %s", r.ID, r.Category, CategoryDangerousPermissions)
		}
	}
}

func TestAllRulesHaveRequiredFields(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	for _, r := range db.AllRules() {
		if r.ID == "" {
			t.Error("rule has empty ID")
		}
		if r.Name == "" {
			t.Errorf("rule %s has empty Name", r.ID)
		}
		if r.Severity == "" {
			t.Errorf("rule %s has empty Severity", r.ID)
		}
		if r.Category == "" {
			t.Errorf("rule %s has empty Category", r.ID)
		}
		if r.Description == "" {
			t.Errorf("rule %s has empty Description", r.ID)
		}
		if r.Message == "" {
			t.Errorf("rule %s has empty Message", r.ID)
		}
		if len(r.DetectionPatterns) == 0 {
			t.Errorf("rule %s has no detection patterns", r.ID)
		}
		if r.Remediation == "" {
			t.Errorf("rule %s has empty Remediation", r.ID)
		}
		if r.PolicyLink == "" {
			t.Errorf("rule %s has empty PolicyLink", r.ID)
		}
	}
}

func TestUniqueRuleIDs(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	seen := map[string]bool{}
	for _, r := range db.AllRules() {
		if seen[r.ID] {
			t.Errorf("duplicate rule ID: %s", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestGetRule_NotFound(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	r := db.GetRule("NONEXISTENT")
	if r != nil {
		t.Error("expected nil for nonexistent rule ID")
	}
}

func TestGetRulesByCategory_Empty(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	rules := db.GetRulesByCategory("nonexistent_category")
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for nonexistent category, got %d", len(rules))
	}
}

func TestAllCategories(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	expectedCategories := []string{
		CategoryDangerousPermissions,
		CategoryPrivacyDataSafety,
		CategorySDKCompliance,
		CategoryAccountManagement,
		CategoryManifestValidation,
		CategorySecurity,
	}

	for _, cat := range expectedCategories {
		rules := db.GetRulesByCategory(cat)
		if len(rules) == 0 {
			t.Errorf("expected rules in category %s, got none", cat)
		}
	}
}

func TestValidSeverityValues(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	validSeverities := map[string]bool{
		SeverityCritical: true,
		SeverityError:    true,
		SeverityWarning:  true,
		SeverityInfo:     true,
	}
	for _, r := range db.AllRules() {
		if !validSeverities[r.Severity] {
			t.Errorf("rule %s has invalid severity: %s", r.ID, r.Severity)
		}
	}
}

func TestDetectionPatternTypes(t *testing.T) {
	db, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	validTypes := map[string]bool{
		"manifest_permission": true,
		"manifest_element":    true,
		"manifest_attribute":  true,
		"code_pattern":        true,
		"file_check":          true,
	}
	for _, r := range db.AllRules() {
		for _, dp := range r.DetectionPatterns {
			if !validTypes[dp.Type] {
				t.Errorf("rule %s has invalid detection pattern type: %s", r.ID, dp.Type)
			}
		}
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := parse([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseEmptyRules(t *testing.T) {
	_, err := parse([]byte(`{"version":"1.0.0","rules":[]}`))
	if err == nil {
		t.Error("expected error for empty rules")
	}
}

func TestLoadCaching(t *testing.T) {
	db1, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	db2, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if db1 != db2 {
		t.Error("expected same database object from cached Load()")
	}
}
