package preflight

// NewDefaultRunner creates a Runner and registers checkers using the provided
// registration function. This avoids import cycles by letting the caller
// (typically the CLI or main package) provide the checker instances.
//
// Usage:
//
//	runner := preflight.NewDefaultRunner(func(r *Runner) {
//	    r.RegisterScanner(manifest.NewScanner())
//	    r.RegisterScanner(datasafety.NewChecker())
//	    r.RegisterScanner(codescan.NewScanner())
//	})
func NewDefaultRunner(register func(*Runner)) *Runner {
	r := &Runner{}
	register(r)
	return r
}
