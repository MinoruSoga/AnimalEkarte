package seedbundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSVPathAccountLayout(t *testing.T) {
	dir := t.TempDir()
	if got := CSVPath(dir, "permission_groups.csv"); got != filepath.Join(dir, "permission_groups.csv") {
		t.Fatalf("legacy path = %s", got)
	}
	if err := os.Mkdir(filepath.Join(dir, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"permission_groups.csv", "permission_group_rules.csv", "staffs.csv", "accounts.csv"} {
		if got := CSVPath(dir, name); got != filepath.Join(dir, "accounts", name) {
			t.Fatalf("account path = %s", got)
		}
	}
	if got := CSVPath(dir, "clinics.csv"); got != filepath.Join(dir, "clinics.csv") {
		t.Fatalf("clinic path = %s", got)
	}
}
