package staff_test

// available_staffs_ban_test.go — TASK-021 Stage B static gate.
// GET /v1/reservations/available-staffs is WONTFILE; product code must not introduce it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAvailableStaffsRouteIsNotIntroduced(t *testing.T) {
	t.Parallel()

	roots := []string{
		filepath.Join("..", "..", "internal"),
		filepath.Join("..", "..", "..", "frontend", "src"),
	}
	// Also scan from module root if cwd differs.
	if wd, err := os.Getwd(); err == nil {
		_ = wd
	}

	banned := []string{
		"available-staffs",
		"available_staffs",
		"AvailableStaffs",
	}

	var hits []string
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			// Try absolute from backend/internal/staff cwd
			continue
		}
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			// Skip this ban test itself and other test-only files that may assert the ban string.
			base := filepath.Base(path)
			if strings.Contains(base, "available_staffs_ban") {
				return nil
			}
			// Only product sources.
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
				return nil
			}
			if strings.HasSuffix(base, "_test.go") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			content := string(data)
			for _, b := range banned {
				if strings.Contains(content, b) {
					hits = append(hits, path+": "+b)
				}
			}
			return nil
		})
	}

	// Resolve roots relative to repo: this package is backend/internal/staff
	// Walk from module root discovered via go.mod.
	moduleRoot := findModuleRoot(t)
	for _, rel := range []string{
		filepath.Join("backend", "internal"),
		filepath.Join("frontend", "src"),
	} {
		root := filepath.Join(moduleRoot, rel)
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			if strings.Contains(base, "available_staffs_ban") {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
				return nil
			}
			if strings.HasSuffix(base, "_test.go") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			content := string(data)
			for _, b := range banned {
				if strings.Contains(content, b) {
					hits = append(hits, path+": "+b)
				}
			}
			return nil
		})
	}

	if len(hits) > 0 {
		t.Fatalf("available-staffs product surface is WONTFILE; found banned tokens:\n%s", strings.Join(hits, "\n"))
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// backend/go.mod may exist; prefer repo root with frontend/
			if _, err := os.Stat(filepath.Join(dir, "frontend")); err == nil {
				return dir
			}
			// if inside backend/, climb once more
			parent := filepath.Dir(dir)
			if _, err := os.Stat(filepath.Join(parent, "frontend")); err == nil {
				return parent
			}
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}
