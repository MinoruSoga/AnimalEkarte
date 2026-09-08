package csvimport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type windowZeroTargetQuerier struct {
	count   int64
	digest  string
	invalid int64
}

func (q windowZeroTargetQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	return staticRow{values: []any{q.count, q.digest, q.invalid}}
}

func TestWindowZeroTargetRequiresExactSet(t *testing.T) {
	dir, digest := writeCutoverFixture(t, nil)
	bundle, err := PreflightCutoverBundle(dir, ExpectedCutoverSource{ManifestSHA256: digest, ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	m := bundle.Manifest
	rowCount := int64(1)
	m.WindowZeroSettlementEvidence = &CutoverWindowZeroEvidence{SchemaVersion: windowZeroSchema,
		ContractSHA256: cutoverWindowZeroEvidenceContractSHA256, RowCount: &rowCount, SHA256: strings.Repeat("a", 64)}
	for _, tc := range []struct {
		name    string
		count   int64
		digest  string
		invalid int64
		pass    bool
	}{
		{"match", 1, strings.Repeat("a", 64), 0, true},
		{"missing", 0, strings.Repeat("a", 64), 0, false},
		{"extra", 2, strings.Repeat("a", 64), 0, false},
		{"same count substitution", 1, strings.Repeat("b", 64), 0, false},
		{"invalid attachment", 1, strings.Repeat("a", 64), 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyWindowZeroTarget(context.Background(), windowZeroTargetQuerier{tc.count, tc.digest, tc.invalid}, &m, CutoverSeedIDs{ClinicID: 1}, CutoverProvenanceContract{})
			if (err == nil) != tc.pass {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
	if err := verifyWindowZeroTarget(context.Background(), errorTargetQuerier{err: errUnexpectedQuery}, &m, CutoverSeedIDs{ClinicID: 1}, CutoverProvenanceContract{}); !errors.Is(err, errUnexpectedQuery) {
		t.Fatalf("query failure not preserved: %v", err)
	}
	if err := verifyWindowZeroTarget(context.Background(), validTargetQuerier{}, &m, CutoverSeedIDs{ClinicID: 1}, CutoverProvenanceContract{Mode: CutoverProvenanceLocalRehearsal}); err == nil {
		t.Fatal("target accepted local rehearsal evidence")
	}
	m.SourceIdentity.SourceBackupSHA256 = nil
	if err := verifyWindowZeroTarget(context.Background(), validTargetQuerier{}, &m, CutoverSeedIDs{ClinicID: 1}, CutoverProvenanceContract{}); err == nil {
		t.Fatal("target accepted incomplete evidence header")
	}
}
