-- 007: cash_register_closes に論理削除 (deleted_at) を追加する。
--
-- 背景 (issue #76): 月次レポート GetMonthlyReport が cash_register_closes を
-- `deleted_at IS NULL` で絞るが、テーブルにカラムが存在せず
-- "column deleted_at does not exist" (SQLSTATE 42703) となり
-- GET /api/v1/reports/monthly が 400 を返していた。
-- 論理削除を基本方針とするため、カラムを追加してモデル (gorm.DeletedAt) と整合させる。
--
-- あわせて UNIQUE(clinic_id, close_date, period) を「未削除行のみ」に効く部分 UNIQUE INDEX へ
-- 置換する。これがないと締めを論理削除した後に同一 (clinic, close_date, period) を再締めできない
-- (削除済み行が UNIQUE 枠を占有してしまう)。

ALTER TABLE cash_register_closes ADD COLUMN deleted_at timestamptz;

ALTER TABLE cash_register_closes DROP CONSTRAINT uq_cash_register_closes_date_period;

CREATE UNIQUE INDEX uq_cash_register_closes_date_period
    ON cash_register_closes (clinic_id, close_date, period)
    WHERE deleted_at IS NULL;
