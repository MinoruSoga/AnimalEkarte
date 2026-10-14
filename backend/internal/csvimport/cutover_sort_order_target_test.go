package csvimport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Codex 4266561071: the candidate medicines source sort_order is a signed
// int32 decimal (DESIGN 3.2 legacy order), and the canonical 001 target
// column is integer. A drifted smaller integer target would fail COPY only
// after earlier tables already committed — unrecoverable under fresh-only
// decision A. The candidate target preflight therefore requires data_type
// integer for public.medicines.sort_order. Legacy21 never inspects it.
type sortOrderTargetQuerier struct {
	validTargetQuerier
	sortType   string
	catalogErr error
	sortProbes int
}

func (q *sortOrderTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if len(args) == 2 && args[0] == "medicines" && strings.Contains(query, "information_schema.columns") {
		if strings.Contains(query, "numeric_precision") {
			return staticRow{values: []any{int64(10), int64(2)}}
		}
		if strings.Contains(query, "data_type") {
			if args[1] == "default_quantity" || args[1] == "tax_rate" {
				return staticRow{values: []any{"numeric"}}
			}
			if args[1] == "sort_order" {
				q.sortProbes++
				if q.catalogErr != nil {
					return staticRow{err: q.catalogErr}
				}
				return staticRow{values: []any{q.sortType}}
			}
		}
	}
	return q.validTargetQuerier.QueryRow(ctx, query, args...)
}

func (q *sortOrderTargetQuerier) runCandidatePreflight(t *testing.T) error {
	t.Helper()
	return PreflightCutoverTargetWithProvenance(
		context.Background(),
		q,
		candidateManifestForTargetTests(),
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	)
}

func TestSortOrderTargetPreflightRejectsNonIntegerTypes(t *testing.T) {
	for _, wrongType := range []string{"smallint", "bigint", "numeric", "text"} {
		t.Run("rejects "+wrongType, func(t *testing.T) {
			q := &sortOrderTargetQuerier{sortType: wrongType}
			err := q.runCandidatePreflight(t)
			if err == nil || !strings.Contains(err.Error(), "public.medicines.sort_order") || !strings.Contains(err.Error(), "integer") {
				t.Fatalf("preflight error = %v, want public.medicines.sort_order integer rejection", err)
			}
			if q.sortProbes == 0 {
				t.Fatal("candidate preflight never probed medicines.sort_order")
			}
		})
	}
}

func TestSortOrderTargetIntegerAccepted(t *testing.T) {
	q := &sortOrderTargetQuerier{sortType: "integer"}
	if err := q.runCandidatePreflight(t); err != nil {
		t.Fatalf("preflight error = %v, want integer sort_order accepted", err)
	}
	if q.sortProbes == 0 {
		t.Fatal("candidate preflight never probed medicines.sort_order")
	}
}

func TestSortOrderTargetCatalogErrorFailsClosed(t *testing.T) {
	q := &sortOrderTargetQuerier{catalogErr: errors.New("catalog lookup failed")}
	err := q.runCandidatePreflight(t)
	if err == nil || !strings.Contains(err.Error(), "medicines.sort_order") {
		t.Fatalf("preflight error = %v, want sort_order catalog failure", err)
	}
}

func TestSortOrderTargetLegacyNeverProbes(t *testing.T) {
	q := &sortOrderTargetQuerier{}
	if err := PreflightCutoverTarget(context.Background(), q, cutoverManifestForTargetTests(), validCutoverSeeds()); err != nil {
		t.Fatalf("PreflightCutoverTarget() error = %v", err)
	}
	if q.sortProbes != 0 {
		t.Fatalf("legacy21 preflight issued %d sort_order probes, want 0", q.sortProbes)
	}
}

// sortOrderApplySession drives the table-per-commit apply with a candidate
// target whose medicines.sort_order data type is configurable.
type sortOrderApplySession struct {
	candidateFreshApplySession
	sortType   string
	sortProbes int
}

func (s *sortOrderApplySession) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if len(args) == 2 && args[0] == "medicines" && strings.Contains(query, "information_schema.columns") &&
		strings.Contains(query, "data_type") && args[1] == "sort_order" {
		s.sortProbes++
		return staticRow{values: []any{s.sortType}}
	}
	return s.candidateFreshApplySession.QueryRow(ctx, query, args...)
}

func candidateFreshBundleWithSortOrder(t *testing.T, value string) CutoverBundle {
	t.Helper()
	spec := cutoverSpecByName(t, "medicines")
	sortCol := columnIndex(spec.Columns, "sort_order")
	dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
		f.rows["medicines"][0][sortCol] = value
	})
	bundle, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestSortOrderTargetApplyRejectsSmallintBeforeBegin(t *testing.T) {
	// Source 40000 is a valid int32 sort order that preflight accepts; a
	// smallint target must be denied before the first Begin/COPY.
	bundle := candidateFreshBundleWithSortOrder(t, "40000")
	session := &sortOrderApplySession{sortType: "smallint"}
	_, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), "public.medicines.sort_order") || !strings.Contains(err.Error(), "integer") {
		t.Fatalf("applyCutoverCommittingEachTable() error = %v, want smallint sort_order rejection", err)
	}
	if len(session.txs) != 0 {
		t.Fatalf("transactions = %d, want 0 before denial", len(session.txs))
	}
}

func TestSortOrderTargetIntegerFreshApplySucceeds(t *testing.T) {
	bundle := candidateFreshBundleWithSortOrder(t, "40000")
	session := &sortOrderApplySession{sortType: "integer"}
	if _, err := applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead); err != nil {
		t.Fatal(err)
	}
	want := len(CutoverTableSpecs(CutoverProfileMedicinesCandidate))
	copyCalls := 0
	for _, tx := range session.txs {
		copyCalls += tx.copyCalls
	}
	if copyCalls != want {
		t.Fatalf("copy calls = %d, want %d", copyCalls, want)
	}
}
