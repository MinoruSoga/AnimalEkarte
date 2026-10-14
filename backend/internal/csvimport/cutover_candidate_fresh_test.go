package csvimport

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// User decision A (2026-10-14, "A: freshのみ を採用"): a candidate22 import
// runs only against an entirely empty clinic band. Resume preflight and
// table-per-commit apply must reject ANY populated prefix — including a
// legitimate interrupted candidate run — before any Begin/COPY. Legacy21
// formal resume semantics are unchanged, and candidate formal activation
// stays blocked before the database.

func candidateFreshBundleForApply(t *testing.T) CutoverBundle {
	t.Helper()
	dir, digest := writeMedicinesCandidateFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

// candidateBandQuerier serves the candidate target schema checks plus
// per-table clinic-band counts and occupancy for preflight band inspection.
type candidateBandQuerier struct {
	validTargetQuerier
	occupied map[string]bool
}

func (q candidateBandQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
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
	if strings.Contains(query, "clinic_id <> $3") {
		return staticRow{values: []any{int64(0)}}
	}
	if strings.Contains(query, "count(*)") && strings.Contains(query, "id >= $1 AND id < $2") {
		if occupied := q.occupied[candidateTableNameFromBandQuery(query)]; occupied {
			return staticRow{values: []any{int64(1)}}
		}
		return staticRow{values: []any{int64(0)}}
	}
	if strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "id >= $1 AND id < $2") {
		return staticRow{values: []any{q.occupied[candidateTableNameFromBandQuery(query)]}}
	}
	return q.validTargetQuerier.QueryRow(ctx, query, args...)
}

func candidateTableNameFromBandQuery(query string) string {
	best := ""
	for _, spec := range CutoverTableSpecs(CutoverProfileMedicinesCandidate) {
		needle := "FROM " + pgx.Identifier{spec.Name}.Sanitize()
		if strings.Contains(query, needle) && len(spec.Name) >= len(best) {
			best = spec.Name
		}
	}
	return best
}

func TestCandidateResumePreflightRejectsAnyPopulatedBandTable(t *testing.T) {
	// Even a legitimate interrupted candidate prefix (first table complete,
	// counts and clinic matching the manifest) is denied under decision A.
	manifest := candidateManifestForTargetTests()
	for i := range manifest.Tables {
		manifest.Tables[i].RowCount = 1
	}
	querier := candidateBandQuerier{occupied: map[string]bool{"staffs": true}}
	err := PreflightCutoverTargetAllowingResumeWithProvenance(
		context.Background(),
		querier,
		manifest,
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("PreflightCutoverTargetAllowingResumeWithProvenance() error = %v, want candidate populated prefix rejection", err)
	}
}

func TestCandidateApplyRejectsExistingPrefixBeforeAnyTransaction(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &candidateFreshApplySession{fakeCutoverSession{inspectCounts: map[string]int64{"staffs": 1}}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want candidate existing-prefix rejection", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

func TestCandidateApplyRejectsLateOccupiedTableBeforeAnyBegin(t *testing.T) {
	// A table occupied late in the parent-before-child order must be denied
	// before the first Begin/COPY, so no earlier partial import can happen.
	bundle := candidateFreshBundleForApply(t)
	session := &candidateFreshApplySession{fakeCutoverSession{inspectCounts: map[string]int64{"estimates": 1}}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want late occupied rejection", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 (no partial import before denial)", len(session.txs))
	}
}

func TestCandidateApplyFreshEmptyBandImportsAllTables(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &candidateFreshApplySession{}
	result, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err != nil {
		t.Fatal(err)
	}
	wantTables := len(CutoverTableSpecs(CutoverProfileMedicinesCandidate))
	copyCalls := 0
	for i, tx := range session.txs {
		if !tx.committed {
			t.Fatalf("transaction %d was not committed", i)
		}
		copyCalls += tx.copyCalls
	}
	if copyCalls != wantTables {
		t.Fatalf("copy calls = %d, want %d", copyCalls, wantTables)
	}
	if result.Counts["staffs"] != 1 {
		t.Fatalf("staffs count = %d, want 1", result.Counts["staffs"])
	}
}

// candidateFreshApplySession is the fresh-empty session: every occupancy and
// count probe reports an empty clinic band while the candidate target schema
// checks (medicines numeric(10,2) quantity column) are served.
type candidateFreshApplySession struct {
	fakeCutoverSession
}

func (s *candidateFreshApplySession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
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
	if strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "id >= $1 AND id < $2") {
		// Mirror the real catalog: inspectCounts drives both the count and
		// the occupancy probes so the fresh-only guard sees the same state.
		occupied := s.inspectCounts[candidateTableNameFromBandQuery(query)] > 0
		return staticRow{values: []any{occupied}}
	}
	return s.fakeCutoverSession.QueryRow(ctx, query, args...)
}

// Empty candidate bands remain admissible through the public resume preflight.
func TestCandidateFreshResumePreflightAcceptsEmptyBand(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	if err := PreflightCutoverTargetAllowingResumeWithProvenance(context.Background(), candidateBandQuerier{}, bundle.Manifest, validCutoverSeeds(), bundle.Provenance); err != nil {
		t.Fatal(err)
	}
}

// Target state can change between the all-band probe and the first table
// inspection. A populated table must never be adopted as a candidate prefix.
func TestCandidateFreshRejectsBandOccupiedAfterInitialProbe(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	session := &candidateBandAppearsSession{candidateFreshApplySession{fakeCutoverSession{inspectCounts: map[string]int64{"staffs": 1}}}}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), CutoverRefBandOccupied) || !strings.Contains(err.Error(), "staffs") {
		t.Fatalf("error = %v, want occupied staffs rejection", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("Begin calls = %d, want zero", len(session.txs))
	}
}

type candidateBandAppearsSession struct {
	candidateFreshApplySession
}

func (s *candidateBandAppearsSession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "SELECT EXISTS") && strings.Contains(query, "id >= $1 AND id < $2") {
		return staticRow{values: []any{false}}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

// Profile failures precede every target query, session Exec and transaction.
func TestCandidateFreshProfileGuardsBeforeDB(t *testing.T) {
	bundle := candidateFreshBundleForApply(t)
	for _, tc := range []struct {
		name     string
		contract CutoverProvenanceContract
		want     string
	}{
		{"formal candidate", CutoverProvenanceContract{Mode: CutoverProvenanceFormal, Profile: CutoverProfileMedicinesCandidate}, "activation"},
		{"unknown profile", CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal, Profile: CutoverProfile("unknown-profile")}, "unknown cutover profile"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &cutoverReviewDeniedSession{}
			err := PreflightCutoverTargetAllowingResumeWithProvenance(context.Background(), q, bundle.Manifest, validCutoverSeeds(), tc.contract)
			if err == nil || !strings.Contains(err.Error(), tc.want) || q.calls != 0 {
				t.Fatalf("preflight error = %v, DB calls = %d", err, q.calls)
			}
			q = &cutoverReviewDeniedSession{}
			b := bundle
			b.Provenance = tc.contract
			_, err = applyCutoverCommittingEachTable(context.Background(), q, b, validCutoverSeeds(), pgx.RepeatableRead)
			if err == nil || !strings.Contains(err.Error(), tc.want) || q.calls != 0 {
				t.Fatalf("apply error = %v, DB calls = %d", err, q.calls)
			}
		})
	}
}
