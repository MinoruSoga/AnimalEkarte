package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readRepositoryMakefile(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	candidates := []string{
		filepath.Join(cwd, "../../../Makefile"),
		"/Makefile",
		filepath.Join(cwd, "../../../../../Makefile"),
	}
	var lastErr error
	for _, path := range candidates {
		raw, readErr := os.ReadFile(path)
		if readErr == nil {
			return string(raw)
		}
		lastErr = readErr
	}
	t.Fatalf("Makefile not found from cwd=%s: %v", cwd, lastErr)
	return ""
}

func TestMakefileSTGUATTargetsWireEnvironmentAndExactConfirmations(t *testing.T) {
	makefile := readRepositoryMakefile(t)
	required := []string{
		`STG_UAT_CSV_IMPORT_ARGS =`, `--confirm-target-host "$${STG_UAT_CSV_IMPORT_CONFIRM_HOST}"`, `--confirm-target-database "$${TARGET_DB_NAME}"`,
		`run ./cmd/stg-uat-skeleton apply`, `--confirm-target-host "$${STG_UAT_SKELETON_CONFIRM_HOST}"`,
		`run ./cmd/stg-uat-staff-attach preflight`, `run ./cmd/stg-uat-staff-attach apply`, `--confirm-target-host "$${STG_UAT_STAFF_ATTACH_CONFIRM_HOST}"`,
		`run ./cmd/csv-import-stg-uat import`,
	}
	for _, want := range required {
		if !strings.Contains(makefile, want) {
			t.Errorf("Makefile missing %q", want)
		}
	}
	for _, target := range []string{"stg-uat-skeleton:", "stg-uat-csv-import-preflight:", "stg-uat-csv-import:", "stg-uat-csv-import-verify:", "stg-uat-import:", "stg-uat-staff-attach-preflight:", "stg-uat-staff-attach:"} {
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
		if strings.Contains(recipe, "--allow-local-rehearsal") || strings.Contains(recipe, "CSV_IMPORT_EXTRA_ARGS") {
			t.Errorf("%s must not pass local-rehearsal extra args", target)
		}
	}
}

func TestMakefileSTGUATImportIsOneShotPreflightApplyVerify(t *testing.T) {
	makefile := readRepositoryMakefile(t)
	argsStart := strings.Index(makefile, "STG_UAT_IMPORT_ARGS =")
	if argsStart < 0 {
		t.Fatal("missing STG_UAT_IMPORT_ARGS")
	}
	argsBlock := makefile[argsStart:]
	if next := strings.Index(argsBlock, "\n\n"); next >= 0 {
		argsBlock = argsBlock[:next]
	}
	for _, want := range []string{
		`--source-dir /migration-input`,
		`--expected-manifest-sha256 "$${CSV_MANIFEST_SHA256}"`,
		`--clinic-code "$${CLINIC_CODE}"`,
		`--clinic-ordinal "$${CLINIC_ORDINAL}"`,
		`--run-id "$${MIGRATION_RUN_ID}"`,
		`--clinic-id "$${TARGET_CLINIC_ID}"`,
		`--fallback-animal-species-id "$${FALLBACK_ANIMAL_SPECIES_ID}"`,
		`--fallback-exam-type-id "$${FALLBACK_EXAM_TYPE_ID}"`,
		`--trimming-reservation-type-id "$${TRIMMING_RESERVATION_TYPE_ID}"`,
		`--cash-payment-method-id "$${PAYMENT_METHOD_CASH_ID}"`,
		`--credit-card-payment-method-id "$${PAYMENT_METHOD_CREDIT_CARD_ID}"`,
		`--confirm-target-host "$${STG_UAT_CSV_IMPORT_CONFIRM_HOST}"`,
		`--confirm-target-database "$${TARGET_DB_NAME}"`,
	} {
		if !strings.Contains(argsBlock, want) {
			t.Errorf("STG_UAT_IMPORT_ARGS missing %q", want)
		}
	}
	for _, forbidden := range []string{"--allow-local-rehearsal"} {
		if strings.Contains(argsBlock, forbidden) {
			t.Errorf("STG_UAT_IMPORT_ARGS must not contain %q", forbidden)
		}
	}

	start := strings.Index(makefile, "stg-uat-import:")
	if start < 0 {
		t.Fatal("missing target stg-uat-import:")
	}
	recipe := makefile[start:]
	if next := strings.Index(recipe, "\n\n"); next >= 0 {
		recipe = recipe[:next]
	}
	for _, want := range []string{
		`run ./cmd/csv-import-stg-uat import`,
		`$(STG_UAT_IMPORT_ARGS)`,
		`--confirm-target-write`,
		`--confirm-backup-ready`,
		`STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL`,
	} {
		if !strings.Contains(recipe, want) {
			t.Errorf("stg-uat-import recipe missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"--allow-local-rehearsal",
		"CSV_IMPORT_EXTRA_ARGS",
		"--fallback-animal-species-id",
		"--fallback-exam-type-id",
		"--trimming-reservation-type-id",
		"--cash-payment-method-id",
		"--credit-card-payment-method-id",
		"stg-uat-staff-attach",
		"stg-uat-skeleton",
	} {
		if strings.Contains(recipe, forbidden) {
			t.Errorf("stg-uat-import recipe must not contain %q", forbidden)
		}
	}
}
