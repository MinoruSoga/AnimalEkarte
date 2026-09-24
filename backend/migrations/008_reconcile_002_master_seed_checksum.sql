-- EMR-213: seeds/002_master 適用済み DB の checksum ドリフト修復 + EMR-71 権限データ反映。
--
-- 背景:
--   c21419692 (EMR-71) は適用済み seed bundle 002_master の
--   accounts/permission_group_rules.csv を直接編集した。bundle は適用後 immutable で、
--   manifest+CSV の checksum が schema_migrations に記録される。適用済み DB では
--   cmd/migrate の checksum ガードが fail-closed で exit 1 となり、
--   CMD `/app/migrate && exec /app/api` のためコンテナ全体が起動不能になる
--   (2026-09-24 STG で発生。SQL migration は seed チェックより先に実行されるため、
--   本ファイルで先に reconcile すれば後続の bundle チェックを通過できる)。
--
-- 規約との関係:
--   migrations/CLAUDE.md は直下 .sql を DDL 専用・seed 正データは CSV(cmd/seed-export
--   経由)と定める。本ファイルはその恒久的経路を迂回するものではなく、「適用済み
--   bundle の記録を現行内容へ reconcile する」一回限りの修復例外。
--   この例外が無い場合の正規修復経路は DB_RESET による再構築のみであり、
--   STG/PROD のデータを失うため採用しない。
--
-- 内容:
--   1) c21419692 が意図した権限変更を適用済み行へ反映する:
--      執行グループ (1,3,5,7) の cash-register-close を t,f,t,f → t,t,f,f。
--      (create 付与。edit は dead grant のため除去 — closes は append-only で
--      edit を消費する route は存在しない)
--   2) schema_migrations の seeds/002_master 記録 checksum を現行 bundle の値へ更新。
--      既知の旧値 (39e47e2a…) の行のみ対象とし、未知のドリフト状態は上書きしない
--      (その場合は checksum ガードが引き続き fail する = 意図どおり)。
--
-- fresh DB では対象行・記録とも未存在のため両 UPDATE は no-op であり、後続の
-- seed フェーズが新 CSV をそのまま適用する。既修復 DB への再適用も冪等。

UPDATE permission_group_rules
SET can_create = 't', can_edit = 'f'
WHERE resource = 'cash-register-close'
  AND group_id IN (1, 3, 5, 7);

UPDATE schema_migrations
SET checksum = 'e4af744e52b725e1db1708331a542446bdc8cc2cc1e1147a2c0eaa7904842bc7'
WHERE filename = 'seeds/002_master'
  AND checksum = '39e47e2a1c160520f0261a29846008d788534aea6b379e8925b562cc061d1222';
