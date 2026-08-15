package main

import (
	"reflect"
	"testing"
)

// TestReconcileMigrationKeys locks the post-apply coverage contract:
// missing expected keys fail-closed; extra recorded keys are report-only.
// Count equality between recorded and on-disk is intentionally not used —
// consolidated/deleted filenames remain legitimately in schema_migrations.
func TestReconcileMigrationKeys(t *testing.T) {
	tests := []struct {
		name        string
		recorded    []string
		ddlFiles    []string
		seedBundles []string
		wantMissing []string
		wantExtra   []string
	}{
		{
			name: "full match",
			recorded: []string{
				"001_init.sql",
				"002_lstep_delivery_trigger_log_daily_unique.sql",
				"seeds/002_master",
				"seeds/003_demo",
				"seeds/004_staging",
			},
			ddlFiles: []string{
				"001_init.sql",
				"002_lstep_delivery_trigger_log_daily_unique.sql",
			},
			seedBundles: []string{"002_master", "003_demo", "004_staging"},
			wantMissing: nil,
			wantExtra:   nil,
		},
		{
			name: "missing DDL key",
			recorded: []string{
				"seeds/002_master",
				"seeds/003_demo",
				"seeds/004_staging",
			},
			ddlFiles: []string{
				"001_init.sql",
			},
			seedBundles: []string{"002_master", "003_demo", "004_staging"},
			wantMissing: []string{"001_init.sql"},
			wantExtra:   nil,
		},
		{
			name: "missing seed key",
			recorded: []string{
				"001_init.sql",
				"seeds/002_master",
				"seeds/004_staging",
			},
			ddlFiles:    []string{"001_init.sql"},
			seedBundles: []string{"002_master", "003_demo", "004_staging"},
			wantMissing: []string{"seeds/003_demo"},
			wantExtra:   nil,
		},
		{
			name: "extra recorded key does not fail",
			recorded: []string{
				"001_init.sql",
				"002_seed_master.sql", // legacy/orphan key still on disk history
				"seeds/002_master",
				"seeds/003_demo",
				"seeds/004_staging",
			},
			ddlFiles:    []string{"001_init.sql"},
			seedBundles: []string{"002_master", "003_demo", "004_staging"},
			wantMissing: nil,
			wantExtra:   []string{"002_seed_master.sql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing, extra := reconcileMigrationKeys(tt.recorded, tt.ddlFiles, tt.seedBundles)
			if !reflect.DeepEqual(missing, tt.wantMissing) {
				t.Fatalf("reconcileMigrationKeys missing = %v, want %v (extra=%v)", missing, tt.wantMissing, extra)
			}
			if !reflect.DeepEqual(extra, tt.wantExtra) {
				t.Fatalf("reconcileMigrationKeys extra = %v, want %v (missing=%v)", extra, tt.wantExtra, missing)
			}
		})
	}
}
