-- #95 vital_records / treatment_plans / care_logs に直接 clinic_id を追加する。
-- 既存データは親 medical_records / hospitalizations / daily_records から復元する。

ALTER TABLE vital_records
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE vital_records vr
SET clinic_id = mr.clinic_id
FROM medical_records mr
WHERE vr.medical_record_id = mr.id
  AND vr.clinic_id IS NULL;

UPDATE vital_records vr
SET clinic_id = dr.clinic_id
FROM daily_records dr
WHERE vr.daily_record_id = dr.id
  AND vr.clinic_id IS NULL;

ALTER TABLE vital_records
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE vital_records
    DROP CONSTRAINT IF EXISTS vital_records_clinic_id_fkey,
    ADD CONSTRAINT vital_records_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_vital_records_clinic_id
    ON vital_records(clinic_id)
    WHERE deleted_at IS NULL;

ALTER TABLE treatment_plans
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE treatment_plans tp
SET clinic_id = mr.clinic_id
FROM medical_records mr
WHERE tp.medical_record_id = mr.id
  AND tp.clinic_id IS NULL;

UPDATE treatment_plans tp
SET clinic_id = h.clinic_id
FROM hospitalizations h
WHERE tp.hospitalization_id = h.id
  AND tp.clinic_id IS NULL;

ALTER TABLE treatment_plans
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE treatment_plans
    DROP CONSTRAINT IF EXISTS treatment_plans_clinic_id_fkey,
    ADD CONSTRAINT treatment_plans_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_treatment_plans_clinic_id
    ON treatment_plans(clinic_id)
    WHERE deleted_at IS NULL;

ALTER TABLE care_logs
    ADD COLUMN IF NOT EXISTS clinic_id bigint;

UPDATE care_logs cl
SET clinic_id = dr.clinic_id
FROM daily_records dr
WHERE cl.daily_record_id = dr.id
  AND cl.clinic_id IS NULL;

ALTER TABLE care_logs
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE care_logs
    DROP CONSTRAINT IF EXISTS care_logs_clinic_id_fkey,
    ADD CONSTRAINT care_logs_clinic_id_fkey
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_care_logs_clinic_id
    ON care_logs(clinic_id);
