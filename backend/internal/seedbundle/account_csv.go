package seedbundle

import (
	"os"
	"path/filepath"
)

// CSVRelativePath groups account data without changing logical manifest names.
func CSVRelativePath(name string) string {
	switch name {
	case "accounts.csv", "staffs.csv", "occupations.csv", "permission_groups.csv", "permission_group_rules.csv", "staff_permission_groups.csv", "staff_clinic_assignments.csv":
		return filepath.Join("accounts", name)
	default:
		return name
	}
}

// CSVPath supports both the account directory and historical flat bundles.
// Choosing the layout by directory avoids falling back to stale flat CSVs when
// a file is missing from an account directory. Manifest bytes remain unchanged,
// preserving already-applied bundle checksums.
func CSVPath(dir, name string) string {
	if _, err := os.Lstat(filepath.Join(dir, "accounts")); !os.IsNotExist(err) {
		return filepath.Join(dir, CSVRelativePath(name))
	}
	return filepath.Join(dir, name)
}
