package csvimport

import (
	"fmt"
	"os"
	"path/filepath"
)

func cutoverCSVPath(sourceDir string, table CutoverManifestTable) string {
	if table.sourcePath != "" {
		return table.sourcePath
	}
	return filepath.Join(sourceDir, table.File)
}

// bindCutoverAccountSource binds physical storage without changing signed data.
func bindCutoverAccountSource(sourceDir, accountDir string, manifest *CutoverManifest) error {
	if accountDir == "" {
		parent := filepath.Dir(sourceDir)
		if filepath.Base(parent) != "_old_db_handoff" {
			return nil
		}
		accountDir = filepath.Join(filepath.Dir(parent), "002_master", "accounts", "_old_db_handoff", filepath.Base(sourceDir))
		if _, err := os.Lstat(accountDir); os.IsNotExist(err) {
			return nil
		}
	}
	accountDir, err := validateCutoverDirectory(accountDir)
	if err != nil {
		return fmt.Errorf("account directory: %w", err)
	}
	for _, local := range []string{"staffs.csv", "accounts"} {
		if _, err := os.Lstat(filepath.Join(sourceDir, local)); !os.IsNotExist(err) {
			return fmt.Errorf("duplicate local account source")
		}
	}
	entries, err := os.ReadDir(accountDir)
	if err != nil {
		return fmt.Errorf("read account directory: %w", err)
	}
	if len(entries) != 1 || entries[0].Name() != "staffs.csv" {
		return fmt.Errorf("account directory must contain only staffs.csv")
	}
	for i := range manifest.Tables {
		if manifest.Tables[i].Table == "staffs" {
			manifest.Tables[i].sourcePath = filepath.Join(accountDir, "staffs.csv")
		}
	}
	return nil
}

// resolveCutoverAccountCSV retains the producer's signed logical filename.
// Only staffs.csv can move; all other cutover files keep their exact layout.
func resolveCutoverAccountCSV(path string) (string, error) {
	if filepath.Base(path) != "staffs.csv" {
		return path, nil
	}
	dir := filepath.Join(filepath.Dir(path), "accounts")
	if _, err := os.Lstat(dir); os.IsNotExist(err) {
		return path, nil
	}
	if _, err := validateCutoverDirectory(dir); err != nil {
		return "", fmt.Errorf("account directory: %w", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("staff CSV must exist only in the account directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read account directory: %w", err)
	}
	if len(entries) != 1 || entries[0].Name() != "staffs.csv" {
		return "", fmt.Errorf("account directory must contain only staffs.csv")
	}
	return filepath.Join(dir, "staffs.csv"), nil
}
