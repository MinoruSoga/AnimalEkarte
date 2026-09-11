package csvimport

import (
	"strings"
	"testing"
)

// These vectors also appear in the producer's test-window-zero-settlement-evidence.mjs.
// Fixed expected bytes guard the Node/Go/SQL boundary independently of fixture builders.
func windowZeroVectorManifest() CutoverManifest {
	a := strings.Repeat("a", 64)
	route := "complete_base"
	return CutoverManifest{ClinicCode: "jouto", ClinicOrdinal: 2, ClinicBandBase: 10000000,
		ClinicBandEndExclusive: 20000000, StageIDOffset: 11000000,
		IDBand:      CutoverIDBand{Base: 10000000, NonOwnerIDOffset: 11000000, EndExclusive: 20000000, OwnerFloor: 10300000, ApplicationIDFloor: 1000000000},
		SourceRunID: "synthetic-run", StageBuildID: "00000000-0000-4000-8000-000000000001",
		StageMappingSHA256: a, CSVContractSHA256: a, OutputDir: "synthetic/output",
		SourceIdentity:      CutoverSourceIdentity{SourceBackupSHA256: &a, BaseArchiveSHA256: &a, KnjoProvenanceRoute: &route},
		SourceSummarySHA256: CutoverLayerDigests{Stage: a},
		Tables:              []CutoverManifestTable{{Table: "billings", SHA256: a}, {Table: "payments", SHA256: a}, {Table: "payment_splits", SHA256: a}},
	}
}

func TestWindowZeroCrossLanguageEncoding(t *testing.T) {
	m := windowZeroVectorManifest()
	count, empty, err := windowZeroCSVSetDigest(&m, nil, nil)
	if err != nil || count != 0 || empty != "b80035d631013450988d5fc02884f6da44b63b2481446dd1d3e6494ec2bb8e0d" {
		t.Fatalf("empty vector mismatch: %v", err)
	}
	facts := map[int64]cutoverBillingFact{11000001: {totalAmount: 5000, status: "completed", completionTimestamp: "2026-09-07T10:02:03.123456+09:00"}}
	count, digest, err := windowZeroCSVSetDigest(&m, facts, nil)
	if err != nil || count != 1 || digest != "4af66e97c0b6d675ee3f820cedd3987043aaae9480008e2481320e176e497872" {
		t.Fatalf("tuple vector mismatch: %v", err)
	}
	for _, mutate := range []func(*CutoverManifest){
		func(m *CutoverManifest) { m.StageIDOffset++ }, func(m *CutoverManifest) { m.IDBand.EndExclusive++ },
		func(m *CutoverManifest) { m.SourceRunID = "different" }, func(m *CutoverManifest) { m.OutputDir = "other/revision" },
		func(m *CutoverManifest) { m.StageBuildID = "00000000-0000-4000-8000-000000000002" },
		func(m *CutoverManifest) { m.ClinicCode = "other" }, func(m *CutoverManifest) { m.SourceSummarySHA256.Stage = strings.Repeat("b", 64) },
	} {
		changed := m
		mutate(&changed)
		_, actual, err := windowZeroCSVSetDigest(&changed, nil, nil)
		if err == nil && actual == empty {
			t.Fatal("empty evidence was not bound to changed context")
		}
	}
}

func TestWindowZeroEncodingRejectsInvalidInputs(t *testing.T) {
	m := windowZeroVectorManifest()
	for _, tc := range []struct {
		id    int64
		stamp string
	}{{1, "2026-01-01T00:00:00Z"}, {11000001, "invalid"}} {
		_, _, err := windowZeroCSVSetDigest(&m, map[int64]cutoverBillingFact{tc.id: {totalAmount: 1, status: "completed", completionTimestamp: tc.stamp}}, nil)
		if err == nil {
			t.Fatal("invalid subject accepted")
		}
	}
	for _, mutate := range []func(*CutoverManifest){
		func(m *CutoverManifest) { m.SourceIdentity.SourceBackupSHA256 = nil },
		func(m *CutoverManifest) { m.Tables = nil }, func(m *CutoverManifest) { m.OutputDir = "bad\nheader" },
		func(m *CutoverManifest) { m.OutputDir = strings.Repeat("a", 513) },
	} {
		changed := m
		mutate(&changed)
		if _, err := windowZeroHeader(&changed); err == nil {
			t.Fatal("invalid header accepted")
		}
	}
}
