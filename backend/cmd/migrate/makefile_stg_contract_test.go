package main

import (
	"os"
	"strings"
	"testing"
)

func TestMakefileSTGUATTargetsWireEnvironmentAndExactConfirmations(t *testing.T) {
	raw, err := os.ReadFile("../../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	makefile := string(raw)
	required := []string{
		`STG_UAT_CSV_IMPORT_ARGS =`, `--confirm-target-host "$${STG_UAT_CSV_IMPORT_CONFIRM_HOST}"`, `--confirm-target-database "$${TARGET_DB_NAME}"`,
		`run ./cmd/stg-uat-skeleton apply`, `--confirm-target-host "$${STG_UAT_SKELETON_CONFIRM_HOST}"`,
		`run ./cmd/stg-uat-staff-attach preflight`, `run ./cmd/stg-uat-staff-attach apply`, `--confirm-target-host "$${STG_UAT_STAFF_ATTACH_CONFIRM_HOST}"`,
	}
	for _, want := range required {
		if !strings.Contains(makefile, want) {
			t.Errorf("Makefile missing %q", want)
		}
	}
	for _, target := range []string{"stg-uat-skeleton:", "stg-uat-csv-import-preflight:", "stg-uat-csv-import:", "stg-uat-csv-import-verify:", "stg-uat-staff-attach-preflight:", "stg-uat-staff-attach:"} {
		start := strings.Index(makefile, target)
		if start < 0 {
			t.Errorf("missing target %s", target)
			continue
		}
		recipe := makefile[start:]
		if next := strings.Index(recipe, "\n\n"); next >= 0 {
			recipe = recipe[:next]
		}
		if !strings.Contains(recipe, "APP_ENV=staging") {
			t.Errorf("%s does not pass APP_ENV=staging", target)
		}
	}
}
