package utils

import (
	"os"
	"path/filepath"
	"strings"
)

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

	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries with errors
		}

		if d.IsDir() {
			if cfg.skipDirs[d.Name()] {
				return filepath.SkipDir
			}
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

	return files, err
}

// FindAndroidManifests locates all AndroidManifest.xml files in the project.
func FindAndroidManifests(root string) ([]string, error) {
	return WalkFiles(root, WithFilenames("AndroidManifest.xml"))
}

// FindGradleFiles locates all build.gradle and build.gradle.kts files in the project.
func FindGradleFiles(root string) ([]string, error) {
	return WalkFiles(root, WithFilenames("build.gradle", "build.gradle.kts"))
}
