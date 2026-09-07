package csvimport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// PathWithinRoot reports whether candidate is a direct child of root.
func PathWithinRoot(root string, candidate string) bool {
	if !filepath.IsAbs(root) || !filepath.IsAbs(candidate) {
		return false
	}
	cleanRoot := filepath.Clean(root)
	cleanCandidate := filepath.Clean(candidate)
	return cleanCandidate != cleanRoot && filepath.Dir(cleanCandidate) == cleanRoot
}

// CreateAuditReport creates an exclusive owner-only JSON report under root.
func CreateAuditReport(path string, root string, report any) (*os.File, error) {
	if !PathWithinRoot(root, path) {
		return nil, fmt.Errorf("audit report path must stay under %s", root)
	}
	cleanRoot := filepath.Clean(root)
	resolvedRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil || resolvedRoot != cleanRoot {
		return nil, fmt.Errorf("audit report directory must not contain symbolic links")
	}
	info, err := os.Lstat(cleanRoot)
	if err != nil {
		return nil, fmt.Errorf("inspect audit report directory: %w", err)
	}
	if !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("audit report directory must be owner-only")
	}
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_EXCL|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600) //nolint:gosec // fixed-root operator path, no-follow and exclusive owner-only creation
	if err != nil {
		return nil, fmt.Errorf("create audit report without overwrite: %w", err)
	}
	file := os.NewFile(uintptr(fd), filepath.Base(path))
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("create audit report: invalid file descriptor")
	}
	if err := ReplaceAuditReport(file, report); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

// ReplaceAuditReport overwrites an open report file with indented JSON.
func ReplaceAuditReport(file *os.File, report any) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek audit report: %w", err)
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate audit report: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write audit report: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync audit report: %w", err)
	}
	return nil
}
