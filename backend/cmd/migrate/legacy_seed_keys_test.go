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
	// (4 rows as of 2026-07-27). None of these are legacy stub filenames.
	applied := []string{
		"001_init.sql",
		"seeds/002_master",
		"seeds/003_demo",
		"seeds/004_staging",
	}

	got := legacyKeysAmong(applied)
	if len(got) != 0 {
		t.Fatalf("legacyKeysAmong(%v) = %v, want empty — current seed keys must never be mistaken for legacy ones", applied, got)
	}
}

func TestValidateBaselineSafety(t *testing.T) {
	tests := []struct {
		name           string
		count          int
		hasSchema      bool
		wantErr        bool
		wantSubstrings []string
	}{
		{name: "fresh database", count: 0, hasSchema: false},
		{name: "migration history with application schema", count: 1, hasSchema: true},
		{
			name:      "migration history without application schema",
			count:     1,
			hasSchema: false,
			wantErr:   true,
			wantSubstrings: []string{
				"migration history exists",
				"application schema is missing",
				"LOCAL_DB_RESET.md",
			},
		},
		{
			name:      "application schema without migration history",
			count:     0,
			hasSchema: true,
			wantErr:   true,
			wantSubstrings: []string{
				"existing application schema",
				"schema_migrations is empty",
				"schema completeness cannot be verified",
				"LOCAL_DB_RESET.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBaselineSafety(tt.count, tt.hasSchema)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("validateBaselineSafety(%d, %t) returned unexpected error: %v", tt.count, tt.hasSchema, err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateBaselineSafety(%d, %t) returned nil, want error", tt.count, tt.hasSchema)
			}
			for _, substring := range tt.wantSubstrings {
				if !strings.Contains(err.Error(), substring) {
					t.Errorf("error %q does not contain %q", err, substring)
				}
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
// returns all three legacy-equivalent bundle keys (PR #186 security review,
// HIGH). Bundles introduced after the stub-SQL era must not be translated:
// they have no legacy applied-history equivalent and must remain eligible for
// normal application.

func TestLegacyTranslationTargetsCoversOnlyLegacyEquivalentBundles(t *testing.T) {
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
