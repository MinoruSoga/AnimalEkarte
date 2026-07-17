-- P3-2/P3-3: PlanetScale Postgres へのマイグレーション適用後のスキーマ互換検証クエリ集。
-- 2026-07-05 に animalekarte-stg/main に対して実行し、全項目確認済み(migration-cloudflare.md 参照)。
-- psql "$DATABASE_URL" -f infra/scripts/validate-schema.sql で再実行可能。

\echo '--- 適用済みマイグレーション一覧(backend/migrations/*.sql と件数一致すること) ---'
SELECT filename, executed_at FROM schema_migrations ORDER BY filename;

\echo '--- インストール済み拡張機能 ---'
\dx

\echo '--- ENUM型(サンプル5件) ---'
SELECT typname FROM pg_type WHERE typtype = 'e' ORDER BY 1 LIMIT 5;

\echo '--- text[] 等の配列カラム(サンプル5件) ---'
SELECT table_name, column_name, udt_name
FROM information_schema.columns
WHERE table_schema = 'public' AND data_type = 'ARRAY'
LIMIT 5;

\echo '--- jsonb カラム(サンプル5件) ---'
SELECT table_name, column_name
FROM information_schema.columns
WHERE table_schema = 'public' AND data_type = 'jsonb'
LIMIT 5;

\echo '--- トリガ一覧 ---'
SELECT trigger_name, event_object_table
FROM information_schema.triggers;

\echo '--- public スキーマのテーブル数 ---'
SELECT count(*) AS table_count
FROM information_schema.tables
WHERE table_schema = 'public';
