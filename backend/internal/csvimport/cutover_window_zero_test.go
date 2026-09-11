package csvimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowZeroEvidencePreflight(t *testing.T) {
	for _, mode := range []string{"valid", "absent", "count", "null-count", "digest", "contract", "version", "amount", "timestamp", "extra", "source", "local"} {
		t.Run(mode, func(t *testing.T) {
			dir, _ := writeCutoverFixture(t, func(f *fixtureBundle) {
				cols := CutoverTableSpecs()[11].Columns
				row := append([]string(nil), f.rows["billings"][0]...)
				row[columnIndex(cols, "id")] = "1000002"
				row[columnIndex(cols, "medical_record_id")] = ""
				row[columnIndex(cols, "total_amount")] = "-123"
				f.rows["billings"] = append(f.rows["billings"], row)
				f.manifest.Tables[11].RowCount++
			})
			manifestPath := filepath.Join(dir, cutoverManifestName)
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := decodeCutoverManifest(data)
			if err != nil {
				t.Fatal(err)
			}
			spec, table, err := cutoverPaymentContractPart(&manifest, "billings")
			if err != nil {
				t.Fatal(err)
			}
			facts, err := loadCutoverBillingFacts(dir, spec, table)
			if err != nil {
				t.Fatal(err)
			}
			proofFacts := map[int64]cutoverBillingFact{1000002: facts[1000002]}
			if mode == "amount" {
				f := proofFacts[1000002]
				f.totalAmount++
				proofFacts[1000002] = f
			}
			if mode == "timestamp" {
				f := proofFacts[1000002]
				f.completionTimestamp = "2020-01-01T00:00:00Z"
				proofFacts[1000002] = f
			}
			if mode == "extra" {
				proofFacts[1000003] = facts[1000002]
			}
			count, digest, err := windowZeroCSVSetDigest(&manifest, proofFacts, nil)
			if err != nil {
				t.Fatal(err)
			}
			manifest.WindowZeroSettlementEvidence = &CutoverWindowZeroEvidence{
				SchemaVersion: windowZeroSchema, ContractSHA256: cutoverWindowZeroEvidenceContractSHA256,
				RowCount: &count, SHA256: digest,
			}
			switch mode {
			case "absent":
				manifest.WindowZeroSettlementEvidence = nil
			case "count":
				*manifest.WindowZeroSettlementEvidence.RowCount++
			case "null-count":
				manifest.WindowZeroSettlementEvidence.RowCount = nil
			case "digest":
				manifest.WindowZeroSettlementEvidence.SHA256 = strings.Repeat("0", 64)
			case "contract":
				manifest.WindowZeroSettlementEvidence.ContractSHA256 = strings.Repeat("0", 64)
			case "version":
				manifest.WindowZeroSettlementEvidence.SchemaVersion = "unknown"
			case "source":
				manifest.SourceIdentity.Verified = false
			}
			data, err = json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			expected := ExpectedCutoverSource{ManifestSHA256: sha256Hex(data), ClinicCode: "hachioji", ClinicOrdinal: 1, RunID: "run-1"}
			if mode == "local" {
				expected.Provenance.Mode = CutoverProvenanceLocalRehearsal
			}
			_, err = PreflightCutoverBundle(dir, expected)
			if (err == nil) != (mode == "valid") {
				t.Fatalf("unexpected preflight result: %v", err)
			}
		})
	}
}

func TestWindowZeroTimestampPrecision(t *testing.T) {
	for _, value := range []string{"invalid", "2026-01-01T00:00:00.0000001Z", "2026-01-01T00:00:00.0000000Z", "2026-01-01T00:00:00,123Z", "2026-01-01T00:00:00+24:00", "2026-01-01T00:00:00+00:60"} {
		if _, err := windowZeroTimestamp(value); err == nil {
			t.Fatal("invalid timestamp accepted")
		}
	}
	got, err := windowZeroTimestamp("2026-01-01T09:00:00.123456+09:00")
	if err != nil || got != "2026-01-01T00:00:00.123456Z" {
		t.Fatalf("normalization: %s %v", got, err)
	}
}
