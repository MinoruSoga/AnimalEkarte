-- vaccinations.medical_record_id を NULL 許可に変更
-- 予防接種単独登録（カルテなし）をサポートするため
ALTER TABLE vaccinations
  ALTER COLUMN medical_record_id DROP NOT NULL;
