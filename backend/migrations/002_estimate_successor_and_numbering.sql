-- 002_estimate_successor_and_numbering.sql
-- TASK-012 FINAL B: 確定見積の後継ドラフト（supersedes）と見積番号採番の前提カラム。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。001_init.sql は編集しない。

ALTER TABLE estimates
  ADD COLUMN IF NOT EXISTS supersedes_estimate_id BIGINT NULL
  REFERENCES estimates(id);

CREATE INDEX IF NOT EXISTS idx_estimates_clinic_supersedes
  ON estimates(clinic_id, supersedes_estimate_id)
  WHERE supersedes_estimate_id IS NOT NULL;

COMMENT ON COLUMN estimates.supersedes_estimate_id IS
  'Successor draft points to the locked (approved/rejected) estimate it corrects. Original row is never unlocked.';
