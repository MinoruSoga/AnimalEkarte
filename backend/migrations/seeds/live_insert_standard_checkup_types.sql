-- One-shot standard checkup-type master insert for every clinic.
-- EMR-169 (スプシNo.33 健診パッケージ化と検査項目マスタ投入): 共有済みの
-- 標準健診パッケージ 8 区分を checkup_types（診療項目マスタの定期健診タブ）
-- の項目マスタとして全医院へ投入する。
--
-- 位置づけ:
--   seeds/002_master の CSV bundle は適用後 immutable で、内容変更の正規
--   経路は使い捨て DB への実適用からの cmd/seed-export 再生成のみ
--   （docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md §3 / migrations/CLAUDE.md）。
--   本スクリプトはその正規経路を恒久的に置き換えるものではなく、適用済み
--   DB（STG / local / 本番）へ同一マスタを投入する明示的な手動手順。
--   cmd/migrate の入力ではない（直下 *.sql の DDL フェーズは seed bundle
--   より先に走るため clinics が無い fresh DB では必ず失敗する）。
--   承認済み runbook から psql ON_ERROR_STOP=1 で適用すること
--   （live_insert_standard_reservation_types.sql と同型）。
--
-- 値の根拠:
--   - 7 件は backend/checkup-packages/*.json の shipped manifest
--     （clinical_approval_ref=EMR-169、TestShippedCheckupPackageManifests_Validate
--     で検証済み）の types[0] と同一値。manifest import API（TASK-374、
--     checkup-packages/README.md）が checkup_type_fields 投入を担う正式
--     経路であり、本スクリプトは checkup_types の項目マスタのみを対象とする
--     （fields の投入は対象外）。
--   - 歯科検診: manifest 未同梱（旧 003_seed_demo.sql J-12 で投入済みの
--     既存区分）のため属性はその暫定 seed 値に揃える
--     （description='歯周病チェック・歯石付着度の確認' / interval='1年' /
--     target_age='成犬'）。sort_order=45 は manifest 群の並び
--     （アドプリット=40 / 皮膚=50）の間で共有リスト掲載順を維持する値。
--   - price は列挙しない（NULL）: codebase 全体に価格定義がなく manifest
--     schema にも price field が無い。NULL は画面上「価格未設定」と解釈される。
--   - interval は SQL の型キーワードのため "interval" と quote する。
--   - is_active=true: manifest 8 件すべて is_active=true で統一。
--
-- 冪等性・fail-closed:
--   (clinic_id, name) の live 行（deleted_at IS NULL）が既に存在する医院
--   では INSERT をスキップする。idx_checkup_types_clinic_name は
--   deleted_at IS NULL 限定の一意のため既存行との衝突は起きない。
--   既存行（既投入の「歯科検診」等）は属性を一切上書きしない。
--   事後条件として全医院に 8 区分の live 行が揃うことを検証し、
--   未達なら例外で transaction 全体を rollback する。
--
-- 適用後に fresh DB へも同内容を含めたい場合は、別途 CSV bundle の
-- 正規再生成（cmd/seed-export）をユーザーが行う。本スクリプトの適用だけ
-- では seeds/002_master は変わらない。

BEGIN;
SELECT pg_advisory_xact_lock(333052);

CREATE TEMP TABLE desired_checkup_types (
  name text PRIMARY KEY,
  description text NOT NULL,
  "interval" text NOT NULL,
  target_age text NOT NULL,
  sort_order integer NOT NULL
) ON COMMIT DROP;
INSERT INTO desired_checkup_types (name, description, "interval", target_age, sort_order) VALUES
  ('年4健診', '春・夏・秋・冬の年4回実施する定期健診。血液検査・レントゲンは外注のため院内測定は尿検査のみ。尿検査の項目マスタは医師確認後に追加予定。', '年4回', '', 10),
  ('バレンタイン健診', '2月実施の季節健診。フィラリア予防注射キャンペーンと同時進行。血液検査・レントゲンは外注のため院内測定は尿検査のみ。尿検査の項目マスタは医師確認後に追加予定。', '年1回', '', 20),
  ('5月健診', '5月実施の季節健診（フィラリア予防開始時期）。血液検査・レントゲンは外注のため院内測定は尿検査のみ。尿検査の項目マスタは医師確認後に追加予定。', '年1回', '', 30),
  ('アドプリット検診', '歯科検診キャンペーンに併施する検査（別料金・セット販売ではない）。結果は0〜5のレベルでお渡しする。', '', '', 40),
  ('歯科検診', '歯周病チェック・歯石付着度の確認', '1年', '成犬', 45),
  ('皮膚検診', '獣医師の視診による皮膚状態チェック。異常部位の図示（イラスト記入）は現行機能外のため、必要時はカルテ添付機能を利用する。状態チェック項目の内訳は医師確認後に追加予定。', '', '', 50),
  ('耳検診', '耳内視鏡による耳の状態チェックと総評価。耳内写真はカルテ添付機能を利用する。状態チェック項目の内訳は医師確認後に追加予定。', '', '', 60),
  ('眼科検診', '傷の有無・眼圧・涙量の測定。傷の部位図示（イラスト記入）は現行機能外のため、必要時はカルテ添付機能を利用する。眼圧・涙量の基準値は医師確認後に設定予定（設定までは異常値自動判定は行われない）。', '', '', 70);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM clinics) THEN
    RAISE EXCEPTION 'no clinics found; apply seed bundle 002_master first';
  END IF;
END $$;

INSERT INTO checkup_types
  (clinic_id, name, is_active, description, "interval", target_age, sort_order)
SELECT c.id, d.name, true, d.description, d."interval", d.target_age, d.sort_order
FROM clinics c
CROSS JOIN desired_checkup_types d
WHERE NOT EXISTS (
  SELECT 1 FROM checkup_types e
  WHERE e.clinic_id = c.id AND e.name = d.name AND e.deleted_at IS NULL
);

DO $$
DECLARE expected bigint; actual bigint;
BEGIN
  SELECT count(*) INTO expected
  FROM clinics c CROSS JOIN desired_checkup_types d;
  SELECT count(*) INTO actual
  FROM clinics c CROSS JOIN desired_checkup_types d
  WHERE EXISTS (
    SELECT 1 FROM checkup_types e
    WHERE e.clinic_id = c.id AND e.name = d.name AND e.deleted_at IS NULL
  );
  IF actual <> expected THEN
    RAISE EXCEPTION 'standard checkup types postcondition mismatch: expected %, got %', expected, actual;
  END IF;
END $$;

COMMIT;
