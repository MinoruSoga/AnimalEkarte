-- job_titles.code カラムを削除する
-- 006_add_fields_to_chief_complaints_job_titles.sql で追加されたが不要と判断
ALTER TABLE job_titles DROP COLUMN IF EXISTS code;
