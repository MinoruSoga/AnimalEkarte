package main

import (
	"reflect"
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
	// Fresh layout schema_migrations keys: DDL 001–004 plus seeds/<bundle>
	// (7 rows). None of these are legacy stub filenames.
	applied := []string{
		"001_init.sql",
		"002_add_checkup_vaccination_indexes.sql",
		"003_add_pets_batch_living_count_index.sql",
		"004_add_billings_hospitalization_id_unique_index.sql",
		"seeds/002_master",
		"seeds/003_demo",
		"seeds/004_staging",
	}

	got := legacyKeysAmong(applied)
	if len(got) != 0 {
		t.Fatalf("legacyKeysAmong(%v) = %v, want empty — current seed keys must never be mistaken for legacy ones", applied, got)
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
