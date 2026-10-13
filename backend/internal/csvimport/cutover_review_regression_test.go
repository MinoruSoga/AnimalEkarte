package csvimport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCutoverReviewMerchandisePriceRequired(t *testing.T) {
	for _, table := range []string{"procedures", "medicines", "vaccines", "merchandise_items"} {
		t.Run(table, func(t *testing.T) {
			column := "price"
			if table == "merchandise_items" {
				column = "unit_price"
			}
			err := validateCutoverNonNegativeInt(table, column, "", 2)
			if table == "merchandise_items" {
				if err == nil || !strings.Contains(err.Error(), "required") {
					t.Fatalf("error = %v, want required price rejection", err)
				}
			} else if err != nil {
				t.Fatalf("nullable price rejected: %v", err)
			}
		})
	}
}

func TestCutoverReviewMedicineQuantityExactNumeric(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"", false}, {"0", true}, {"0.01", true}, {"99999999.99", true}, {"1.2300", true}, {"000000001.00", true},
		{"0.001", false}, {"99999999.999", false}, {"100000000", false}, {"-1", false}, {"NaN", false}, {"1e2", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			// Lexical numeric(10,2) precision is a property of the general
			// helper, so zero and other representable values stay valid
			// here; the public row contract additionally pins one (see the
			// fixed-mapping bundle tests).
			err := validateCutoverMedicineQuantity(tc.value, 2)
			if (err == nil) != tc.valid {
				t.Fatalf("quantity valid=%v, error=%v", tc.valid, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "default_quantity") || !strings.Contains(err.Error(), "row 2")) {
				t.Fatalf("unsafe diagnostic: %v", err)
			}
		})
	}
}

// EMR-342 DESIGN 3.2/3.3: the old database has no medicine_unit source, so
// every imported row must carry an empty medicine_unit. Even target enum
// labels have no producer source and fail preflight. Diagnostics name the
// table, column, and row only — never the rejected cell value.
func TestCutoverReviewMedicineUnitMustStayEmpty(t *testing.T) {
	indexes := map[string]int{"dosage_form": 0, "medicine_unit": 1, "default_quantity": 2, "is_active": 3, "is_non_insurance": 4, "tax_type": 5, "tax_rate": 6}
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"", true},
		{"per_tablet", false}, {"per_ml", false}, {"per_dose", false}, {"per_gram", false},
		{"per_pill", false}, {"PER_TABLET", false}, {"tablet", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			err := validateCutoverMedicinesRow([]string{"", tc.value, "1.0", "t", "f", "excluded", "0.10"}, indexes, 2)
			if (err == nil) != tc.valid {
				t.Fatalf("medicine_unit valid=%v, error=%v", tc.valid, err)
			}
			if err != nil {
				if !strings.Contains(err.Error(), "medicine_unit") || !strings.Contains(err.Error(), "row 2") {
					t.Fatalf("unsafe diagnostic: %v", err)
				}
				if tc.value != "" && strings.Contains(err.Error(), tc.value) {
					t.Fatal("error leaks the source cell value")
				}
			}
		})
	}
}

func TestCutoverReviewUnapprovedProducerBeforeDatabase(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil, CutoverProfileMedicinesCandidate)
	contents, err := os.ReadFile(filepath.Join(dir, cutoverManifestName))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeCutoverManifest(contents)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []CutoverProvenanceMode{"", CutoverProvenanceFormal, CutoverProvenanceLocalRehearsal, CutoverProvenanceStagingRehearsal} {
		t.Run(string(mode), func(t *testing.T) {
			contract := stagingExpectedCutoverSource(digest).Provenance
			contract.Mode = mode
			contract.Profile = CutoverProfileMedicinesCandidate
			expected := formalExpectedSource(digest)
			expected.Provenance = contract
			if _, err := PreflightCutoverBundle(dir, expected); err == nil || !strings.Contains(err.Error(), "activation") {
				t.Fatalf("preflight error=%v, want activation rejection", err)
			}
			bundle := CutoverBundle{Manifest: manifest, Provenance: contract}
			began := false
			_, err := applyCutoverWithBegin(context.Background(), func(context.Context) (cutoverTransaction, error) {
				began = true
				return nil, errors.New("unexpected begin")
			}, bundle, validCutoverSeeds())
			if began || err == nil || !strings.Contains(err.Error(), "activation") {
				t.Fatalf("began=%v, error=%v", began, err)
			}
			session := &cutoverReviewDeniedSession{}
			_, err = applyCutoverCommittingEachTable(context.Background(), session, bundle, validCutoverSeeds(), pgx.RepeatableRead)
			if session.calls != 0 || err == nil || !strings.Contains(err.Error(), "activation") {
				t.Fatalf("session calls=%d, error=%v", session.calls, err)
			}
		})
	}
}

type cutoverReviewDeniedSession struct{ calls int }

func (s *cutoverReviewDeniedSession) QueryRow(context.Context, string, ...any) pgx.Row {
	s.calls++
	return staticRow{values: []any{int64(0)}}
}
func (s *cutoverReviewDeniedSession) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	s.calls++
	return pgconn.CommandTag{}, errors.New("unexpected exec")
}
func (s *cutoverReviewDeniedSession) Begin(context.Context, pgx.TxIsoLevel) (cutoverTransaction, error) {
	s.calls++
	return nil, errors.New("unexpected begin")
}

func TestCutoverReviewValuePreflight(t *testing.T) {
	for _, tc := range []struct {
		name, table, column, value string
		valid                      bool
	}{
		{"missing merchandise price", "merchandise_items", "unit_price", "", false},
		{"zero merchandise price", "merchandise_items", "unit_price", "0", true},
		{"quantity rounding", "medicines", "default_quantity", "0.001", false},
		{"quantity overflow", "medicines", "default_quantity", "100000000", false},
		// 4264495015 contract change: the public bundle pins the quantity
		// to one, so zero and other representable values are rejected even
		// though they remain lexically valid in the numeric(10,2) helper.
		{"quantity zero", "medicines", "default_quantity", "0", false},
		{"quantity exact trailing zeros", "medicines", "default_quantity", "1.2300", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeMedicinesCandidateFixture(t, func(f *fixtureBundle) {
				spec := cutoverSpecByName(t, tc.table)
				f.rows[tc.table][0][columnIndex(spec.Columns, tc.column)] = tc.value
			})
			_, err := PreflightCutoverBundle(dir, candidateExpectedSource(digest))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
			if err != nil && (!strings.Contains(err.Error(), tc.table) || !strings.Contains(err.Error(), tc.column) || !strings.Contains(err.Error(), "row 2")) {
				t.Fatalf("diagnostic=%v", err)
			}
		})
	}
}

func TestCutoverReviewPublicApplyRejectsBeforePoolAcquire(t *testing.T) {
	bundle := CutoverBundle{Manifest: CutoverManifest{Status: "PASS", HandoffEligibility: "TRUSTED_CANDIDATE"}, Provenance: CutoverProvenanceContract{Profile: CutoverProfileMedicinesCandidate}}
	// A nil pool is safe only when activation rejects before resource access.
	_, err := ApplyCutoverCommittingEachTable(context.Background(), nil, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), "activation") {
		t.Fatalf("error=%v", err)
	}
}
