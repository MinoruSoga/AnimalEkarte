-- 主訴区分マスタに description カラムを追加
ALTER TABLE chief_complaint_categories ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

-- 職種マスタに code / description カラムを追加
ALTER TABLE job_titles ADD COLUMN IF NOT EXISTS code        VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE job_titles ADD COLUMN IF NOT EXISTS description TEXT        NOT NULL DEFAULT '';
