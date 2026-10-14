package csvimport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Codex 4265970828: the contract pins medicines.tax_rate to exactly 0.10, so
// the target column must be numeric capable of storing 0.10 exactly. The
// canonical 001 schema keeps the column as unconstrained numeric (NULL
// precision/scale), so unbounded numeric is accepted; a bounded numeric
// needs scale >= 1 and precision >= scale (numeric(1,1) stores 0.1 exactly).
// A drifted numeric(...,0) would round 0.10 to 0 and must be denied before
// apply. Legacy21 never inspects medicines columns.
type taxScaleTargetQuerier struct {
	validTargetQuerier
	precision    *int64
	scale        *int64
	taxRateType  string
	probeErr     error
	taxProbes    int
	taxRateTyped int
}

func (q *taxScaleTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if len(args) == 2 && args[0] == "medicines" && strings.Contains(query, "information_schema.columns") {
		if strings.Contains(query, "numeric_precision") {
			if args[1] == "tax_rate" {
				q.taxProbes++
				if q.probeErr != nil {
					return staticRow{err: q.probeErr}
				}
				if q.precision == nil && q.scale == nil {
					return staticRow{values: []any{nil, nil}}
				}
				var p, sc any
				if q.precision != nil {
					p = *q.precision
				}
				if q.scale != nil {
					sc = *q.scale
				}
				return staticRow{values: []any{p, sc}}
			}
			if args[1] == "default_quantity" {
				return staticRow{values: []any{int64(10), int64(2)}}
			}
		}
		if strings.Contains(query, "data_type") && args[1] == "tax_rate" {
			q.taxRateTyped++
			t := q.taxRateType
			if t == "" {
				t = "numeric"
			}
			return staticRow{values: []any{t}}
		}
		if strings.Contains(query, "data_type") && args[1] == "default_quantity" {
			return staticRow{values: []any{"numeric"}}
		}
	}
	return q.validTargetQuerier.QueryRow(ctx, query, args...)
}

func (q *taxScaleTargetQuerier) runCandidate(t *testing.T) error {
	t.Helper()
	return PreflightCutoverTargetWithProvenance(
		context.Background(),
		q,
		candidateManifestForTargetTests(),
		validCutoverSeeds(),
		CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
	)
}

func TestCandidateTargetTaxRateScale(t *testing.T) {
	t.Run("accepts unconstrained numeric", func(t *testing.T) {
		if err := (&taxScaleTargetQuerier{}).runCandidate(t); err != nil {
			t.Fatalf("error = %v, want canonical unbounded numeric accepted", err)
		}
	})
	t.Run("accepts bounded numeric(1,1)", func(t *testing.T) {
		if err := (&taxScaleTargetQuerier{precision: int64Pointer(1), scale: int64Pointer(1)}).runCandidate(t); err != nil {
			t.Fatalf("error = %v, want numeric(1,1) accepted", err)
		}
	})
	t.Run("accepts bounded numeric(3,2)", func(t *testing.T) {
		if err := (&taxScaleTargetQuerier{precision: int64Pointer(3), scale: int64Pointer(2)}).runCandidate(t); err != nil {
			t.Fatalf("error = %v, want numeric(3,2) accepted", err)
		}
	})
	for _, tc := range []struct {
		name             string
		precision, scale *int64
		taxRateType      string
		probeErr         bool
		want             string
	}{
		{name: "rejects scale zero", precision: int64Pointer(2), scale: int64Pointer(0), want: "tax_rate"},
		{name: "rejects negative scale", precision: int64Pointer(3), scale: int64Pointer(-1), want: "tax_rate"},
		{name: "rejects precision below scale", precision: int64Pointer(1), scale: int64Pointer(2), want: "tax_rate"},
		{name: "rejects precision without scale", precision: int64Pointer(3), want: "tax_rate"},
		{name: "rejects scale without precision", scale: int64Pointer(2), want: "tax_rate"},
		{name: "rejects non-numeric type", taxRateType: "double precision", want: "tax_rate"},
		{name: "fails closed on catalog error", probeErr: true, want: "tax_rate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &taxScaleTargetQuerier{precision: tc.precision, scale: tc.scale, taxRateType: tc.taxRateType}
			if tc.probeErr {
				q.probeErr = errors.New("catalog lookup failed")
			}
			err := q.runCandidate(t)
			if err == nil {
				t.Fatal("PreflightCutoverTargetWithProvenance accepted a tax_rate schema that cannot store 0.10 exactly")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q named", err, tc.want)
			}
			if tc.probeErr && !errors.Is(err, q.probeErr) {
				t.Fatalf("error = %v, want wrapped catalog error", err)
			}
		})
	}

	t.Run("legacy never probes medicines tax_rate", func(t *testing.T) {
		q := &taxScaleTargetQuerier{}
		if err := PreflightCutoverTarget(context.Background(), q, cutoverManifestForTargetTests(), validCutoverSeeds()); err != nil {
			t.Fatalf("PreflightCutoverTarget() error = %v", err)
		}
		if q.taxProbes != 0 || q.taxRateTyped != 0 {
			t.Fatalf("legacy probes tax_rate %d/%d times, want 0", q.taxProbes, q.taxRateTyped)
		}
	})
}
