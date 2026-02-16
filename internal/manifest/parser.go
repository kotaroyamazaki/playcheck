package manifest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// AndroidManifest represents the parsed AndroidManifest.xml.
type AndroidManifest struct {
	Package           string
	VersionCode       int
	VersionName       string
	MinSdkVersion     int
	TargetSdkVersion  int
	CompileSdkVersion int

	UsesCleartext bool // android:usesCleartextTraffic
	HasCleartext  bool // whether the attribute was explicitly set

	Permissions []Permission
	Activities  []Activity
	Services    []Service
	Receivers   []Receiver
	Providers   []Provider

	// Raw lines for line-number tracking.
	rawContent []byte
	filePath   string
}

// Permission represents a <uses-permission> element.
type Permission struct {
	Name     string
	MaxSdk   int
	Line     int
	Required bool // android:required
}

// IntentFilter represents an <intent-filter> element.
type IntentFilter struct {
	Actions    []string
	Categories []string
	Line       int
}

// Activity represents an <activity> element.
type Activity struct {
	Name          string
	Exported      *bool // nil if not explicitly set
	IntentFilters []IntentFilter
	Line          int
}

// Service represents a <service> element.
type Service struct {
	Name          string
	Exported      *bool
	IntentFilters []IntentFilter
	Line          int
}

// Receiver represents a <receiver> element.
type Receiver struct {
	Name          string
	Exported      *bool
	IntentFilters []IntentFilter
	Line          int
}

// Provider represents a <provider> element.
type Provider struct {
	Name          string
	Exported      *bool
	IntentFilters []IntentFilter
	Line          int
}

// HasLauncherActivity returns true if any activity has a launcher intent filter.
func (m *AndroidManifest) HasLauncherActivity() bool {
	for _, a := range m.Activities {
		if isLauncherActivity(a) {
			return true
		}
	}
	return false
}

func isLauncherActivity(a Activity) bool {
	for _, f := range a.IntentFilters {
		hasMain := false
		hasLauncher := false
		for _, action := range f.Actions {
			if action == "android.intent.action.MAIN" {
				hasMain = true
			}
		}
		for _, cat := range f.Categories {
			if cat == "android.intent.category.LAUNCHER" {
				hasLauncher = true
			}
		}
		if hasMain && hasLauncher {
			return true
		}
	}
	return false
}

// FilePath returns the file path of the parsed manifest.
func (m *AndroidManifest) FilePath() string {
	return m.filePath
}

// ParseFile parses an AndroidManifest.xml file at the given path.
func ParseFile(path string) (*AndroidManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	m, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	m.filePath = path
	return m, nil
}

// FindAndParse locates AndroidManifest.xml in a project directory and parses it.
func FindAndParse(projectDir string) (*AndroidManifest, error) {
	candidates := []string{
		filepath.Join(projectDir, "app", "src", "main", "AndroidManifest.xml"),
		filepath.Join(projectDir, "AndroidManifest.xml"),
		filepath.Join(projectDir, "src", "main", "AndroidManifest.xml"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return ParseFile(path)
		}
	}
	return nil, fmt.Errorf("AndroidManifest.xml not found in %s", projectDir)
}

