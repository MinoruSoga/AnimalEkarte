-- FEAT-381-2 Commit 2: recommendation_reason (Lstep phase 2)
-- 受診推奨理由: revisit / checkup / prevention / exam の4値（NULL = 未設定）
-- アプリ層で whitelist バリデーション実施（DB CHECK 制約は使用しない）

ALTER TABLE medical_records
    ADD COLUMN IF NOT EXISTS recommendation_reason varchar(100);
