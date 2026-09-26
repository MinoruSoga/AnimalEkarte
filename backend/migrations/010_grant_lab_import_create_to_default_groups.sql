-- EMR-176: lab-import:create の既定権限付与を適用済み行へ反映 + seeds/002_master
-- checksum reconcile（EMR-213 / 009_reconcile_002_master_seed_checksum.sql と同型）。
--
-- 背景:
--   本変更は適用済み seed bundle 002_master の
--   accounts/permission_group_rules.csv を直接編集した（lab-import 行の
--   can_create f→t）。bundle は適用後 immutable で、manifest+CSV の checksum が
--   schema_migrations に記録される。reconcile しないまま適用済み DB で
--   cmd/migrate が走ると checksum ガードが fail-closed で exit 1 となり、
--   CMD `/app/migrate && exec /app/api` のためコンテナ全体が起動不能になる
--   （EMR-71 の CSV 直編集 → 2026-09-24 STG crash loop / EMR-213 で修復済みの
--   同型事故を繰り返さないための必須の対。SQL migration は seed チェックより
--   先に実行されるため、本ファイルで先に reconcile すれば後続の bundle チェックを
--   通過できる）。
--
-- 規約との関係:
--   migrations/CLAUDE.md は直下 .sql を DDL 専用・seed 正データは CSV(cmd/seed-export
--   経由)と定める。本ファイルはその恒久的経路を迂回するものではなく、「適用済み
--   bundle の記録を現行内容へ reconcile し、bundle 変更が意図した権限差分を
--   bundle 所有行へ反映する」一回限りの修復例外（009 と同じ位置づけ）。
--
-- 内容:
--   1) lab-import:create の既定付与を適用済み行へ反映する:
--      執行 (1,3,5,7) t,f,t,f → t,t,t,f、一般 (2,4,6,8) t,f,f,f → t,t,f,f。
--      can_view='t' を保持する行のみ can_create を立てる
--      （can_view='f' への個別調整や create-only 行という fail-closed 違反を
--      作らない）。閲覧専用 (9) は変更しない。
--      CREATE 権限は検査結果の受信・検査機器登録・取込起動を許可するが、
--      結果のカルテ紐付け/解除/切戻し（edit）は執行のみのまま。
--   2) schema_migrations の seeds/002_master 記録 checksum を現行 bundle の値
--      へ更新。既知の旧値 (e4af744e… = 009 reconcile 後の値) の行のみ対象とし、
--      未知のドリフト状態は上書きしない（その場合は checksum ガードが引き続き
--      fail する = 意図どおり）。
--
-- 対象外:
--   seed bundle が所有する既定グループ (1–8) 以外、すなわち CreateClinic や
--   権限グループ設定 UI で後から作られた「執行」「一般」名グループ (id > 9) は
--   本 migration では変更しない。それらへの適用は運用判断を要するため、
--   seeds/live_insert_lab_import_create_grant.sql（手動・承認済み runbook 経由）
--   または /settings/permission-groups の UI で個別に行う。
--
-- fresh DB では対象行・記録とも未存在のため両 UPDATE は no-op であり、後続の
-- seed フェーズが新 CSV をそのまま適用する。既修復 DB への再適用も冪等。

UPDATE permission_group_rules
SET can_create = 't'
WHERE resource = 'lab-import'
  AND group_id IN (1, 2, 3, 4, 5, 6, 7, 8)
  AND can_view = 't'
  AND can_create = 'f'
  AND deleted_at IS NULL;

-- 意図した既定値差分が残っていないことを検証（view 保持行で create=f が
-- 残存 = UPDATE が効いていない、または未知の状態）。
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM permission_group_rules
    WHERE resource = 'lab-import'
      AND group_id IN (1, 2, 3, 4, 5, 6, 7, 8)
      AND can_view = 't'
      AND can_create = 'f'
      AND deleted_at IS NULL
  ) THEN
    RAISE EXCEPTION 'EMR-176: lab-import default groups still deny create after reconcile';
  END IF;
END $$;

-- schema_migrations は cmd/migrate バイナリが ensureMigrationsTable で作成する
-- bookkeeping テーブルであり、直下 DDL 群には含まれない。CI のテスト DB
-- ブートストラップ(psql -f 直適用)では同テーブルが存在しないため、存在する
-- 環境(= cmd/migrate 経由で管理される実 DB)でのみ reconcile を実行する。
DO $$
BEGIN
  IF to_regclass('public.schema_migrations') IS NOT NULL THEN
    UPDATE schema_migrations
    SET checksum = 'b38a15124e0a75694871d3ad63c621425342b3ea7553c15c8795189e37a0e659'
    WHERE filename = 'seeds/002_master'
      AND checksum = 'e4af744e52b725e1db1708331a542446bdc8cc2cc1e1147a2c0eaa7904842bc7';
  END IF;
END $$;
