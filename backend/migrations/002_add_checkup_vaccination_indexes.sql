-- 002_add_checkup_vaccination_indexes.sql
-- G7-4: checkups/vaccinations の日付レンジ一覧・期限アラートクエリに対応する複合インデックスを追加する。
--
-- checkup_repository.go / vaccination_repository.go の一覧・アラートクエリは
-- clinic_id + date（または next_date）でのレンジ絞り込み + ソートを行うが、
-- 既存インデックスは idx_checkups_clinic_id / idx_vaccinations_clinic_id の単列のみで、
-- medical_records(idx_medical_records_clinic_date)・billings(idx_billings_clinic_date)・
-- appointments(idx_appointments_clinic_date) と異なり clinic_id+date 複合を欠いていた。
--
-- 挙動保存: インデックス追加のみ。既存の単列 idx_checkups_clinic_id / idx_vaccinations_clinic_id は
-- 複合インデックスの prefix に包含されるため DROP 候補だが、削除は EXPLAIN での利用確認後に
-- 別コミットで行う(本マイグレーションでは追加のみ)。

CREATE INDEX idx_checkups_clinic_date
  ON checkups(clinic_id, date)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_checkups_clinic_next_date
  ON checkups(clinic_id, next_date)
  WHERE deleted_at IS NULL AND next_date IS NOT NULL;

CREATE INDEX idx_vaccinations_clinic_date
  ON vaccinations(clinic_id, date)
  WHERE deleted_at IS NULL;
