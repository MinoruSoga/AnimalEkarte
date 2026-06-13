-- Migration 006: 冗長インデックス削除
--
-- 削除対象:
--   idx_vital_records_deleted_at  - idx_vital_records_clinic_id (clinic_id WHERE deleted_at IS NULL)
--                                   が同じ述語をより選択的な先頭列付きでカバー済み。
--   idx_billing_confirmations_status - status 単体インデックスは全クエリが
--                                       medical_record_id JOIN を先頭に使うため planner に選ばれない。
--
-- 保留:
--   idx_shift_entries_staff_date  - UNIQUE インデックスとして (staff_id, date) の
--                                   グローバル一意性を別途保証しているため要 PO 判断。

DROP INDEX IF EXISTS idx_vital_records_deleted_at;
DROP INDEX IF EXISTS idx_billing_confirmations_status;
