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
		{"", true}, {"0", true}, {"0.01", true}, {"99999999.99", true}, {"1.2300", true}, {"000000001.00", true},
		{"0.001", false}, {"99999999.999", false}, {"100000000", false}, {"-1", false}, {"NaN", false}, {"1e2", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			err := validateCutoverMedicinesRow([]string{"", tc.value}, map[string]int{"dosage_form": 0, "default_quantity": 1}, 2)
			if (err == nil) != tc.valid {
				t.Fatalf("quantity valid=%v, error=%v", tc.valid, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "default_quantity") || !strings.Contains(err.Error(), "row 2")) {
				t.Fatalf("unsafe diagnostic: %v", err)
			}
		})
	}
}

func TestCutoverReviewUnapprovedProducerBeforeDatabase(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
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
		{"quantity zero", "medicines", "default_quantity", "0", true},
		{"quantity exact trailing zeros", "medicines", "default_quantity", "1.2300", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, digest := writeCutoverRehearsalFixture(t, func(f *fixtureBundle) {
				spec := cutoverSpecByName(t, tc.table)
				f.rows[tc.table][0][columnIndex(spec.Columns, tc.column)] = tc.value
			})
			_, err := PreflightCutoverBundle(dir, localRehearsalExpectedSource(digest))
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
	bundle := CutoverBundle{Manifest: CutoverManifest{Status: "PASS", HandoffEligibility: "TRUSTED_CANDIDATE"}}
	// A nil pool is safe only when activation rejects before resource access.
	_, err := ApplyCutoverCommittingEachTable(context.Background(), nil, bundle, validCutoverSeeds(), pgx.RepeatableRead)
	if err == nil || !strings.Contains(err.Error(), "activation") {
		t.Fatalf("error=%v", err)
	}
}
