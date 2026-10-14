package csvimport

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Codex 4260555439: the four master price columns carry int64 producer
// values (int64 can be rounded by a double precision target column), so the
// target must expose each price column as bigint before apply. Legacy21
// guards the three legacy masters; candidate22 also guards medicines.price.
func TestTargetRequiresBigintMasterPriceColumns(t *testing.T) {
	prices := [][2]string{
		{"procedures", "price"},
		{"vaccines", "price"},
		{"merchandise_items", "unit_price"},
	}
	for _, tableColumn := range prices {
		for _, wrongType := range []string{"double precision", "numeric", "integer"} {
			t.Run("legacy rejects "+tableColumn[0]+"."+tableColumn[1]+" as "+wrongType, func(t *testing.T) {
				q := &masterPriceTypeTargetQuerier{columnTypes: map[[2]string]string{tableColumn: wrongType}}
				err := PreflightCutoverTarget(context.Background(), q, cutoverManifestForTargetTests(), validCutoverSeeds())
				if err == nil || !strings.Contains(err.Error(), "public."+tableColumn[0]+"."+tableColumn[1]) || !strings.Contains(err.Error(), "bigint") {
					t.Fatalf("PreflightCutoverTarget() error = %v, want %s.%s bigint rejection", err, tableColumn[0], tableColumn[1])
				}
			})
		}
	}
	t.Run("legacy accepts bigint prices", func(t *testing.T) {
		if err := PreflightCutoverTarget(context.Background(), &masterPriceTypeTargetQuerier{}, cutoverManifestForTargetTests(), validCutoverSeeds()); err != nil {
			t.Fatalf("PreflightCutoverTarget() error = %v, want bigint master prices accepted", err)
		}
	})
	t.Run("legacy rejects only the tampered table", func(t *testing.T) {
		q := &masterPriceTypeTargetQuerier{columnTypes: map[[2]string]string{{"vaccines", "price"}: "double precision"}}
		err := PreflightCutoverTarget(context.Background(), q, cutoverManifestForTargetTests(), validCutoverSeeds())
		if err == nil || !strings.Contains(err.Error(), "public.vaccines.price") || strings.Contains(err.Error(), "procedures") {
			t.Fatalf("PreflightCutoverTarget() error = %v, want a vaccines.price-only rejection", err)
		}
	})

	t.Run("candidate rejects medicines.price as double precision", func(t *testing.T) {
		q := &masterPriceTypeTargetQuerier{columnTypes: map[[2]string]string{{"medicines", "price"}: "double precision"}}
		err := PreflightCutoverTargetWithProvenance(
			context.Background(), q, candidateManifestForTargetTests(), validCutoverSeeds(),
			CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
		)
		if err == nil || !strings.Contains(err.Error(), "public.medicines.price") || !strings.Contains(err.Error(), "bigint") {
			t.Fatalf("PreflightCutoverTargetWithProvenance() error = %v, want public.medicines.price bigint rejection", err)
		}
	})
	t.Run("candidate accepts bigint medicines.price", func(t *testing.T) {
		if err := PreflightCutoverTargetWithProvenance(
			context.Background(), &masterPriceTypeTargetQuerier{}, candidateManifestForTargetTests(), validCutoverSeeds(),
			CutoverProvenanceContract{Mode: CutoverProvenanceStagingRehearsal, Profile: CutoverProfileMedicinesCandidate},
		); err != nil {
			t.Fatalf("PreflightCutoverTargetWithProvenance() error = %v, want candidate master prices accepted", err)
		}
	})
	t.Run("legacy never probes medicines.price", func(t *testing.T) {
		q := &masterPriceTypeTargetQuerier{}
		if err := PreflightCutoverTarget(context.Background(), q, cutoverManifestForTargetTests(), validCutoverSeeds()); err != nil {
			t.Fatalf("PreflightCutoverTarget() error = %v", err)
		}
		if q.dataTypeQueries[[2]string{"medicines", "price"}] != 0 {
			t.Fatal("legacy21 target preflight probed public.medicines.price")
		}
	})
}

