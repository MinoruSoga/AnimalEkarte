-- =============================================================================
-- Animal Ekarte - ISSUE-010: 監査ログにメタデータ・件数を永続化する
-- PostgreSQL 18
-- 依存: 001_init.sql
-- 内容: audit_logs に jsonb metadata カラムを追加する。
--       LSTEP 操作（健診対象抽出、一括タグ付与、no-show / dormant バッチなど）の
--       件数・抽出条件を構造化メタデータとして恒久保存できるようにする。
--
-- 本マイグレーションの安全性:
--   - ALTER TABLE ADD COLUMN ... jsonb NULL は既存行に影響しない（DEFAULT 不要）。
--   - 既存の audit_logs.{action, resource, resource_id, old_value, new_value} には変更を加えない。
--   - 既存の参照クエリ・インデックスはそのまま動作する。
-- =============================================================================

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS metadata jsonb NULL;

COMMENT ON COLUMN audit_logs.metadata IS 'LSTEP 監査メタデータ。件数・抽出条件など resource_id 単一 ID では表現できない多次元情報を JSON で保存する（ISSUE-010）';
