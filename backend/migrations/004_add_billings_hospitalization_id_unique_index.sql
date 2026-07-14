-- 004_add_billings_hospitalization_id_unique_index.sql
-- billings: hospitalization_id がある場合は active 行で 1 対 1（退院会計の二重永続化を DB で防ぐ）
-- 001 の idx_billings_medical_record_id_unique と対称だが、soft-delete 後の再作成を許すため
-- deleted_at IS NULL を述語に含める（medical_record 側との意図的非対称）。
--
-- 挙動保存: インデックス追加のみ。hospitalization_id IS NULL（診療会計等）は UNIQUE 対象外で複数可。
-- soft-deleted 行も UNIQUE 対象外。
--
-- 適用前必須チェック（重複ありなら BLOCKED・修復は別 unit）:
--   SELECT hospitalization_id, COUNT(*)
--   FROM billings
--   WHERE hospitalization_id IS NOT NULL AND deleted_at IS NULL
--   GROUP BY hospitalization_id
--   HAVING COUNT(*) > 1;

CREATE UNIQUE INDEX idx_billings_hospitalization_id_unique
  ON billings(hospitalization_id)
  WHERE hospitalization_id IS NOT NULL AND deleted_at IS NULL;
