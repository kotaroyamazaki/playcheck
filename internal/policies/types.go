package policies

// Severity levels for policy rules.
const (
	SeverityCritical = "CRITICAL"
	SeverityError    = "ERROR"
	SeverityWarning  = "WARNING"
	SeverityInfo     = "INFO"
)

// Category groups for policy rules.
const (
	CategoryDangerousPermissions = "dangerous_permissions"
	CategoryPrivacyDataSafety    = "privacy_data_safety"
	CategorySDKCompliance        = "sdk_compliance"
	CategoryAccountManagement    = "account_management"
	CategoryManifestValidation   = "manifest_validation"
	CategoryContentPolicy        = "content_policy"
	CategoryMonetization         = "monetization"
	CategorySecurity             = "security"
)

// DetectionPattern describes how to detect a policy violation.
type DetectionPattern struct {
	Type    string `json:"type"`    // "manifest_permission", "manifest_element", "code_pattern", "file_check", "manifest_attribute"
	Value   string `json:"value"`   // The value to match (permission name, regex, XPath, etc.)
	Context string `json:"context"` // Additional context for the match (e.g., file type filter)
}

// Rule represents a single Google Play Store compliance rule.
type Rule struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Severity          string             `json:"severity"`
	Category          string             `json:"category"`
	Description       string             `json:"description"`
	Message           string             `json:"message"`
	DetectionPatterns []DetectionPattern  `json:"detection_patterns"`
	Remediation       string             `json:"remediation"`
	PolicyLink        string             `json:"policy_link"`
	Metadata          map[string]string  `json:"metadata,omitempty"`
}

// PolicyDatabase holds all compliance rules loaded from the embedded JSON.
type PolicyDatabase struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`

	// Indexes built at load time for fast lookup.
	byID       map[string]*Rule
	byCategory map[string][]*Rule
}

// GetRule returns a rule by its ID, or nil if not found.
func (db *PolicyDatabase) GetRule(id string) *Rule {
	return db.byID[id]
}

// GetRulesByCategory returns all rules in the given category.
func (db *PolicyDatabase) GetRulesByCategory(category string) []*Rule {
	return db.byCategory[category]
}

// AllRules returns all rules in the database.
func (db *PolicyDatabase) AllRules() []Rule {
	return db.Rules
}

// buildIndexes populates the lookup maps from the Rules slice.
func (db *PolicyDatabase) buildIndexes() {
	db.byID = make(map[string]*Rule, len(db.Rules))
	db.byCategory = make(map[string][]*Rule)
	for i := range db.Rules {
		r := &db.Rules[i]
		db.byID[r.ID] = r
		db.byCategory[r.Category] = append(db.byCategory[r.Category], r)
	}
}
