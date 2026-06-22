-- 006_add_owner_report_fields.sql
-- Issue #158: レガシー EMR（Figma 37:142）準拠の飼主レポート項目を補完する。
--
-- レガシー EMR が表示する以下の項目に対応する列を追加する:
--   - pets.blood_type        : ペット血液型（任意記録・未記録は NULL）
--   - pets.microchip_number  : マイクロチップ番号（任意記録・未記録は NULL）
--   - owners.dm_preference   : DM（ダイレクトメール）送付希望（NULL=未設定 / true=必要 / false=不要）
--
-- 新規導入のレガシー項目は nullable とし、既存レコードを「記録済み」と誤表示しない（unknown-safe）。
-- 既存行は NULL のまま残り、レポート上は「-」または非表示となる（捏造しない）。
-- 列追加のみで既存データへの影響は無い。

ALTER TABLE pets ADD COLUMN IF NOT EXISTS blood_type text NULL;
ALTER TABLE pets ADD COLUMN IF NOT EXISTS microchip_number text NULL;

ALTER TABLE owners ADD COLUMN IF NOT EXISTS dm_preference boolean NULL;

COMMENT ON COLUMN pets.blood_type IS 'ペット血液型。NULL=未記録。';
COMMENT ON COLUMN pets.microchip_number IS 'マイクロチップ番号。NULL=未記録。';
COMMENT ON COLUMN owners.dm_preference IS 'DM（ダイレクトメール）送付希望。NULL=未設定 / true=必要 / false=不要。';
