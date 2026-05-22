-- 005: add is_non_insurance to exam_types/medicines and visit_type to medical_records
-- These columns were added directly to 001_init.sql in commits c1d4e110 and 8b886d00
-- without incremental migrations. This file makes existing DBs consistent.

ALTER TABLE exam_types
    ADD COLUMN IF NOT EXISTS is_non_insurance boolean NOT NULL DEFAULT false;

ALTER TABLE medicines
    ADD COLUMN IF NOT EXISTS is_non_insurance boolean NOT NULL DEFAULT false;

ALTER TABLE medical_records
    ADD COLUMN IF NOT EXISTS visit_type visit_type NULL;
