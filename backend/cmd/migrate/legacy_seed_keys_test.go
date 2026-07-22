package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLegacyKeysAmongDetectsAllLegacyFilenames(t *testing.T) {
	applied := []string{
		"001_init.sql",
		"002_seed_master.sql",
		"003_seed_demo.sql",
		"004_seed_staging.sql",
	}

	got := legacyKeysAmong(applied)
	want := []string{"002_seed_master.sql", "003_seed_demo.sql", "004_seed_staging.sql"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacyKeysAmong(%v) = %v, want %v", applied, got, want)
	}
}

func TestLegacyKeysAmongEmptyOnFreshLayout(t *testing.T) {
	// Fresh layout schema_migrations keys: DDL filenames plus seeds/<bundle>
	// (6 rows as of 2026-07-22). None of these are legacy stub filenames.
	applied := []string{
		"001_init.sql",
		"002_lstep_snapshot_import_clinic_fk.sql",
		"003_medical_records_appointment_id_index.sql",
		"seeds/002_master",
		"seeds/003_demo",
		"seeds/004_staging",
	}

	got := legacyKeysAmong(applied)
	if len(got) != 0 {
		t.Fatalf("legacyKeysAmong(%v) = %v, want empty — current seed keys must never be mistaken for legacy ones", applied, got)
	}
}

func TestShouldBaselineSQLMigrationOnlyMarksInitialSchema(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "001_init.sql", want: true},
		{filename: "002_lstep_snapshot_import_clinic_fk.sql", want: false},
		{filename: "003_future_incremental.sql", want: false},
		{filename: "seeds/002_master", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := shouldBaselineSQLMigration(tt.filename); got != tt.want {
				t.Fatalf("shouldBaselineSQLMigration(%q) = %t, want %t", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLegacyKeysAmongEmptyWhenNoMigrationsApplied(t *testing.T) {
	got := legacyKeysAmong(nil)
	if len(got) != 0 {
		t.Fatalf("legacyKeysAmong(nil) = %v, want empty", got)
	}
}

func TestLegacyKeysAmongPartialDetection(t *testing.T) {
	// A DB baselined only against a partial legacy history should still be
	// caught — this must not require all three legacy keys to be present.
	applied := []string{"001_init.sql", "003_seed_demo.sql"}

	got := legacyKeysAmong(applied)
	want := []string{"003_seed_demo.sql"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacyKeysAmong(%v) = %v, want %v", applied, got, want)
	}
}

// P1-3 (PR #186 review): detectLegacySeedKeys used to fail-fast on any legacy
// key, blocking normal upgrades (e.g. main->staging) unless db_reset=true.
// legacyTranslationTargets is the pure function backing the
// translate-not-fail-fast path — the same DB-free boundary legacyKeysAmong
// already established for this package.
//
// It intentionally takes no "which legacy keys were found" input and always
// returns ALL bundle keys (PR #186 security review, HIGH): baselining only
// the bundles whose specific legacy filename was found would leave the rest
// "unapplied" for a DB whose legacy key set is genuinely partial, letting the
// following runSeedBundles call auto-load those CSV bundles onto what may be
// a real database — see the doc comment on legacyTranslationTargets in
// main.go for the full hazard.

func TestLegacyTranslationTargetsCoversAllBundlesInOrder(t *testing.T) {
	got := legacyTranslationTargets()
	want := []string{"seeds/002_master", "seeds/003_demo", "seeds/004_staging"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacyTranslationTargets() = %v, want %v", got, want)
	}
}

func TestLegacyTranslationTargetsNeverProducesLegacyKeys(t *testing.T) {
	// Every target must carry the "seeds/" prefix that BundleMigrationKey
	// guarantees is disjoint from *.sql filenames — translation must never
	// collide with a real DDL migration filename or a legacy stub filename.
	got := legacyTranslationTargets()
	for _, key := range got {
		if !strings.HasPrefix(key, "seeds/") {
			t.Fatalf("legacyTranslationTargets() contains %q, want a seeds/<bundle> key", key)
		}
	}
}
