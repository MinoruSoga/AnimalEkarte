-- UAT-R2-EXCLUSIVE-LOCK: 治療・バイタル・処方・接種の子レコードに楽観的ロック用
-- version 列を追加する。medical_records.version / clinical_plans.version
-- （001_init.sql）と同型の INTEGER NOT NULL DEFAULT 1。
-- Update は WHERE version = expectedVersion の CAS + 成功時 version+1 を行う
-- （backend/internal/medicalrecord の各 repository Update）。
-- DEFAULT 1 により既存行は適用時に自動で 1 へバックフィルされる。
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE treatments
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE vital_records
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE prescriptions
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE vaccinations
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
