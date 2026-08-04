-- 007_lab_import_job_status_reverted.sql
-- TASK-032: lab import job 状態機械に terminal compensation 値 'reverted' を追加する。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。既適用 migration / seed は編集しない。
--
-- PostgreSQL 制約: 同一 transaction 内で ADD VALUE した新 enum 値は参照できない。
-- 本ファイルは ADD VALUE のみ。新値を使う DDL・DML は 008 以降に分離する。
-- ファイル内に BEGIN;/COMMIT; を書かない（cmd/migrate が各ファイルを自前 tx で包む）。

ALTER TYPE lab_import_job_status ADD VALUE IF NOT EXISTS 'reverted';
