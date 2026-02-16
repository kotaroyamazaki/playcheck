package utils

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func setupWalkDir(t *testing.T, files map[string]string) string {
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

func TestWalkFiles_NoFilters(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"a.txt":        "hello",
		"b.java":       "class B {}",
		"sub/c.kt":     "fun main() {}",
		"sub/d.xml":    "<root/>",
	})

	files, err := WalkFiles(dir)
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 4 {
		t.Errorf("expected 4 files, got %d", len(files))
	}
}

func TestWalkFiles_WithExtensions(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"Main.java":    "class Main {}",
		"App.kt":       "fun main() {}",
		"config.xml":   "<config/>",
		"readme.txt":   "readme",
		"sub/Util.java": "class Util {}",
	})

	files, err := WalkFiles(dir, WithExtensions(".java", ".kt"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 java/kt files, got %d", len(files))
	}
	for _, f := range files {
		ext := filepath.Ext(f)
		if ext != ".java" && ext != ".kt" {
			t.Errorf("unexpected extension: %s", ext)
		}
	}
}

func TestWalkFiles_WithExtensions_CaseInsensitive(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"Main.JAVA": "class Main {}",
		"App.Kt":    "fun main() {}",
		"other.txt": "txt",
	})

	files, err := WalkFiles(dir, WithExtensions(".java", ".kt"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files (case insensitive), got %d", len(files))
	}
}

func TestWalkFiles_WithExtensions_NoDot(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"Main.java": "class Main {}",
		"other.txt": "txt",
	})

	// Extensions without leading dot should still work
	files, err := WalkFiles(dir, WithExtensions("java"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestWalkFiles_WithFilenames(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"AndroidManifest.xml":                    "<manifest/>",
		"app/src/main/AndroidManifest.xml":       "<manifest/>",
		"other.xml":                              "<other/>",
		"build.gradle":                           "apply plugin",
	})

	files, err := WalkFiles(dir, WithFilenames("AndroidManifest.xml"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 AndroidManifest.xml files, got %d", len(files))
	}
	for _, f := range files {
		if filepath.Base(f) != "AndroidManifest.xml" {
			t.Errorf("unexpected file: %s", f)
		}
	}
}

func TestWalkFiles_DefaultSkipDirs(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"src/Main.java":           "class Main {}",
		".git/config":             "[core]",
		"build/output.class":      "bytecode",
		".gradle/cache.bin":       "cache",
		".idea/workspace.xml":     "<idea/>",
		"node_modules/pkg/a.js":   "module",
		".cxx/cmake/build.ninja":  "ninja",
	})

	files, err := WalkFiles(dir)
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	// Only src/Main.java should be found; all others are in skip dirs
	if len(files) != 1 {
		t.Errorf("expected 1 file (default skip dirs), got %d: %v", len(files), files)
	}
}

func TestWalkFiles_WithSkipDirs(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"src/Main.java":        "class Main {}",
		"test/MainTest.java":   "class MainTest {}",
		"vendor/lib.java":      "class Lib {}",
	})

	files, err := WalkFiles(dir, WithSkipDirs("test", "vendor"), WithExtensions(".java"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (custom skip dirs), got %d", len(files))
	}
}

func TestWalkFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	files, err := WalkFiles(dir)
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty dir, got %d", len(files))
	}
}

func TestWalkFiles_NonexistentDir(t *testing.T) {
	_, err := WalkFiles("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestFindAndroidManifests(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"app/src/main/AndroidManifest.xml": "<manifest/>",
		"lib/src/main/AndroidManifest.xml": "<manifest/>",
		"app/src/main/other.xml":           "<other/>",
	})

	files, err := FindAndroidManifests(dir)
	if err != nil {
		t.Fatalf("FindAndroidManifests error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 manifests, got %d", len(files))
	}
}

func TestFindGradleFiles(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"build.gradle":             "apply plugin",
		"app/build.gradle":         "apply plugin",
		"app/build.gradle.kts":     "plugins {}",
		"settings.gradle":          "include",
		"other.txt":                "txt",
	})

	files, err := FindGradleFiles(dir)
	if err != nil {
		t.Fatalf("FindGradleFiles error: %v", err)
	}
	sort.Strings(files)
	if len(files) != 3 {
		t.Errorf("expected 3 gradle files (build.gradle + build.gradle.kts), got %d: %v", len(files), files)
	}
}

func TestWalkFiles_NestedDirectories(t *testing.T) {
	dir := setupWalkDir(t, map[string]string{
		"a/b/c/d/deep.java":    "class Deep {}",
		"a/b/shallow.java":     "class Shallow {}",
		"top.java":             "class Top {}",
	})

	files, err := WalkFiles(dir, WithExtensions(".java"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files from nested dirs, got %d", len(files))
	}
}

func TestWalkFiles_FilenamesTakePrecedence(t *testing.T) {
	// When both filenames and extensions are set, filenames filter matches by name,
	// and extension filter matches remaining files.
	dir := setupWalkDir(t, map[string]string{
		"AndroidManifest.xml": "<manifest/>",
		"other.xml":           "<other/>",
		"main.java":           "class Main {}",
	})

	// With filenames set, only exact name matches are returned (for files matching name).
	// Files not matching name are checked against extension filter.
	files, err := WalkFiles(dir, WithFilenames("AndroidManifest.xml"))
	if err != nil {
		t.Fatalf("WalkFiles error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file by filename, got %d", len(files))
	}
}
