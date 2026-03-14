-- 004_add_clinic_id_to_vaccinations.sql
-- vaccinations テーブルに clinic_id を追加
-- standalone vaccination（medical_record_id = NULL）のマルチテナント対応

ALTER TABLE vaccinations
    ADD COLUMN IF NOT EXISTS clinic_id bigint REFERENCES clinics(id) ON DELETE RESTRICT;

-- 既存レコード: medical_records から clinic_id をバックフィル
UPDATE vaccinations v
SET clinic_id = mr.clinic_id
FROM medical_records mr
WHERE v.medical_record_id = mr.id
  AND v.clinic_id IS NULL;

-- pet_id がある場合は pets から clinic_id をバックフィル（medical_record_id が NULL のケース）
UPDATE vaccinations v
SET clinic_id = p.clinic_id
FROM pets p
WHERE v.pet_id = p.id
  AND v.clinic_id IS NULL;

-- clinic_id を NOT NULL に変更（バックフィル後）
ALTER TABLE vaccinations
    ALTER COLUMN clinic_id SET NOT NULL;

-- インデックス追加
CREATE INDEX IF NOT EXISTS idx_vaccinations_clinic_id ON vaccinations(clinic_id);
