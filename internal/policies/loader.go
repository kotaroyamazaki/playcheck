package policies

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed policies.json
var policiesJSON []byte

var (
	defaultDB   *PolicyDatabase
	defaultOnce sync.Once
	loadErr     error
)

// Load parses the embedded policies JSON and returns a PolicyDatabase.
// The result is cached after the first call.
func Load() (*PolicyDatabase, error) {
	defaultOnce.Do(func() {
		defaultDB, loadErr = parse(policiesJSON)
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return defaultDB, nil
}

// parse decodes raw JSON into a PolicyDatabase and builds indexes.
func parse(data []byte) (*PolicyDatabase, error) {
	var db PolicyDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("parsing policy database: %w", err)
	}
	if len(db.Rules) == 0 {
		return nil, fmt.Errorf("policy database contains no rules")
	}
	db.buildIndexes()
	return &db, nil
}
