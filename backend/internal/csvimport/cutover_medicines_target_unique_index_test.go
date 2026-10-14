package csvimport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// EMR-342 target contract (001_init.sql:2452): medicines(clinic_id, name) is
// unique WHERE deleted_at IS NULL. The candidate22 target preflight must
// require a valid partial unique index with the exact key and no extra
// predicate narrowing. Legacy21 never queries or requires a medicines index.
func TestMedicinesClinicNameUniqueIndexQueryRejectsWeakIndexes(t *testing.T) {
	for _, fragment := range []string{
		"target_schema.nspname = 'public'",
		"target_table.relname = 'medicines'",
		"index_definition.indisunique = true",
		"index_definition.indisvalid = true",
		"index_definition.indisready = true",
		"index_definition.indimmediate = true",
		"index_definition.indexprs IS NULL",
		"index_definition.indnkeyatts = 2",
		") = ARRAY['clinic_id', 'name']::text[]",
		") = 'deleted_atisnull'",
	} {
		if !strings.Contains(medicinesClinicNameUniqueIndexQuery, fragment) {
			t.Errorf("medicines index query is missing %q", fragment)
		}
	}
	if !strings.Contains(medicinesClinicNameUniqueIndexQuery, "pg_get_expr(index_definition.indpred") {
		t.Error("medicines index query must compare the normalized partial predicate")
	}
}

// INCLUDE columns are payload, not uniqueness keys. Inspect only the two
// key attributes so covering indexes preserve the same uniqueness contract.
func TestMedicinesUniqueIndexQueryAllowsIncludedAttributes(t *testing.T) {
	if strings.Contains(medicinesClinicNameUniqueIndexQuery, "index_definition.indnatts = 2") {
		t.Error("medicines unique index query rejects equivalent INCLUDE indexes")
	}
	if !strings.Contains(medicinesClinicNameUniqueIndexQuery, "key_column.ordinality <= index_definition.indnkeyatts") {
		t.Error("medicines unique index query must exclude INCLUDE attributes from key comparison")
	}
}

func candidateManifestForTargetTests() CutoverManifest {
	manifest := CutoverManifest{
		profile:            CutoverProfileMedicinesCandidate,
		HandoffEligibility: "REHEARSAL_ONLY",
		IDBand: CutoverIDBand{
			Base: 0, EndExclusive: 10_000_000, NonOwnerIDOffset: 1_000_000,
			OwnerFloor: 300_000, ApplicationIDFloor: applicationIDFloor,
		},
	}
	for _, spec := range CutoverTableSpecs(CutoverProfileMedicinesCandidate) {
		manifest.Tables = append(manifest.Tables, CutoverManifestTable{Table: spec.Name, File: spec.Name + ".csv"})
	}
	return manifest
}

type medicinesIndexTargetQuerier struct {
	validTargetQuerier
	exists  bool
	err     error
	queries int
}

func (q *medicinesIndexTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if len(args) == 2 && args[0] == "medicines" && strings.Contains(query, "information_schema.columns") {
		if strings.Contains(query, "numeric_precision") {
			// Canonical 001 schema: tax_rate is unconstrained numeric.
			if args[1] == "tax_rate" {
				return staticRow{values: []any{nil, nil}}
			}
			return staticRow{values: []any{int64(10), int64(2)}}
		}
		if strings.Contains(query, "data_type") && args[1] == "tax_rate" {
			// Canonical 001 schema: tax_rate is numeric (unbounded).
			return staticRow{values: []any{"numeric"}}
		}
		if strings.Contains(query, "data_type") && args[1] == "default_quantity" {
			return staticRow{values: []any{"numeric"}}
		}
	}
	if strings.Contains(query, "FROM pg_index") && strings.Contains(query, "relname = 'medicines'") {
		q.queries++
		if q.err != nil {
			return staticRow{err: q.err}
		}
		return staticRow{values: []any{q.exists}}
	}
	return q.validTargetQuerier.QueryRow(ctx, query, args...)
}

func TestPreflightTargetRequiresMedicinesUniqueIndexForCandidate(t *testing.T) {
	missing := &medicinesIndexTargetQuerier{exists: false}
	err := PreflightCutoverTargetWithProvenance(
		context.Background(),
		missing,
		candidateManifestForTargetTests(),
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	)
	if err == nil || !strings.Contains(err.Error(), "medicines(clinic_id, name)") {
		t.Fatalf("PreflightCutoverTargetWithProvenance() error = %v, want medicines(clinic_id, name) unique index rejection", err)
	}
	if missing.queries == 0 {
		t.Fatal("candidate target preflight never probed the medicines unique index")
	}
}

func TestPreflightTargetAcceptsExactMedicinesUniqueIndexForCandidate(t *testing.T) {
	present := &medicinesIndexTargetQuerier{exists: true}
	if err := PreflightCutoverTargetWithProvenance(
		context.Background(),
		present,
		candidateManifestForTargetTests(),
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	); err != nil {
		t.Fatalf("PreflightCutoverTargetWithProvenance() error = %v, want exact medicines unique index accepted", err)
	}
}

func TestPreflightTargetMedicinesIndexQueryErrorFailsClosed(t *testing.T) {
	queryErr := errors.New("catalog lookup failed")
	probing := &medicinesIndexTargetQuerier{err: queryErr}
	err := PreflightCutoverTargetWithProvenance(
		context.Background(),
		probing,
		candidateManifestForTargetTests(),
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	)
	if !errors.Is(err, queryErr) {
		t.Fatalf("PreflightCutoverTargetWithProvenance() error = %v, want wrapped query error", err)
	}
}

func TestLegacyTargetNeverProbesMedicinesUniqueIndex(t *testing.T) {
	probing := &medicinesIndexTargetQuerier{exists: false}
	if err := PreflightCutoverTarget(
		context.Background(),
		probing,
		cutoverManifestForTargetTests(),
		validCutoverSeeds(),
	); err != nil {
		t.Fatalf("PreflightCutoverTarget() error = %v, want legacy21 preflight untouched by the medicines guard", err)
	}
	if probing.queries != 0 {
		t.Fatalf("legacy21 target preflight issued %d medicines index queries, want 0", probing.queries)
	}
}
