package codescan

import (
	"regexp"
	"sync"
)

// patternCache provides thread-safe caching of compiled regex patterns.
var patternCache struct {
	sync.RWMutex
	m map[string]*regexp.Regexp
}

func init() {
	patternCache.m = make(map[string]*regexp.Regexp)
}

// compilePattern returns a compiled regex, using the cache when possible.
func compilePattern(pattern string) (*regexp.Regexp, error) {
	patternCache.RLock()
	if re, ok := patternCache.m[pattern]; ok {
		patternCache.RUnlock()
		return re, nil
	}
	patternCache.RUnlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	patternCache.Lock()
	patternCache.m[pattern] = re
	patternCache.Unlock()

	return re, nil
}

// compiledRule holds a code rule with its pre-compiled regex patterns.
type compiledRule struct {
	rule     codeRule
	patterns []*regexp.Regexp
}

// compileRules compiles all pattern strings in the rule set into regexps.
// Invalid patterns are silently skipped.
func compileRules(rules []codeRule) []compiledRule {
	compiled := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		cr := compiledRule{rule: r}
		for _, p := range r.Patterns {
			re, err := compilePattern(p)
			if err != nil {
				continue
			}
			cr.patterns = append(cr.patterns, re)
		}
		if len(cr.patterns) > 0 {
			compiled = append(compiled, cr)
		}
	}
	return compiled
}
