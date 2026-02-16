package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxFileSize is the maximum file size (10 MB) that will be read during scanning.
// Files exceeding this limit are skipped to prevent memory exhaustion from
// maliciously crafted project files.
const MaxFileSize = 10 * 1024 * 1024

// DefaultSkipDirs contains directories that should be skipped when walking Android projects.
var DefaultSkipDirs = map[string]bool{
	".git":         true,
	".gradle":      true,
	".idea":        true,
	"build":        true,
	"node_modules": true,
	".cxx":         true,
}

// WalkOption configures the file walker behavior.
type WalkOption func(*walkConfig)

type walkConfig struct {
	extensions []string
	skipDirs   map[string]bool
	filenames  []string
}

// WithExtensions limits the walk to files matching the given extensions (e.g., ".xml", ".kt").
func WithExtensions(exts ...string) WalkOption {
	return func(c *walkConfig) {
		c.extensions = append(c.extensions, exts...)
	}
}

// WithSkipDirs adds directories to skip during traversal.
func WithSkipDirs(dirs ...string) WalkOption {
	return func(c *walkConfig) {
		for _, d := range dirs {
			c.skipDirs[d] = true
		}
	}
}

// WithFilenames limits the walk to files with specific names (e.g., "AndroidManifest.xml").
func WithFilenames(names ...string) WalkOption {
	return func(c *walkConfig) {
		c.filenames = append(c.filenames, names...)
	}
}

// WalkFiles traverses the project directory and returns file paths matching the given options.
func WalkFiles(root string, opts ...WalkOption) ([]string, error) {
	cfg := &walkConfig{
		skipDirs: make(map[string]bool),
	}
	for k, v := range DefaultSkipDirs {
		cfg.skipDirs[k] = v
	}
	for _, opt := range opts {
		opt(cfg)
	}

	extSet := make(map[string]bool, len(cfg.extensions))
	for _, ext := range cfg.extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[strings.ToLower(ext)] = true
	}

	nameSet := make(map[string]bool, len(cfg.filenames))
	for _, name := range cfg.filenames {
		nameSet[name] = true
	}

	// Verify root exists before walking.
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("cannot access root directory: %w", err)
	}

	// Resolve the root to an absolute path for symlink containment checks.
	absRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		absRoot = root
	}
	absRoot, _ = filepath.Abs(absRoot)

	var files []string

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries with errors
		}

		if d.IsDir() {
			if cfg.skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip symlinks to prevent path traversal outside the project root.
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if len(nameSet) > 0 && nameSet[d.Name()] {
			files = append(files, path)
			return nil
		}

		if len(extSet) > 0 {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if extSet[ext] {
				files = append(files, path)
			}
			return nil
		}

		// No filters means collect all files.
		if len(nameSet) == 0 && len(extSet) == 0 {
			files = append(files, path)
		}

		return nil
	})

	_ = absRoot // used for documentation of intent; symlinks are skipped above

	return files, err
}

// ReadFileWithLimit reads a file up to MaxFileSize bytes. Returns an error if
// the file exceeds the limit, preventing memory exhaustion from oversized files.
func ReadFileWithLimit(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file %s exceeds maximum size (%d bytes > %d bytes)", filepath.Base(path), info.Size(), MaxFileSize)
	}
	return os.ReadFile(path)
}

// FindAndroidManifests locates all AndroidManifest.xml files in the project.
func FindAndroidManifests(root string) ([]string, error) {
	return WalkFiles(root, WithFilenames("AndroidManifest.xml"))
}

// FindGradleFiles locates all build.gradle and build.gradle.kts files in the project.
func FindGradleFiles(root string) ([]string, error) {
	return WalkFiles(root, WithFilenames("build.gradle", "build.gradle.kts"))
}
