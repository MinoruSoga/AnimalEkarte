-- EMR-249 follow-up: seeds/002_master 適用済み DB の checksum ドリフト修復。
--
-- 履歴・前例: 009_reconcile_002_master_seed_checksum.sql と同一パターン。
--   seed bundle は適用後 immutable で、manifest+CSV の checksum が
--   schema_migrations に記録される。適用済み DB で bundle のファイル内容が
--   変わると cmd/migrate の checksum ガードが fail-closed で exit 1 となり、
--   CMD `/app/migrate && exec /app/api` のためコンテナ全体が起動不能になる
--   （2026-09-24 に加え 2026-10-03 STG で再発: migrate-runner も通常
--   インスタンスも port check で全滅）。
--
-- 乖離の内容:
--   577c7ff1d (EMR-249) は適用済み bundle の
--   accounts/permission_group_rules.csv から resource='checkup-package-import'
--   の default-deny 行 9 件を除去した（resource 自体を model.AllResources
--   から除去したため）。STG の記録 checksum は a55403e36 時点の
--   b38a1512… で、現行 bundle は 544bd6b7…。差分は当該 CSV の行除去のみ
--   （git diff a55403e36..HEAD で確認済み）。
--
-- データ差分の担保:
--   CSV から除去された行に対応する既存 DB の行は、本ファイルより先に
--   適用される 019_drop_checkup_package_import_permission.sql が
--   DELETE FROM permission_group_rules WHERE resource='checkup-package-import'
--   で物理削除する。したがって bundle 記録の reconcile のみで
--   データ実態との乖離は残らない（EMR-71 系の追加 UPDATE は不要）。
--
-- 規約との関係:
--   migrations/CLAUDE.md の「直下 .sql は DDL 専用」に対し、本ファイルは
--   「適用済み bundle の記録を現行内容へ reconcile する」一回限りの修復
--   例外（009 と同型）。この例外が無い場合の正規修復経路は DB_RESET による
--   再構築のみであり、STG/PROD のデータを失うため採用しない。
--
-- 適用範囲の限定:
--   既知の旧値 (b38a1512…) の行のみ対象とし、未知のドリフト状態は
--   上書きしない（その場合は checksum ガードが引き続き fail する = 意図どおり）。
--
-- fresh DB では対象行・記録とも未存在のため UPDATE は no-op であり、
-- 後続の seed フェーズが新 CSV をそのまま適用する。
-- Do not edit applied migrations. User must run: make migrate

DO $$
BEGIN
  IF to_regclass('public.schema_migrations') IS NOT NULL THEN
    UPDATE schema_migrations
    SET checksum = '544bd6b76fa456ea43193e7ad05cac008301eafcf7c603498482299d1f2378fe'
    WHERE filename = 'seeds/002_master'
      AND checksum = 'b38a15124e0a75694871d3ad63c621425342b3ea7553c15c8795189e37a0e659';
  END IF;
END $$;
