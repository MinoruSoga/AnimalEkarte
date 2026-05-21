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
-- STG 2026-05-21 反映分: STG dump (prodData/stg_5-21/ekarte) の完全再現を目的に、
-- 003_seed_demo.sql の clinic_id=1〜3 既存 seed と重複しない差分のみ追加。
--
-- ⚠️ セキュリティ警告: 本 seed はユーザー許可のもと、STG 環境の Lステップ
-- 連携シークレット (LINE channel_access_token / channel_secret / liff_id) を
-- 平文で含む。本 seed を流す DB は STG 専用であること、また STG の
-- LINE 公式アカウント設定が変更された場合は seed もローテートすること。
--
-- 追加対象:
--   1. clinics.id=4 (Hako bu neco 猫専門病院)
--   2. permission_groups.id=7,8 (clinic_id=4 用、CreateClinic が auto-gen する
--      '執行'/'一般' と同一構成)
--   3. accounts.id=17 (clinic_id=4 用ログインアカウント)
--   4. staffs.id=36 (青山 純子 → accounts.id=17 リンク)
--   5. staff_clinic_assignments id=38 (soft-deleted, clinic_1 旧割当),
--      id=39 (clinic_2 現割当), id=40 (clinic_4 現割当)
--   6. staff_permission_groups (staff_36 → group_id=2)
--   7. clinic_integrations id=1〜4 (clinic_1 用 Lステップ連携設定)
--
-- 対象外 (003 既存 seed と重複):
--   * companies / animal_species / 既存 clinics 1-3 / 既存 staffs 1-35 /
--     既存 owners / pets / medical_records / appointments / マスタ全般
-- -----------------------------------------------------------------------------

-- -----------------------------------------------------------------------------
-- 1. clinics (id=4)
-- -----------------------------------------------------------------------------
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

-- -----------------------------------------------------------------------------
-- 2. permission_groups for clinic_id=4 (id=7,8)
--    STG 実値: color='#6B7280' (CreateClinic デフォルト)。
--    003 の groups 1-6 は color='#6366F1'/'#10B981' で seed 済み。
-- -----------------------------------------------------------------------------
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

-- -----------------------------------------------------------------------------
-- 3. accounts (id=17) — clinic_id=4 用ログインアカウント
--    ⚠️ 実 email + 実 password_hash を含む (STG 環境専用)
-- -----------------------------------------------------------------------------
INSERT INTO accounts (id, email, password_hash, is_active, is_system_admin) VALUES
    (17, 'chunzishan72@gmail.com',
     '$2a$12$83ztnVW/5NSm1kDq3ZiqXOuu41J2MLSrQ40b.v2/e6PpxheZ.4kIK',
     true, false)
ON CONFLICT (id) DO UPDATE
    SET email           = EXCLUDED.email,
        password_hash   = EXCLUDED.password_hash,
        is_active       = EXCLUDED.is_active,
        is_system_admin = EXCLUDED.is_system_admin,
        updated_at      = now();

SELECT setval(pg_get_serial_sequence('accounts', 'id'), (SELECT MAX(id) FROM accounts));

-- -----------------------------------------------------------------------------
-- 4. staffs (id=36) — accounts.id=17 と紐付け
--    occupation_id=2 は clinic_1 の '看護師'。STG 実値どおり (clinic 跨ぎの
--    occupation 参照は STG の異動履歴の結果)。
-- -----------------------------------------------------------------------------
INSERT INTO staffs (
    id, account_id, name, is_active, license_number, occupation_id,
    sort_order, staff_type, reservation_visible
) VALUES
    (36, 17, '青山 純子', true, '', 2, 0, 'nurse', true)
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

-- -----------------------------------------------------------------------------
-- 5. staff_clinic_assignments — staff_36 の clinic 割当 3 件
--    id=38: clinic_1 旧割当 (soft-deleted at 2026-05-17 07:22:17)
--    id=39: clinic_2 現割当 (is_main=true)
--    id=40: clinic_4 現割当 (is_main=false)
--    deleted_at の固定時刻は STG dump 実値どおり。
-- -----------------------------------------------------------------------------
INSERT INTO staff_clinic_assignments (id, staff_id, clinic_id, is_main, deleted_at) VALUES
    (38, 36, 1, true,  '2026-05-17 07:22:17.205168+00'),
    (39, 36, 2, true,  NULL),
    (40, 36, 4, false, NULL)
ON CONFLICT (id) DO UPDATE
    SET staff_id   = EXCLUDED.staff_id,
        clinic_id  = EXCLUDED.clinic_id,
        is_main    = EXCLUDED.is_main,
        deleted_at = EXCLUDED.deleted_at,
        updated_at = now();

SELECT setval(
    pg_get_serial_sequence('staff_clinic_assignments', 'id'),
    (SELECT MAX(id) FROM staff_clinic_assignments)
);

-- -----------------------------------------------------------------------------
-- 6. staff_permission_groups — staff_36 → group_id=2 (clinic_1 '一般')
--    clinic_1 への旧割当が soft-delete された後も権限付与だけ残留している
--    STG 状態をそのまま反映。
-- -----------------------------------------------------------------------------
INSERT INTO staff_permission_groups (staff_id, group_id) VALUES
    (36, 2)
ON CONFLICT (staff_id, group_id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7. clinic_integrations — clinic_id=1 用 Lステップ (LINE) 連携設定
--    ⚠️ 実 LINE channel_access_token / channel_secret / liff_id を含む
--    (STG 環境専用)。本番環境では絶対に投入しない。
--    UNIQUE 制約: (clinic_id, service, key_name)
-- -----------------------------------------------------------------------------
INSERT INTO clinic_integrations (id, clinic_id, service, key_name, key_value) VALUES
    (1, 1, 'lstep', 'line_channel_access_token',
     'pwMi3yP6jhRa0xbmnR0IPEcE5l+OIp21a7ia3hmoiuFSCvqkR5Tmmfm6fLoSTB1Bt7uQjAe9NN7fZ+LBDtNKLGnrqBrjDmhTnws9PVxQKLyinomNzUAb61KADX7NJmFBfEsLQQ9VmlU+tMJcWh+zswdB04t89/1O/w1cDnyilFU='),
    (2, 1, 'lstep', 'line_channel_secret', '5344ef84eb7072b5894f7e087db28827'),
    (3, 1, 'lstep', 'liff_id',             '2009755581-w5NOA3EW'),
    (4, 1, 'lstep', 'line_account_name',   'テスト-八王子')
ON CONFLICT (clinic_id, service, key_name) DO UPDATE
    SET key_value  = EXCLUDED.key_value,
        updated_at = now();

SELECT setval(
    pg_get_serial_sequence('clinic_integrations', 'id'),
    (SELECT MAX(id) FROM clinic_integrations)
);
