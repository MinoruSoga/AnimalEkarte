-- =============================================================================
-- Animal Ekarte - password_hash カラム追加
-- PostgreSQL 18
-- 冪等性保証: IF NOT EXISTS
-- =============================================================================

ALTER TABLE user_accounts
    ADD COLUMN IF NOT EXISTS password_hash text NOT NULL DEFAULT '';
