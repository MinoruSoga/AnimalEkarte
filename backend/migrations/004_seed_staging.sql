-- =============================================================================
-- Animal Ekarte - Staging Additions (clinic_id=4,5 exclusive data)
-- PostgreSQL 18
-- 依存: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- 内容: clinic_id=4,5 新規追加クリニック用の最小限データセット
-- =============================================================================

-- Staging clinics: clinic_id=4,5 (master only)
-- デモ用permission_groups・rules は GORM CreateClinic が動的生成するため手動定義不要

-- NOTE: clinic, company, permission_groups, permission_group_rules, staff_clinic_assignments
-- などは 003_seed_demo.sql で clinic_id=1,2,3 のみ処理。clinic_id=4,5 の運用データ
-- が必要な場合は、以降のマイグレーション（005以降）で段階的に追加可能。
-- 現在の004は「最小限のデータセットで動作確認可能」な状態を目指す。

-- -----------------------------------------------------------------------------
-- STG 2026-05-21 反映分: clinic_id=4 'Hako bu neco 猫専門病院' を STG 環境で
-- 追加したため、fresh DB reset 時に再現できるよう seed に反映。
--
-- 対象範囲:
--   1. clinics.id=4 (Hako bu neco 猫専門病院)
--   2. permission_groups.id=7,8 (clinic_id=4 用、CreateClinic が自動生成する
--      '執行'/'一般' と同一構成)
--   3. staffs.id=36 (青山 純子) — account 連携は STG 環境で手動運用
--   4. staff_clinic_assignments id=39,40 (staff_36 を clinic_2/4 にアサイン)
--   5. staff_permission_groups (staff_36 → group_id=2)
--
-- 除外項目 (理由付き):
--   * accounts.id=17 — 実 Gmail アドレス + 実 password_hash。認証情報のため
--     seed 対象外。STG 環境では手動で追加運用する。
--   * staff_clinic_assignments id=38 — 同レコードは soft-deleted のため
--     アクティブクエリに影響しない。sequence のみ setval で整合。
--   * clinic_integrations (id=1〜4) — clinic_id=1 用 LINE channel_access_token
--     と channel_secret に実値が含まれていたため seed 対象外。STG 環境では
--     設定 API 経由で個別投入する運用 (該当 token のローテート推奨)。
--   * permission_group_rules for groups 7,8 — STG dump にも未投入のため
--     既存挙動と整合させて空のままとする (CreateClinic は rules を自動生成
--     しない既存仕様に追従)。
-- -----------------------------------------------------------------------------

-- 1. clinics (id=4)
INSERT INTO clinics (
    id, company_id, name, postal_code, address,
    is_active, standard_tax_rate, reduced_tax_rate
) VALUES
    (4, 1, 'Hako bu neco 猫専門病院', '400-0861', '山梨県甲府市城東1-13-17',
     true, 0.10, 0.08)
ON CONFLICT (id) DO UPDATE
    SET company_id         = EXCLUDED.company_id,
        name               = EXCLUDED.name,
        postal_code        = EXCLUDED.postal_code,
        address            = EXCLUDED.address,
        is_active          = EXCLUDED.is_active,
        standard_tax_rate  = EXCLUDED.standard_tax_rate,
        reduced_tax_rate   = EXCLUDED.reduced_tax_rate,
        updated_at         = now();

SELECT setval(pg_get_serial_sequence('clinics', 'id'), (SELECT MAX(id) FROM clinics));

-- 2. permission_groups for clinic_id=4 (id=7,8)
--    STG 実値: color='#6B7280' (CreateClinic デフォルト)。
--    003 の groups 1-6 は color='#6366F1'/'#10B981' で seed 済み。
INSERT INTO permission_groups (
    id, clinic_id, name, description, color, is_active, sort_order
) VALUES
    (7, 4, '執行', '執行権限',           '#6B7280', true, 1),
    (8, 4, '一般', '一般スタッフ権限',   '#6B7280', true, 2)
ON CONFLICT (id) DO UPDATE
    SET clinic_id   = EXCLUDED.clinic_id,
        name        = EXCLUDED.name,
        description = EXCLUDED.description,
        color       = EXCLUDED.color,
        is_active   = EXCLUDED.is_active,
        sort_order  = EXCLUDED.sort_order,
        updated_at  = now();

SELECT setval(
    pg_get_serial_sequence('permission_groups', 'id'),
    (SELECT MAX(id) FROM permission_groups)
);

-- 3. staffs (id=36) — account_id は意図的に NULL (STG 認証アカウントは seed 除外)
--    occupation_id=2 (clinic_1 看護師) は STG 実値どおり。clinic 跨ぎの構造は
--    STG での移動履歴の結果なのでそのまま反映する。
INSERT INTO staffs (
    id, account_id, name, is_active, license_number, occupation_id,
    sort_order, staff_type, reservation_visible
) VALUES
    (36, NULL, '青山 純子', true, '', 2, 0, 'nurse', true)
ON CONFLICT (id) DO UPDATE
    SET account_id          = EXCLUDED.account_id,
        name                = EXCLUDED.name,
        is_active           = EXCLUDED.is_active,
        license_number      = EXCLUDED.license_number,
        occupation_id       = EXCLUDED.occupation_id,
        sort_order          = EXCLUDED.sort_order,
        staff_type          = EXCLUDED.staff_type,
        reservation_visible = EXCLUDED.reservation_visible,
        updated_at          = now();

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- 4. staff_clinic_assignments — staff_36 を clinic_2 と clinic_4 にアサイン。
--    STG dump の id=38 (clinic_1 への割当) は soft-deleted のため省略する。
--    id を明示することで sequence/PK の整合を保つ。
INSERT INTO staff_clinic_assignments (id, staff_id, clinic_id, is_main) VALUES
    (39, 36, 2, true),
    (40, 36, 4, false)
ON CONFLICT (id) DO UPDATE
    SET staff_id  = EXCLUDED.staff_id,
        clinic_id = EXCLUDED.clinic_id,
        is_main   = EXCLUDED.is_main,
        updated_at = now();

SELECT setval(
    pg_get_serial_sequence('staff_clinic_assignments', 'id'),
    (SELECT MAX(id) FROM staff_clinic_assignments)
);

-- 5. staff_permission_groups — staff_36 → group_id=2 (clinic_1 '一般')。
--    STG dump にあるレコードをそのまま反映 (clinic_1 への割当が soft-delete
--    される前の権限付与が残留している状態)。
INSERT INTO staff_permission_groups (staff_id, group_id) VALUES
    (36, 2)
ON CONFLICT (staff_id, group_id) DO NOTHING;
