-- 004_add_columns.sql
-- incremental migration: add fields introduced by BE-MEDI-010 and BE-EXAM-001
-- Safe to run on existing DBs (IF NOT EXISTS guards)

-- BE-MEDI-010: treatments.admin_route
ALTER TABLE treatments
  ADD COLUMN IF NOT EXISTS admin_route varchar(50) NOT NULL DEFAULT '';

-- BE-EXAM-001: examination_items reference range + abnormal flag
ALTER TABLE examination_items
  ADD COLUMN IF NOT EXISTS ref_min     decimal(10,4),
  ADD COLUMN IF NOT EXISTS ref_max     decimal(10,4),
  ADD COLUMN IF NOT EXISTS is_abnormal boolean DEFAULT false;
