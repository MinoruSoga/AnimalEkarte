package csvimport

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const windowZeroSchema = "window-zero-settlement-v1"

var windowZeroTimestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?(?:Z|[+-]\d{2}:\d{2})$`)

// Pinned to the producer's canonical schema and evidence implementation.
// Updated together with scripts/test-support/test-stage-csv-output.mjs.
const cutoverWindowZeroEvidenceContractSHA256 = "c7db2104428a4370b8f7536adb9cacf0968bacad035ba6cb2e8348218ac63699"

// CutoverWindowZeroEvidence attests the exact source-proven no-tender subset,
// not arbitrary completed billings lacking payment records.
type CutoverWindowZeroEvidence struct {
	SchemaVersion  string `json:"schemaVersion"`
	ContractSHA256 string `json:"contractSha256"`
	RowCount       *int64 `json:"rowCount"`
	SHA256         string `json:"sha256"`
}

func validateWindowZeroEvidence(m CutoverManifest, p CutoverProvenanceContract) error {
	e := m.WindowZeroSettlementEvidence
	if e == nil {
		return nil
	}
	if p.Mode == CutoverProvenanceLocalRehearsal || m.Status != "PASS" ||
		m.HandoffEligibility != "TRUSTED_CANDIDATE" || !m.SourceComplete ||
		!m.SourceProvenanceVerified || !m.SourceIdentity.Verified || m.SourceCompletenessStatus != "PASS" {
		return fmt.Errorf("window-zero evidence requires fully verified trusted source")
	}
	if e.SchemaVersion != windowZeroSchema || e.ContractSHA256 != cutoverWindowZeroEvidenceContractSHA256 ||
		e.RowCount == nil || *e.RowCount < 0 || *e.RowCount > maxCutoverPaymentRows || !validSHA256(e.SHA256) || strings.ToLower(e.SHA256) != e.SHA256 {
		return fmt.Errorf("window-zero evidence contract is invalid")
	}
	return nil
}

func windowZeroHeader(m *CutoverManifest) (string, error) {
	identity := m.SourceIdentity
	if identity.SourceBackupSHA256 == nil || identity.BaseArchiveSHA256 == nil {
		return "", fmt.Errorf("window-zero evidence source identity is missing")
	}
	knjo := ""
	if identity.KNJOArchiveSHA256 != nil {
		knjo = *identity.KNJOArchiveSHA256
	}
	fields := []string{windowZeroSchema, m.ClinicCode, strconv.FormatInt(m.ClinicOrdinal, 10),
		strconv.FormatInt(m.ClinicBandBase, 10), strconv.FormatInt(m.ClinicBandEndExclusive, 10),
		strconv.FormatInt(m.StageIDOffset, 10), strconv.FormatInt(m.IDBand.Base, 10),
		strconv.FormatInt(m.IDBand.NonOwnerIDOffset, 10), strconv.FormatInt(m.IDBand.EndExclusive, 10),
		strconv.FormatInt(m.IDBand.OwnerFloor, 10), strconv.FormatInt(m.IDBand.ApplicationIDFloor, 10),
		m.SourceRunID, m.StageBuildID, m.StageMappingSHA256, m.CSVContractSHA256,
		m.OutputDir, *identity.SourceBackupSHA256, *identity.BaseArchiveSHA256, knjo, m.SourceSummarySHA256.Stage}
	for _, name := range []string{"billings", "payments", "payment_splits"} {
		_, table, err := cutoverPaymentContractPart(m, name)
		if err != nil {
			return "", err
		}
		fields = append(fields, table.SHA256)
	}
	for _, field := range fields {
		if len(field) > 512 {
			return "", fmt.Errorf("window-zero evidence header exceeds limit")
		}
		for _, c := range field {
			if c < 32 || c > 126 {
				return "", fmt.Errorf("window-zero evidence header is not printable ASCII")
			}
		}
	}
	return strings.Join(fields, "\n") + "\n", nil
}

func windowZeroTimestamp(value string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || !windowZeroTimestampPattern.MatchString(value) || t.Nanosecond()%1000 != 0 || t.UTC().Year() < 1 || t.UTC().Year() > 9999 {
		return "", fmt.Errorf("window-zero completion timestamp is invalid")
	}
	if !strings.HasSuffix(value, "Z") && (value[len(value)-5:len(value)-3] > "23" || value[len(value)-2:] > "59") {
		return "", fmt.Errorf("window-zero completion timestamp offset is invalid")
	}
	return t.UTC().Format("2006-01-02T15:04:05.000000Z"), nil
}

func windowZeroCSVSetDigest(m *CutoverManifest, billings map[int64]cutoverBillingFact, parents map[int64]cutoverPaymentParent) (int64, string, error) {
	header, err := windowZeroHeader(m)
	if err != nil {
		return 0, "", err
	}
	ids := make([]int64, 0)
	for id, b := range billings {
		if _, found := parents[id]; found || b.status != "completed" || b.totalAmount == 0 {
			continue
		}
		if id < m.IDBand.NonOwnerIDOffset || id >= m.IDBand.EndExclusive || id <= 0 {
			return 0, "", fmt.Errorf("window-zero billing id is outside the clinic band")
		}
		ids = append(ids, id)
		if int64(len(ids)) > maxCutoverPaymentRows {
			return 0, "", fmt.Errorf("window-zero subject set exceeds limit")
		}
	}
	slices.Sort(ids)
	h := sha256.New()
	_, _ = h.Write([]byte(header))
	for _, id := range ids {
		b := billings[id]
		stamp, err := windowZeroTimestamp(b.completionTimestamp)
		if err != nil {
			return 0, "", err
		}
		_, _ = fmt.Fprintf(h, "%d\t%d\t%s\n", id, b.totalAmount, stamp)
	}
	return int64(len(ids)), hex.EncodeToString(h.Sum(nil)), nil
}

func verifyWindowZeroCSVSet(m *CutoverManifest, p CutoverProvenanceContract, billings map[int64]cutoverBillingFact, parents map[int64]cutoverPaymentParent) error {
	if err := validateWindowZeroEvidence(*m, p); err != nil {
		return err
	}
	if m.WindowZeroSettlementEvidence == nil {
		return nil
	}
	count, digest, err := windowZeroCSVSetDigest(m, billings, parents)
	if err != nil {
		return err
	}
	if count != *m.WindowZeroSettlementEvidence.RowCount || digest != m.WindowZeroSettlementEvidence.SHA256 {
		return fmt.Errorf("window-zero evidence does not match the exact CSV subject set")
	}
	return nil
}