// Parse parses AndroidManifest.xml content from raw bytes.
func Parse(data []byte) (*AndroidManifest, error) {
	m := &AndroidManifest{
		rawContent: data,
	}

	// Build a line offset index for accurate line number reporting.
	lineOffsets := buildLineOffsets(data)

	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose

	var elementStack []string

	// Track current component being parsed.
	type componentCtx struct {
		kind          string // "activity", "service", "receiver", "provider"
		name          string
		exported      *bool
		intentFilters []IntentFilter
		line          int
	}
	var currentComponent *componentCtx
	var currentIntentFilter *IntentFilter

	for {
		offset := decoder.InputOffset()
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML parse error at offset %d: %w", offset, err)
		}

		line := offsetToLine(lineOffsets, offset)

		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			elementStack = append(elementStack, name)

			switch name {
			case "manifest":
				m.parseManifestAttrs(t.Attr)

			case "uses-sdk":
				m.parseUsesSdkAttrs(t.Attr)

			case "application":
				m.parseApplicationAttrs(t.Attr)

			case "uses-permission":
				perm := parsePermission(t.Attr, line)
				m.Permissions = append(m.Permissions, perm)

			case "activity", "activity-alias":
				currentComponent = &componentCtx{
					kind: "activity",
					line: line,
				}
				currentComponent.name, currentComponent.exported = parseComponentAttrs(t.Attr)

			case "service":
				currentComponent = &componentCtx{
					kind: "service",
					line: line,
				}
				currentComponent.name, currentComponent.exported = parseComponentAttrs(t.Attr)

			case "receiver":
				currentComponent = &componentCtx{
					kind: "receiver",
					line: line,
				}
				currentComponent.name, currentComponent.exported = parseComponentAttrs(t.Attr)

			case "provider":
				currentComponent = &componentCtx{
					kind: "provider",
					line: line,
				}
				currentComponent.name, currentComponent.exported = parseComponentAttrs(t.Attr)

			case "intent-filter":
				currentIntentFilter = &IntentFilter{
					Line: line,
				}

			case "action":
				if currentIntentFilter != nil {
					for _, attr := range t.Attr {
						if attr.Name.Local == "name" {
							currentIntentFilter.Actions = append(currentIntentFilter.Actions, attr.Value)
						}
					}
				}

			case "category":
				if currentIntentFilter != nil {
					for _, attr := range t.Attr {
						if attr.Name.Local == "name" {
							currentIntentFilter.Categories = append(currentIntentFilter.Categories, attr.Value)
						}
					}
				}
			}

		case xml.EndElement:
			name := t.Name.Local

			switch name {
			case "intent-filter":
				if currentIntentFilter != nil && currentComponent != nil {
					currentComponent.intentFilters = append(currentComponent.intentFilters, *currentIntentFilter)
				}
				currentIntentFilter = nil

			case "activity", "activity-alias":
				if currentComponent != nil && currentComponent.kind == "activity" {
					m.Activities = append(m.Activities, Activity{
						Name:          currentComponent.name,
						Exported:      currentComponent.exported,
						IntentFilters: currentComponent.intentFilters,
						Line:          currentComponent.line,
					})
					currentComponent = nil
				}

			case "service":
				if currentComponent != nil && currentComponent.kind == "service" {
					m.Services = append(m.Services, Service{
						Name:          currentComponent.name,
						Exported:      currentComponent.exported,
						IntentFilters: currentComponent.intentFilters,
						Line:          currentComponent.line,
					})
					currentComponent = nil
				}

			case "receiver":
				if currentComponent != nil && currentComponent.kind == "receiver" {
					m.Receivers = append(m.Receivers, Receiver{
						Name:          currentComponent.name,
						Exported:      currentComponent.exported,
						IntentFilters: currentComponent.intentFilters,
						Line:          currentComponent.line,
					})
					currentComponent = nil
				}

			case "provider":
				if currentComponent != nil && currentComponent.kind == "provider" {
					m.Providers = append(m.Providers, Provider{
						Name:          currentComponent.name,
						Exported:      currentComponent.exported,
						IntentFilters: currentComponent.intentFilters,
						Line:          currentComponent.line,
					})
					currentComponent = nil
				}
			}

			if len(elementStack) > 0 {
				elementStack = elementStack[:len(elementStack)-1]
			}
		}
	}

	return m, nil
}

func (m *AndroidManifest) parseManifestAttrs(attrs []xml.Attr) {
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "package":
			m.Package = attr.Value
		case "versionCode":
			m.VersionCode, _ = strconv.Atoi(attr.Value)
		case "versionName":
			m.VersionName = attr.Value
		case "compileSdkVersion":
			m.CompileSdkVersion, _ = strconv.Atoi(attr.Value)
		}
	}
}

func (m *AndroidManifest) parseUsesSdkAttrs(attrs []xml.Attr) {
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "minSdkVersion":
			m.MinSdkVersion, _ = strconv.Atoi(attr.Value)
		case "targetSdkVersion":
			m.TargetSdkVersion, _ = strconv.Atoi(attr.Value)
		}
	}
}

func (m *AndroidManifest) parseApplicationAttrs(attrs []xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Local == "usesCleartextTraffic" {
			m.HasCleartext = true
			m.UsesCleartext = strings.EqualFold(attr.Value, "true")
		}
	}
}

func parsePermission(attrs []xml.Attr, line int) Permission {
	p := Permission{Line: line, Required: true}
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "name":
			p.Name = attr.Value
		case "maxSdkVersion":
			p.MaxSdk, _ = strconv.Atoi(attr.Value)
		case "required":
			p.Required = strings.EqualFold(attr.Value, "true")
		}
	}
	return p
}

func parseComponentAttrs(attrs []xml.Attr) (name string, exported *bool) {
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "name":
			name = attr.Value
		case "exported":
			val := strings.EqualFold(attr.Value, "true")
			exported = &val
		}
	}
	return
}

// buildLineOffsets creates an index of byte offsets for the start of each line.
func buildLineOffsets(data []byte) []int {
	offsets := []int{0}
	for i, b := range data {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

// offsetToLine converts a byte offset to a 1-based line number.
func offsetToLine(lineOffsets []int, offset int64) int {
	lo, hi := 0, len(lineOffsets)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if int64(lineOffsets[mid]) <= offset {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return lo // 1-based since offsets[0]=0 and lo ends up at correct 1-based line
}
