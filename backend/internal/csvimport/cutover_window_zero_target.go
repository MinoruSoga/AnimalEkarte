package csvimport

import (
	"context"
	"fmt"
)

// The entire no-payment subject set is rechecked before enabling the narrow
// missing-graph exemption. Any split attachment, even outside the target band
// or clinic, invalidates the evidence. Existing payment guards remain active.
const verifyWindowZeroTargetQuery = `
WITH window_zero_subjects AS MATERIALIZED (
  SELECT billing.id, billing.total_amount, billing.completed_at, billing.deleted_at
  FROM billings billing
  WHERE billing.id >= $1 AND billing.id < $2
    AND billing.clinic_id = $3
    AND billing.status = 'completed' AND billing.total_amount <> 0
    AND NOT EXISTS (SELECT 1 FROM payments attached WHERE attached.billing_id = billing.id)
  ORDER BY billing.id LIMIT 1000001
)
SELECT count(*),
  encode(sha256(convert_to($4::text || COALESCE(string_agg(
    subject.id::text || E'\t' || subject.total_amount::text || E'\t' ||
    to_char(subject.completed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"') || E'\n',
    '' ORDER BY subject.id), ''), 'UTF8')), 'hex'),
  count(*) FILTER (WHERE subject.deleted_at IS NOT NULL OR subject.completed_at IS NULL
    OR NOT isfinite(subject.completed_at)
    OR extract(year FROM subject.completed_at AT TIME ZONE 'UTC') NOT BETWEEN 1 AND 9999
    OR EXISTS (SELECT 1 FROM payment_splits attached WHERE attached.billing_id = subject.id))
FROM window_zero_subjects subject`

func verifyWindowZeroTarget(ctx context.Context, q cutoverQuerier, m *CutoverManifest, seeds CutoverSeedIDs, p CutoverProvenanceContract) error {
	if err := validateWindowZeroEvidence(*m, p); err != nil {
		return err
	}
	if m.WindowZeroSettlementEvidence == nil {
		return nil
	}
	header, err := windowZeroHeader(m)
	if err != nil {
		return err
	}
	var count, invalid int64
	var digest string
	if err := q.QueryRow(ctx, verifyWindowZeroTargetQuery, m.IDBand.NonOwnerIDOffset,
		m.IDBand.EndExclusive, seeds.ClinicID, header).Scan(&count, &digest, &invalid); err != nil {
		return fmt.Errorf("verify window-zero target evidence: %w", err)
	}
	e := m.WindowZeroSettlementEvidence
	if invalid != 0 || count != *e.RowCount || digest != e.SHA256 {
		return fmt.Errorf("window-zero evidence does not match the exact target subject set")
	}
	return nil
}