// Codex 4260555456: EMR-342 DESIGN 3.2 pins medicines is_active=true and
// is_non_insurance=false. Source preflight accepts the producer canonical
// t/f and the semantic true/false lexical forms only; wrong semantics,
// empty, and non-boolean lexical values fail closed without coercion and
// without leaking the rejected cell value.
func TestCutoverMedicinesFlagsMatchDesignPins(t *testing.T) {
	spec := cutoverSpecByName(t, "medicines")
	activeCol := columnIndex(spec.Columns, "is_active")
	nonInsuranceCol := columnIndex(spec.Columns, "is_non_insurance")

	t.Run("accepts canonical and semantic lexical forms", func(t *testing.T) {
		for _, tc := range []struct{ active, nonInsurance string }{
			{"t", "f"}, {"true", "false"},
		} {
			t.Run(tc.active+"/"+tc.nonInsurance, func(t *testing.T) {
				dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
					f.rows["medicines"][0][activeCol] = tc.active
					f.rows["medicines"][0][nonInsuranceCol] = tc.nonInsurance
				})
				if _, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest)); err != nil {
					t.Fatalf("PreflightCutoverBundle error = %v, want %q/%q accepted", err, tc.active, tc.nonInsurance)
				}
			})
		}
	})

	rejections := []struct{ label, active, nonInsurance string }{
		{"empty active", "", "f"},
		{"empty non insurance", "t", ""},
		{"inverted active", "f", "f"},
		{"inverted non insurance", "t", "t"},
		{"numeric active", "1", "f"},
		{"numeric non insurance", "t", "0"},
		{"uppercase active", "TRUE", "f"},
		{"unknown non insurance", "t", "yes"},
	}
	for _, tc := range rejections {
		t.Run("rejects "+tc.label, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				f.rows["medicines"][0][activeCol] = tc.active
				f.rows["medicines"][0][nonInsuranceCol] = tc.nonInsurance
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if err == nil {
				t.Fatalf("PreflightCutoverBundle accepted %q/%q", tc.active, tc.nonInsurance)
			}
			if !strings.Contains(err.Error(), "table medicines") || !strings.Contains(err.Error(), "row 2") {
				t.Fatalf("PreflightCutoverBundle error = %v, want a table/column/row diagnostic", err)
			}
			for _, leaked := range []string{tc.active, tc.nonInsurance} {
				// Single-letter canonical forms appear in the allowed-value
				// hint, so the no-leak guard only applies to longer values.
				if len(leaked) > 1 && strings.Contains(err.Error(), leaked) {
					t.Fatal("error leaks the rejected cell value")
				}
			}
		})
	}
}

type masterPriceTypeTargetQuerier struct {
	validTargetQuerier
	columnTypes     map[[2]string]string
	dataTypeQueries map[[2]string]int
}

func (q *masterPriceTypeTargetQuerier) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if len(args) == 2 && strings.Contains(query, "data_type FROM information_schema.columns") {
		key := [2]string{args[0].(string), args[1].(string)}
		if q.dataTypeQueries == nil {
			q.dataTypeQueries = map[[2]string]int{}
		}
		q.dataTypeQueries[key]++
		if override, ok := q.columnTypes[key]; ok {
			return staticRow{values: []any{override}}
		}
		if key == [2]string{"medicines", "default_quantity"} {
			return staticRow{values: []any{"numeric"}}
		}
		if key == [2]string{"medicines", "tax_rate"} {
			// Canonical 001 schema: tax_rate is numeric (unbounded).
			return staticRow{values: []any{"numeric"}}
		}
	}
	if len(args) == 2 && strings.Contains(query, "numeric_precision") && args[0] == "medicines" {
		// Canonical 001 schema: tax_rate is unconstrained numeric.
		if args[1] == "tax_rate" {
			return staticRow{values: []any{nil, nil}}
		}
		return staticRow{values: []any{int64(10), int64(2)}}
	}
	return q.validTargetQuerier.QueryRow(ctx, query, args...)
}
