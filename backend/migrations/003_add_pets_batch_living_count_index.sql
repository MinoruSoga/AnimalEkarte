-- 003_add_pets_batch_living_count_index.sql
-- PERF-FOLLOWUP-01: pet_repository.go CountLivingByOwnerIDs (owner_id IN (...) 用バッチ集計)
-- に対応する複合インデックスを追加する。
--
-- 対象クエリ:
--   WHERE clinic_id = ? AND owner_id IN (...) AND deceased_at IS NULL AND deleted_at IS NULL
--   GROUP BY owner_id
--
-- 既存の idx_pets_owner_id(owner_id) WHERE deleted_at IS NULL や
-- idx_pets_deceased(clinic_id, deceased_at) はいずれも owner_id IN (...) + clinic_id の
-- 等値絞り込みと deceased_at IS NULL の組み合わせを一度にカバーしない。
-- clinic_id・owner_id を先頭の等値/IN 条件、deceased_at を末尾の NULL 判定として並べる。
--
-- owners 側は FindByIDs が主キー id IN (...) で解決されるため対象外（追加不要）。

CREATE INDEX idx_pets_clinic_owner_living
  ON pets(clinic_id, owner_id, deceased_at)
  WHERE deleted_at IS NULL;
