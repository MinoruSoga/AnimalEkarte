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
     true, 0.1, 0.08)
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

-- -----------------------------------------------------------------------------
-- 8. permission_group_rules — STG 突合: 003 が STG リセット後に追加した 6 リソース削除
--    dump は 24 resources × 6 groups = 144 行。003 がその後 6 リソース追加で 180 行になった。
--    差分を削除して dump に合わせる。
-- -----------------------------------------------------------------------------
DELETE FROM permission_group_rules
WHERE resource IN (
    'lstep-csv-import',
    'lstep-analytics',
    'cash-register-close',
    'accounting-reports',
    'closing-settings',
    'master-payment-method'
);

-- -----------------------------------------------------------------------------
-- 9. appointments — STG 突合: 過去予約が cron で no_show に更新された実態に合わせる
-- -----------------------------------------------------------------------------
UPDATE appointments
SET    status = 'no_show'
WHERE  id IN (5, 6, 7, 8, 9, 10, 11, 12, 13, 102, 106, 108);

-- -----------------------------------------------------------------------------
-- 10. appointments (id=109〜117) — STG リセット後に実際に作成された予約
-- -----------------------------------------------------------------------------
INSERT INTO appointments (
    id, clinic_id, start_time, end_time, owner_id, pet_id,
    visit_type, reservation_type_id, doctor_id, is_designated,
    status, notes, source, created_by, is_staff_delegated,
    customer_fields, created_at, updated_at, deleted_at, line_customer_id
) VALUES
    (109, 1, '2026-05-07 07:30:00+00', '2026-05-07 09:45:00+00', 1, 1, 'revisit', 1, 4, false, 'checked_in', '', 'manual', 4, false, '{}', '2026-05-07 05:44:20.90667+00',   '2026-05-13 05:26:48.697607+00', NULL, NULL),
    (110, 1, '2026-05-09 06:00:00+00', '2026-05-09 06:15:00+00', 1, 1, 'revisit', 2, NULL, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-09 04:37:04.699057+00', '2026-05-09 11:00:00.538006+00', NULL, NULL),
    (111, 1, '2026-05-09 01:30:00+00', '2026-05-09 01:45:00+00', 1, 2, 'revisit', 2, 4, false, 'in_consultation', '', 'manual', 4, false, '{}', '2026-05-09 08:50:41.468507+00', '2026-05-09 08:52:06.54822+00',  NULL, NULL),
    (112, 1, '2026-05-11 16:45:00+09', '2026-05-11 17:45:00+09', 1, 1, 'revisit', 2, 1, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-11 08:33:53.134392+00', '2026-05-12 01:00:00.734922+00', NULL, NULL),
    (113, 1, '2026-05-11 16:30:00+09', '2026-05-11 17:30:00+09', 1, 1, 'revisit', 2, NULL, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-11 08:34:46.519909+00', '2026-05-12 01:00:00.785954+00', NULL, NULL),
    (114, 1, '2026-05-13 01:00:00+00', '2026-05-13 01:15:00+00', 1, 1, 'revisit', 2, 1, false, 'checked_in', '', 'manual', 4, false, '{}', '2026-05-12 04:58:04.581474+00', '2026-05-13 05:29:31.75124+00',  NULL, NULL),
    (115, 1, '2026-05-12 05:30:00+00', '2026-05-12 05:45:00+00', 1, 1, 'revisit', 2, 1, false, 'checked_in', '', 'manual', 4, false, '{}', '2026-05-12 05:05:01.562877+00', '2026-05-12 05:10:46.021868+00', NULL, NULL),
    (116, 1, '2026-05-20 01:00:00+00', '2026-05-20 02:00:00+00', 1, 1, 'revisit', 3, NULL, false, 'accounting', '', 'manual', 4, false, '{}', '2026-05-13 05:24:00.074152+00', '2026-05-20 02:19:05.324388+00', NULL, NULL),
    (117, 1, '2026-05-20 01:00:00+00', '2026-05-20 02:00:00+00', 1, 2, 'revisit', 3, NULL, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-13 05:24:00.104721+00', '2026-05-20 06:00:00.741487+00', NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    start_time          = EXCLUDED.start_time,
    end_time            = EXCLUDED.end_time,
    owner_id            = EXCLUDED.owner_id,
    pet_id              = EXCLUDED.pet_id,
    visit_type          = EXCLUDED.visit_type,
    reservation_type_id = EXCLUDED.reservation_type_id,
    doctor_id           = EXCLUDED.doctor_id,
    is_designated       = EXCLUDED.is_designated,
    status              = EXCLUDED.status,
    notes               = EXCLUDED.notes,
    source              = EXCLUDED.source,
    created_by          = EXCLUDED.created_by,
    is_staff_delegated  = EXCLUDED.is_staff_delegated,
    customer_fields     = EXCLUDED.customer_fields,
    deleted_at          = EXCLUDED.deleted_at,
    line_customer_id    = EXCLUDED.line_customer_id,
    updated_at          = now();

SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT MAX(id) FROM appointments));

-- -----------------------------------------------------------------------------
-- 10b. appointments (id=118-119) — STG dump 2026-05-25 差分
-- 2026-05-21 に実際に作成されたクリニック1の予約
-- -----------------------------------------------------------------------------
INSERT INTO appointments (
    id, clinic_id, start_time, end_time, owner_id, pet_id,
    visit_type, reservation_type_id, doctor_id, is_designated,
    status, notes, source, created_by, is_staff_delegated,
    customer_fields, created_at, updated_at, deleted_at, line_customer_id
) VALUES
    (118, 1, '2026-05-21 08:30:00+00', '2026-05-21 09:30:00+00', 1, 1, 'revisit', 3, 1, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-21 08:29:00+00', '2026-05-21 08:30:00+00', NULL, NULL),
    (119, 1, '2026-05-21 08:00:00+00', '2026-05-21 08:15:00+00', 1, 1, 'first', 3, 1, false, 'checked_in', '', 'manual', 4, false, '{}', '2026-05-21 07:59:00+00', '2026-05-21 08:00:00+00', NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    start_time          = EXCLUDED.start_time,
    end_time            = EXCLUDED.end_time,
    owner_id            = EXCLUDED.owner_id,
    pet_id              = EXCLUDED.pet_id,
    visit_type          = EXCLUDED.visit_type,
    reservation_type_id = EXCLUDED.reservation_type_id,
    doctor_id           = EXCLUDED.doctor_id,
    is_designated       = EXCLUDED.is_designated,
    status              = EXCLUDED.status,
    notes               = EXCLUDED.notes,
    source              = EXCLUDED.source,
    created_by          = EXCLUDED.created_by,
    is_staff_delegated  = EXCLUDED.is_staff_delegated,
    customer_fields     = EXCLUDED.customer_fields,
    deleted_at          = EXCLUDED.deleted_at,
    line_customer_id    = EXCLUDED.line_customer_id,
    updated_at          = now();

SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT MAX(id) FROM appointments));

-- -----------------------------------------------------------------------------
-- 11. medical_records (id=37〜60) — STG リセット後に作成されたカルテ
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (
    id, clinic_id, record_no, date, owner_id, pet_id,
    doctor_id, appointment_id, status, version, entered_by,
    next_visit_recommended_date, created_at, updated_at, deleted_at
) VALUES
    (37, 1, 'MR-20260507-1-ez8XhN', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-07 04:10:02.867522+00', '2026-05-07 04:10:02.867522+00', NULL),
    (38, 1, 'MR-20260507-1-uUJgtF', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-07 05:44:37.876649+00', '2026-05-07 05:44:37.876649+00', NULL),
    (39, 1, 'MR-20260507-1-QHeBtx', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-07 05:44:51.894538+00', '2026-05-07 05:44:51.894538+00', NULL),
    (40, 1, 'MR-20260507-1-A0cYzP', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-07 05:45:37.458156+00', '2026-05-07 05:45:37.458156+00', NULL),
    (41, 1, 'MR-20260507-1-iznyqy', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-07 05:45:50.493979+00', '2026-05-07 05:45:50.493979+00', NULL),
    (42, 1, 'MR-20260509-1-zRC1Qb', '2026-05-09', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-09 08:48:12.285328+00', '2026-05-09 08:48:12.285328+00', NULL),
    (43, 1, 'MR-20260509-1-bx40GM', '2026-05-09', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-09 08:48:44.370087+00', '2026-05-09 08:48:44.370087+00', NULL),
    (44, 1, 'MR-20260509-1-mytfsU', '2026-05-09', 1, 2, 4,    111,  'draft', 1, NULL, NULL, '2026-05-09 08:51:27.788135+00', '2026-05-09 08:51:27.788135+00', NULL),
    (45, 1, 'MR-20260509-1-bGpxAZ', '2026-05-09', 1, 2, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-09 08:51:37.186341+00', '2026-05-09 08:51:37.186341+00', NULL),
    (46, 1, 'MR-20260509-1-YUsMoD', '2026-05-09', 1, 2, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-09 08:52:06.833723+00', '2026-05-09 08:52:06.833723+00', NULL),
    (47, 1, 'MR-20260511-1-huxJux', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-11 08:19:30.853282+00', '2026-05-11 08:19:30.853282+00', NULL),
    (48, 1, 'MR-20260511-1-ehfPeh', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-11 08:30:32.830175+00', '2026-05-11 08:30:32.830175+00', NULL),
    (49, 1, 'MR-20260511-1-FLYxqO', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-11 08:32:10.19843+00',  '2026-05-11 08:32:10.19843+00',  NULL),
    (50, 1, 'MR-20260512-1-f34Chb', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-12 05:08:01.542107+00', '2026-05-12 05:08:01.542107+00', NULL),
    (51, 1, 'MR-20260512-1-rjpx3a', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-12 05:10:53.599995+00', '2026-05-12 05:10:53.599995+00', NULL),
    (52, 1, 'MR-20260512-1-PesMFD', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-12 05:12:01.708601+00', '2026-05-12 05:12:01.708601+00', NULL),
    (53, 1, 'MR-20260512-1-2G93IH', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-12 05:12:14.263886+00', '2026-05-12 05:12:14.263886+00', NULL),
    (54, 1, 'MR-20260512-1-wYohCA', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-12 05:12:27.850393+00', '2026-05-12 05:12:27.850393+00', NULL),
    (55, 1, 'MR-20260513-1-DGmrwv', '2026-05-13', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-13 05:27:04.142696+00', '2026-05-13 05:27:04.142696+00', NULL),
    (56, 1, 'MR-20260513-1-UVuDVo', '2026-05-13', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-13 05:29:38.405567+00', '2026-05-13 05:29:38.405567+00', NULL),
    (57, 1, 'MR-20260520-1-R3kBq1', '2026-05-20', 1, 1, NULL, 116,  'draft', 1, NULL, NULL, '2026-05-20 02:16:03.121269+00', '2026-05-20 02:16:03.121269+00', NULL),
    (58, 1, 'MR-20260520-1-nuLiSC', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-20 02:16:14.089542+00', '2026-05-20 02:16:14.089542+00', NULL),
    (59, 1, 'MR-20260520-1-uIYcc1', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-20 02:16:34.475044+00', '2026-05-20 02:16:34.475044+00', NULL),
    (60, 1, 'MR-20260520-1-W3bBXH', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4,    NULL, '2026-05-20 02:16:52.992284+00', '2026-05-20 02:16:52.992284+00', NULL)
ON CONFLICT (id) DO UPDATE SET
    clinic_id                   = EXCLUDED.clinic_id,
    record_no                   = EXCLUDED.record_no,
    date                        = EXCLUDED.date,
    owner_id                    = EXCLUDED.owner_id,
    pet_id                      = EXCLUDED.pet_id,
    doctor_id                   = EXCLUDED.doctor_id,
    appointment_id              = EXCLUDED.appointment_id,
    status                      = EXCLUDED.status,
    version                     = EXCLUDED.version,
    entered_by                  = EXCLUDED.entered_by,
    next_visit_recommended_date = EXCLUDED.next_visit_recommended_date,
    deleted_at                  = EXCLUDED.deleted_at,
    updated_at                  = now();

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- -----------------------------------------------------------------------------
-- 12. clinical_plans (id=37〜60) — medical_records 37-60 対応の空プラン
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (
    id, medical_record_id, physical_exam,
    diagnosis_type_id, diagnosis_name_id, diagnosis_2_type_id, diagnosis_2_name_id,
    diagnosis_details, treatment_policy, created_at, updated_at, deleted_at
) VALUES
    (37, 37, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 04:10:02.89686+00',  '2026-05-07 04:10:02.89686+00',  NULL),
    (38, 38, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:44:37.891701+00', '2026-05-07 05:44:37.891701+00', NULL),
    (39, 39, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:44:51.908098+00', '2026-05-07 05:44:51.908098+00', NULL),
    (40, 40, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:45:37.465514+00', '2026-05-07 05:45:37.465514+00', NULL),
    (41, 41, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:45:50.501602+00', '2026-05-07 05:45:50.501602+00', NULL),
    (42, 42, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:48:12.303789+00', '2026-05-09 08:48:12.303789+00', NULL),
    (43, 43, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:48:44.383347+00', '2026-05-09 08:48:44.383347+00', NULL),
    (44, 44, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:51:27.802325+00', '2026-05-09 08:51:27.802325+00', NULL),
    (45, 45, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:51:37.195867+00', '2026-05-09 08:51:37.195867+00', NULL),
    (46, 46, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:52:06.842019+00', '2026-05-09 08:52:06.842019+00', NULL),
    (47, 47, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:19:30.872507+00', '2026-05-11 08:19:30.872507+00', NULL),
    (48, 48, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:30:32.843875+00', '2026-05-11 08:30:32.843875+00', NULL),
    (49, 49, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:32:10.206396+00', '2026-05-11 08:32:10.206396+00', NULL),
    (50, 50, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:08:01.563109+00', '2026-05-12 05:08:01.563109+00', NULL),
    (51, 51, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:10:53.608981+00', '2026-05-12 05:10:53.608981+00', NULL),
    (52, 52, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:01.764254+00', '2026-05-12 05:12:01.764254+00', NULL),
    (53, 53, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:14.276389+00', '2026-05-12 05:12:14.276389+00', NULL),
    (54, 54, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:27.858983+00', '2026-05-12 05:12:27.858983+00', NULL),
    (55, 55, '', NULL, NULL, NULL, NULL, '', '', '2026-05-13 05:27:04.157177+00', '2026-05-13 05:27:04.157177+00', NULL),
    (56, 56, '', NULL, NULL, NULL, NULL, '', '', '2026-05-13 05:29:38.41457+00',  '2026-05-13 05:29:38.41457+00',  NULL),
    (57, 57, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:03.143892+00', '2026-05-20 02:16:03.143892+00', NULL),
    (58, 58, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:14.103071+00', '2026-05-20 02:16:14.103071+00', NULL),
    (59, 59, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:34.48364+00',  '2026-05-20 02:16:34.48364+00',  NULL),
    (60, 60, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:53.008248+00', '2026-05-20 02:16:53.008248+00', NULL)
ON CONFLICT (id) DO UPDATE SET
    medical_record_id   = EXCLUDED.medical_record_id,
    physical_exam       = EXCLUDED.physical_exam,
    diagnosis_type_id   = EXCLUDED.diagnosis_type_id,
    diagnosis_name_id   = EXCLUDED.diagnosis_name_id,
    diagnosis_2_type_id = EXCLUDED.diagnosis_2_type_id,
    diagnosis_2_name_id = EXCLUDED.diagnosis_2_name_id,
    diagnosis_details   = EXCLUDED.diagnosis_details,
    treatment_policy    = EXCLUDED.treatment_policy,
    deleted_at          = EXCLUDED.deleted_at,
    updated_at          = now();

SELECT setval(pg_get_serial_sequence('clinical_plans', 'id'), (SELECT MAX(id) FROM clinical_plans));

-- -----------------------------------------------------------------------------
-- 13. inquiries (id=41〜64) — medical_records 37-60 対応の空問診
--     ⚠️ inquiries id=21〜24 は dump に存在しない (gap): seed に含めない
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (
    id, medical_record_id, chief_complaint_type_id,
    chief_complaint, history, current_medications, allergy_info,
    last_meal, last_defecation, last_urination,
    appetite, water_intake, owner_observations, notes,
    staff_id, created_at, updated_at
) VALUES
    (41, 37, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 04:10:02.879897+00', '2026-05-07 04:10:02.885494+00'),
    (42, 38, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:44:37.883924+00', '2026-05-07 05:44:37.886174+00'),
    (43, 39, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:44:51.901744+00', '2026-05-07 05:44:51.904033+00'),
    (44, 40, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:45:37.461418+00', '2026-05-07 05:45:43.378725+00'),
    (45, 41, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:45:50.497277+00', '2026-05-07 05:45:50.498997+00'),
    (46, 42, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:48:12.294211+00', '2026-05-09 08:48:12.29652+00'),
    (47, 43, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:48:44.376885+00', '2026-05-09 08:48:44.379088+00'),
    (48, 44, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:51:27.795388+00', '2026-05-09 08:51:27.798253+00'),
    (49, 45, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:51:37.1897+00',   '2026-05-09 08:51:37.192864+00'),
    (50, 46, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:52:06.837309+00', '2026-05-09 08:52:06.839046+00'),
    (51, 47,    1, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:19:30.863395+00', '2026-05-11 08:22:59.398691+00'),
    (52, 48,    1, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:30:32.836651+00', '2026-05-11 08:31:51.424634+00'),
    (53, 49, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:32:10.20185+00',  '2026-05-11 08:32:15.234448+00'),
    (54, 50, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:08:01.550843+00', '2026-05-12 05:08:01.554051+00'),
    (55, 51,    2, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:10:53.603986+00', '2026-05-12 05:11:45.172348+00'),
    (56, 52, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:01.715127+00', '2026-05-12 05:12:01.751826+00'),
    (57, 53, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:14.269626+00', '2026-05-12 05:12:14.272251+00'),
    (58, 54, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:27.854086+00', '2026-05-12 05:12:27.855968+00'),
    (59, 55, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-13 05:27:04.148505+00', '2026-05-13 05:27:35.8547+00'),
    (60, 56, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-13 05:29:38.409629+00', '2026-05-13 05:29:38.41183+00'),
    (61, 57, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:03.133988+00', '2026-05-20 02:16:03.13918+00'),
    (62, 58, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:14.096188+00', '2026-05-20 02:16:14.098507+00'),
    (63, 59, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:34.478667+00', '2026-05-20 02:16:34.480523+00'),
    (64, 60, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:53.000132+00', '2026-05-20 02:16:53.002662+00')
ON CONFLICT (id) DO UPDATE SET
    medical_record_id       = EXCLUDED.medical_record_id,
    chief_complaint_type_id = EXCLUDED.chief_complaint_type_id,
    chief_complaint         = EXCLUDED.chief_complaint,
    history                 = EXCLUDED.history,
    current_medications     = EXCLUDED.current_medications,
    allergy_info            = EXCLUDED.allergy_info,
    last_meal               = EXCLUDED.last_meal,
    last_defecation         = EXCLUDED.last_defecation,
    last_urination          = EXCLUDED.last_urination,
    appetite                = EXCLUDED.appetite,
    water_intake            = EXCLUDED.water_intake,
    owner_observations      = EXCLUDED.owner_observations,
    notes                   = EXCLUDED.notes,
    staff_id                = EXCLUDED.staff_id,
    updated_at              = now();

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));


-- =============================================================================
-- STG 実データ（prodData/ekarte-stg-6-5.sql 取得、2026-06-05）
-- 内容: STG 実データ（設定 + トランザクション）
-- ⚠️ マスタ系データは 002/003 を優先。このセクションにはトランザクションデータ + STG 固有設定のみ
-- =============================================================================

-- -----------------------------------------------------------------------------
-- clinic_settings
-- -----------------------------------------------------------------------------
INSERT INTO clinic_settings ("clinic_id", "closing_am_pm_boundary", "closing_weekday_end", "closing_sunday_end", "closed_weekdays", "cpm_version", "dormant_prevention_180_days", "dormant_prevention_210_days", "dormant_prevention_240_days", "dormant_prevention_365_days", "cpm_v2_coming_threshold", "cpm_v2_good_threshold", "cpm_v2_family_threshold", "cpm_v2_noah_threshold", "cpm_v1_dormant_days", "cpm_v1_noah_days", "cpm_v1_noah_annual_visits", "cpm_v1_noah_ltv", "cpm_v1_core_days", "cpm_v1_core_annual_visits", "cpm_v1_core_ltv", "cpm_v1_spot_min_amount", "cpm_v1_spot_inactive_days", "cpm_v1_growing_max_days", "cpm_v1_growing_min_visits", "cpm_v1_growing_max_visits", "cpm_v1_ltv_break_low", "health_prevention_lookback_days", "vaccine_deadline_days", "created_at", "updated_at") VALUES
(1, '14:00:00', '18:30:00', '17:30:00', '{0}', 'v1', 180, 210, 240, 365, 2, 4, 8, 13, 240, 365, 3, 80000, 180, 2, 50000, 30000, 90, 90, 2, 3, 20000, 365, 60, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, '13:30:00', '19:00:00', '18:00:00', '{4}', 'v1', 180, 210, 240, 365, 2, 4, 8, 13, 240, 365, 3, 80000, 180, 2, 50000, 30000, 90, 90, 2, 3, 20000, 365, 60, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, '14:00:00', '18:00:00', '17:00:00', '{0,6}', 'v1', 180, 210, 240, 365, 2, 4, 8, 13, 240, 365, 3, 80000, 180, 2, 50000, 30000, 90, 90, 2, 3, 20000, 365, 60, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- closing_special_periods
-- -----------------------------------------------------------------------------
INSERT INTO closing_special_periods ("id", "clinic_id", "start_date", "end_date", "am_pm_boundary", "pm_end", "note", "created_at", "updated_at") VALUES
(1, 1, '2026-08-13', '2026-08-16', '12:00:00', '18:00:00', '夏季休暇（お盆休み）', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, '2026-12-30', '2027-01-03', '12:00:00', '18:00:00', '年末年始休暇', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- lstep_trigger_priorities
-- -----------------------------------------------------------------------------
INSERT INTO lstep_trigger_priorities ("id", "clinic_id", "trigger_type", "priority", "created_at", "updated_at") VALUES
(1, 1, 'AFTER_SURGERY_FOLLOW', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 'HOSPITALIZATION_LOG', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 'RESERVATION_REMINDER', 3, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 1, 'VACCINE_ANNOUNCEMENT', 4, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- reservation_type_unavailable_times
-- -----------------------------------------------------------------------------
INSERT INTO reservation_type_unavailable_times ("id", "clinic_id", "reservation_type_id", "unavailable_type", "day_of_week", "specific_date", "start_time", "end_time", "created_at", "updated_at") VALUES
(1, 1, 1, 'weekly', 1, NULL, '13:00', '14:00', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 1, 'weekly', 2, NULL, '13:00', '14:00', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 1, 'weekly', 3, NULL, '13:00', '14:00', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 1, 1, 'weekly', 4, NULL, '13:00', '14:00', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 1, 1, 'weekly', 5, NULL, '13:00', '14:00', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- owners
-- -----------------------------------------------------------------------------
INSERT INTO owners ("id", "clinic_id", "name", "name_kana", "birth_date", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks", "is_dangerous", "discount_rate", "membership_type", "line_user_id", "lstep_opt_out", "lstep_opt_out_at", "lstep_opt_out_reason", "line_followed_at", "line_blocked_at", "line_id_confirmed_by", "line_id_confirmed_at", "delivery_excluded", "delivery_excluded_reason", "is_transferred", "transfer_at", "delivery_caution", "delivery_caution_reason", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, '林 文明', 'はやし ふみあき', '1980-05-15', 'サンプル株式会社', '150-0001', '東京都渋谷区神宮前1-2-3', '', '', '', '', '090-1111-2222', '03-1234-5678', 'hayashi@example.com', '定期検診を希望', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, '田中 花子', 'たなか はなこ', '1985-03-20', '', '160-0022', '東京都新宿区新宿1-1-1', '', '', '', '', '080-3333-4444', '', 'tanaka@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, '鈴木 一郎', 'すずき いちろう', '1978-11-03', '', '170-0001', '東京都豊島区西巣鴨1-3-5', '', '', '', '', '070-5555-6666', '', 'suzuki@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, '田中 美咲', 'たなか みさき', '1990-07-22', '', '153-0044', '東京都目黒区大橋2-4-6', '', '', '', '', '090-9999-8888', '', 'misaki.tanaka@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, '佐藤 花子', 'さとう はなこ', '1975-02-14', '', '140-0001', '東京都品川区北品川3-5-7', '', '', '', '', '080-2222-3333', '', 'hanako.sato@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, '伊藤 次郎', 'いとう じろう', '1983-09-30', '', '166-0013', '東京都杉並区堀ノ内1-7-9', '', '', '', '', '090-1234-5678', '', 'jiro.ito@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, '小林 さくら', 'こばやし さくら', '1992-04-05', '', '176-0012', '東京都練馬区豊玉北4-2-8', '', '', '', '', '080-9876-5432', '', 'sakura.kobayashi@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 1, '中村 勇気', 'なかむら ゆうき', '1987-12-18', '', '174-0041', '東京都板橋区舟渡2-6-10', '', '', '', '', '090-1122-3344', '', 'yuuki.nakamura@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 1, '加藤 恵', 'かとう めぐみ', '1995-06-25', '', '134-0083', '東京都江戸川区中葛西5-3-2', '', '', '', '', '080-5566-7788', '', 'megumi.kato@example.com', '', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, '山田 太郎', 'やまだ たろう', '1970-01-10', '', '144-0051', '東京都大田区西蒲田6-8-4', '', '', '', '', '090-2233-4455', '', 'taro.yamada@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, '高橋 由美', 'たかはし ゆみ', '1988-08-15', '', '110-0005', '東京都台東区上野5-1-3', '', '', '', '', '080-6677-8899', '', 'yumi.takahashi@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 1, '松本 隆', 'まつもと たかし', '1965-03-28', '', '125-0061', '東京都葛飾区亀有3-9-7', '', '', '', '', '090-3344-5566', '', 'takashi.matsumoto@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 1, '吉田 誠', 'よしだ まこと', '1982-11-05', '', '123-0845', '東京都足立区西新井7-4-6', '', '', '', '', '080-7788-9900', '', 'makoto.yoshida@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 1, '井上 京子', 'いのうえ きょうこ', '1973-05-19', '', '189-0023', '東京都東村山市美住町1-5-2', '', '', '', '', '090-4455-6677', '', 'kyoko.inoue@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, '木村 拓也', 'きむら たくや', '1991-07-14', '', '179-0081', '東京都練馬区北町3-6-9', '', '', '', '', '080-8899-0011', '', 'takuya.kimura@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 1, '佐々木 亮', 'ささき りょう', '1986-02-23', '', '207-0013', '東京都東大和市清水2-4-8', '', '', '', '', '090-5566-7788', '', 'ryo.sasaki@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 1, '山本 健太', 'やまもと けんた', '1998-09-12', '', '206-0802', '東京都稲城市東長沼2-8-3', '', '', '', '', '090-1234-9876', '', 'kenta.yamamoto@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 1, '青木 麻衣', 'あおき まい', '1993-03-10', '', '150-0002', '東京都渋谷区渋谷2-1-1', '', '', '', '', '090-1111-1111', '', 'mai.aoki@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 1, '橋本 俊介', 'はしもと しゅんすけ', '1980-07-25', '', '130-0001', '東京都墨田区吾妻橋1-3-5', '', '', '', '', '080-2222-2222', '', 'shunsuke.h@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 1, '福田 裕子', 'ふくだ ゆうこ', '1977-11-14', '', '145-0062', '東京都大田区北千束2-5-8', '', '', '', '', '090-3333-3333', '', 'yuko.fukuda@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 1, '石川 大輔', 'いしかわ だいすけ', '1989-04-02', '', '167-0041', '東京都杉並区善福寺3-2-6', '', '', '', '', '080-4444-4444', '', 'daisuke.ishikawa@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(22, 1, '村田 奈々', 'むらた なな', '1996-09-19', '', '182-0021', '東京都調布市調布ヶ丘1-4-7', '', '', '', '', '090-5555-5555', '', 'nana.murata@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 2, '大野 健司', 'おおの けんじ', '1979-06-10', '', '136-0071', '東京都江東区亀戸3-5-8', '', '', '', '', '090-6601-2233', '', 'kenji.ono@example.com', '定期通院', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(24, 2, '松田 香織', 'まつだ かおり', '1988-02-14', '', '135-0044', '東京都江東区越中島2-1-4', '', '', '', '', '080-7702-3344', '', 'kaori.matsuda@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(25, 2, '渡辺 直樹', 'わたなべ なおき', '1972-10-28', '渡辺商事', '132-0025', '東京都江戸川区松江3-7-2', '', '', '', '', '090-8803-4455', '', 'naoki.watanabe@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(26, 2, '中島 奈緒', 'なかじま なお', '1994-04-17', '', '133-0065', '東京都江戸川区南篠崎町1-6-3', '', '', '', '', '080-9904-5566', '', 'nao.nakajima@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(27, 2, '斎藤 浩二', 'さいとう こうじ', '1967-12-05', '', '131-0031', '東京都墨田区墨田1-4-9', '', '', '', '', '090-1105-6677', '', 'koji.saito@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(28, 2, '坂本 真由美', 'さかもと まゆみ', '1991-08-22', '', '130-0022', '東京都墨田区横川4-2-7', '', '', '', '', '080-2206-7788', '', 'mayumi.sakamoto@example.com', '猫アレルギーに注意', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(29, 2, '岡田 俊雄', 'おかだ としお', '1983-03-30', 'オカダ工業', '132-0034', '東京都江戸川区小松川1-8-5', '', '', '', '', '090-3307-8899', '', 'toshio.okada@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(30, 2, '藤田 彩', 'ふじた あや', '1997-11-11', '', '136-0076', '東京都江東区南砂3-9-1', '', '', '', '', '080-4408-9900', '', 'aya.fujita@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(31, 3, '村上 俊平', 'むらかみ しゅんぺい', '1976-04-12', '', '400-0031', '山梨県甲府市丸の内2-3-4', '', '', '', '', '090-5501-1122', '', 'shunpei.murakami@example.com', '', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(32, 3, '長谷川 恵子', 'はせがわ けいこ', '1989-09-07', '', '400-0032', '山梨県甲府市中央3-5-6', '', '', '', '', '080-6602-2233', '', 'keiko.hasegawa@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(33, 3, '野口 正樹', 'のぐち まさき', '1971-01-25', '野口設計', '400-0035', '山梨県甲府市丸の内3-7-8', '', '', '', '', '090-7703-3344', '', 'masaki.noguchi@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(34, 3, '石田 沙織', 'いしだ さおり', '1993-07-18', '', '400-0801', '山梨県甲府市横根町5-2-1', '', '', '', '', '080-8804-4455', '', 'saori.ishida@example.com', '猫2匹飼い', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(35, 3, '前田 修', 'まえだ おさむ', '1968-11-02', '', '400-0031', '山梨県甲府市丸の内1-9-3', '', '', '', '', '090-9905-5566', '', 'osamu.maeda@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(36, 3, '菊池 里奈', 'きくち りな', '1995-06-14', '', '400-0032', '山梨県甲府市中央5-4-7', '', '', '', '', '080-1106-6677', '', 'rina.kikuchi@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(37, 3, '清水 和彦', 'しみず かずひこ', '1980-03-28', '清水工務店', '400-0034', '山梨県甲府市北口1-6-2', '', '', '', '', '090-2207-7788', '', 'kazuhiko.shimizu@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(38, 3, '岩崎 美穂', 'いわさき みほ', '1998-12-05', '', '400-0803', '山梨県甲府市横根町2-8-9', '', '', '', '', '080-3308-8899', '', 'miho.iwasaki@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(39, 1, '渡辺 健', 'わたなべ けん', '1975-06-12', '', '101-0001', '東京都千代田区千代田1-1', '', '', '', '', '090-6666-6666', '', 'ken.w@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(40, 1, '斎藤 由紀', 'さいとう ゆき', '1982-08-24', '', '102-0002', '東京都千代田区麹町2-2', '', '', '', '', '080-7777-7777', '', 'yuki.s@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(41, 1, '岡田 准一', 'おかだ じゅんいち', '1980-11-18', '', '103-0003', '東京都中央区日本橋3-3', '', '', '', '', '070-8888-8888', '', 'junichi.o@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(42, 1, '松嶋 菜々子', 'まつしま ななこ', '1973-10-13', '', '104-0004', '東京都中央区築地4-4', '', '', '', '', '090-9999-9999', '', 'nanako.m@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(43, 1, '二宮 和也', 'にのみや かずなり', '1983-06-17', '', '105-0005', '東京都港区新橋5-5', '', '', '', '', '080-1111-2222', '', 'kazunari.n@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(44, 1, '石原 さとみ', 'いしはら さとみ', '1986-12-24', '', '106-0006', '東京都港区六本木6-6', '', '', '', '', '070-3333-4444', '', 'satomi.i@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(45, 1, '菅田 将暉', 'すだ まさき', '1993-02-21', '', '107-0007', '東京都港区赤坂7-7', '', '', '', '', '090-5555-6666', '', 'masaki.s@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(46, 1, '小松 菜奈', 'こまつ なな', '1996-02-16', '', '108-0008', '東京都港区三田8-8', '', '', '', '', '080-7777-8888', '', 'nana.k@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(47, 1, '綾瀬 はるか', 'あやせ はるか', '1985-03-24', '', '109-0009', '東京都港区虎ノ門9-9', '', '', '', '', '070-9999-0000', '', 'haruka.a@example.com', '', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(48, 1, '長澤 まさみ', 'ながさわ まさみ', '1987-06-03', '', '110-0010', '東京都台東区台東10-10', '', '', '', '', '090-1234-1234', '', 'masami.n@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(49, 1, '阿部 寛', 'あべ ひろし', '1964-06-22', '', '111-0011', '東京都台東区浅草11-11', '', '', '', '', '080-2345-2345', '', 'hiroshi.a@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(50, 1, '仲間 由紀恵', 'なかま ゆきえ', '1979-10-30', '', '112-0012', '東京都文京区本駒込12-12', '', '', '', '', '070-3456-3456', '', 'yukie.n@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(51, 1, '堺 雅人', 'さかい まさと', '1973-10-14', '', '113-0013', '東京都文京区弥生13-13', '', '', '', '', '090-4567-4567', '', 'masato.s@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(52, 1, '新垣 結衣', 'あらがき ゆい', '1988-06-11', '', '114-0014', '東京都北区王子14-14', '', '', '', '', '080-5678-5678', '', 'yui.a@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(53, 1, '星野 源', 'ほしの げん', '1981-01-28', '', '115-0015', '東京都北区赤羽15-15', '', '', '', '', '070-6789-6789', '', 'gen.h@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(54, 1, '吉沢 亮', 'よしざわ りょう', '1994-02-01', '', '116-0016', '東京都荒川区荒川16-16', '', '', '', '', '090-7890-7890', '', 'ryo.y@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(55, 1, '広瀬 すず', 'ひろせ すず', '1998-06-19', '', '117-0017', '東京都荒川区東日暮里17-17', '', '', '', '', '080-8901-8901', '', 'suzu.h@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(56, 1, '橋本 環奈', 'はしもと かんな', '1999-02-03', '', '118-0018', '東京都練馬区北町18-18', '', '', '', '', '070-9012-9012', '', 'kanna.h@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(57, 1, 'ムロ ツヨシ', 'むろ つよし', '1976-01-23', '', '119-0019', '東京都豊島区南池袋19-19', '', '', '', '', '090-0123-0123', '', 'muro.t@example.com', '', 'f', 0.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(58, 1, '戸田 恵梨香', 'とだ えりか', '1988-08-17', '', '120-0020', '東京都足立区千住20-20', '', '', '', '', '080-1234-5678', '', 'erika.t@example.com', '', 'f', 5.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(59, 1, '松坂 桃李', 'まつざか とおり', '1988-10-17', '', '121-0021', '東京都足立区保木間21-21', '', '', '', '', '070-2345-6789', '', 'tori.m@example.com', '', 'f', 0.00, 'non_member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(60, 1, '北川 景子', 'きたがわ けいこ', '1986-08-22', '', '122-0022', '東京都港区北青山22-22', '', '', '', '', '090-3456-7890', '', 'keiko.k@example.com', '', 'f', 10.00, 'member', NULL, 'f', NULL, NULL, NULL, NULL, NULL, NULL, 'f', NULL, 'f', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- pets
-- -----------------------------------------------------------------------------
INSERT INTO pets ("id", "clinic_id", "owner_id", "pet_number", "name", "name_kana", "animal_species_id", "gender", "status", "birth_date", "breed", "color", "weight", "neutered_date", "acquisition_type", "danger_level", "food", "environment", "phone", "last_visit", "insurance_id", "remarks", "deceased_at", "deceased_reason", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 1, '1-1', 'Iris(イリス)', 'いりす', 1, 'male', 'alive', '2015-04-14', 'ゴールデンレトリーバー', '茶色', 26.50, NULL, NULL, 'low', '', '', '', '2015-08-28', 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 1, '1-2', 'Max(マックス)', 'まっくす', 1, 'male', 'alive', '2018-06-20', 'ラブラドール', 'ゴールデン', 15.20, NULL, NULL, 'low', '', '', '', '2024-11-15', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 2, '2-1', 'ミケ', 'みけ', 2, 'female', 'alive', '2020-03-10', '三毛猫', '三毛', 4.20, NULL, NULL, 'low', '', '', '', '2024-11-18', 2, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 3, '3-1', 'タロウ', 'たろう', 1, 'male', 'alive', '2019-05-15', '柴犬', 'レッド', 8.30, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 3, '3-2', 'ジロウ', 'じろう', 1, 'male', 'alive', '2021-08-10', '柴犬', 'ブラック', 7.10, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, 4, '4-1', 'チョコ', 'ちょこ', 1, 'female', 'alive', '2017-11-20', 'トイプードル', 'チョコ', 3.80, NULL, NULL, 'low', '', '', '', NULL, 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, 5, '5-1', 'レオ', 'れお', 2, 'male', 'alive', '2016-07-04', 'スコティッシュフォールド', 'グレー', 5.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 1, 6, '6-1', 'ハチ', 'はち', 1, 'male', 'alive', '2018-03-25', '秋田犬', 'ホワイト', 22.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 1, 7, '7-1', 'モモ', 'もも', 2, 'female', 'alive', '2022-01-15', 'マンチカン', 'キャリコ', 3.20, NULL, NULL, 'low', '', '', '', NULL, 2, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, 8, '8-1', 'ロッキー', 'ろっきー', 1, 'male', 'alive', '2014-09-08', 'ボーダーコリー', 'ブラックホワイト', 18.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, 9, '9-1', 'ルナ', 'るな', 2, 'female', 'alive', '2021-02-28', 'ペルシャ', 'シルバー', 4.80, NULL, NULL, 'low', '', '', '', NULL, 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 1, 10, '10-1', 'ケン', 'けん', 1, 'male', 'alive', '2013-06-18', 'ジャーマンシェパード', 'ブラックタン', 32.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 1, 11, '11-1', 'ソラ', 'そら', 2, 'male', 'alive', '2023-04-01', 'アメリカンショートヘア', 'タビー', 3.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 1, 12, '12-1', 'ゴン', 'ごん', 1, 'male', 'alive', '2016-12-05', '紀州犬', 'ホワイト', 19.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, 13, '13-1', 'シロ', 'しろ', 1, 'male', 'alive', '2020-08-10', 'ミックス犬', 'ホワイト', 6.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 1, 14, '14-1', 'トラ', 'とら', 2, 'male', 'alive', '2019-10-22', 'トラ猫', 'トラ', 5.10, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 1, 15, '15-1', 'ベロ', 'べろ', 1, 'male', 'alive', '2018-05-03', 'ビーグル', 'トライカラー', 13.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 1, 16, '16-1', 'チビ', 'ちび', 2, 'female', 'alive', '2022-06-20', 'ミックス猫', 'サビ', 3.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 1, 17, '17-1', 'ポチ', 'ぽち', 1, 'male', 'alive', '2017-02-14', 'ダックスフンド', 'チョコ', 7.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 1, 18, '18-1', 'モカ', 'もか', 2, 'female', 'alive', '2022-05-10', 'ミックス猫', 'ホワイト', 4.10, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 1, 18, '18-2', 'クルミ', 'くるみ', 1, 'male', 'alive', '2020-08-20', 'ミックス犬', 'ベージュ', 8.30, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(22, 1, 19, '19-1', 'ハル', 'はる', 1, 'male', 'alive', '2019-03-15', 'ミックス犬', 'ブラック', 12.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 1, 19, '19-2', 'ユキ', 'ゆき', 2, 'female', 'alive', '2021-12-01', 'ミックス猫', 'ホワイト', 3.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(24, 1, 20, '20-1', 'ピーチ', 'ぴーち', 2, 'female', 'alive', '2023-01-07', 'ミックス猫', 'オレンジ', 3.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(25, 1, 21, '21-1', 'コタ', 'こた', 1, 'male', 'alive', '2018-09-23', 'ミックス犬', 'ブラウン', 22.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(26, 1, 21, '21-2', 'アン', 'あん', 2, 'female', 'alive', '2020-04-11', 'ミックス猫', 'キャリコ', 4.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(27, 1, 22, '22-1', 'ゴマ', 'ごま', 2, 'male', 'alive', '2022-11-30', 'ミックス猫', 'グレー', 5.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(28, 1, 22, '22-2', 'マル', 'まる', 1, 'female', 'alive', '2021-06-18', 'ミックス犬', 'ゴールデン', 9.70, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(29, 2, 23, '23-1', 'クロ', 'くろ', 1, 'male', 'alive', '2019-03-20', 'ラブラドール', 'ブラック', 28.00, NULL, NULL, 'low', '', '', '', '2025-11-10', 6, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(30, 2, 23, '23-2', 'シナモン', 'しなもん', 2, 'female', 'alive', '2021-07-15', 'アビシニアン', 'レッド', 4.10, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(31, 2, 24, '24-1', 'ポポ', 'ぽぽ', 2, 'female', 'alive', '2020-05-10', 'ロシアンブルー', 'グレー', 3.80, NULL, NULL, 'low', '', '', '', '2025-10-20', 7, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(32, 2, 25, '25-1', 'ダン', 'だん', 1, 'male', 'alive', '2017-09-05', 'ウェルシュコーギー', 'セーブルホワイト', 13.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(33, 2, 26, '26-1', 'キナ', 'きな', 2, 'male', 'alive', '2022-02-28', 'ミックス猫', 'オレンジ', 4.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(34, 2, 27, '27-1', 'バロン', 'ばろん', 1, 'male', 'alive', '2015-11-18', 'シェパード', 'ブラックタン', 30.50, NULL, NULL, 'low', '', '', '', '2025-12-01', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(35, 2, 28, '28-1', 'ユズ', 'ゆず', 2, 'female', 'alive', '2023-03-03', 'ミックス猫', 'キャリコ', 3.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(36, 2, 28, '28-2', 'レン', 'れん', 1, 'male', 'alive', '2021-06-14', 'ビーグル', 'トライカラー', 12.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(37, 2, 29, '29-1', 'ナナ', 'なな', 2, 'female', 'alive', '2018-08-07', 'メインクーン', 'ブラウンタビー', 6.20, NULL, NULL, 'low', '', '', '', '2026-01-15', 6, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(38, 2, 30, '30-1', 'コウ', 'こう', 1, 'male', 'alive', '2020-12-25', 'トイプードル', 'アプリコット', 3.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(39, 3, 31, '31-1', 'フク', 'ふく', 1, 'male', 'alive', '2018-05-05', 'シベリアンハスキー', 'グレーホワイト', 24.00, NULL, NULL, 'low', '', '', '', '2025-10-05', 9, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(40, 3, 32, '32-1', 'アズキ', 'あずき', 2, 'female', 'alive', '2021-11-20', 'ノルウェージャンフォレストキャット', 'ブラウンタビー', 4.80, NULL, NULL, 'low', '', '', '', '2025-12-18', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(41, 3, 33, '33-1', 'カイ', 'かい', 1, 'male', 'alive', '2016-07-10', 'ゴールデンレトリーバー', 'ゴールデン', 31.20, NULL, NULL, 'low', '', '', '', '2026-01-08', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(42, 3, 34, '34-1', 'キキ', 'きき', 2, 'female', 'alive', '2022-04-25', 'ミックス猫', 'トラ', 3.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(43, 3, 34, '34-2', 'ニコ', 'にこ', 2, 'female', 'alive', '2020-09-12', 'ミックス猫', 'サビ', 4.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(44, 3, 35, '35-1', 'ジャック', 'じゃっく', 1, 'male', 'alive', '2014-02-28', 'ジャックラッセルテリア', 'ホワイトタン', 6.80, NULL, NULL, 'low', '', '', '', '2025-09-20', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(45, 3, 36, '36-1', 'ミル', 'みる', 2, 'female', 'alive', '2023-01-15', 'ミックス猫', 'ホワイト', 3.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(46, 3, 37, '37-1', 'ガイア', 'がいあ', 1, 'male', 'alive', '2017-08-30', 'アラスカンマラミュート', 'グレー', 36.50, NULL, NULL, 'low', '', '', '', '2025-11-25', NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(47, 3, 38, '38-1', 'ハナ', 'はな', 2, 'female', 'alive', '2021-03-08', 'アメリカンショートヘア', 'シルバータビー', 4.00, NULL, NULL, 'low', '', '', '', '2026-02-10', 9, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(48, 3, 38, '38-2', 'ソウ', 'そう', 1, 'male', 'alive', '2019-10-01', 'ミックス犬', 'ベージュ', 9.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(49, 1, 39, '39-1', 'コタロウ', 'こたろう', 1, 'male', 'alive', '2021-05-12', '柴犬', '赤', 9.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(50, 1, 40, '40-1', 'モモコ', 'ももこ', 2, 'female', 'alive', '2022-08-24', 'ミックス猫', '白黒', 4.00, NULL, NULL, 'low', '', '', '', NULL, 2, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(51, 1, 41, '41-1', 'リュウ', 'りゅう', 1, 'male', 'alive', '2020-11-18', 'フレンチブルドッグ', 'パイド', 12.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(52, 1, 42, '42-1', 'サクラ', 'さくら', 2, 'female', 'alive', '2023-10-13', 'アメリカンショートヘア', 'シルバータビー', 3.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(53, 1, 43, '43-1', 'カズ', 'かず', 1, 'male', 'alive', '2019-06-17', 'チワワ', 'ブラックタン', 2.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(54, 1, 44, '44-1', 'サト', 'さと', 2, 'female', 'alive', '2021-12-24', 'ベンガル', 'スポッテッド', 4.20, NULL, NULL, 'low', '', '', '', NULL, 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(55, 1, 45, '45-1', 'マサ', 'まさ', 1, 'male', 'alive', '2022-02-21', 'パグ', 'フォーン', 7.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(56, 1, 46, '46-1', 'ナナ', 'なな', 2, 'female', 'alive', '2023-02-16', 'ラグドール', 'ブルーポイント', 5.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(57, 1, 47, '47-1', 'ハル', 'はる', 2, 'female', 'alive', '2021-03-24', 'ミックス猫', '三毛', 4.50, NULL, NULL, 'low', '', '', '', NULL, 2, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(58, 1, 48, '48-1', 'マサミ', 'まさみ', 1, 'female', 'alive', '2022-06-03', 'トイプードル', 'アプリコット', 3.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(59, 1, 49, '49-1', 'ヒロシ', 'ひろし', 1, 'male', 'alive', '2020-06-22', '秋田犬', '白', 28.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(60, 1, 50, '50-1', 'ユキエ', 'ゆきえ', 2, 'female', 'alive', '2021-10-30', 'シャム', 'ポイント', 3.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(61, 1, 51, '51-1', 'マサト', 'まさと', 1, 'male', 'alive', '2022-10-14', 'ポメラニアン', 'オレンジ', 4.50, NULL, NULL, 'low', '', '', '', NULL, 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(62, 1, 52, '52-1', 'ユイ', 'ゆい', 2, 'female', 'alive', '2023-06-11', 'スコティッシュフォールド', 'ブルー', 4.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(63, 1, 53, '53-1', 'ゲン', 'げん', 1, 'male', 'alive', '2021-01-28', 'ミックス犬', '黒白', 10.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(64, 1, 54, '54-1', 'リョウ', 'りょう', 1, 'male', 'alive', '2022-02-01', 'コーギー', 'レッドホワイト', 11.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(65, 1, 55, '55-1', 'スズ', 'すず', 2, 'female', 'alive', '2023-06-19', 'ミックス猫', 'サバトラ', 3.50, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(66, 1, 56, '56-1', 'カンナ', 'かんな', 2, 'female', 'alive', '2023-02-03', 'マンチカン', 'キャリコ', 3.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(67, 1, 57, '57-1', 'ムロ', 'むろ', 1, 'male', 'alive', '2021-01-23', 'ボストンテリア', 'ボストンカラー', 8.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(68, 1, 58, '58-1', 'エリカ', 'えりか', 2, 'female', 'alive', '2022-08-17', 'アビシニアン', 'ルディ', 3.60, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(69, 1, 59, '59-1', 'トオリ', 'とおり', 1, 'male', 'alive', '2022-10-17', 'ドーベルマン', 'ブラックタン', 30.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(70, 1, 60, '60-1', 'ケイコ', 'けいこ', 2, 'female', 'alive', '2021-08-22', 'ノルウェージャンフォレスト', 'ブラウンタビー', 5.50, NULL, NULL, 'low', '', '', '', NULL, 1, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(71, 1, 39, '39-2', 'レオ', 'れお', 1, 'male', 'alive', '2022-01-10', 'ミニチュアピンシャー', 'レッド', 4.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(72, 1, 40, '40-2', 'ヒメ', 'ひめ', 2, 'female', 'alive', '2023-03-05', 'ペルシャ', 'ホワイト', 4.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(73, 1, 41, '41-2', 'コテツ', 'こてつ', 1, 'male', 'alive', '2021-07-20', 'パピヨン', '白黒', 4.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(74, 1, 42, '42-2', 'ルナ', 'るな', 2, 'female', 'alive', '2022-12-15', 'ソマリ', 'レッド', 3.80, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(75, 1, 43, '43-2', 'チョコ', 'ちょこ', 1, 'female', 'alive', '2023-05-30', 'トイプードル', 'ブラウン', 3.00, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(80, 1, 5, '5-2', 'ピーター', 'ぴーたー', 4, 'male', 'alive', '2024-05-10', 'ネザーランドドワーフ', '', 1.20, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(81, 1, 10, '10-2', 'ハム助', 'はむすけ', 5, 'female', 'alive', '2025-01-15', 'ジャンガリアン', '', 0.04, NULL, NULL, 'low', '', '', '', NULL, NULL, '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- pet_chronic_conditions
-- -----------------------------------------------------------------------------
INSERT INTO pet_chronic_conditions ("id", "clinic_id", "pet_id", "condition_code", "condition_name", "diagnosed_at", "notes", "is_active", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 1, 'ATOPY', 'アトピー性皮膚炎', '2023-04-15', '通年性の痒み。ステロイド・シャンプーで維持。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 11, 'CKD', '慢性腎臓病', '2025-11-10', 'ステージ2。療法食と皮下補液継続。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 13, 'OTITIS', '外耳炎（再発性）', '2026-01-20', '左耳が特に悪化しやすい。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 5, 'DIABETES', '糖尿病', '2025-06-01', 'インスリン投与中。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 4, 'LUXATION', '膝蓋骨脱臼', '2024-08-20', 'グレード2。体重管理注意。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 2, 29, 'HEART', '僧帽弁閉鎖不全症', '2025-01-10', '投薬継続中。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 2, 35, 'ASTHMA', '猫喘息', '2025-12-05', '季節性に注意。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 3, 39, 'ALLERGY', '食物アレルギー', '2024-11-20', '鶏肉除去食。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 3, 47, 'EPILEPSY', 'てんかん', '2025-03-15', '抗てんかん薬継続。', 't', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- line_customers
-- -----------------------------------------------------------------------------
INSERT INTO line_customers ("id", "clinic_id", "line_user_id", "display_name", "real_name", "additional_fields", "owner_id", "created_at", "updated_at") VALUES
(1, 1, 'U_test_hachioji_001', 'テスト 太郎', '執行 太郎', '{"phone": "090-1234-5678", "pet_name": "ポチ", "pet_type": "柴犬", "owner_name": "執行 太郎"}', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 'U_test_hachioji_002', 'テスト 花子', '一般 花子', '{"phone": "080-9876-5432", "pet_name": "ミケ", "pet_type": "三毛猫", "owner_name": "一般 花子"}', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 2, 'U_test_joto_001', 'テスト 次郎', '城東テスト', '{"phone": "070-1111-2222", "pet_name": "チョコ", "pet_type": "トイプードル", "owner_name": "城東テスト"}', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 1, 'U1234567890abcdef1234567890abcdef', 'HAYASHI', '林 文明', '{}', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 1, 'U9876543210fedcba9876543210fedcba', 'HANA', '田中 花子', '{}', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 1, 'Uabcdef1234567890abcdef1234567890', 'SUZUKI', '鈴木 一郎', '{}', 3, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 1, 'Ufedcba9876543210fedcba9876543210', 'MISAKI', '田中 美咲', '{}', 4, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 1, 'U55555555555555555555555555555555', 'SATO', '佐藤 花子', '{}', 5, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- lstep_tag_cache
-- -----------------------------------------------------------------------------
INSERT INTO lstep_tag_cache ("id", "clinic_id", "owner_id", "tag_name", "category", "synced_at", "reason") VALUES
(1, 1, 1, '優良顧客', 'manual', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 1, '大型犬', 'auto', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 2, 'ワクチン案内中', 'auto', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 4, '優良顧客', 'manual', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- lstep_delivery_trigger_log
-- -----------------------------------------------------------------------------
INSERT INTO lstep_delivery_trigger_log ("id", "owner_id", "clinic_id", "trigger_type", "scheduled_at", "status", "fired_at", "excluded_reason", "suppressed_by_priority", "suppression_reason", "created_at", "updated_at") VALUES
(1, 1, 1, 'RESERVATION_REMINDER', '2026-06-01 04:33:17.574774+00', 'scheduled', NULL, NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 2, 1, 'VACCINE_ANNOUNCEMENT', '2026-05-29 04:33:17.574774+00', 'fired', '2026-05-29 04:33:17.574774+00', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 3, 1, 'BIRTHDAY_MESSAGE', '2026-05-26 04:33:17.574774+00', 'fired', '2026-05-26 04:33:17.574774+00', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 4, 1, 'HOSPITALIZATION_LOG', '2026-05-31 03:33:17.574774+00', 'fired', '2026-05-31 03:33:17.574774+00', NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 5, 1, 'AFTER_SURGERY_FOLLOW', '2026-06-03 04:33:17.574774+00', 'scheduled', NULL, NULL, 'f', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- lstep_csv_imports
-- -----------------------------------------------------------------------------
INSERT INTO lstep_csv_imports ("id", "clinic_id", "csv_type", "file_name", "uploaded_by_user_id", "row_count", "success_count", "error_count", "status", "error_log", "imported_at", "created_at") VALUES
('0a2a1621-65ed-449b-9d6a-4e2cd474241b', 1, 'friend_attribute', 'owners_export_20260520.csv', 1, 100, 100, 0, 'completed', NULL, '2026-05-28 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
('31481c9f-2db9-4ddc-b047-aed4529fb155', 1, 'friend_attribute', 'invalid_data_test.csv', 1, 10, 0, 0, 'failed', NULL, '2026-05-30 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
('cc7e1133-0324-49e4-97d0-5aa926195855', 1, 'friend_attribute', 'new_members_20260521.csv', 1, 50, 48, 0, 'completed', NULL, '2026-05-29 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- lstep_friend_attribute_snapshots
-- -----------------------------------------------------------------------------
INSERT INTO lstep_friend_attribute_snapshots ("id", "clinic_id", "line_user_id", "display_name", "registered_at", "tags", "scenarios", "traffic_source", "block_status", "last_message_at", "snapshot_taken_at", "csv_import_id", "created_at", "updated_at") VALUES
(1, 1, 'U1234567890abcdef1234567890abcdef', NULL, NULL, '["puppy", "last_visit_20251220"]', NULL, NULL, NULL, NULL, '2025-11-30 04:33:17.574774+00', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 'U1234567890abcdef1234567890abcdef', NULL, NULL, '["adult", "last_visit_20260522"]', NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 'U9876543210fedcba9876543210fedcba', NULL, NULL, '["senior", "care_needed"]', NULL, NULL, NULL, NULL, '2026-05-30 04:33:17.574774+00', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- line_send_logs
-- -----------------------------------------------------------------------------
INSERT INTO line_send_logs ("id", "clinic_id", "owner_id", "sent_by_user_id", "message_type", "content_summary", "line_message_id", "status", "error_message", "sent_at") VALUES
(1, 1, 1, 1, 'text', '明日の予約リマインド：Iris(イリス)様 9:00〜', NULL, 'success', NULL, '2026-05-31 04:33:17.574774+00'),
(2, 1, 2, 1, 'text', '狂犬病ワクチンのご案内（ミケちゃん）', NULL, 'success', NULL, '2026-05-31 04:33:17.574774+00'),
(3, 1, 3, 2, 'image', '【画像】院内の改装工事のお知らせ', NULL, 'success', NULL, '2026-05-31 04:33:17.574774+00'),
(4, 1, 4, 1, 'text', '診察完了のお知らせ', NULL, 'success', NULL, '2026-05-31 04:33:17.574774+00'),
(5, 1, 5, 3, 'text', 'お薬の準備ができました。ご来院をお待ちしております。', NULL, 'success', NULL, '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- staff_reservation_exclusions
-- -----------------------------------------------------------------------------
INSERT INTO staff_reservation_exclusions ("id", "staff_id", "reservation_type_id") VALUES
(1, 1, 4),
(2, 1, 13),
(3, 1, 9),
(4, 1, 10),
(5, 1, 11),
(6, 1, 12),
(7, 2, 4),
(8, 2, 13),
(9, 2, 9),
(10, 2, 10),
(11, 2, 11),
(12, 2, 12),
(13, 3, 4),
(14, 3, 13),
(15, 3, 9),
(16, 3, 10),
(17, 3, 11),
(18, 3, 12),
(19, 4, 4),
(20, 4, 13),
(21, 4, 9),
(22, 4, 10),
(23, 4, 11),
(24, 4, 12),
(25, 4, 33),
(26, 4, 34),
(27, 4, 35),
(28, 4, 52),
(29, 16, 33),
(30, 16, 34),
(31, 16, 35),
(32, 17, 33),
(33, 17, 34),
(34, 17, 35),
(35, 18, 33),
(36, 18, 34),
(37, 18, 35),
(38, 26, 52),
(39, 27, 52),
(40, 28, 52),
(41, 34, 33),
(42, 34, 34),
(43, 34, 35),
(44, 35, 52),
(47, 16, 28)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- appointments
-- -----------------------------------------------------------------------------
INSERT INTO appointments ("id", "clinic_id", "start_time", "end_time", "owner_id", "pet_id", "visit_type", "reservation_type_id", "doctor_id", "is_designated", "status", "notes", "source", "created_by", "is_staff_delegated", "customer_fields", "reservation_route", "actual_reservation_at", "created_at", "updated_at", "deleted_at", "line_customer_id") VALUES
(1, 1, '2026-05-22 00:00:00+00', '2026-05-22 00:15:00+00', 1, 1, 'revisit', 1, 1, 't', 'completed', '皮膚の経過観察', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(2, 1, '2026-05-22 00:15:00+00', '2026-05-22 00:30:00+00', 2, 3, 'revisit', 7, 2, 'f', 'accounting', '猫の定期健診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(3, 1, '2026-05-22 01:00:00+00', '2026-05-22 01:15:00+00', 3, 4, 'revisit', 1, 1, 't', 'in_consultation', '足を引きずっている', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(4, 1, '2026-05-22 01:15:00+00', '2026-05-22 01:30:00+00', 4, 6, 'first', 3, 2, 'f', 'checked_in', 'ワクチン接種希望', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(5, 1, '2026-05-22 05:00:00+00', '2026-05-22 05:15:00+00', 6, 8, 'revisit', 1, 1, 'f', 'no_show', '食欲低下が続いている', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(6, 1, '2026-05-23 00:00:00+00', '2026-05-23 00:15:00+00', 7, 9, 'revisit', 1, 2, 't', 'no_show', '耳の治療経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(7, 1, '2026-05-23 01:00:00+00', '2026-05-23 01:15:00+00', 8, 10, 'first', 1, 1, 'f', 'no_show', '嘔吐が続いている', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(8, 1, '2026-05-24 00:30:00+00', '2026-05-24 00:45:00+00', 9, 11, 'revisit', 1, 2, 'f', 'no_show', 'ルナの経過観察', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(9, 1, '2026-05-25 02:00:00+00', '2026-05-25 02:15:00+00', 10, 12, 'first', 3, 1, 'f', 'no_show', '初回ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(10, 1, '2026-05-26 05:00:00+00', '2026-05-26 05:15:00+00', 11, 13, 'revisit', 1, 2, 't', 'no_show', '腎臓値の経過観察', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(11, 2, '2026-05-22 01:00:00+00', '2026-05-22 01:15:00+00', 23, 29, 'revisit', 26, 16, 't', 'no_show', 'クロの定期診察', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(12, 2, '2026-05-23 05:00:00+00', '2026-05-23 05:15:00+00', 24, 31, 'first', 28, 16, 'f', 'no_show', 'ポポのワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(13, 2, '2026-05-24 02:00:00+00', '2026-05-24 02:15:00+00', 25, 32, 'revisit', 31, 16, 'f', 'no_show', 'ダンの健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(101, 1, '2025-12-20 01:00:00+00', '2025-12-20 02:30:00+00', 1, 1, 'first', 9, 6, 'f', 'completed', 'サマーカット希望', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(102, 1, '2025-12-25 01:00:00+00', '2025-12-25 02:30:00+00', 1, 2, 'first', 9, 12, 'f', 'no_show', 'ふんわりカット', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(103, 1, '2025-12-22 01:00:00+00', '2025-12-22 02:30:00+00', 2, 3, 'first', 9, 6, 'f', 'in_consultation', '毛玉カット', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(104, 1, '2026-03-18 01:00:00+00', '2026-03-18 02:00:00+00', 4, 6, 'first', 11, 6, 'f', 'completed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(105, 1, '2026-03-18 02:00:00+00', '2026-03-18 03:30:00+00', 15, 17, 'first', 9, 12, 'f', 'completed', '全体カット', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(106, 1, '2026-03-18 04:00:00+00', '2026-03-18 05:30:00+00', 8, 10, 'first', 10, 12, 'f', 'no_show', '爪切り・ブラッシング', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(107, 1, '2026-03-18 05:00:00+00', '2026-03-18 06:00:00+00', 13, 15, 'first', 11, 6, 'f', 'completed', 'シャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(108, 1, '2026-03-18 06:00:00+00', '2026-03-18 07:30:00+00', 4, 6, 'first', 9, 6, 'f', 'no_show', 'トリミング', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(109, 1, '2026-05-07 07:30:00+00', '2026-05-07 09:45:00+00', 1, 1, 'revisit', 1, 4, 'f', 'checked_in', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-07 05:44:20.90667+00', '2026-05-13 05:26:48.697607+00', NULL, NULL),
(110, 1, '2026-05-09 06:00:00+00', '2026-05-09 06:15:00+00', 1, 1, 'revisit', 2, NULL, 'f', 'no_show', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-09 04:37:04.699057+00', '2026-05-09 11:00:00.538006+00', NULL, NULL),
(111, 1, '2026-05-09 01:30:00+00', '2026-05-09 01:45:00+00', 1, 2, 'revisit', 2, 4, 'f', 'in_consultation', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-09 08:50:41.468507+00', '2026-05-09 08:52:06.54822+00', NULL, NULL),
(112, 1, '2026-05-11 16:45:00+09', '2026-05-11 17:45:00+09', 1, 1, 'revisit', 2, 1, 'f', 'no_show', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-11 08:33:53.134392+00', '2026-05-12 01:00:00.734922+00', NULL, NULL),
(113, 1, '2026-05-11 16:30:00+09', '2026-05-11 17:30:00+09', 1, 1, 'revisit', 2, NULL, 'f', 'no_show', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-11 08:34:46.519909+00', '2026-05-12 01:00:00.785954+00', NULL, NULL),
(114, 1, '2026-05-13 01:00:00+00', '2026-05-13 01:15:00+00', 1, 1, 'revisit', 2, 1, 'f', 'checked_in', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-12 04:58:04.581474+00', '2026-05-13 05:29:31.75124+00', NULL, NULL),
(115, 1, '2026-05-12 05:30:00+00', '2026-05-12 05:45:00+00', 1, 1, 'revisit', 2, 1, 'f', 'checked_in', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-12 05:05:01.562877+00', '2026-05-12 05:10:46.021868+00', NULL, NULL),
(116, 1, '2026-05-20 01:00:00+00', '2026-05-20 02:00:00+00', 1, 1, 'revisit', 3, NULL, 'f', 'accounting', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-13 05:24:00.074152+00', '2026-05-20 02:19:05.324388+00', NULL, NULL),
(117, 1, '2026-05-20 01:00:00+00', '2026-05-20 02:00:00+00', 1, 2, 'revisit', 3, NULL, 'f', 'no_show', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-13 05:24:00.104721+00', '2026-05-20 06:00:00.741487+00', NULL, NULL),
(118, 1, '2026-05-21 08:30:00+00', '2026-05-21 09:30:00+00', 1, 1, 'revisit', 3, 1, 'f', 'no_show', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-21 08:29:00+00', '2026-05-21 08:30:00+00', NULL, NULL),
(119, 1, '2026-05-21 08:00:00+00', '2026-05-21 08:15:00+00', 1, 1, 'first', 3, 1, 'f', 'checked_in', '', 'manual', 4, 'f', '{}', NULL, NULL, '2026-05-21 07:59:00+00', '2026-05-21 08:00:00+00', NULL, NULL),
(201, 1, '2026-05-22 00:30:00+00', '2026-05-22 00:45:00+00', 39, 49, 'revisit', 1, 1, 'f', 'confirmed', '下痢が続いている', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(202, 1, '2026-05-22 00:30:00+00', '2026-05-22 00:45:00+00', 40, 50, 'first', 3, 2, 'f', 'confirmed', '混合ワクチン希望', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(203, 1, '2026-05-22 00:45:00+00', '2026-05-22 01:00:00+00', 41, 51, 'revisit', 1, 3, 'f', 'confirmed', '皮膚の赤み', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(204, 1, '2026-05-22 01:00:00+00', '2026-05-22 01:15:00+00', 42, 52, 'revisit', 1, 2, 'f', 'confirmed', '目やにが出る', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(205, 1, '2026-05-22 01:15:00+00', '2026-05-22 01:30:00+00', 43, 53, 'first', 1, 1, 'f', 'confirmed', '咳をしている', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(206, 1, '2026-05-22 01:30:00+00', '2026-05-22 01:45:00+00', 44, 54, 'revisit', 7, 3, 'f', 'confirmed', '心臓の定期検査', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(207, 1, '2026-05-22 01:45:00+00', '2026-05-22 02:00:00+00', 45, 55, 'first', 5, 2, 'f', 'confirmed', '狂犬病ワクチン', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(208, 1, '2026-05-22 02:00:00+00', '2026-05-22 02:15:00+00', 46, 56, 'revisit', 1, 1, 'f', 'confirmed', '耳を痒がる', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(209, 1, '2026-05-22 02:15:00+00', '2026-05-22 02:30:00+00', 47, 57, 'revisit', 1, 2, 'f', 'confirmed', 'アレルギー治療', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(210, 1, '2026-05-22 02:30:00+00', '2026-05-22 02:45:00+00', 48, 58, 'first', 7, 3, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(211, 1, '2026-05-22 02:45:00+00', '2026-05-22 03:00:00+00', 49, 59, 'revisit', 1, 1, 'f', 'confirmed', '術後経過観察', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(212, 1, '2026-05-22 03:00:00+00', '2026-05-22 03:15:00+00', 50, 60, 'revisit', 1, 2, 'f', 'confirmed', '腎臓ケア相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(213, 1, '2026-05-22 00:00:00+00', '2026-05-22 00:30:00+00', 51, 61, 'revisit', 7, 13, 'f', 'confirmed', '定期健診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(214, 1, '2026-05-22 00:30:00+00', '2026-05-22 01:00:00+00', 52, 62, 'revisit', 3, 13, 'f', 'confirmed', 'ワクチンと検診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(215, 1, '2026-05-22 01:00:00+00', '2026-05-22 01:30:00+00', 53, 63, 'revisit', 7, 13, 'f', 'confirmed', 'シニア健診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(216, 1, '2026-05-22 01:30:00+00', '2026-05-22 02:00:00+00', 54, 64, 'first', 7, 13, 'f', 'confirmed', '初回検診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(217, 1, '2026-05-22 02:00:00+00', '2026-05-22 02:30:00+00', 55, 65, 'revisit', 1, 13, 'f', 'confirmed', 'ダイエット相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(218, 1, '2026-05-22 00:00:00+00', '2026-05-22 00:15:00+00', 56, 66, 'revisit', 1, 3, 'f', 'confirmed', '吐血した', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(219, 1, '2026-05-22 00:15:00+00', '2026-05-22 00:30:00+00', 57, 67, 'revisit', 1, 3, 'f', 'confirmed', '誤飲疑い', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(220, 1, '2026-05-22 03:15:00+00', '2026-05-22 03:30:00+00', 58, 68, 'revisit', 6, 1, 'f', 'confirmed', 'ノミダニ予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(221, 1, '2026-05-22 03:30:00+00', '2026-05-22 03:45:00+00', 59, 69, 'revisit', 1, 2, 'f', 'confirmed', '足の爪が折れた', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(222, 1, '2026-05-22 03:45:00+00', '2026-05-22 04:00:00+00', 60, 70, 'revisit', 6, 3, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(223, 1, '2026-05-22 04:00:00+00', '2026-05-22 04:15:00+00', 39, 71, 'revisit', 4, 1, 'f', 'confirmed', '耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(224, 1, '2026-05-22 04:15:00+00', '2026-05-22 04:30:00+00', 40, 72, 'revisit', 1, 2, 'f', 'confirmed', '食欲不振', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(225, 1, '2026-05-22 04:30:00+00', '2026-05-22 04:45:00+00', 41, 73, 'revisit', 1, 3, 'f', 'confirmed', '歯石の相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(226, 1, '2026-05-22 04:45:00+00', '2026-05-22 05:00:00+00', 42, 74, 'revisit', 1, 1, 'f', 'confirmed', '再診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(227, 1, '2026-05-22 00:00:00+00', '2026-05-22 01:30:00+00', 43, 75, 'revisit', 9, 12, 'f', 'confirmed', 'サマーカット', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(228, 1, '2026-05-22 01:30:00+00', '2026-05-22 03:00:00+00', 1, 2, 'revisit', 11, 12, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(229, 1, '2026-05-22 03:00:00+00', '2026-05-22 04:30:00+00', 2, 3, 'revisit', 9, 12, 'f', 'confirmed', 'トリミング', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(230, 1, '2026-05-22 00:00:00+00', '2026-05-22 00:15:00+00', 3, 5, 'revisit', 1, 2, 'f', 'confirmed', '下痢', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(231, 1, '2026-05-22 05:15:00+00', '2026-05-22 05:30:00+00', 4, 6, 'revisit', 1, 2, 'f', 'confirmed', '皮膚病の経過', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(232, 1, '2026-05-22 05:30:00+00', '2026-05-22 05:45:00+00', 5, 7, 'revisit', 1, 3, 'f', 'confirmed', '目薬の追加', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(233, 1, '2026-05-22 05:45:00+00', '2026-05-22 06:00:00+00', 6, 8, 'revisit', 3, 1, 'f', 'confirmed', 'ワクチン', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(234, 1, '2026-05-22 06:00:00+00', '2026-05-22 06:15:00+00', 7, 9, 'revisit', 1, 2, 'f', 'confirmed', '耳の赤み', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(235, 1, '2026-05-22 06:15:00+00', '2026-05-22 06:30:00+00', 8, 10, 'revisit', 7, 3, 'f', 'confirmed', '血液検査', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(236, 1, '2026-05-22 06:30:00+00', '2026-05-22 06:45:00+00', 9, 11, 'revisit', 1, 1, 'f', 'confirmed', '定期処方', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(237, 1, '2026-05-22 06:45:00+00', '2026-05-22 07:00:00+00', 10, 12, 'revisit', 1, 2, 'f', 'confirmed', '再診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(238, 1, '2026-05-22 07:00:00+00', '2026-05-22 07:15:00+00', 11, 13, 'revisit', 4, 3, 'f', 'confirmed', '爪切り', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(239, 1, '2026-05-22 07:15:00+00', '2026-05-22 07:30:00+00', 12, 14, 'revisit', 1, 1, 'f', 'confirmed', '外耳炎', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(240, 1, '2026-05-22 07:30:00+00', '2026-05-22 07:45:00+00', 13, 15, 'revisit', 6, 2, 'f', 'confirmed', 'フィラリア', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(241, 1, '2026-05-22 07:45:00+00', '2026-05-22 08:00:00+00', 14, 16, 'revisit', 6, 3, 'f', 'confirmed', 'ノミダニ', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(242, 1, '2026-05-22 08:00:00+00', '2026-05-22 08:15:00+00', 15, 17, 'revisit', 1, 1, 'f', 'confirmed', '健康相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(243, 1, '2026-05-22 08:15:00+00', '2026-05-22 08:30:00+00', 16, 18, 'revisit', 3, 2, 'f', 'confirmed', 'ワクチンの相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(244, 1, '2026-05-22 08:30:00+00', '2026-05-22 08:45:00+00', 17, 19, 'revisit', 4, 3, 'f', 'confirmed', '耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(245, 1, '2026-05-22 08:45:00+00', '2026-05-22 09:00:00+00', 18, 20, 'revisit', 1, 1, 'f', 'confirmed', '再診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(246, 1, '2026-05-22 05:00:00+00', '2026-05-22 06:30:00+00', 19, 21, 'revisit', 9, 12, 'f', 'confirmed', '全体カット', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(247, 1, '2026-05-22 06:30:00+00', '2026-05-22 08:00:00+00', 20, 22, 'revisit', 11, 12, 'f', 'confirmed', 'シャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(248, 1, '2026-05-22 08:00:00+00', '2026-05-22 09:30:00+00', 39, 49, 'revisit', 9, 12, 'f', 'confirmed', 'トリミング', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(249, 1, '2026-05-22 05:00:00+00', '2026-05-22 05:30:00+00', 40, 50, 'revisit', 7, 13, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(250, 1, '2026-05-22 05:30:00+00', '2026-05-22 06:00:00+00', 41, 51, 'revisit', 7, 13, 'f', 'confirmed', '検診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(251, 1, '2026-05-22 06:00:00+00', '2026-05-22 06:30:00+00', 42, 52, 'revisit', 7, 13, 'f', 'confirmed', 'シニア検査', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(252, 1, '2026-05-22 06:30:00+00', '2026-05-22 07:00:00+00', 43, 53, 'revisit', 7, 13, 'f', 'confirmed', '血液検査', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(253, 1, '2026-05-22 07:00:00+00', '2026-05-22 07:30:00+00', 44, 54, 'revisit', 7, 13, 'f', 'confirmed', '定期健診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(254, 1, '2026-05-22 07:30:00+00', '2026-05-22 08:00:00+00', 45, 55, 'revisit', 5, 13, 'f', 'confirmed', '狂犬病ワクチン', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(255, 1, '2026-05-22 08:00:00+00', '2026-05-22 08:30:00+00', 46, 56, 'revisit', 3, 13, 'f', 'confirmed', '混合ワクチン', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(256, 1, '2026-05-22 08:30:00+00', '2026-05-22 09:00:00+00', 47, 57, 'revisit', 1, 13, 'f', 'confirmed', '健康相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(257, 1, '2026-05-22 05:00:00+00', '2026-05-22 05:15:00+00', 48, 58, 'revisit', 1, 3, 'f', 'confirmed', '再診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(258, 1, '2026-05-22 05:15:00+00', '2026-05-22 05:30:00+00', 49, 59, 'revisit', 1, 3, 'f', 'confirmed', '処方のみ', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(259, 1, '2026-05-22 09:00:00+00', '2026-05-22 09:15:00+00', 50, 60, 'revisit', 1, 3, 'f', 'confirmed', '下痢の相談', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(260, 1, '2026-05-22 09:15:00+00', '2026-05-22 09:30:00+00', 51, 61, 'revisit', 1, 3, 'f', 'confirmed', '再診', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(501, 1, '2026-05-22 01:00:00+00', '2026-05-22 01:15:00+00', 20, 22, 'revisit', 1, 1, 'f', 'cancelled', '飼い主急病のためキャンセル', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(502, 1, '2026-05-22 02:30:00+00', '2026-05-22 02:45:00+00', 21, 25, 'revisit', 1, 2, 'f', 'no_show', '連絡なし不在', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(503, 1, '2026-05-22 05:00:00+00', '2026-05-22 05:30:00+00', 22, 27, 'revisit', 1, 8, 't', 'checked_in', '急患：呼吸困難', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(504, 1, '2026-05-01 06:00:00+00', '2026-05-01 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(505, 1, '2026-05-02 06:00:00+00', '2026-05-02 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(506, 1, '2026-05-03 06:00:00+00', '2026-05-03 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(507, 1, '2026-05-04 06:00:00+00', '2026-05-04 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(508, 1, '2026-05-05 06:00:00+00', '2026-05-05 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(509, 1, '2026-05-06 06:00:00+00', '2026-05-06 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(510, 1, '2026-05-07 06:00:00+00', '2026-05-07 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(511, 1, '2026-05-08 06:00:00+00', '2026-05-08 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(512, 1, '2026-05-09 06:00:00+00', '2026-05-09 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(513, 1, '2026-05-10 06:00:00+00', '2026-05-10 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(514, 1, '2026-05-11 06:00:00+00', '2026-05-11 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(515, 1, '2026-05-12 06:00:00+00', '2026-05-12 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(516, 1, '2026-05-13 06:00:00+00', '2026-05-13 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(517, 1, '2026-05-14 06:00:00+00', '2026-05-14 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(518, 1, '2026-05-15 06:00:00+00', '2026-05-15 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(519, 1, '2026-05-16 06:00:00+00', '2026-05-16 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(520, 1, '2026-05-17 06:00:00+00', '2026-05-17 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(521, 1, '2026-05-18 06:00:00+00', '2026-05-18 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(522, 1, '2026-05-19 06:00:00+00', '2026-05-19 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(523, 1, '2026-05-20 06:00:00+00', '2026-05-20 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(524, 1, '2026-05-21 06:00:00+00', '2026-05-21 06:15:00+00', 1, 1, 'revisit', 1, 33, 'f', 'completed', '', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1000, 1, '2026-05-30 09:00:00+09', '2026-05-30 10:30:00+09', 7, 9, 'first', 9, 1, 'f', 'completed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1001, 1, '2026-05-30 14:00:00+09', '2026-05-30 14:15:00+09', 50, 60, 'revisit', 2, 18, 't', 'completed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1002, 1, '2026-05-30 09:20:00+00', '2026-05-30 09:35:00+00', 42, 52, 'first', 3, 1, 't', 'completed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1003, 1, '2026-05-30 14:20:00+09', '2026-05-30 14:35:00+09', 48, 58, 'first', 4, 3, 'f', 'completed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1004, 1, '2026-05-30 09:40:00+00', '2026-05-30 09:55:00+00', 39, 71, 'revisit', 5, 15, 'f', 'completed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1005, 1, '2026-05-30 14:40:00+09', '2026-05-30 14:55:00+09', 20, 24, 'revisit', 6, 2, 'f', 'completed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1006, 1, '2026-05-30 10:00:00+09', '2026-05-30 10:15:00+09', 21, 25, 'revisit', 7, 24, 'f', 'completed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1007, 1, '2026-05-30 15:00:00+09', '2026-05-30 15:15:00+09', 9, 11, 'revisit', 8, 29, 'f', 'completed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1008, 1, '2026-05-30 10:20:00+09', '2026-05-30 10:35:00+09', 8, 10, 'first', 1, 33, 'f', 'completed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1009, 1, '2026-05-30 15:20:00+09', '2026-05-30 15:35:00+09', 14, 16, 'first', 2, 12, 'f', 'completed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1010, 1, '2026-05-30 10:40:00+09', '2026-05-30 10:55:00+09', 10, 12, 'first', 3, 26, 'f', 'completed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1011, 1, '2026-05-30 15:40:00+09', '2026-05-30 15:55:00+09', 24, 31, 'revisit', 4, 34, 'f', 'completed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1012, 1, '2026-05-30 11:00:00+09', '2026-05-30 11:15:00+09', 9, 11, 'revisit', 5, 11, 't', 'completed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1013, 1, '2026-05-30 16:00:00+09', '2026-05-30 16:15:00+09', 37, 46, 'revisit', 6, 5, 'f', 'completed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1014, 1, '2026-05-30 11:20:00+09', '2026-05-30 11:35:00+09', 57, 67, 'revisit', 7, 5, 'f', 'completed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1015, 1, '2026-05-30 16:20:00+09', '2026-05-30 16:35:00+09', 38, 47, 'revisit', 8, 10, 'f', 'completed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1016, 1, '2026-05-30 11:40:00+09', '2026-05-30 11:55:00+09', 2, 3, 'first', 1, 18, 'f', 'completed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1017, 1, '2026-05-30 16:40:00+09', '2026-05-30 16:55:00+09', 9, 11, 'revisit', 2, 6, 't', 'completed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1018, 1, '2026-05-30 12:00:00+09', '2026-05-30 12:15:00+09', 40, 50, 'revisit', 3, 29, 'f', 'completed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1019, 1, '2026-05-30 17:00:00+09', '2026-05-30 17:15:00+09', 50, 60, 'revisit', 4, 3, 'f', 'completed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1020, 1, '2026-05-31 09:00:00+09', '2026-05-31 10:30:00+09', 29, 37, 'revisit', 9, 30, 'f', 'completed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1021, 1, '2026-05-31 14:00:00+09', '2026-05-31 14:15:00+09', 41, 51, 'revisit', 6, 17, 't', 'checked_in', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:40:33.934139+00', NULL, NULL),
(1022, 1, '2026-05-31 09:20:00+00', '2026-05-31 09:35:00+00', 41, 51, 'revisit', 7, 2, 'f', 'completed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1023, 1, '2026-05-31 14:20:00+09', '2026-05-31 14:35:00+09', 38, 47, 'revisit', 8, 16, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1024, 1, '2026-05-31 09:40:00+00', '2026-05-31 09:55:00+00', 32, 40, 'revisit', 1, 24, 't', 'completed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1025, 1, '2026-05-31 14:40:00+09', '2026-05-31 14:55:00+09', 34, 43, 'revisit', 2, 18, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1026, 1, '2026-05-31 10:00:00+09', '2026-05-31 10:15:00+09', 21, 26, 'revisit', 3, 25, 'f', 'completed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1027, 1, '2026-05-31 15:00:00+09', '2026-05-31 15:15:00+09', 1, 2, 'revisit', 4, 22, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1028, 1, '2026-05-31 10:20:00+09', '2026-05-31 10:35:00+09', 54, 64, 'revisit', 5, 30, 't', 'completed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1029, 1, '2026-05-31 15:20:00+09', '2026-05-31 15:35:00+09', 17, 19, 'revisit', 6, 4, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1030, 1, '2026-05-31 10:40:00+09', '2026-05-31 10:55:00+09', 43, 53, 'revisit', 7, 14, 'f', 'completed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1031, 1, '2026-05-31 15:40:00+09', '2026-05-31 15:55:00+09', 12, 14, 'revisit', 8, 16, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1032, 1, '2026-05-31 11:00:00+09', '2026-05-31 11:15:00+09', 5, 7, 'revisit', 1, 18, 't', 'completed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1033, 1, '2026-05-31 16:00:00+09', '2026-05-31 16:15:00+09', 21, 25, 'first', 2, 4, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1034, 1, '2026-05-31 11:20:00+09', '2026-05-31 11:35:00+09', 2, 3, 'revisit', 3, 17, 'f', 'completed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1035, 1, '2026-05-31 16:20:00+09', '2026-05-31 16:35:00+09', 3, 4, 'revisit', 4, 9, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1036, 1, '2026-05-31 11:40:00+09', '2026-05-31 11:55:00+09', 21, 25, 'revisit', 5, 23, 'f', 'completed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1037, 1, '2026-05-31 16:40:00+09', '2026-05-31 16:55:00+09', 5, 7, 'revisit', 6, 10, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1038, 1, '2026-05-31 12:00:00+09', '2026-05-31 12:15:00+09', 31, 39, 'revisit', 7, 23, 'f', 'completed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1039, 1, '2026-05-31 17:00:00+09', '2026-05-31 17:15:00+09', 28, 35, 'revisit', 8, 14, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1040, 1, '2026-05-31 12:20:00+09', '2026-05-31 12:35:00+09', 1, 1, 'first', 1, 34, 'f', 'completed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1041, 1, '2026-06-01 09:00:00+09', '2026-06-01 10:30:00+09', 42, 52, 'revisit', 10, 1, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1042, 1, '2026-06-01 14:00:00+09', '2026-06-01 14:15:00+09', 37, 46, 'revisit', 3, 6, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1043, 1, '2026-06-01 09:20:00+00', '2026-06-01 09:35:00+00', 10, 12, 'revisit', 4, 21, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1044, 1, '2026-06-01 14:20:00+09', '2026-06-01 14:35:00+09', 36, 45, 'revisit', 5, 26, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1045, 1, '2026-06-01 09:40:00+00', '2026-06-01 09:55:00+00', 52, 62, 'revisit', 6, 33, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1046, 1, '2026-06-01 14:40:00+09', '2026-06-01 14:55:00+09', 4, 6, 'revisit', 7, 31, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1047, 1, '2026-06-01 10:00:00+09', '2026-06-01 10:15:00+09', 10, 81, 'revisit', 8, 20, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1048, 1, '2026-06-01 15:00:00+09', '2026-06-01 15:15:00+09', 2, 3, 'revisit', 1, 34, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1049, 1, '2026-06-01 10:20:00+09', '2026-06-01 10:35:00+09', 27, 34, 'revisit', 2, 18, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1050, 1, '2026-06-01 15:20:00+09', '2026-06-01 15:35:00+09', 38, 47, 'revisit', 3, 7, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1051, 1, '2026-06-01 10:40:00+09', '2026-06-01 10:55:00+09', 14, 16, 'revisit', 4, 14, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1052, 1, '2026-06-01 15:40:00+09', '2026-06-01 15:55:00+09', 7, 9, 'revisit', 5, 32, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1053, 1, '2026-06-01 11:00:00+09', '2026-06-01 11:15:00+09', 19, 22, 'revisit', 6, 13, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1054, 1, '2026-06-01 16:00:00+09', '2026-06-01 16:15:00+09', 21, 26, 'revisit', 7, 12, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1055, 1, '2026-06-01 11:20:00+09', '2026-06-01 11:35:00+09', 59, 69, 'revisit', 8, 9, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1056, 1, '2026-06-01 16:20:00+09', '2026-06-01 16:35:00+09', 34, 43, 'first', 1, 34, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1057, 1, '2026-06-01 11:40:00+09', '2026-06-01 11:55:00+09', 32, 40, 'revisit', 2, 21, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1058, 1, '2026-06-01 16:40:00+09', '2026-06-01 16:55:00+09', 20, 24, 'revisit', 3, 20, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1059, 1, '2026-06-01 12:00:00+09', '2026-06-01 12:15:00+09', 60, 70, 'revisit', 4, 6, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1060, 1, '2026-06-01 17:00:00+09', '2026-06-01 17:15:00+09', 39, 71, 'first', 5, 10, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1061, 1, '2026-06-01 12:20:00+09', '2026-06-01 12:35:00+09', 47, 57, 'revisit', 6, 10, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1062, 1, '2026-06-02 09:00:00+09', '2026-06-02 10:30:00+09', 2, 3, 'revisit', 11, 28, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1063, 1, '2026-06-02 14:00:00+09', '2026-06-02 14:15:00+09', 10, 12, 'revisit', 8, 13, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1064, 1, '2026-06-02 09:20:00+00', '2026-06-02 09:35:00+00', 26, 33, 'revisit', 1, 5, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1065, 1, '2026-06-02 14:20:00+09', '2026-06-02 14:35:00+09', 40, 72, 'first', 2, 28, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1066, 1, '2026-06-02 09:40:00+00', '2026-06-02 09:55:00+00', 25, 32, 'revisit', 3, 31, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1067, 1, '2026-06-02 14:40:00+09', '2026-06-02 14:55:00+09', 41, 73, 'first', 4, 16, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1068, 1, '2026-06-02 10:00:00+09', '2026-06-02 10:15:00+09', 40, 72, 'revisit', 5, 16, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1069, 1, '2026-06-02 15:00:00+09', '2026-06-02 15:15:00+09', 25, 32, 'revisit', 6, 13, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1070, 1, '2026-06-02 10:20:00+09', '2026-06-02 10:35:00+09', 40, 50, 'revisit', 7, 13, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1071, 1, '2026-06-02 15:20:00+09', '2026-06-02 15:35:00+09', 32, 40, 'first', 8, 1, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1072, 1, '2026-06-02 10:40:00+09', '2026-06-02 10:55:00+09', 16, 18, 'revisit', 1, 15, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1073, 1, '2026-06-02 15:40:00+09', '2026-06-02 15:55:00+09', 14, 16, 'revisit', 2, 19, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1074, 1, '2026-06-02 11:00:00+09', '2026-06-02 11:15:00+09', 21, 26, 'revisit', 3, 35, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1075, 1, '2026-06-02 16:00:00+09', '2026-06-02 16:15:00+09', 25, 32, 'first', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1076, 1, '2026-06-02 11:20:00+09', '2026-06-02 11:35:00+09', 32, 40, 'revisit', 5, 25, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1077, 1, '2026-06-02 16:20:00+09', '2026-06-02 16:35:00+09', 9, 11, 'first', 6, 22, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1078, 1, '2026-06-02 11:40:00+09', '2026-06-02 11:55:00+09', 28, 35, 'revisit', 7, 22, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1079, 1, '2026-06-02 16:40:00+09', '2026-06-02 16:55:00+09', 53, 63, 'revisit', 8, 4, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1080, 1, '2026-06-02 12:00:00+09', '2026-06-02 12:15:00+09', 4, 6, 'first', 1, 34, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1081, 1, '2026-06-02 17:00:00+09', '2026-06-02 17:15:00+09', 9, 11, 'first', 2, 26, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1082, 1, '2026-06-02 12:20:00+09', '2026-06-02 12:35:00+09', 40, 72, 'revisit', 3, 20, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1083, 1, '2026-06-02 17:20:00+09', '2026-06-02 17:35:00+09', 22, 28, 'revisit', 4, 8, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1084, 1, '2026-06-02 12:40:00+09', '2026-06-02 12:55:00+09', 38, 47, 'first', 5, 12, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1085, 1, '2026-06-03 09:00:00+09', '2026-06-03 10:30:00+09', 57, 67, 'revisit', 10, 31, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1086, 1, '2026-06-03 14:00:00+09', '2026-06-03 14:15:00+09', 10, 81, 'revisit', 7, 12, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1087, 1, '2026-06-03 09:20:00+00', '2026-06-03 09:35:00+00', 21, 25, 'first', 8, 14, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1088, 1, '2026-06-03 14:20:00+09', '2026-06-03 14:35:00+09', 9, 11, 'revisit', 1, 19, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1089, 1, '2026-06-03 09:40:00+00', '2026-06-03 09:55:00+00', 27, 34, 'revisit', 2, 28, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1090, 1, '2026-06-03 14:40:00+09', '2026-06-03 14:55:00+09', 18, 20, 'first', 3, 34, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1091, 1, '2026-06-03 10:00:00+09', '2026-06-03 10:15:00+09', 10, 12, 'revisit', 4, 17, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1092, 1, '2026-06-03 15:00:00+09', '2026-06-03 15:15:00+09', 28, 35, 'first', 5, 7, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1093, 1, '2026-06-03 10:20:00+09', '2026-06-03 10:35:00+09', 10, 81, 'revisit', 6, 18, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1094, 1, '2026-06-03 15:20:00+09', '2026-06-03 15:35:00+09', 28, 35, 'first', 7, 1, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1095, 1, '2026-06-03 10:40:00+09', '2026-06-03 10:55:00+09', 30, 38, 'revisit', 8, 13, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1096, 1, '2026-06-03 15:40:00+09', '2026-06-03 15:55:00+09', 15, 17, 'first', 1, 26, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1097, 1, '2026-06-03 11:00:00+09', '2026-06-03 11:15:00+09', 10, 12, 'revisit', 2, 11, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1098, 1, '2026-06-03 16:00:00+09', '2026-06-03 16:15:00+09', 34, 42, 'revisit', 3, 1, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1099, 1, '2026-06-03 11:20:00+09', '2026-06-03 11:35:00+09', 52, 62, 'first', 4, 7, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1100, 1, '2026-06-03 16:20:00+09', '2026-06-03 16:35:00+09', 57, 67, 'revisit', 5, 25, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1101, 1, '2026-06-03 11:40:00+09', '2026-06-03 11:55:00+09', 44, 54, 'revisit', 6, 13, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1102, 1, '2026-06-03 16:40:00+09', '2026-06-03 16:55:00+09', 3, 5, 'revisit', 7, 7, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1103, 1, '2026-06-03 12:00:00+09', '2026-06-03 12:15:00+09', 37, 46, 'revisit', 8, 12, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1104, 1, '2026-06-03 17:00:00+09', '2026-06-03 17:15:00+09', 6, 8, 'first', 1, 30, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1105, 1, '2026-06-04 09:00:00+09', '2026-06-04 10:30:00+09', 22, 28, 'revisit', 10, 24, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1106, 1, '2026-06-04 14:00:00+09', '2026-06-04 15:30:00+09', 20, 24, 'first', 11, 11, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1107, 1, '2026-06-04 09:20:00+00', '2026-06-04 09:35:00+00', 49, 59, 'revisit', 4, 9, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1108, 1, '2026-06-04 14:20:00+09', '2026-06-04 14:35:00+09', 23, 30, 'first', 5, 29, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1109, 1, '2026-06-04 09:40:00+00', '2026-06-04 09:55:00+00', 1, 1, 'revisit', 6, 34, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1110, 1, '2026-06-04 14:40:00+09', '2026-06-04 14:55:00+09', 22, 27, 'revisit', 7, 4, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1111, 1, '2026-06-04 10:00:00+09', '2026-06-04 10:15:00+09', 2, 3, 'revisit', 8, 26, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1112, 1, '2026-06-04 15:00:00+09', '2026-06-04 15:15:00+09', 10, 81, 'first', 1, 4, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1113, 1, '2026-06-04 10:20:00+09', '2026-06-04 10:35:00+09', 38, 47, 'revisit', 2, 22, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1114, 1, '2026-06-04 15:20:00+09', '2026-06-04 15:35:00+09', 45, 55, 'revisit', 3, 26, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1115, 1, '2026-06-04 10:40:00+09', '2026-06-04 10:55:00+09', 21, 25, 'revisit', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1116, 1, '2026-06-04 15:40:00+09', '2026-06-04 15:55:00+09', 22, 27, 'revisit', 5, 17, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1117, 1, '2026-06-04 11:00:00+09', '2026-06-04 11:15:00+09', 26, 33, 'revisit', 6, 28, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1118, 1, '2026-06-04 16:00:00+09', '2026-06-04 16:15:00+09', 60, 70, 'revisit', 7, 10, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1119, 1, '2026-06-04 11:20:00+09', '2026-06-04 11:35:00+09', 15, 17, 'revisit', 8, 10, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1120, 1, '2026-06-04 16:20:00+09', '2026-06-04 16:35:00+09', 2, 3, 'first', 1, 21, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1121, 1, '2026-06-04 11:40:00+09', '2026-06-04 11:55:00+09', 28, 35, 'first', 2, 34, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1122, 1, '2026-06-04 16:40:00+09', '2026-06-04 16:55:00+09', 18, 21, 'revisit', 3, 27, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1123, 1, '2026-06-04 12:00:00+09', '2026-06-04 12:15:00+09', 39, 49, 'first', 4, 21, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1124, 1, '2026-06-04 17:00:00+09', '2026-06-04 17:15:00+09', 42, 52, 'revisit', 5, 24, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1125, 1, '2026-06-04 12:20:00+09', '2026-06-04 12:35:00+09', 49, 59, 'revisit', 6, 26, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1126, 1, '2026-06-04 17:20:00+09', '2026-06-04 17:35:00+09', 10, 12, 'first', 7, 2, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1127, 1, '2026-06-05 09:00:00+09', '2026-06-05 10:30:00+09', 14, 16, 'revisit', 12, 2, 'f', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1128, 1, '2026-06-05 14:00:00+09', '2026-06-05 14:15:00+09', 41, 51, 'first', 1, 30, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1129, 1, '2026-06-05 09:20:00+00', '2026-06-05 09:35:00+00', 42, 74, 'revisit', 2, 21, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1130, 1, '2026-06-05 14:20:00+09', '2026-06-05 14:35:00+09', 4, 6, 'revisit', 3, 4, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1131, 1, '2026-06-05 09:40:00+00', '2026-06-05 09:55:00+00', 12, 14, 'revisit', 4, 29, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1132, 1, '2026-06-05 14:40:00+09', '2026-06-05 14:55:00+09', 23, 30, 'revisit', 5, 28, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1133, 1, '2026-06-05 10:00:00+09', '2026-06-05 10:15:00+09', 51, 61, 'first', 6, 18, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1134, 1, '2026-06-05 15:00:00+09', '2026-06-05 15:15:00+09', 14, 16, 'revisit', 7, 4, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1135, 1, '2026-06-05 10:20:00+09', '2026-06-05 10:35:00+09', 41, 73, 'first', 8, 19, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1136, 1, '2026-06-05 15:20:00+09', '2026-06-05 15:35:00+09', 52, 62, 'revisit', 1, 24, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1137, 1, '2026-06-05 10:40:00+09', '2026-06-05 10:55:00+09', 6, 8, 'revisit', 2, 14, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1138, 1, '2026-06-05 15:40:00+09', '2026-06-05 15:55:00+09', 41, 73, 'revisit', 3, 20, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1139, 1, '2026-06-05 11:00:00+09', '2026-06-05 11:15:00+09', 19, 23, 'revisit', 4, 8, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1140, 1, '2026-06-05 16:00:00+09', '2026-06-05 16:15:00+09', 3, 5, 'revisit', 5, 19, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1141, 1, '2026-06-05 11:20:00+09', '2026-06-05 11:35:00+09', 38, 48, 'first', 6, 20, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1142, 1, '2026-06-05 16:20:00+09', '2026-06-05 16:35:00+09', 60, 70, 'revisit', 7, 14, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1143, 1, '2026-06-05 11:40:00+09', '2026-06-05 11:55:00+09', 19, 22, 'revisit', 8, 22, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1144, 1, '2026-06-05 16:40:00+09', '2026-06-05 16:55:00+09', 23, 30, 'revisit', 1, 33, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1145, 1, '2026-06-05 12:00:00+09', '2026-06-05 12:15:00+09', 30, 38, 'first', 2, 32, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1146, 1, '2026-06-05 17:00:00+09', '2026-06-05 17:15:00+09', 21, 26, 'revisit', 3, 31, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1147, 1, '2026-06-05 12:20:00+09', '2026-06-05 12:35:00+09', 52, 62, 'revisit', 4, 5, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1148, 1, '2026-06-06 09:00:00+09', '2026-06-06 10:30:00+09', 25, 32, 'first', 9, 29, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1149, 1, '2026-06-06 14:00:00+09', '2026-06-06 15:30:00+09', 40, 50, 'first', 10, 21, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1150, 1, '2026-06-06 09:20:00+00', '2026-06-06 09:35:00+00', 37, 46, 'revisit', 7, 14, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1151, 1, '2026-06-06 14:20:00+09', '2026-06-06 14:35:00+09', 35, 44, 'revisit', 8, 13, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1152, 1, '2026-06-06 09:40:00+00', '2026-06-06 09:55:00+00', 34, 43, 'revisit', 1, 30, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1153, 1, '2026-06-06 14:40:00+09', '2026-06-06 14:55:00+09', 8, 10, 'revisit', 2, 32, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1154, 1, '2026-06-06 10:00:00+09', '2026-06-06 10:15:00+09', 37, 46, 'first', 3, 10, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1155, 1, '2026-06-06 15:00:00+09', '2026-06-06 15:15:00+09', 37, 46, 'revisit', 4, 25, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1156, 1, '2026-06-06 10:20:00+09', '2026-06-06 10:35:00+09', 21, 26, 'revisit', 5, 24, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1157, 1, '2026-06-06 15:20:00+09', '2026-06-06 15:35:00+09', 41, 73, 'revisit', 6, 31, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1158, 1, '2026-06-06 10:40:00+09', '2026-06-06 10:55:00+09', 15, 17, 'revisit', 7, 10, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1159, 1, '2026-06-06 15:40:00+09', '2026-06-06 15:55:00+09', 14, 16, 'revisit', 8, 30, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1160, 1, '2026-06-06 11:00:00+09', '2026-06-06 11:15:00+09', 15, 17, 'revisit', 1, 21, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1161, 1, '2026-06-06 16:00:00+09', '2026-06-06 16:15:00+09', 40, 72, 'revisit', 2, 30, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1162, 1, '2026-06-06 11:20:00+09', '2026-06-06 11:35:00+09', 32, 40, 'revisit', 3, 11, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1163, 1, '2026-06-06 16:20:00+09', '2026-06-06 16:35:00+09', 24, 31, 'revisit', 4, 22, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1164, 1, '2026-06-06 11:40:00+09', '2026-06-06 11:55:00+09', 60, 70, 'revisit', 5, 15, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1165, 1, '2026-06-06 16:40:00+09', '2026-06-06 16:55:00+09', 40, 72, 'revisit', 6, 2, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1166, 1, '2026-06-06 12:00:00+09', '2026-06-06 12:15:00+09', 17, 19, 'revisit', 7, 30, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1167, 1, '2026-06-06 17:00:00+09', '2026-06-06 17:15:00+09', 24, 31, 'revisit', 8, 33, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1168, 1, '2026-06-06 12:20:00+09', '2026-06-06 12:35:00+09', 39, 49, 'revisit', 1, 19, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1169, 1, '2026-06-07 09:00:00+09', '2026-06-07 10:30:00+09', 16, 18, 'revisit', 10, 2, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1170, 1, '2026-06-07 14:00:00+09', '2026-06-07 15:30:00+09', 13, 15, 'revisit', 11, 33, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1171, 1, '2026-06-07 09:20:00+00', '2026-06-07 09:35:00+00', 41, 51, 'revisit', 4, 19, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1172, 1, '2026-06-07 14:20:00+09', '2026-06-07 14:35:00+09', 45, 55, 'revisit', 5, 22, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1173, 1, '2026-06-07 09:40:00+00', '2026-06-07 09:55:00+00', 22, 27, 'revisit', 6, 1, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1174, 1, '2026-06-07 14:40:00+09', '2026-06-07 14:55:00+09', 30, 38, 'revisit', 7, 15, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1175, 1, '2026-06-07 10:00:00+09', '2026-06-07 10:15:00+09', 50, 60, 'first', 8, 25, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1176, 1, '2026-06-07 15:00:00+09', '2026-06-07 15:15:00+09', 24, 31, 'revisit', 1, 14, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1177, 1, '2026-06-07 10:20:00+09', '2026-06-07 10:35:00+09', 39, 71, 'revisit', 2, 31, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1178, 1, '2026-06-07 15:20:00+09', '2026-06-07 15:35:00+09', 38, 48, 'revisit', 3, 23, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1179, 1, '2026-06-07 10:40:00+09', '2026-06-07 10:55:00+09', 34, 43, 'revisit', 4, 4, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1180, 1, '2026-06-07 15:40:00+09', '2026-06-07 15:55:00+09', 3, 4, 'revisit', 5, 16, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1181, 1, '2026-06-07 11:00:00+09', '2026-06-07 11:15:00+09', 30, 38, 'revisit', 6, 18, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1182, 1, '2026-06-07 16:00:00+09', '2026-06-07 16:15:00+09', 17, 19, 'revisit', 7, 20, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1183, 1, '2026-06-07 11:20:00+09', '2026-06-07 11:35:00+09', 44, 54, 'first', 8, 17, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1184, 1, '2026-06-07 16:20:00+09', '2026-06-07 16:35:00+09', 5, 7, 'revisit', 1, 18, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1185, 1, '2026-06-07 11:40:00+09', '2026-06-07 11:55:00+09', 8, 10, 'first', 2, 13, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1186, 1, '2026-06-07 16:40:00+09', '2026-06-07 16:55:00+09', 21, 26, 'revisit', 3, 12, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1187, 1, '2026-06-07 12:00:00+09', '2026-06-07 12:15:00+09', 56, 66, 'revisit', 4, 6, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1188, 1, '2026-06-07 17:00:00+09', '2026-06-07 17:15:00+09', 23, 30, 'revisit', 5, 16, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1189, 1, '2026-06-08 09:00:00+09', '2026-06-08 10:30:00+09', 11, 13, 'first', 10, 9, 't', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1190, 1, '2026-06-08 14:00:00+09', '2026-06-08 14:15:00+09', 41, 51, 'revisit', 7, 12, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1191, 1, '2026-06-08 09:20:00+00', '2026-06-08 09:35:00+00', 17, 19, 'revisit', 8, 18, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1192, 1, '2026-06-08 14:20:00+09', '2026-06-08 14:35:00+09', 44, 54, 'revisit', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1193, 1, '2026-06-08 09:40:00+00', '2026-06-08 09:55:00+00', 42, 74, 'revisit', 2, 32, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1194, 1, '2026-06-08 14:40:00+09', '2026-06-08 14:55:00+09', 36, 45, 'revisit', 3, 34, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1195, 1, '2026-06-08 10:00:00+09', '2026-06-08 10:15:00+09', 53, 63, 'revisit', 4, 19, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1196, 1, '2026-06-08 15:00:00+09', '2026-06-08 15:15:00+09', 42, 52, 'revisit', 5, 28, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1197, 1, '2026-06-08 10:20:00+09', '2026-06-08 10:35:00+09', 10, 12, 'first', 6, 31, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1198, 1, '2026-06-08 15:20:00+09', '2026-06-08 15:35:00+09', 55, 65, 'first', 7, 1, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1199, 1, '2026-06-08 10:40:00+09', '2026-06-08 10:55:00+09', 42, 52, 'first', 8, 33, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1200, 1, '2026-06-08 15:40:00+09', '2026-06-08 15:55:00+09', 24, 31, 'revisit', 1, 26, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1201, 1, '2026-06-08 11:00:00+09', '2026-06-08 11:15:00+09', 57, 67, 'revisit', 2, 20, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1202, 1, '2026-06-08 16:00:00+09', '2026-06-08 16:15:00+09', 17, 19, 'revisit', 3, 31, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1203, 1, '2026-06-08 11:20:00+09', '2026-06-08 11:35:00+09', 1, 2, 'revisit', 4, 17, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1204, 1, '2026-06-08 16:20:00+09', '2026-06-08 16:35:00+09', 55, 65, 'revisit', 5, 7, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1205, 1, '2026-06-08 11:40:00+09', '2026-06-08 11:55:00+09', 31, 39, 'revisit', 6, 3, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1206, 1, '2026-06-08 16:40:00+09', '2026-06-08 16:55:00+09', 39, 71, 'first', 7, 11, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1207, 1, '2026-06-08 12:00:00+09', '2026-06-08 12:15:00+09', 15, 17, 'first', 8, 6, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1208, 1, '2026-06-08 17:00:00+09', '2026-06-08 17:15:00+09', 45, 55, 'revisit', 1, 31, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1209, 1, '2026-06-08 12:20:00+09', '2026-06-08 12:35:00+09', 1, 1, 'first', 2, 8, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1210, 1, '2026-06-08 17:20:00+09', '2026-06-08 17:35:00+09', 21, 26, 'revisit', 3, 34, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1211, 1, '2026-06-08 12:40:00+09', '2026-06-08 12:55:00+09', 41, 51, 'revisit', 4, 11, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1212, 1, '2026-06-09 09:00:00+09', '2026-06-09 10:30:00+09', 22, 28, 'revisit', 9, 27, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1213, 1, '2026-06-09 14:00:00+09', '2026-06-09 14:15:00+09', 28, 35, 'revisit', 6, 4, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1214, 1, '2026-06-09 09:20:00+00', '2026-06-09 09:35:00+00', 8, 10, 'first', 7, 21, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1215, 1, '2026-06-09 14:20:00+09', '2026-06-09 14:35:00+09', 42, 52, 'first', 8, 7, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1216, 1, '2026-06-09 09:40:00+00', '2026-06-09 09:55:00+00', 36, 45, 'revisit', 1, 23, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1217, 1, '2026-06-09 14:40:00+09', '2026-06-09 14:55:00+09', 35, 44, 'revisit', 2, 16, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1218, 1, '2026-06-09 10:00:00+09', '2026-06-09 10:15:00+09', 37, 46, 'revisit', 3, 11, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1219, 1, '2026-06-09 15:00:00+09', '2026-06-09 15:15:00+09', 59, 69, 'first', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1220, 1, '2026-06-09 10:20:00+09', '2026-06-09 10:35:00+09', 5, 7, 'revisit', 5, 5, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1221, 1, '2026-06-09 15:20:00+09', '2026-06-09 15:35:00+09', 50, 60, 'revisit', 6, 7, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1222, 1, '2026-06-09 10:40:00+09', '2026-06-09 10:55:00+09', 28, 35, 'revisit', 7, 31, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1223, 1, '2026-06-09 15:40:00+09', '2026-06-09 15:55:00+09', 16, 18, 'revisit', 8, 19, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1224, 1, '2026-06-09 11:00:00+09', '2026-06-09 11:15:00+09', 47, 57, 'revisit', 1, 3, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1225, 1, '2026-06-09 16:00:00+09', '2026-06-09 16:15:00+09', 38, 47, 'revisit', 2, 17, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1226, 1, '2026-06-09 11:20:00+09', '2026-06-09 11:35:00+09', 26, 33, 'first', 3, 7, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1227, 1, '2026-06-09 16:20:00+09', '2026-06-09 16:35:00+09', 47, 57, 'revisit', 4, 19, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1228, 1, '2026-06-09 11:40:00+09', '2026-06-09 11:55:00+09', 39, 71, 'revisit', 5, 15, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1229, 1, '2026-06-09 16:40:00+09', '2026-06-09 16:55:00+09', 22, 28, 'revisit', 6, 2, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1230, 1, '2026-06-09 12:00:00+09', '2026-06-09 12:15:00+09', 60, 70, 'revisit', 7, 4, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1231, 1, '2026-06-09 17:00:00+09', '2026-06-09 17:15:00+09', 40, 72, 'revisit', 8, 7, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1232, 1, '2026-06-09 12:20:00+09', '2026-06-09 12:35:00+09', 17, 19, 'revisit', 1, 2, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1233, 1, '2026-06-10 09:00:00+09', '2026-06-10 10:30:00+09', 28, 36, 'revisit', 10, 20, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1234, 1, '2026-06-10 14:00:00+09', '2026-06-10 14:15:00+09', 38, 48, 'revisit', 3, 6, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1235, 1, '2026-06-10 09:20:00+00', '2026-06-10 09:35:00+00', 47, 57, 'revisit', 4, 23, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1236, 1, '2026-06-10 14:20:00+09', '2026-06-10 14:35:00+09', 41, 73, 'revisit', 5, 20, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1237, 1, '2026-06-10 09:40:00+00', '2026-06-10 09:55:00+00', 33, 41, 'revisit', 6, 3, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1238, 1, '2026-06-10 14:40:00+09', '2026-06-10 14:55:00+09', 51, 61, 'first', 7, 27, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1239, 1, '2026-06-10 10:00:00+09', '2026-06-10 10:15:00+09', 15, 17, 'revisit', 8, 24, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1240, 1, '2026-06-10 15:00:00+09', '2026-06-10 15:15:00+09', 24, 31, 'revisit', 1, 25, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1241, 1, '2026-06-10 10:20:00+09', '2026-06-10 10:35:00+09', 45, 55, 'revisit', 2, 34, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1242, 1, '2026-06-10 15:20:00+09', '2026-06-10 15:35:00+09', 13, 15, 'revisit', 3, 6, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1243, 1, '2026-06-10 10:40:00+09', '2026-06-10 10:55:00+09', 16, 18, 'revisit', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1244, 1, '2026-06-10 15:40:00+09', '2026-06-10 15:55:00+09', 3, 4, 'revisit', 5, 19, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1245, 1, '2026-06-10 11:00:00+09', '2026-06-10 11:15:00+09', 41, 51, 'revisit', 6, 2, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1246, 1, '2026-06-10 16:00:00+09', '2026-06-10 16:15:00+09', 33, 41, 'first', 7, 1, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1247, 1, '2026-06-10 11:20:00+09', '2026-06-10 11:35:00+09', 5, 80, 'revisit', 8, 23, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1248, 1, '2026-06-10 16:20:00+09', '2026-06-10 16:35:00+09', 55, 65, 'revisit', 1, 21, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1249, 1, '2026-06-10 11:40:00+09', '2026-06-10 11:55:00+09', 49, 59, 'revisit', 2, 3, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1250, 1, '2026-06-10 16:40:00+09', '2026-06-10 16:55:00+09', 4, 6, 'revisit', 3, 12, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1251, 1, '2026-06-10 12:00:00+09', '2026-06-10 12:15:00+09', 39, 71, 'revisit', 4, 8, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1252, 1, '2026-06-10 17:00:00+09', '2026-06-10 17:15:00+09', 18, 21, 'revisit', 5, 35, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1253, 1, '2026-06-10 12:20:00+09', '2026-06-10 12:35:00+09', 46, 56, 'revisit', 6, 1, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1254, 1, '2026-06-10 17:20:00+09', '2026-06-10 17:35:00+09', 47, 57, 'first', 7, 6, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1255, 1, '2026-06-10 12:40:00+09', '2026-06-10 12:55:00+09', 46, 56, 'revisit', 8, 10, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1256, 1, '2026-06-11 09:00:00+09', '2026-06-11 10:30:00+09', 53, 63, 'first', 9, 26, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1257, 1, '2026-06-11 14:00:00+09', '2026-06-11 15:30:00+09', 21, 25, 'first', 10, 4, 't', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1258, 1, '2026-06-11 09:20:00+00', '2026-06-11 09:35:00+00', 55, 65, 'revisit', 3, 10, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1259, 1, '2026-06-11 14:20:00+09', '2026-06-11 14:35:00+09', 51, 61, 'first', 4, 7, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1260, 1, '2026-06-11 09:40:00+00', '2026-06-11 09:55:00+00', 43, 53, 'first', 5, 19, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1261, 1, '2026-06-11 14:40:00+09', '2026-06-11 14:55:00+09', 19, 23, 'revisit', 6, 19, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1262, 1, '2026-06-11 10:00:00+09', '2026-06-11 10:15:00+09', 42, 74, 'revisit', 7, 29, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1263, 1, '2026-06-11 15:00:00+09', '2026-06-11 15:15:00+09', 17, 19, 'revisit', 8, 26, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1264, 1, '2026-06-11 10:20:00+09', '2026-06-11 10:35:00+09', 14, 16, 'revisit', 1, 9, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1265, 1, '2026-06-11 15:20:00+09', '2026-06-11 15:35:00+09', 57, 67, 'revisit', 2, 30, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1266, 1, '2026-06-11 10:40:00+09', '2026-06-11 10:55:00+09', 54, 64, 'revisit', 3, 23, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1267, 1, '2026-06-11 15:40:00+09', '2026-06-11 15:55:00+09', 47, 57, 'revisit', 4, 27, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1268, 1, '2026-06-11 11:00:00+09', '2026-06-11 11:15:00+09', 40, 72, 'revisit', 5, 35, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1269, 1, '2026-06-11 16:00:00+09', '2026-06-11 16:15:00+09', 3, 5, 'revisit', 6, 20, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1270, 1, '2026-06-11 11:20:00+09', '2026-06-11 11:35:00+09', 22, 27, 'first', 7, 21, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1271, 1, '2026-06-11 16:20:00+09', '2026-06-11 16:35:00+09', 52, 62, 'revisit', 8, 22, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1272, 1, '2026-06-11 11:40:00+09', '2026-06-11 11:55:00+09', 20, 24, 'first', 1, 3, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1273, 1, '2026-06-11 16:40:00+09', '2026-06-11 16:55:00+09', 14, 16, 'revisit', 2, 21, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1274, 1, '2026-06-11 12:00:00+09', '2026-06-11 12:15:00+09', 36, 45, 'first', 3, 16, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1275, 1, '2026-06-11 17:00:00+09', '2026-06-11 17:15:00+09', 23, 29, 'revisit', 4, 6, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1276, 1, '2026-06-11 12:20:00+09', '2026-06-11 12:35:00+09', 39, 71, 'revisit', 5, 20, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1277, 1, '2026-06-12 09:00:00+09', '2026-06-12 10:30:00+09', 47, 57, 'revisit', 10, 19, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1278, 1, '2026-06-12 14:00:00+09', '2026-06-12 14:15:00+09', 3, 4, 'first', 7, 4, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1279, 1, '2026-06-12 09:20:00+00', '2026-06-12 09:35:00+00', 31, 39, 'revisit', 8, 28, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1280, 1, '2026-06-12 14:20:00+09', '2026-06-12 14:35:00+09', 19, 22, 'revisit', 1, 33, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1281, 1, '2026-06-12 09:40:00+00', '2026-06-12 09:55:00+00', 37, 46, 'revisit', 2, 10, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1282, 1, '2026-06-12 14:40:00+09', '2026-06-12 14:55:00+09', 44, 54, 'revisit', 3, 27, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1283, 1, '2026-06-12 10:00:00+09', '2026-06-12 10:15:00+09', 41, 51, 'revisit', 4, 21, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1284, 1, '2026-06-12 15:00:00+09', '2026-06-12 15:15:00+09', 58, 68, 'revisit', 5, 35, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1285, 1, '2026-06-12 10:20:00+09', '2026-06-12 10:35:00+09', 35, 44, 'first', 6, 22, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1286, 1, '2026-06-12 15:20:00+09', '2026-06-12 15:35:00+09', 1, 2, 'first', 7, 35, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1287, 1, '2026-06-12 10:40:00+09', '2026-06-12 10:55:00+09', 30, 38, 'revisit', 8, 24, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1288, 1, '2026-06-12 15:40:00+09', '2026-06-12 15:55:00+09', 45, 55, 'revisit', 1, 19, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1289, 1, '2026-06-12 11:00:00+09', '2026-06-12 11:15:00+09', 10, 12, 'revisit', 2, 35, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1290, 1, '2026-06-12 16:00:00+09', '2026-06-12 16:15:00+09', 47, 57, 'revisit', 3, 20, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1291, 1, '2026-06-12 11:20:00+09', '2026-06-12 11:35:00+09', 5, 80, 'first', 4, 11, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1292, 1, '2026-06-12 16:20:00+09', '2026-06-12 16:35:00+09', 47, 57, 'revisit', 5, 32, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1293, 1, '2026-06-12 11:40:00+09', '2026-06-12 11:55:00+09', 20, 24, 'revisit', 6, 21, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1294, 1, '2026-06-12 16:40:00+09', '2026-06-12 16:55:00+09', 36, 45, 'revisit', 7, 9, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1295, 1, '2026-06-12 12:00:00+09', '2026-06-12 12:15:00+09', 42, 74, 'first', 8, 14, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1296, 1, '2026-06-12 17:00:00+09', '2026-06-12 17:15:00+09', 34, 43, 'revisit', 1, 34, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1297, 1, '2026-06-12 12:20:00+09', '2026-06-12 12:35:00+09', 54, 64, 'revisit', 2, 4, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1298, 1, '2026-06-13 09:00:00+09', '2026-06-13 10:30:00+09', 23, 29, 'revisit', 11, 18, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1299, 1, '2026-06-13 14:00:00+09', '2026-06-13 15:30:00+09', 43, 53, 'revisit', 12, 14, 't', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1300, 1, '2026-06-13 09:20:00+00', '2026-06-13 09:35:00+00', 36, 45, 'first', 5, 26, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1301, 1, '2026-06-13 14:20:00+09', '2026-06-13 14:35:00+09', 1, 2, 'revisit', 6, 4, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1302, 1, '2026-06-13 09:40:00+00', '2026-06-13 09:55:00+00', 42, 52, 'revisit', 7, 33, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1303, 1, '2026-06-13 14:40:00+09', '2026-06-13 14:55:00+09', 2, 3, 'revisit', 8, 4, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1304, 1, '2026-06-13 10:00:00+09', '2026-06-13 10:15:00+09', 41, 51, 'revisit', 1, 8, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1305, 1, '2026-06-13 15:00:00+09', '2026-06-13 15:15:00+09', 39, 71, 'first', 2, 10, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1306, 1, '2026-06-13 10:20:00+09', '2026-06-13 10:35:00+09', 5, 80, 'first', 3, 2, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1307, 1, '2026-06-13 15:20:00+09', '2026-06-13 15:35:00+09', 32, 40, 'first', 4, 17, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1308, 1, '2026-06-13 10:40:00+09', '2026-06-13 10:55:00+09', 23, 30, 'first', 5, 28, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1309, 1, '2026-06-13 15:40:00+09', '2026-06-13 15:55:00+09', 42, 52, 'revisit', 6, 21, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1310, 1, '2026-06-13 11:00:00+09', '2026-06-13 11:15:00+09', 34, 43, 'revisit', 7, 17, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1311, 1, '2026-06-13 16:00:00+09', '2026-06-13 16:15:00+09', 10, 12, 'first', 8, 32, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1312, 1, '2026-06-13 11:20:00+09', '2026-06-13 11:35:00+09', 22, 28, 'revisit', 1, 35, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1313, 1, '2026-06-13 16:20:00+09', '2026-06-13 16:35:00+09', 15, 17, 'revisit', 2, 14, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1314, 1, '2026-06-13 11:40:00+09', '2026-06-13 11:55:00+09', 41, 73, 'first', 3, 3, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1315, 1, '2026-06-13 16:40:00+09', '2026-06-13 16:55:00+09', 18, 20, 'revisit', 4, 9, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1316, 1, '2026-06-13 12:00:00+09', '2026-06-13 12:15:00+09', 8, 10, 'first', 5, 5, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1317, 1, '2026-06-13 17:00:00+09', '2026-06-13 17:15:00+09', 28, 35, 'revisit', 6, 29, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1318, 1, '2026-06-13 12:20:00+09', '2026-06-13 12:35:00+09', 28, 36, 'revisit', 7, 23, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1319, 1, '2026-06-13 17:20:00+09', '2026-06-13 17:35:00+09', 48, 58, 'revisit', 8, 29, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1320, 1, '2026-06-14 09:00:00+09', '2026-06-14 10:30:00+09', 52, 62, 'first', 9, 24, 't', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1321, 1, '2026-06-14 14:00:00+09', '2026-06-14 14:15:00+09', 6, 8, 'revisit', 2, 2, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1322, 1, '2026-06-14 09:20:00+00', '2026-06-14 09:35:00+00', 34, 42, 'revisit', 3, 32, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1323, 1, '2026-06-14 14:20:00+09', '2026-06-14 14:35:00+09', 46, 56, 'revisit', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1324, 1, '2026-06-14 09:40:00+00', '2026-06-14 09:55:00+00', 40, 50, 'revisit', 5, 22, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1325, 1, '2026-06-14 14:40:00+09', '2026-06-14 14:55:00+09', 5, 80, 'first', 6, 1, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1326, 1, '2026-06-14 10:00:00+09', '2026-06-14 10:15:00+09', 43, 75, 'first', 7, 25, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1327, 1, '2026-06-14 15:00:00+09', '2026-06-14 15:15:00+09', 39, 49, 'first', 8, 22, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1328, 1, '2026-06-14 10:20:00+09', '2026-06-14 10:35:00+09', 27, 34, 'revisit', 1, 24, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1329, 1, '2026-06-14 15:20:00+09', '2026-06-14 15:35:00+09', 38, 48, 'revisit', 2, 6, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1330, 1, '2026-06-14 10:40:00+09', '2026-06-14 10:55:00+09', 19, 22, 'revisit', 3, 2, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1331, 1, '2026-06-14 15:40:00+09', '2026-06-14 15:55:00+09', 32, 40, 'revisit', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1332, 1, '2026-06-14 11:00:00+09', '2026-06-14 11:15:00+09', 44, 54, 'revisit', 5, 14, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1333, 1, '2026-06-14 16:00:00+09', '2026-06-14 16:15:00+09', 45, 55, 'revisit', 6, 34, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1334, 1, '2026-06-14 11:20:00+09', '2026-06-14 11:35:00+09', 43, 75, 'revisit', 7, 30, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1335, 1, '2026-06-14 16:20:00+09', '2026-06-14 16:35:00+09', 9, 11, 'revisit', 8, 9, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1336, 1, '2026-06-14 11:40:00+09', '2026-06-14 11:55:00+09', 30, 38, 'revisit', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1337, 1, '2026-06-14 16:40:00+09', '2026-06-14 16:55:00+09', 21, 26, 'first', 2, 18, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1338, 1, '2026-06-14 12:00:00+09', '2026-06-14 12:15:00+09', 52, 62, 'revisit', 3, 30, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1339, 1, '2026-06-14 17:00:00+09', '2026-06-14 17:15:00+09', 12, 14, 'revisit', 4, 9, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1340, 1, '2026-06-15 09:00:00+09', '2026-06-15 10:30:00+09', 45, 55, 'revisit', 9, 3, 't', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1341, 1, '2026-06-15 14:00:00+09', '2026-06-15 15:30:00+09', 38, 48, 'revisit', 10, 14, 't', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1342, 1, '2026-06-15 09:20:00+00', '2026-06-15 09:35:00+00', 23, 29, 'revisit', 7, 11, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1343, 1, '2026-06-15 14:20:00+09', '2026-06-15 14:35:00+09', 7, 9, 'first', 8, 15, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1344, 1, '2026-06-15 09:40:00+00', '2026-06-15 09:55:00+00', 43, 53, 'first', 1, 31, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1345, 1, '2026-06-15 14:40:00+09', '2026-06-15 14:55:00+09', 5, 7, 'revisit', 2, 18, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1346, 1, '2026-06-15 10:00:00+09', '2026-06-15 10:15:00+09', 40, 50, 'revisit', 3, 6, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1347, 1, '2026-06-15 15:00:00+09', '2026-06-15 15:15:00+09', 27, 34, 'revisit', 4, 5, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1348, 1, '2026-06-15 10:20:00+09', '2026-06-15 10:35:00+09', 1, 1, 'revisit', 5, 28, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1349, 1, '2026-06-15 15:20:00+09', '2026-06-15 15:35:00+09', 28, 36, 'revisit', 6, 24, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1350, 1, '2026-06-15 10:40:00+09', '2026-06-15 10:55:00+09', 19, 22, 'revisit', 7, 28, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1351, 1, '2026-06-15 15:40:00+09', '2026-06-15 15:55:00+09', 54, 64, 'revisit', 8, 12, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1352, 1, '2026-06-15 11:00:00+09', '2026-06-15 11:15:00+09', 31, 39, 'revisit', 1, 11, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1353, 1, '2026-06-15 16:00:00+09', '2026-06-15 16:15:00+09', 42, 52, 'first', 2, 3, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1354, 1, '2026-06-15 11:20:00+09', '2026-06-15 11:35:00+09', 38, 48, 'revisit', 3, 31, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1355, 1, '2026-06-15 16:20:00+09', '2026-06-15 16:35:00+09', 15, 17, 'revisit', 4, 22, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1356, 1, '2026-06-15 11:40:00+09', '2026-06-15 11:55:00+09', 13, 15, 'revisit', 5, 23, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1357, 1, '2026-06-15 16:40:00+09', '2026-06-15 16:55:00+09', 27, 34, 'revisit', 6, 16, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1358, 1, '2026-06-15 12:00:00+09', '2026-06-15 12:15:00+09', 56, 66, 'revisit', 7, 14, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1359, 1, '2026-06-15 17:00:00+09', '2026-06-15 17:15:00+09', 59, 69, 'revisit', 8, 8, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1360, 1, '2026-06-15 12:20:00+09', '2026-06-15 12:35:00+09', 3, 4, 'revisit', 1, 1, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1361, 1, '2026-06-15 17:20:00+09', '2026-06-15 17:35:00+09', 3, 4, 'first', 2, 32, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1362, 1, '2026-06-15 12:40:00+09', '2026-06-15 12:55:00+09', 38, 48, 'revisit', 3, 26, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1363, 1, '2026-06-16 09:00:00+09', '2026-06-16 10:30:00+09', 29, 37, 'revisit', 12, 5, 'f', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1364, 1, '2026-06-16 14:00:00+09', '2026-06-16 15:30:00+09', 43, 53, 'revisit', 9, 17, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1365, 1, '2026-06-16 09:20:00+00', '2026-06-16 09:35:00+00', 39, 71, 'revisit', 6, 10, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1366, 1, '2026-06-16 14:20:00+09', '2026-06-16 14:35:00+09', 32, 40, 'revisit', 7, 25, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1367, 1, '2026-06-16 09:40:00+00', '2026-06-16 09:55:00+00', 21, 26, 'revisit', 8, 13, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1368, 1, '2026-06-16 14:40:00+09', '2026-06-16 14:55:00+09', 53, 63, 'revisit', 1, 31, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1369, 1, '2026-06-16 10:00:00+09', '2026-06-16 10:15:00+09', 28, 35, 'revisit', 2, 17, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1370, 1, '2026-06-16 15:00:00+09', '2026-06-16 15:15:00+09', 22, 27, 'revisit', 3, 15, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1371, 1, '2026-06-16 10:20:00+09', '2026-06-16 10:35:00+09', 24, 31, 'revisit', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1372, 1, '2026-06-16 15:20:00+09', '2026-06-16 15:35:00+09', 47, 57, 'revisit', 5, 18, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1373, 1, '2026-06-16 10:40:00+09', '2026-06-16 10:55:00+09', 59, 69, 'revisit', 6, 19, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1374, 1, '2026-06-16 15:40:00+09', '2026-06-16 15:55:00+09', 35, 44, 'first', 7, 2, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1375, 1, '2026-06-16 11:00:00+09', '2026-06-16 11:15:00+09', 1, 1, 'revisit', 8, 15, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1376, 1, '2026-06-16 16:00:00+09', '2026-06-16 16:15:00+09', 15, 17, 'first', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1377, 1, '2026-06-16 11:20:00+09', '2026-06-16 11:35:00+09', 47, 57, 'first', 2, 35, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1378, 1, '2026-06-16 16:20:00+09', '2026-06-16 16:35:00+09', 40, 50, 'revisit', 3, 30, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1379, 1, '2026-06-16 11:40:00+09', '2026-06-16 11:55:00+09', 42, 52, 'revisit', 4, 1, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1380, 1, '2026-06-16 16:40:00+09', '2026-06-16 16:55:00+09', 44, 54, 'first', 5, 23, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1381, 1, '2026-06-16 12:00:00+09', '2026-06-16 12:15:00+09', 50, 60, 'first', 6, 5, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1382, 1, '2026-06-16 17:00:00+09', '2026-06-16 17:15:00+09', 27, 34, 'revisit', 7, 24, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1383, 1, '2026-06-16 12:20:00+09', '2026-06-16 12:35:00+09', 33, 41, 'revisit', 8, 1, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO appointments ("id", "clinic_id", "start_time", "end_time", "owner_id", "pet_id", "visit_type", "reservation_type_id", "doctor_id", "is_designated", "status", "notes", "source", "created_by", "is_staff_delegated", "customer_fields", "reservation_route", "actual_reservation_at", "created_at", "updated_at", "deleted_at", "line_customer_id") VALUES
(1384, 1, '2026-06-16 17:20:00+09', '2026-06-16 17:35:00+09', 44, 54, 'revisit', 1, 2, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1385, 1, '2026-06-17 09:00:00+09', '2026-06-17 10:30:00+09', 3, 4, 'first', 10, 19, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1386, 1, '2026-06-17 14:00:00+09', '2026-06-17 14:15:00+09', 32, 40, 'revisit', 3, 33, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1387, 1, '2026-06-17 09:20:00+00', '2026-06-17 09:35:00+00', 41, 73, 'revisit', 4, 4, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1388, 1, '2026-06-17 14:20:00+09', '2026-06-17 14:35:00+09', 5, 80, 'revisit', 5, 7, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1389, 1, '2026-06-17 09:40:00+00', '2026-06-17 09:55:00+00', 11, 13, 'revisit', 6, 35, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1390, 1, '2026-06-17 14:40:00+09', '2026-06-17 14:55:00+09', 33, 41, 'revisit', 7, 25, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1391, 1, '2026-06-17 10:00:00+09', '2026-06-17 10:15:00+09', 46, 56, 'revisit', 8, 15, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1392, 1, '2026-06-17 15:00:00+09', '2026-06-17 15:15:00+09', 29, 37, 'first', 1, 22, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1393, 1, '2026-06-17 10:20:00+09', '2026-06-17 10:35:00+09', 44, 54, 'revisit', 2, 5, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1394, 1, '2026-06-17 15:20:00+09', '2026-06-17 15:35:00+09', 38, 48, 'revisit', 3, 30, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1395, 1, '2026-06-17 10:40:00+09', '2026-06-17 10:55:00+09', 52, 62, 'revisit', 4, 19, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1396, 1, '2026-06-17 15:40:00+09', '2026-06-17 15:55:00+09', 37, 46, 'revisit', 5, 34, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1397, 1, '2026-06-17 11:00:00+09', '2026-06-17 11:15:00+09', 22, 27, 'first', 6, 17, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1398, 1, '2026-06-17 16:00:00+09', '2026-06-17 16:15:00+09', 18, 21, 'revisit', 7, 31, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1399, 1, '2026-06-17 11:20:00+09', '2026-06-17 11:35:00+09', 6, 8, 'first', 8, 11, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1400, 1, '2026-06-17 16:20:00+09', '2026-06-17 16:35:00+09', 32, 40, 'first', 1, 22, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1401, 1, '2026-06-17 11:40:00+09', '2026-06-17 11:55:00+09', 25, 32, 'revisit', 2, 13, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1402, 1, '2026-06-17 16:40:00+09', '2026-06-17 16:55:00+09', 10, 81, 'revisit', 3, 8, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1403, 1, '2026-06-17 12:00:00+09', '2026-06-17 12:15:00+09', 55, 65, 'revisit', 4, 22, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1404, 1, '2026-06-17 17:00:00+09', '2026-06-17 17:15:00+09', 45, 55, 'revisit', 5, 18, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1405, 1, '2026-06-17 12:20:00+09', '2026-06-17 12:35:00+09', 13, 15, 'revisit', 6, 29, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1406, 1, '2026-06-17 17:20:00+09', '2026-06-17 17:35:00+09', 10, 81, 'first', 7, 28, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1407, 1, '2026-06-17 12:40:00+09', '2026-06-17 12:55:00+09', 24, 31, 'first', 8, 33, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1408, 1, '2026-06-18 09:00:00+09', '2026-06-18 10:30:00+09', 32, 40, 'first', 9, 28, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1409, 1, '2026-06-18 14:00:00+09', '2026-06-18 14:15:00+09', 21, 26, 'revisit', 2, 35, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1410, 1, '2026-06-18 09:20:00+00', '2026-06-18 09:35:00+00', 17, 19, 'revisit', 3, 35, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1411, 1, '2026-06-18 14:20:00+09', '2026-06-18 14:35:00+09', 1, 1, 'revisit', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1412, 1, '2026-06-18 09:40:00+00', '2026-06-18 09:55:00+00', 55, 65, 'revisit', 5, 29, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1413, 1, '2026-06-18 14:40:00+09', '2026-06-18 14:55:00+09', 40, 50, 'revisit', 6, 27, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1414, 1, '2026-06-18 10:00:00+09', '2026-06-18 10:15:00+09', 59, 69, 'revisit', 7, 30, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1415, 1, '2026-06-18 15:00:00+09', '2026-06-18 15:15:00+09', 18, 21, 'first', 8, 29, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1416, 1, '2026-06-18 10:20:00+09', '2026-06-18 10:35:00+09', 13, 15, 'revisit', 1, 5, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1417, 1, '2026-06-18 15:20:00+09', '2026-06-18 15:35:00+09', 12, 14, 'revisit', 2, 13, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1418, 1, '2026-06-18 10:40:00+09', '2026-06-18 10:55:00+09', 55, 65, 'revisit', 3, 4, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1419, 1, '2026-06-18 15:40:00+09', '2026-06-18 15:55:00+09', 55, 65, 'first', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1420, 1, '2026-06-18 11:00:00+09', '2026-06-18 11:15:00+09', 13, 15, 'first', 5, 24, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1421, 1, '2026-06-18 16:00:00+09', '2026-06-18 16:15:00+09', 22, 27, 'revisit', 6, 32, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1422, 1, '2026-06-18 11:20:00+09', '2026-06-18 11:35:00+09', 26, 33, 'first', 7, 34, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1423, 1, '2026-06-18 16:20:00+09', '2026-06-18 16:35:00+09', 56, 66, 'revisit', 8, 7, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1424, 1, '2026-06-18 11:40:00+09', '2026-06-18 11:55:00+09', 9, 11, 'revisit', 1, 14, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1425, 1, '2026-06-18 16:40:00+09', '2026-06-18 16:55:00+09', 6, 8, 'first', 2, 33, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1426, 1, '2026-06-18 12:00:00+09', '2026-06-18 12:15:00+09', 32, 40, 'revisit', 3, 3, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1427, 1, '2026-06-18 17:00:00+09', '2026-06-18 17:15:00+09', 40, 50, 'revisit', 4, 34, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1428, 1, '2026-06-18 12:20:00+09', '2026-06-18 12:35:00+09', 7, 9, 'revisit', 5, 9, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1429, 1, '2026-06-18 17:20:00+09', '2026-06-18 17:35:00+09', 28, 36, 'revisit', 6, 7, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1430, 1, '2026-06-19 09:00:00+09', '2026-06-19 10:30:00+09', 5, 80, 'revisit', 11, 19, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1431, 1, '2026-06-19 14:00:00+09', '2026-06-19 14:15:00+09', 34, 42, 'revisit', 8, 19, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1432, 1, '2026-06-19 09:20:00+00', '2026-06-19 09:35:00+00', 17, 19, 'first', 1, 19, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1433, 1, '2026-06-19 14:20:00+09', '2026-06-19 14:35:00+09', 49, 59, 'revisit', 2, 12, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1434, 1, '2026-06-19 09:40:00+00', '2026-06-19 09:55:00+00', 19, 23, 'revisit', 3, 25, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1435, 1, '2026-06-19 14:40:00+09', '2026-06-19 14:55:00+09', 22, 28, 'revisit', 4, 1, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1436, 1, '2026-06-19 10:00:00+09', '2026-06-19 10:15:00+09', 22, 27, 'revisit', 5, 17, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1437, 1, '2026-06-19 15:00:00+09', '2026-06-19 15:15:00+09', 36, 45, 'revisit', 6, 19, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1438, 1, '2026-06-19 10:20:00+09', '2026-06-19 10:35:00+09', 13, 15, 'first', 7, 22, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1439, 1, '2026-06-19 15:20:00+09', '2026-06-19 15:35:00+09', 22, 28, 'revisit', 8, 25, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1440, 1, '2026-06-19 10:40:00+09', '2026-06-19 10:55:00+09', 1, 1, 'first', 1, 20, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1441, 1, '2026-06-19 15:40:00+09', '2026-06-19 15:55:00+09', 1, 1, 'first', 2, 4, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1442, 1, '2026-06-19 11:00:00+09', '2026-06-19 11:15:00+09', 45, 55, 'revisit', 3, 14, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1443, 1, '2026-06-19 16:00:00+09', '2026-06-19 16:15:00+09', 5, 7, 'revisit', 4, 12, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1444, 1, '2026-06-19 11:20:00+09', '2026-06-19 11:35:00+09', 55, 65, 'revisit', 5, 31, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1445, 1, '2026-06-19 16:20:00+09', '2026-06-19 16:35:00+09', 41, 51, 'first', 6, 4, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1446, 1, '2026-06-19 11:40:00+09', '2026-06-19 11:55:00+09', 26, 33, 'first', 7, 3, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1447, 1, '2026-06-19 16:40:00+09', '2026-06-19 16:55:00+09', 10, 81, 'revisit', 8, 9, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1448, 1, '2026-06-19 12:00:00+09', '2026-06-19 12:15:00+09', 29, 37, 'revisit', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1449, 1, '2026-06-19 17:00:00+09', '2026-06-19 17:15:00+09', 19, 23, 'revisit', 2, 23, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1450, 1, '2026-06-20 09:00:00+09', '2026-06-20 10:30:00+09', 13, 15, 'first', 11, 27, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1451, 1, '2026-06-20 14:00:00+09', '2026-06-20 15:30:00+09', 44, 54, 'revisit', 12, 18, 'f', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1452, 1, '2026-06-20 09:20:00+00', '2026-06-20 09:35:00+00', 6, 8, 'first', 5, 10, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1453, 1, '2026-06-20 14:20:00+09', '2026-06-20 14:35:00+09', 21, 26, 'revisit', 6, 33, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1454, 1, '2026-06-20 09:40:00+00', '2026-06-20 09:55:00+00', 5, 80, 'revisit', 7, 11, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1455, 1, '2026-06-20 14:40:00+09', '2026-06-20 14:55:00+09', 11, 13, 'revisit', 8, 14, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1456, 1, '2026-06-20 10:00:00+09', '2026-06-20 10:15:00+09', 1, 1, 'first', 1, 21, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1457, 1, '2026-06-20 15:00:00+09', '2026-06-20 15:15:00+09', 32, 40, 'first', 2, 8, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1458, 1, '2026-06-20 10:20:00+09', '2026-06-20 10:35:00+09', 47, 57, 'first', 3, 14, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1459, 1, '2026-06-20 15:20:00+09', '2026-06-20 15:35:00+09', 19, 23, 'revisit', 4, 8, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1460, 1, '2026-06-20 10:40:00+09', '2026-06-20 10:55:00+09', 20, 24, 'revisit', 5, 9, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1461, 1, '2026-06-20 15:40:00+09', '2026-06-20 15:55:00+09', 46, 56, 'revisit', 6, 22, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1462, 1, '2026-06-20 11:00:00+09', '2026-06-20 11:15:00+09', 25, 32, 'first', 7, 31, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1463, 1, '2026-06-20 16:00:00+09', '2026-06-20 16:15:00+09', 42, 74, 'revisit', 8, 2, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1464, 1, '2026-06-20 11:20:00+09', '2026-06-20 11:35:00+09', 49, 59, 'revisit', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1465, 1, '2026-06-20 16:20:00+09', '2026-06-20 16:35:00+09', 11, 13, 'first', 2, 20, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1466, 1, '2026-06-20 11:40:00+09', '2026-06-20 11:55:00+09', 21, 26, 'first', 3, 16, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1467, 1, '2026-06-20 16:40:00+09', '2026-06-20 16:55:00+09', 40, 72, 'first', 4, 16, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1468, 1, '2026-06-20 12:00:00+09', '2026-06-20 12:15:00+09', 31, 39, 'revisit', 5, 1, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1469, 1, '2026-06-20 17:00:00+09', '2026-06-20 17:15:00+09', 40, 50, 'first', 6, 17, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1470, 1, '2026-06-20 12:20:00+09', '2026-06-20 12:35:00+09', 30, 38, 'first', 7, 31, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1471, 1, '2026-06-21 09:00:00+09', '2026-06-21 10:30:00+09', 33, 41, 'revisit', 12, 33, 't', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1472, 1, '2026-06-21 14:00:00+09', '2026-06-21 15:30:00+09', 36, 45, 'revisit', 9, 27, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1473, 1, '2026-06-21 09:20:00+00', '2026-06-21 09:35:00+00', 9, 11, 'revisit', 2, 20, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1474, 1, '2026-06-21 14:20:00+09', '2026-06-21 14:35:00+09', 5, 80, 'revisit', 3, 16, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1475, 1, '2026-06-21 09:40:00+00', '2026-06-21 09:55:00+00', 59, 69, 'first', 4, 32, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1476, 1, '2026-06-21 14:40:00+09', '2026-06-21 14:55:00+09', 17, 19, 'revisit', 5, 31, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1477, 1, '2026-06-21 10:00:00+09', '2026-06-21 10:15:00+09', 30, 38, 'first', 6, 24, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1478, 1, '2026-06-21 15:00:00+09', '2026-06-21 15:15:00+09', 40, 72, 'first', 7, 34, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1479, 1, '2026-06-21 10:20:00+09', '2026-06-21 10:35:00+09', 7, 9, 'revisit', 8, 9, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1480, 1, '2026-06-21 15:20:00+09', '2026-06-21 15:35:00+09', 60, 70, 'revisit', 1, 15, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1481, 1, '2026-06-21 10:40:00+09', '2026-06-21 10:55:00+09', 1, 1, 'revisit', 2, 34, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1482, 1, '2026-06-21 15:40:00+09', '2026-06-21 15:55:00+09', 17, 19, 'revisit', 3, 5, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1483, 1, '2026-06-21 11:00:00+09', '2026-06-21 11:15:00+09', 29, 37, 'revisit', 4, 5, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1484, 1, '2026-06-21 16:00:00+09', '2026-06-21 16:15:00+09', 6, 8, 'revisit', 5, 1, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1485, 1, '2026-06-21 11:20:00+09', '2026-06-21 11:35:00+09', 14, 16, 'revisit', 6, 12, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1486, 1, '2026-06-21 16:20:00+09', '2026-06-21 16:35:00+09', 35, 44, 'first', 7, 19, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1487, 1, '2026-06-21 11:40:00+09', '2026-06-21 11:55:00+09', 32, 40, 'revisit', 8, 27, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1488, 1, '2026-06-21 16:40:00+09', '2026-06-21 16:55:00+09', 38, 47, 'first', 1, 9, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1489, 1, '2026-06-21 12:00:00+09', '2026-06-21 12:15:00+09', 19, 23, 'first', 2, 34, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1490, 1, '2026-06-21 17:00:00+09', '2026-06-21 17:15:00+09', 45, 55, 'revisit', 3, 8, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1491, 1, '2026-06-21 12:20:00+09', '2026-06-21 12:35:00+09', 21, 25, 'first', 4, 7, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1492, 1, '2026-06-21 17:20:00+09', '2026-06-21 17:35:00+09', 4, 6, 'first', 5, 10, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1493, 1, '2026-06-21 12:40:00+09', '2026-06-21 12:55:00+09', 19, 23, 'revisit', 6, 35, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1494, 1, '2026-06-22 09:00:00+09', '2026-06-22 10:30:00+09', 13, 15, 'revisit', 11, 19, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1495, 1, '2026-06-22 14:00:00+09', '2026-06-22 15:30:00+09', 34, 42, 'first', 12, 1, 'f', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1496, 1, '2026-06-22 09:20:00+00', '2026-06-22 09:35:00+00', 16, 18, 'revisit', 1, 1, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1497, 1, '2026-06-22 14:20:00+09', '2026-06-22 14:35:00+09', 18, 20, 'revisit', 2, 21, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1498, 1, '2026-06-22 09:40:00+00', '2026-06-22 09:55:00+00', 5, 80, 'revisit', 3, 9, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1499, 1, '2026-06-22 14:40:00+09', '2026-06-22 14:55:00+09', 4, 6, 'first', 4, 20, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1500, 1, '2026-06-22 10:00:00+09', '2026-06-22 10:15:00+09', 48, 58, 'first', 5, 28, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1501, 1, '2026-06-22 15:00:00+09', '2026-06-22 15:15:00+09', 8, 10, 'revisit', 6, 18, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1502, 1, '2026-06-22 10:20:00+09', '2026-06-22 10:35:00+09', 10, 12, 'revisit', 7, 12, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1503, 1, '2026-06-22 15:20:00+09', '2026-06-22 15:35:00+09', 40, 50, 'first', 8, 26, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1504, 1, '2026-06-22 10:40:00+09', '2026-06-22 10:55:00+09', 5, 80, 'first', 1, 21, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1505, 1, '2026-06-22 15:40:00+09', '2026-06-22 15:55:00+09', 42, 74, 'revisit', 2, 21, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1506, 1, '2026-06-22 11:00:00+09', '2026-06-22 11:15:00+09', 39, 71, 'first', 3, 27, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1507, 1, '2026-06-22 16:00:00+09', '2026-06-22 16:15:00+09', 42, 52, 'first', 4, 18, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1508, 1, '2026-06-22 11:20:00+09', '2026-06-22 11:35:00+09', 60, 70, 'first', 5, 32, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1509, 1, '2026-06-22 16:20:00+09', '2026-06-22 16:35:00+09', 35, 44, 'first', 6, 27, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1510, 1, '2026-06-22 11:40:00+09', '2026-06-22 11:55:00+09', 40, 50, 'revisit', 7, 33, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1511, 1, '2026-06-22 16:40:00+09', '2026-06-22 16:55:00+09', 52, 62, 'revisit', 8, 22, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1512, 1, '2026-06-22 12:00:00+09', '2026-06-22 12:15:00+09', 20, 24, 'revisit', 1, 20, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1513, 1, '2026-06-22 17:00:00+09', '2026-06-22 17:15:00+09', 25, 32, 'revisit', 2, 21, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1514, 1, '2026-06-22 12:20:00+09', '2026-06-22 12:35:00+09', 46, 56, 'first', 3, 25, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1515, 1, '2026-06-22 17:20:00+09', '2026-06-22 17:35:00+09', 25, 32, 'first', 4, 8, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1516, 1, '2026-06-23 09:00:00+09', '2026-06-23 10:30:00+09', 3, 5, 'revisit', 9, 26, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1517, 1, '2026-06-23 14:00:00+09', '2026-06-23 14:15:00+09', 34, 42, 'revisit', 6, 23, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1518, 1, '2026-06-23 09:20:00+00', '2026-06-23 09:35:00+00', 21, 26, 'revisit', 7, 33, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1519, 1, '2026-06-23 14:20:00+09', '2026-06-23 14:35:00+09', 48, 58, 'first', 8, 32, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1520, 1, '2026-06-23 09:40:00+00', '2026-06-23 09:55:00+00', 56, 66, 'revisit', 1, 8, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1521, 1, '2026-06-23 14:40:00+09', '2026-06-23 14:55:00+09', 19, 23, 'first', 2, 6, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1522, 1, '2026-06-23 10:00:00+09', '2026-06-23 10:15:00+09', 13, 15, 'revisit', 3, 3, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1523, 1, '2026-06-23 15:00:00+09', '2026-06-23 15:15:00+09', 56, 66, 'first', 4, 18, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1524, 1, '2026-06-23 10:20:00+09', '2026-06-23 10:35:00+09', 26, 33, 'revisit', 5, 9, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1525, 1, '2026-06-23 15:20:00+09', '2026-06-23 15:35:00+09', 14, 16, 'first', 6, 18, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1526, 1, '2026-06-23 10:40:00+09', '2026-06-23 10:55:00+09', 2, 3, 'revisit', 7, 9, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1527, 1, '2026-06-23 15:40:00+09', '2026-06-23 15:55:00+09', 52, 62, 'revisit', 8, 1, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1528, 1, '2026-06-23 11:00:00+09', '2026-06-23 11:15:00+09', 6, 8, 'revisit', 1, 22, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1529, 1, '2026-06-23 16:00:00+09', '2026-06-23 16:15:00+09', 50, 60, 'revisit', 2, 12, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1530, 1, '2026-06-23 11:20:00+09', '2026-06-23 11:35:00+09', 28, 35, 'revisit', 3, 34, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1531, 1, '2026-06-23 16:20:00+09', '2026-06-23 16:35:00+09', 23, 30, 'revisit', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1532, 1, '2026-06-23 11:40:00+09', '2026-06-23 11:55:00+09', 1, 2, 'revisit', 5, 9, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1533, 1, '2026-06-23 16:40:00+09', '2026-06-23 16:55:00+09', 6, 8, 'revisit', 6, 10, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1534, 1, '2026-06-23 12:00:00+09', '2026-06-23 12:15:00+09', 11, 13, 'first', 7, 19, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1535, 1, '2026-06-23 17:00:00+09', '2026-06-23 17:15:00+09', 18, 20, 'first', 8, 5, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1536, 1, '2026-06-24 09:00:00+09', '2026-06-24 10:30:00+09', 43, 75, 'revisit', 9, 6, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1537, 1, '2026-06-24 14:00:00+09', '2026-06-24 15:30:00+09', 23, 30, 'first', 10, 18, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1538, 1, '2026-06-24 09:20:00+00', '2026-06-24 09:35:00+00', 21, 25, 'revisit', 3, 8, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1539, 1, '2026-06-24 14:20:00+09', '2026-06-24 14:35:00+09', 60, 70, 'revisit', 4, 14, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1540, 1, '2026-06-24 09:40:00+00', '2026-06-24 09:55:00+00', 43, 75, 'first', 5, 21, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1541, 1, '2026-06-24 14:40:00+09', '2026-06-24 14:55:00+09', 53, 63, 'revisit', 6, 7, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1542, 1, '2026-06-24 10:00:00+09', '2026-06-24 10:15:00+09', 28, 35, 'revisit', 7, 29, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1543, 1, '2026-06-24 15:00:00+09', '2026-06-24 15:15:00+09', 40, 50, 'revisit', 8, 22, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1544, 1, '2026-06-24 10:20:00+09', '2026-06-24 10:35:00+09', 39, 71, 'revisit', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1545, 1, '2026-06-24 15:20:00+09', '2026-06-24 15:35:00+09', 21, 25, 'revisit', 2, 9, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1546, 1, '2026-06-24 10:40:00+09', '2026-06-24 10:55:00+09', 30, 38, 'revisit', 3, 31, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1547, 1, '2026-06-24 15:40:00+09', '2026-06-24 15:55:00+09', 40, 50, 'revisit', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1548, 1, '2026-06-24 11:00:00+09', '2026-06-24 11:15:00+09', 7, 9, 'revisit', 5, 35, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1549, 1, '2026-06-24 16:00:00+09', '2026-06-24 16:15:00+09', 31, 39, 'revisit', 6, 7, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1550, 1, '2026-06-24 11:20:00+09', '2026-06-24 11:35:00+09', 12, 14, 'first', 7, 35, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1551, 1, '2026-06-24 16:20:00+09', '2026-06-24 16:35:00+09', 34, 42, 'revisit', 8, 18, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1552, 1, '2026-06-24 11:40:00+09', '2026-06-24 11:55:00+09', 57, 67, 'revisit', 1, 22, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1553, 1, '2026-06-24 16:40:00+09', '2026-06-24 16:55:00+09', 25, 32, 'revisit', 2, 10, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1554, 1, '2026-06-24 12:00:00+09', '2026-06-24 12:15:00+09', 42, 74, 'revisit', 3, 30, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1555, 1, '2026-06-24 17:00:00+09', '2026-06-24 17:15:00+09', 53, 63, 'revisit', 4, 34, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1556, 1, '2026-06-25 09:00:00+09', '2026-06-25 10:30:00+09', 7, 9, 'revisit', 9, 10, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1557, 1, '2026-06-25 14:00:00+09', '2026-06-25 15:30:00+09', 20, 24, 'revisit', 10, 26, 't', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1558, 1, '2026-06-25 09:20:00+00', '2026-06-25 09:35:00+00', 27, 34, 'revisit', 7, 15, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1559, 1, '2026-06-25 14:20:00+09', '2026-06-25 14:35:00+09', 40, 50, 'revisit', 8, 11, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1560, 1, '2026-06-25 09:40:00+00', '2026-06-25 09:55:00+00', 3, 4, 'revisit', 1, 5, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1561, 1, '2026-06-25 14:40:00+09', '2026-06-25 14:55:00+09', 54, 64, 'revisit', 2, 12, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1562, 1, '2026-06-25 10:00:00+09', '2026-06-25 10:15:00+09', 20, 24, 'revisit', 3, 4, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1563, 1, '2026-06-25 15:00:00+09', '2026-06-25 15:15:00+09', 8, 10, 'first', 4, 22, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1564, 1, '2026-06-25 10:20:00+09', '2026-06-25 10:35:00+09', 18, 21, 'revisit', 5, 32, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1565, 1, '2026-06-25 15:20:00+09', '2026-06-25 15:35:00+09', 31, 39, 'first', 6, 16, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1566, 1, '2026-06-25 10:40:00+09', '2026-06-25 10:55:00+09', 28, 36, 'revisit', 7, 26, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1567, 1, '2026-06-25 15:40:00+09', '2026-06-25 15:55:00+09', 19, 22, 'first', 8, 6, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1568, 1, '2026-06-25 11:00:00+09', '2026-06-25 11:15:00+09', 60, 70, 'revisit', 1, 29, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1569, 1, '2026-06-25 16:00:00+09', '2026-06-25 16:15:00+09', 18, 20, 'revisit', 2, 14, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1570, 1, '2026-06-25 11:20:00+09', '2026-06-25 11:35:00+09', 17, 19, 'revisit', 3, 28, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1571, 1, '2026-06-25 16:20:00+09', '2026-06-25 16:35:00+09', 28, 35, 'revisit', 4, 27, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1572, 1, '2026-06-25 11:40:00+09', '2026-06-25 11:55:00+09', 2, 3, 'revisit', 5, 10, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1573, 1, '2026-06-25 16:40:00+09', '2026-06-25 16:55:00+09', 25, 32, 'revisit', 6, 31, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1574, 1, '2026-06-25 12:00:00+09', '2026-06-25 12:15:00+09', 12, 14, 'first', 7, 3, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1575, 1, '2026-06-25 17:00:00+09', '2026-06-25 17:15:00+09', 56, 66, 'revisit', 8, 21, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1576, 1, '2026-06-26 09:00:00+09', '2026-06-26 10:30:00+09', 45, 55, 'revisit', 9, 19, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1577, 1, '2026-06-26 14:00:00+09', '2026-06-26 15:30:00+09', 32, 40, 'revisit', 10, 4, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1578, 1, '2026-06-26 09:20:00+00', '2026-06-26 09:35:00+00', 18, 21, 'revisit', 3, 32, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1579, 1, '2026-06-26 14:20:00+09', '2026-06-26 14:35:00+09', 56, 66, 'revisit', 4, 3, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1580, 1, '2026-06-26 09:40:00+00', '2026-06-26 09:55:00+00', 29, 37, 'first', 5, 32, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1581, 1, '2026-06-26 14:40:00+09', '2026-06-26 14:55:00+09', 3, 4, 'revisit', 6, 18, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1582, 1, '2026-06-26 10:00:00+09', '2026-06-26 10:15:00+09', 19, 23, 'revisit', 7, 17, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1583, 1, '2026-06-26 15:00:00+09', '2026-06-26 15:15:00+09', 9, 11, 'first', 8, 12, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1584, 1, '2026-06-26 10:20:00+09', '2026-06-26 10:35:00+09', 1, 2, 'revisit', 1, 16, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1585, 1, '2026-06-26 15:20:00+09', '2026-06-26 15:35:00+09', 39, 71, 'first', 2, 15, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1586, 1, '2026-06-26 10:40:00+09', '2026-06-26 10:55:00+09', 55, 65, 'revisit', 3, 27, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1587, 1, '2026-06-26 15:40:00+09', '2026-06-26 15:55:00+09', 29, 37, 'revisit', 4, 19, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1588, 1, '2026-06-26 11:00:00+09', '2026-06-26 11:15:00+09', 59, 69, 'revisit', 5, 35, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1589, 1, '2026-06-26 16:00:00+09', '2026-06-26 16:15:00+09', 20, 24, 'revisit', 6, 28, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1590, 1, '2026-06-26 11:20:00+09', '2026-06-26 11:35:00+09', 3, 5, 'revisit', 7, 22, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1591, 1, '2026-06-26 16:20:00+09', '2026-06-26 16:35:00+09', 41, 73, 'first', 8, 32, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1592, 1, '2026-06-26 11:40:00+09', '2026-06-26 11:55:00+09', 51, 61, 'first', 1, 32, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1593, 1, '2026-06-26 16:40:00+09', '2026-06-26 16:55:00+09', 44, 54, 'revisit', 2, 13, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1594, 1, '2026-06-26 12:00:00+09', '2026-06-26 12:15:00+09', 38, 48, 'revisit', 3, 5, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1595, 1, '2026-06-26 17:00:00+09', '2026-06-26 17:15:00+09', 36, 45, 'revisit', 4, 1, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1596, 1, '2026-06-26 12:20:00+09', '2026-06-26 12:35:00+09', 3, 4, 'first', 5, 16, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1597, 1, '2026-06-26 17:20:00+09', '2026-06-26 17:35:00+09', 19, 23, 'revisit', 6, 17, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1598, 1, '2026-06-26 12:40:00+09', '2026-06-26 12:55:00+09', 41, 73, 'revisit', 7, 23, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1599, 1, '2026-06-27 09:00:00+09', '2026-06-27 10:30:00+09', 25, 32, 'revisit', 12, 23, 't', 'confirmed', 'クイックシャンプー', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1600, 1, '2026-06-27 14:00:00+09', '2026-06-27 14:15:00+09', 7, 9, 'revisit', 1, 14, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1601, 1, '2026-06-27 09:20:00+00', '2026-06-27 09:35:00+00', 59, 69, 'revisit', 2, 6, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1602, 1, '2026-06-27 14:20:00+09', '2026-06-27 14:35:00+09', 60, 70, 'revisit', 3, 3, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1603, 1, '2026-06-27 09:40:00+00', '2026-06-27 09:55:00+00', 39, 71, 'revisit', 4, 2, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1604, 1, '2026-06-27 14:40:00+09', '2026-06-27 14:55:00+09', 59, 69, 'revisit', 5, 30, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1605, 1, '2026-06-27 10:00:00+09', '2026-06-27 10:15:00+09', 10, 81, 'revisit', 6, 15, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1606, 1, '2026-06-27 15:00:00+09', '2026-06-27 15:15:00+09', 10, 81, 'revisit', 7, 1, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1607, 1, '2026-06-27 10:20:00+09', '2026-06-27 10:35:00+09', 24, 31, 'first', 8, 24, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1608, 1, '2026-06-27 15:20:00+09', '2026-06-27 15:35:00+09', 10, 81, 'revisit', 1, 20, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1609, 1, '2026-06-27 10:40:00+09', '2026-06-27 10:55:00+09', 38, 47, 'revisit', 2, 12, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1610, 1, '2026-06-27 15:40:00+09', '2026-06-27 15:55:00+09', 26, 33, 'revisit', 3, 16, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1611, 1, '2026-06-27 11:00:00+09', '2026-06-27 11:15:00+09', 50, 60, 'revisit', 4, 29, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1612, 1, '2026-06-27 16:00:00+09', '2026-06-27 16:15:00+09', 19, 23, 'revisit', 5, 23, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1613, 1, '2026-06-27 11:20:00+09', '2026-06-27 11:35:00+09', 10, 81, 'revisit', 6, 5, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1614, 1, '2026-06-27 16:20:00+09', '2026-06-27 16:35:00+09', 25, 32, 'first', 7, 1, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1615, 1, '2026-06-27 11:40:00+09', '2026-06-27 11:55:00+09', 22, 28, 'first', 8, 34, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1616, 1, '2026-06-27 16:40:00+09', '2026-06-27 16:55:00+09', 3, 4, 'revisit', 1, 13, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1617, 1, '2026-06-27 12:00:00+09', '2026-06-27 12:15:00+09', 20, 24, 'revisit', 2, 26, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1618, 1, '2026-06-27 17:00:00+09', '2026-06-27 17:15:00+09', 28, 35, 'revisit', 3, 21, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1619, 1, '2026-06-27 12:20:00+09', '2026-06-27 12:35:00+09', 12, 14, 'revisit', 4, 32, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1620, 1, '2026-06-28 09:00:00+09', '2026-06-28 10:30:00+09', 55, 65, 'revisit', 9, 24, 'f', 'confirmed', 'トリミングコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1621, 1, '2026-06-28 14:00:00+09', '2026-06-28 14:15:00+09', 40, 50, 'revisit', 6, 12, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1622, 1, '2026-06-28 09:20:00+00', '2026-06-28 09:35:00+00', 22, 28, 'revisit', 7, 30, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1623, 1, '2026-06-28 14:20:00+09', '2026-06-28 14:35:00+09', 19, 23, 'revisit', 8, 5, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1624, 1, '2026-06-28 09:40:00+00', '2026-06-28 09:55:00+00', 21, 26, 'revisit', 1, 12, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1625, 1, '2026-06-28 14:40:00+09', '2026-06-28 14:55:00+09', 18, 21, 'revisit', 2, 18, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1626, 1, '2026-06-28 10:00:00+09', '2026-06-28 10:15:00+09', 34, 42, 'revisit', 3, 28, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1627, 1, '2026-06-28 15:00:00+09', '2026-06-28 15:15:00+09', 28, 36, 'first', 4, 33, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1628, 1, '2026-06-28 10:20:00+09', '2026-06-28 10:35:00+09', 40, 50, 'revisit', 5, 18, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1629, 1, '2026-06-28 15:20:00+09', '2026-06-28 15:35:00+09', 1, 1, 'first', 6, 2, 't', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1630, 1, '2026-06-28 10:40:00+09', '2026-06-28 10:55:00+09', 18, 20, 'revisit', 7, 28, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1631, 1, '2026-06-28 15:40:00+09', '2026-06-28 15:55:00+09', 12, 14, 'revisit', 8, 34, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1632, 1, '2026-06-28 11:00:00+09', '2026-06-28 11:15:00+09', 39, 49, 'first', 1, 2, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1633, 1, '2026-06-28 16:00:00+09', '2026-06-28 16:15:00+09', 12, 14, 'revisit', 2, 35, 't', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1634, 1, '2026-06-28 11:20:00+09', '2026-06-28 11:35:00+09', 1, 1, 'revisit', 3, 35, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1635, 1, '2026-06-28 16:20:00+09', '2026-06-28 16:35:00+09', 59, 69, 'revisit', 4, 29, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1636, 1, '2026-06-28 11:40:00+09', '2026-06-28 11:55:00+09', 10, 12, 'revisit', 5, 1, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1637, 1, '2026-06-28 16:40:00+09', '2026-06-28 16:55:00+09', 58, 68, 'first', 6, 30, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1638, 1, '2026-06-28 12:00:00+09', '2026-06-28 12:15:00+09', 44, 54, 'revisit', 7, 4, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1639, 1, '2026-06-28 17:00:00+09', '2026-06-28 17:15:00+09', 40, 72, 'revisit', 8, 15, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1640, 1, '2026-06-28 12:20:00+09', '2026-06-28 12:35:00+09', 16, 18, 'revisit', 1, 28, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1641, 1, '2026-06-29 09:00:00+09', '2026-06-29 10:30:00+09', 42, 74, 'revisit', 10, 4, 'f', 'confirmed', '部分カットコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1642, 1, '2026-06-29 14:00:00+09', '2026-06-29 15:30:00+09', 41, 51, 'revisit', 11, 20, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1643, 1, '2026-06-29 09:20:00+00', '2026-06-29 09:35:00+00', 26, 33, 'revisit', 4, 33, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1644, 1, '2026-06-29 14:20:00+09', '2026-06-29 14:35:00+09', 11, 13, 'revisit', 5, 18, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1645, 1, '2026-06-29 09:40:00+00', '2026-06-29 09:55:00+00', 29, 37, 'revisit', 6, 25, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1646, 1, '2026-06-29 14:40:00+09', '2026-06-29 14:55:00+09', 19, 22, 'revisit', 7, 12, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1647, 1, '2026-06-29 10:00:00+09', '2026-06-29 10:15:00+09', 38, 48, 'revisit', 8, 12, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1648, 1, '2026-06-29 15:00:00+09', '2026-06-29 15:15:00+09', 59, 69, 'revisit', 1, 22, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1649, 1, '2026-06-29 10:20:00+09', '2026-06-29 10:35:00+09', 45, 55, 'revisit', 2, 11, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1650, 1, '2026-06-29 15:20:00+09', '2026-06-29 15:35:00+09', 4, 6, 'first', 3, 34, 't', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1651, 1, '2026-06-29 10:40:00+09', '2026-06-29 10:55:00+09', 31, 39, 'revisit', 4, 15, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1652, 1, '2026-06-29 15:40:00+09', '2026-06-29 15:55:00+09', 8, 10, 'revisit', 5, 35, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1653, 1, '2026-06-29 11:00:00+09', '2026-06-29 11:15:00+09', 29, 37, 'first', 6, 7, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1654, 1, '2026-06-29 16:00:00+09', '2026-06-29 16:15:00+09', 38, 47, 'first', 7, 30, 't', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1655, 1, '2026-06-29 11:20:00+09', '2026-06-29 11:35:00+09', 60, 70, 'revisit', 8, 33, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1656, 1, '2026-06-29 16:20:00+09', '2026-06-29 16:35:00+09', 4, 6, 'revisit', 1, 3, 't', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1657, 1, '2026-06-29 11:40:00+09', '2026-06-29 11:55:00+09', 48, 58, 'revisit', 2, 16, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1658, 1, '2026-06-29 16:40:00+09', '2026-06-29 16:55:00+09', 16, 18, 'revisit', 3, 11, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1659, 1, '2026-06-29 12:00:00+09', '2026-06-29 12:15:00+09', 12, 14, 'first', 4, 33, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1660, 1, '2026-06-29 17:00:00+09', '2026-06-29 17:15:00+09', 23, 30, 'revisit', 5, 5, 't', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1661, 1, '2026-06-29 12:20:00+09', '2026-06-29 12:35:00+09', 57, 67, 'revisit', 6, 32, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1662, 1, '2026-06-30 09:00:00+09', '2026-06-30 10:30:00+09', 35, 44, 'revisit', 11, 14, 'f', 'confirmed', 'シャンプーコース', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1663, 1, '2026-06-30 14:00:00+09', '2026-06-30 14:15:00+09', 18, 20, 'first', 8, 3, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1664, 1, '2026-06-30 09:20:00+00', '2026-06-30 09:35:00+00', 6, 8, 'first', 1, 26, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1665, 1, '2026-06-30 14:20:00+09', '2026-06-30 14:35:00+09', 38, 48, 'first', 2, 11, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1666, 1, '2026-06-30 09:40:00+00', '2026-06-30 09:55:00+00', 42, 52, 'first', 3, 10, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1667, 1, '2026-06-30 14:40:00+09', '2026-06-30 14:55:00+09', 46, 56, 'revisit', 4, 10, 't', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1668, 1, '2026-06-30 10:00:00+09', '2026-06-30 10:15:00+09', 58, 68, 'revisit', 5, 35, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1669, 1, '2026-06-30 15:00:00+09', '2026-06-30 15:15:00+09', 30, 38, 'revisit', 6, 25, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1670, 1, '2026-06-30 10:20:00+09', '2026-06-30 10:35:00+09', 7, 9, 'revisit', 7, 35, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1671, 1, '2026-06-30 15:20:00+09', '2026-06-30 15:35:00+09', 45, 55, 'revisit', 8, 20, 't', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1672, 1, '2026-06-30 10:40:00+09', '2026-06-30 10:55:00+09', 27, 34, 'revisit', 1, 12, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1673, 1, '2026-06-30 15:40:00+09', '2026-06-30 15:55:00+09', 42, 52, 'revisit', 2, 11, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1674, 1, '2026-06-30 11:00:00+09', '2026-06-30 11:15:00+09', 49, 59, 'revisit', 3, 13, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1675, 1, '2026-06-30 16:00:00+09', '2026-06-30 16:15:00+09', 42, 74, 'revisit', 4, 6, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1676, 1, '2026-06-30 11:20:00+09', '2026-06-30 11:35:00+09', 3, 4, 'revisit', 5, 29, 'f', 'confirmed', '狂犬病予防接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1677, 1, '2026-06-30 16:20:00+09', '2026-06-30 16:35:00+09', 9, 11, 'revisit', 6, 6, 'f', 'confirmed', 'フィラリア予防', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1678, 1, '2026-06-30 11:40:00+09', '2026-06-30 11:55:00+09', 35, 44, 'revisit', 7, 30, 'f', 'confirmed', '健康診断', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1679, 1, '2026-06-30 16:40:00+09', '2026-06-30 16:55:00+09', 18, 21, 'revisit', 8, 10, 'f', 'confirmed', '健康診断結果報告', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1680, 1, '2026-06-30 12:00:00+09', '2026-06-30 12:15:00+09', 39, 49, 'revisit', 1, 17, 'f', 'confirmed', '一般診察：体調確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1681, 1, '2026-06-30 17:00:00+09', '2026-06-30 17:15:00+09', 24, 31, 'revisit', 2, 17, 'f', 'confirmed', '再診：経過確認', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1682, 1, '2026-06-30 12:20:00+09', '2026-06-30 12:35:00+09', 31, 39, 'revisit', 3, 17, 'f', 'confirmed', '混合ワクチン接種', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL),
(1683, 1, '2026-06-30 17:20:00+09', '2026-06-30 17:35:00+09', 26, 33, 'revisit', 4, 29, 'f', 'confirmed', 'お手入れ：爪切り・耳掃除', 'manual', NULL, 'f', '{}', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL)
ON CONFLICT DO NOTHING;


-- -----------------------------------------------------------------------------
-- appointment_trimming_details
-- -----------------------------------------------------------------------------
INSERT INTO appointment_trimming_details ("id", "clinic_id", "appointment_id", "course_id", "style_request", "body_weight", "bw_unit", "body_temperature", "used_shampoo", "used_ribbon", "remarks", "style_image", "completed_image", "created_at", "updated_at") VALUES
(1, 1, 101, 5, 'サマーカット希望', 26.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 102, 4, 'ふんわりカット', 15.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 103, 1, '毛玉カット', 4.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 1, 104, 1, 'シャンプーコース', 3800.00, 'g', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 1, 105, 4, '全体カット', 12.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 1, 106, 2, '爪切り・ブラッシング', 8.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 1, 107, 1, 'シャンプー', 5.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 1, 108, 3, 'トリミング', 3800.00, 'g', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 1, 227, 5, 'サマーカット', 3.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 1, 228, 1, 'シャンプーコース', 4.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(11, 1, 229, 3, 'トリミング', 6.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(12, 1, 246, 4, '全体カット', 8.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(13, 1, 247, 1, 'シャンプー', 5.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(14, 1, 248, 3, 'トリミング', 4.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1000, 1, 1000, 2, 'サマーカット希望', 7.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1001, 1, 1020, 2, 'サマーカット希望', 13.60, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1002, 1, 1041, 5, 'サマーカット希望', 5.70, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1003, 1, 1062, 2, 'サマーカット希望', 13.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1004, 1, 1085, 2, 'サマーカット希望', 15.90, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1005, 1, 1105, 1, 'サマーカット希望', 7.40, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1006, 1, 1106, 3, 'サマーカット希望', 14.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1007, 1, 1127, 3, 'サマーカット希望', 5.10, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1008, 1, 1148, 1, 'サマーカット希望', 12.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1009, 1, 1149, 5, 'サマーカット希望', 5.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1010, 1, 1169, 2, 'サマーカット希望', 10.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1011, 1, 1170, 3, 'サマーカット希望', 5.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1012, 1, 1189, 2, 'サマーカット希望', 16.10, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1013, 1, 1212, 1, 'サマーカット希望', 13.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1014, 1, 1233, 1, 'サマーカット希望', 10.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1015, 1, 1256, 4, 'サマーカット希望', 5.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1016, 1, 1257, 1, 'サマーカット希望', 4.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1017, 1, 1277, 5, 'サマーカット希望', 14.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1018, 1, 1298, 3, 'サマーカット希望', 14.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1019, 1, 1299, 3, 'サマーカット希望', 2.60, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1020, 1, 1320, 2, 'サマーカット希望', 10.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1021, 1, 1340, 4, 'サマーカット希望', 12.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1022, 1, 1341, 4, 'サマーカット希望', 9.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1023, 1, 1363, 4, 'サマーカット希望', 13.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1024, 1, 1364, 5, 'サマーカット希望', 11.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1025, 1, 1385, 5, 'サマーカット希望', 13.90, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1026, 1, 1408, 1, 'サマーカット希望', 9.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1027, 1, 1430, 1, 'サマーカット希望', 6.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1028, 1, 1450, 3, 'サマーカット希望', 3.90, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1029, 1, 1451, 4, 'サマーカット希望', 8.10, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1030, 1, 1471, 2, 'サマーカット希望', 10.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1031, 1, 1472, 3, 'サマーカット希望', 10.10, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1032, 1, 1494, 5, 'サマーカット希望', 13.40, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1033, 1, 1495, 3, 'サマーカット希望', 4.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1034, 1, 1516, 4, 'サマーカット希望', 15.90, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1035, 1, 1536, 1, 'サマーカット希望', 10.90, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1036, 1, 1537, 4, 'サマーカット希望', 7.20, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1037, 1, 1556, 3, 'サマーカット希望', 11.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1038, 1, 1557, 1, 'サマーカット希望', 3.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1039, 1, 1576, 5, 'サマーカット希望', 3.00, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1040, 1, 1577, 1, 'サマーカット希望', 13.40, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1041, 1, 1599, 3, 'サマーカット希望', 4.80, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1042, 1, 1620, 4, 'サマーカット希望', 10.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1043, 1, 1641, 1, 'サマーカット希望', 4.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1044, 1, 1642, 2, 'サマーカット希望', 6.30, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(1045, 1, 1662, 4, 'サマーカット希望', 3.50, 'Kg', NULL, '', '', '', '', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- appointment_trimming_options
-- -----------------------------------------------------------------------------
INSERT INTO appointment_trimming_options ("id", "appointment_id", "option_id", "sort_order", "created_at") VALUES
(1, 227, 1, 1, '2026-05-31 04:33:17.574774+00'),
(2, 227, 2, 2, '2026-05-31 04:33:17.574774+00'),
(3, 228, 1, 1, '2026-05-31 04:33:17.574774+00'),
(4, 248, 1, 1, '2026-05-31 04:33:17.574774+00'),
(5, 248, 2, 2, '2026-05-31 04:33:17.574774+00'),
(6, 248, 4, 3, '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- hospitalizations
-- -----------------------------------------------------------------------------
INSERT INTO hospitalizations ("id", "clinic_id", "owner_id", "pet_id", "hospitalization_type", "start_date", "end_date", "status", "cage_id", "doctor_id", "memo", "owner_request", "staff_notes", "insurance_company_name", "insurance_number", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 3, 5, 'hospitalization', '2026-05-20', '2026-05-24', 'admitted', 5, 1, '急性胃腸炎による脱水治療。点滴管理中。', '食事のアレルギーに注意してほしい（鶏肉不可）', '5/20入院開始。静脈点滴開始。5/21嘔吐1回。5/22状態改善傾向。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 6, 8, 'hospitalization', '2026-05-07', '2026-05-10', 'discharged', 4, 1, '外耳炎重症化に伴う入院治療。', '怖がりなので優しく接してほしい', '耳道洗浄を毎日実施. 5/10退院時, 症状改善。点耳薬処方。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 17, 19, 'hospitalization', '2026-04-22', '2026-05-02', 'discharged', NULL, 1, '骨折治療による入院。手術後経過観察。', '', '4/22手術実施。4/27抜糸。5/2退院。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 4, 6, 'hotel', '2026-05-25', '2026-05-28', 'reserved', NULL, 1, '旅行中のホテル預かり。', 'フードはロイヤルカナンのみ', '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 1, 1, 'hospitalization', '2026-05-30', '2026-06-04', 'reserved', NULL, 1, '膝蓋骨脱臼手術予定。術前検査済み。', '怖がりなので静かな環境を希望', '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, 9, 11, 'hospitalization', '2026-05-15', '2026-05-22', 'admitted', 1, 2, '慢性腎臓病の集中治療。点滴管理中。', 'ペルシャ猫のため温度管理に注意', '5/15入院。毎日皮下補液実施。5/22現在状態安定。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, 3, 4, 'hospitalization', '2026-03-15', '2026-03-18', 'discharged', NULL, 1, '急性胃腸炎による脱水治療。', 'チキンアレルギーあり', '3/15入院。点滴開始。3/18状態改善し退院。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 1, 40, 50, 'hospitalization', '2026-05-21', '2026-05-25', 'admitted', 2, 1, '異物摂取による開腹手術後。', '安静にさせてほしい', '5/21手術。5/22現在自力採食なし。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 1, 42, 52, 'hotel', '2026-05-21', '2026-05-23', 'admitted', NULL, 2, '引越しに伴う一時預かり。', '', '5/21入庫。元気食欲あり。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, 44, 54, 'hospitalization', '2026-05-22', '2026-05-24', 'admitted', 3, 2, '喘息発作の集中管理。酸素室。', 'パニックになりやすい', '5/22朝、呼吸困難で入院。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, 46, 56, 'hospitalization', '2026-05-10', '2026-05-22', 'admitted', 6, 1, '糖尿病のインスリン用量調整。', '性格は穏やか', '5/22退院予定。血糖値安定。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 1, 48, 58, 'hotel', '2026-05-22', '2026-05-24', 'reserved', NULL, 1, '法事のため預かり。', '', '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 1, 50, 60, 'hospitalization', '2026-05-18', '2026-05-23', 'admitted', 7, 2, '肝不全の集中治療。', '薬が苦手', '5/22点滴継続中。黄疸軽減。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 1, 52, 62, 'hospitalization', '2026-05-22', '2026-05-25', 'admitted', 8, 1, '咬傷の感染症治療。', '他犬が苦手', '5/22排膿処置実施。', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, 54, 64, 'hotel', '2026-05-26', '2026-05-28', 'reserved', NULL, 1, '出張のため預かり。', '', '', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1000, 1, 31, 39, 'hotel', '2026-05-30', '2026-06-01', 'admitted', NULL, 24, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1001, 1, 26, 33, 'hotel', '2026-05-30', '2026-06-02', 'admitted', NULL, 21, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1002, 1, 55, 65, 'hotel', '2026-05-31', '2026-06-03', 'admitted', NULL, 24, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1003, 1, 6, 8, 'hotel', '2026-05-31', '2026-06-02', 'admitted', NULL, 25, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1004, 1, 11, 13, 'hotel', '2026-06-01', '2026-06-03', 'reserved', NULL, 11, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1005, 1, 38, 47, 'hotel', '2026-06-01', '2026-06-03', 'reserved', NULL, 6, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1006, 1, 30, 38, 'hotel', '2026-06-02', '2026-06-05', 'reserved', NULL, 2, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1007, 1, 56, 66, 'hotel', '2026-06-03', '2026-06-06', 'reserved', NULL, 27, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1008, 1, 3, 5, 'hotel', '2026-06-04', '2026-06-06', 'reserved', NULL, 29, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1009, 1, 10, 12, 'hotel', '2026-06-04', '2026-06-07', 'reserved', NULL, 31, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1010, 1, 15, 17, 'hotel', '2026-06-05', '2026-06-07', 'reserved', NULL, 15, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1011, 1, 1, 1, 'hotel', '2026-06-05', '2026-06-08', 'reserved', NULL, 25, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1012, 1, 17, 19, 'hotel', '2026-06-06', '2026-06-08', 'reserved', NULL, 31, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1013, 1, 4, 6, 'hotel', '2026-06-07', '2026-06-09', 'reserved', NULL, 29, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1014, 1, 45, 55, 'hotel', '2026-06-08', '2026-06-11', 'reserved', NULL, 28, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1015, 1, 18, 20, 'hotel', '2026-06-09', '2026-06-11', 'reserved', NULL, 16, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1016, 1, 15, 17, 'hotel', '2026-06-10', '2026-06-13', 'reserved', NULL, 2, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1017, 1, 15, 17, 'hotel', '2026-06-10', '2026-06-13', 'reserved', NULL, 12, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1018, 1, 6, 8, 'hotel', '2026-06-11', '2026-06-13', 'reserved', NULL, 11, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1019, 1, 38, 47, 'hotel', '2026-06-11', '2026-06-14', 'reserved', NULL, 28, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1020, 1, 5, 80, 'hotel', '2026-06-12', '2026-06-15', 'reserved', NULL, 7, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1021, 1, 44, 54, 'hotel', '2026-06-12', '2026-06-14', 'reserved', NULL, 10, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1022, 1, 37, 46, 'hotel', '2026-06-13', '2026-06-16', 'reserved', NULL, 19, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1023, 1, 34, 42, 'hotel', '2026-06-14', '2026-06-16', 'reserved', NULL, 4, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1024, 1, 1, 2, 'hotel', '2026-06-14', '2026-06-16', 'reserved', NULL, 4, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1025, 1, 18, 21, 'hotel', '2026-06-15', '2026-06-18', 'reserved', NULL, 24, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1026, 1, 46, 56, 'hotel', '2026-06-16', '2026-06-19', 'reserved', NULL, 17, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1027, 1, 42, 52, 'hotel', '2026-06-17', '2026-06-19', 'reserved', NULL, 26, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1028, 1, 18, 20, 'hotel', '2026-06-17', '2026-06-19', 'reserved', NULL, 13, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1029, 1, 4, 6, 'hotel', '2026-06-18', '2026-06-20', 'reserved', NULL, 21, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1030, 1, 41, 73, 'hotel', '2026-06-19', '2026-06-22', 'reserved', NULL, 16, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1031, 1, 42, 74, 'hotel', '2026-06-20', '2026-06-23', 'reserved', NULL, 21, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1032, 1, 31, 39, 'hotel', '2026-06-20', '2026-06-22', 'reserved', NULL, 8, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1033, 1, 3, 5, 'hotel', '2026-06-21', '2026-06-24', 'reserved', NULL, 1, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1034, 1, 44, 54, 'hotel', '2026-06-22', '2026-06-24', 'reserved', NULL, 20, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1035, 1, 51, 61, 'hotel', '2026-06-22', '2026-06-24', 'reserved', NULL, 10, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1036, 1, 22, 27, 'hotel', '2026-06-23', '2026-06-26', 'reserved', NULL, 3, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1037, 1, 28, 36, 'hotel', '2026-06-23', '2026-06-26', 'reserved', NULL, 27, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1038, 1, 39, 49, 'hotel', '2026-06-24', '2026-06-27', 'reserved', NULL, 25, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1039, 1, 28, 36, 'hotel', '2026-06-25', '2026-06-28', 'reserved', NULL, 16, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1040, 1, 18, 21, 'hotel', '2026-06-25', '2026-06-27', 'reserved', NULL, 26, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1041, 1, 17, 19, 'hotel', '2026-06-26', '2026-06-28', 'reserved', NULL, 5, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1042, 1, 25, 32, 'hotel', '2026-06-26', '2026-06-28', 'reserved', NULL, 4, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1043, 1, 51, 61, 'hotel', '2026-06-27', '2026-06-30', 'reserved', NULL, 24, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1044, 1, 10, 12, 'hotel', '2026-06-27', '2026-06-30', 'reserved', NULL, 23, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1045, 1, 19, 23, 'hotel', '2026-06-28', '2026-06-30', 'reserved', NULL, 8, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1046, 1, 38, 48, 'hotel', '2026-06-29', '2026-07-02', 'reserved', NULL, 35, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1047, 1, 5, 7, 'hotel', '2026-06-29', '2026-07-02', 'reserved', NULL, 22, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(1048, 1, 18, 21, 'hotel', '2026-06-30', '2026-07-02', 'reserved', NULL, 32, 'ペットホテル預かりデモ', 'ご飯持ち込みあり', '自動生成データ', NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- daily_records
-- -----------------------------------------------------------------------------
INSERT INTO daily_records ("id", "hospitalization_id", "clinic_id", "date", "created_at", "updated_at") VALUES
(1, 1, 1, '2026-05-20', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 8, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 8, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 10, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 9, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 9, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 11, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 11, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(11, 13, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(12, 13, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(13, 14, 1, '2026-05-22', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(14, 6, 1, '2026-05-20', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(15, 6, 1, '2026-05-21', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- medical_records
-- -----------------------------------------------------------------------------
INSERT INTO medical_records ("id", "clinic_id", "record_no", "date", "owner_id", "pet_id", "doctor_id", "appointment_id", "status", "version", "entered_by", "next_visit_recommended_date", "recommendation_reason", "visit_type", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 'R-2025-001', '2025-12-20', 1, 1, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 'R-2025-002', '2026-02-24', 1, 1, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 'R-2026-001', '2026-04-01', 1, 1, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 'R-2025-003', '2026-01-15', 1, 2, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 'R-2025-004', '2025-11-25', 2, 3, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, 'R-2026-002', '2026-03-18', 2, 3, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, 'R-2025-005', '2025-11-01', 3, 4, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 1, 'R-2025-006', '2025-12-28', 4, 6, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 1, 'R-2025-007', '2025-10-09', 5, 7, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, 'R-2026-003', '2026-03-27', 6, 8, 1, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, 'R-2025-008', '2026-02-10', 7, 9, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 1, 'R-2025-009', '2026-01-30', 8, 10, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 1, 'R-2026-004', '2026-04-22', 9, 11, 2, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 1, 'R-2025-010', '2025-08-25', 10, 12, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, 'R-2026-005', '2026-03-18', 11, 13, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 1, 'R-2025-011', '2025-11-18', 12, 14, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 1, 'R-2026-006', '2026-05-10', 13, 15, 2, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 1, 'R-2025-012', '2025-10-30', 14, 16, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 1, 'R-2026-007', '2026-03-15', 15, 17, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 1, 'R-2026-008', '2026-03-18', 16, 18, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 2, 'C-2025-001', '2026-01-20', 23, 29, 16, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(22, 2, 'C-2025-002', '2025-12-30', 24, 31, 17, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 2, 'C-2025-003', '2026-02-24', 25, 32, 16, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(24, 2, 'C-2026-001', '2026-03-20', 26, 33, 18, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(25, 2, 'C-2025-004', '2026-02-10', 27, 34, 16, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(26, 2, 'C-2025-005', '2025-11-15', 28, 35, 17, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(27, 2, 'C-2026-002', '2026-03-27', 29, 37, 16, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(28, 2, 'C-2026-003', '2026-05-02', 30, 38, 18, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(29, 3, 'S-2025-001', '2025-12-15', 31, 39, 26, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(30, 3, 'S-2025-002', '2026-02-27', 32, 40, 27, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(31, 3, 'S-2026-001', '2026-03-20', 33, 41, 26, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(32, 3, 'S-2026-002', '2026-04-27', 34, 42, 27, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(33, 3, 'S-2025-003', '2025-11-30', 35, 44, 26, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(34, 3, 'S-2026-003', '2026-05-11', 36, 45, 26, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(35, 3, 'S-2025-004', '2026-02-04', 37, 46, 27, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(36, 3, 'S-2026-004', '2026-04-22', 38, 47, 26, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(37, 1, 'MR-20260507-1-ez8XhN', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-07 04:10:02.867522+00', '2026-05-07 04:10:02.867522+00', NULL),
(38, 1, 'MR-20260507-1-uUJgtF', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-07 05:44:37.876649+00', '2026-05-07 05:44:37.876649+00', NULL),
(39, 1, 'MR-20260507-1-QHeBtx', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-07 05:44:51.894538+00', '2026-05-07 05:44:51.894538+00', NULL),
(40, 1, 'MR-20260507-1-A0cYzP', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-07 05:45:37.458156+00', '2026-05-07 05:45:37.458156+00', NULL),
(41, 1, 'MR-20260507-1-iznyqy', '2026-05-07', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-07 05:45:50.493979+00', '2026-05-07 05:45:50.493979+00', NULL),
(42, 1, 'MR-20260509-1-zRC1Qb', '2026-05-09', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-09 08:48:12.285328+00', '2026-05-09 08:48:12.285328+00', NULL),
(43, 1, 'MR-20260509-1-bx40GM', '2026-05-09', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-09 08:48:44.370087+00', '2026-05-09 08:48:44.370087+00', NULL),
(44, 1, 'MR-20260509-1-mytfsU', '2026-05-09', 1, 2, 4, 111, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-09 08:51:27.788135+00', '2026-05-09 08:51:27.788135+00', NULL),
(45, 1, 'MR-20260509-1-bGpxAZ', '2026-05-09', 1, 2, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-09 08:51:37.186341+00', '2026-05-09 08:51:37.186341+00', NULL),
(46, 1, 'MR-20260509-1-YUsMoD', '2026-05-09', 1, 2, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-09 08:52:06.833723+00', '2026-05-09 08:52:06.833723+00', NULL),
(47, 1, 'MR-20260511-1-huxJux', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-11 08:19:30.853282+00', '2026-05-11 08:19:30.853282+00', NULL),
(48, 1, 'MR-20260511-1-ehfPeh', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-11 08:30:32.830175+00', '2026-05-11 08:30:32.830175+00', NULL),
(49, 1, 'MR-20260511-1-FLYxqO', '2026-05-11', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-11 08:32:10.19843+00', '2026-05-11 08:32:10.19843+00', NULL),
(50, 1, 'MR-20260512-1-f34Chb', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-12 05:08:01.542107+00', '2026-05-12 05:08:01.542107+00', NULL),
(51, 1, 'MR-20260512-1-rjpx3a', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-12 05:10:53.599995+00', '2026-05-12 05:10:53.599995+00', NULL),
(52, 1, 'MR-20260512-1-PesMFD', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-12 05:12:01.708601+00', '2026-05-12 05:12:01.708601+00', NULL),
(53, 1, 'MR-20260512-1-2G93IH', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-12 05:12:14.263886+00', '2026-05-12 05:12:14.263886+00', NULL),
(54, 1, 'MR-20260512-1-wYohCA', '2026-05-12', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-12 05:12:27.850393+00', '2026-05-12 05:12:27.850393+00', NULL),
(55, 1, 'MR-20260513-1-DGmrwv', '2026-05-13', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-13 05:27:04.142696+00', '2026-05-13 05:27:04.142696+00', NULL),
(56, 1, 'MR-20260513-1-UVuDVo', '2026-05-13', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-13 05:29:38.405567+00', '2026-05-13 05:29:38.405567+00', NULL),
(57, 1, 'MR-20260520-1-R3kBq1', '2026-05-20', 1, 1, NULL, 116, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-20 02:16:03.121269+00', '2026-05-20 02:16:03.121269+00', NULL),
(58, 1, 'MR-20260520-1-nuLiSC', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-20 02:16:14.089542+00', '2026-05-20 02:16:14.089542+00', NULL),
(59, 1, 'MR-20260520-1-uIYcc1', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-20 02:16:34.475044+00', '2026-05-20 02:16:34.475044+00', NULL),
(60, 1, 'MR-20260520-1-W3bBXH', '2026-05-20', 1, 1, NULL, NULL, 'draft', 1, 4, NULL, NULL, NULL, '2026-05-20 02:16:52.992284+00', '2026-05-20 02:16:52.992284+00', NULL),
(61, 1, 'R-2026-021', '2026-05-22', 2, 3, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(62, 1, 'R-2026-022', '2026-05-22', 4, 6, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(63, 1, 'R-2026-023', '2026-05-22', 8, 10, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(64, 1, 'R-2026-024', '2026-05-22', 9, 11, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(65, 1, 'R-2026-025', '2026-05-22', 11, 13, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(66, 1, 'R-2026-026', '2026-05-22', 12, 14, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(67, 1, 'R-2026-027', '2026-05-22', 13, 15, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(68, 1, 'R-2026-028', '2026-05-22', 14, 16, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(69, 1, 'R-2026-029', '2026-05-22', 15, 17, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(70, 1, 'R-2026-030', '2026-05-22', 16, 18, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(71, 1, 'R-2026-031', '2026-05-22', 17, 19, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(72, 1, 'R-2026-032', '2026-05-22', 18, 20, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(73, 1, 'R-2026-033', '2026-05-22', 19, 21, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(74, 1, 'R-2026-034', '2026-05-22', 20, 22, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(75, 1, 'R-2026-035', '2026-05-22', 21, 25, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(76, 1, 'R-2026-036', '2026-05-22', 22, 27, 2, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(77, 1, 'R-2026-037', '2026-05-22', 39, 49, 1, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(78, 1, 'R-2026-038', '2026-05-22', 40, 50, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(79, 1, 'R-2026-039', '2026-05-22', 41, 51, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(80, 1, 'R-2026-040', '2026-05-22', 42, 52, 2, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(81, 1, 'R-2026-041', '2026-05-22', 43, 53, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(82, 1, 'R-2026-042', '2026-05-22', 44, 54, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(83, 1, 'R-2026-043', '2026-05-22', 45, 55, 1, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(84, 1, 'R-2026-044', '2026-05-22', 46, 56, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(85, 1, 'R-2026-045', '2026-05-22', 47, 57, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(86, 1, 'R-2026-046', '2026-05-22', 48, 58, 2, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(87, 1, 'R-2026-047', '2026-05-22', 49, 59, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(88, 1, 'R-2026-048', '2026-05-22', 50, 60, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(89, 1, 'R-2026-049', '2026-05-22', 51, 61, 1, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(90, 1, 'R-2026-050', '2026-05-22', 52, 62, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(301, 2, 'C-2026-004', '2026-05-22', 23, 29, 16, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(302, 2, 'C-2026-005', '2026-05-22', 24, 31, 17, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(303, 2, 'C-2026-006', '2026-05-22', 25, 32, 16, NULL, 'draft', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(401, 3, 'S-2026-005', '2026-05-22', 31, 39, 26, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(402, 3, 'S-2026-006', '2026-05-22', 32, 40, 27, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(501, 1, 'EX-001', '2026-05-22', 5, 80, 1, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(502, 1, 'EX-002', '2026-05-22', 10, 81, 2, NULL, 'finalized', 1, NULL, NULL, NULL, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(503, 1, 'MR-20260531-1-G2xCz3', '2026-05-31', 41, 51, 17, 1021, 'draft', 1, NULL, NULL, NULL, 'revisit', '2026-05-31 04:40:33.964912+00', '2026-05-31 04:40:33.964912+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- prescriptions
-- -----------------------------------------------------------------------------
INSERT INTO prescriptions ("id", "clinic_id", "owner_id", "pet_id", "medical_record_id", "prescribed_at", "duration_days", "deleted_at", "created_at", "updated_at") VALUES
(1, 1, 8, 10, 63, '2026-05-22', 14, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 4, 6, 62, '2026-05-22', 7, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 11, 13, 65, '2026-05-22', 5, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 1, 14, 16, 68, '2026-05-22', 21, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 1, 19, 21, 73, '2026-05-22', 10, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 1, 40, 50, 78, '2026-05-22', 7, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 1, 44, 54, 82, '2026-05-22', 14, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 1, 50, 60, 88, '2026-05-22', 7, NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- medical_record_addenda
-- -----------------------------------------------------------------------------
INSERT INTO medical_record_addenda ("id", "medical_record_id", "clinic_id", "author_user_id", "before_text", "after_text", "reason", "created_at") VALUES
(1, 61, 1, 1, '体重 26.0kg', '体重 26.4kg', '入力ミス。前回の値を誤って記載。', '2026-05-31 04:33:17.574774+00'),
(2, 63, 1, 1, '嘔吐2回', '嘔吐3回', '飼い主様からの追加情報により修正。', '2026-05-31 04:33:17.574774+00'),
(3, 3, 1, 2, '治療継続', '週1回の通院に変更', '経過良好のため来院頻度を調整。', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- vaccinations
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations ("id", "clinic_id", "medical_record_id", "pet_id", "vaccine_id", "date", "next_date", "next_schedule_type", "doctor_id", "supplemental", "lot1", "lot2", "lot3", "lot4", "remarks", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 1, 1, 1, '2025-12-20', '2026-12-20', '1year', 1, '', 'LOT-2025-A001', '', '', '', '5種混合ワクチン接種。体調良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 5, 3, 5, '2025-11-25', '2026-11-25', '1year', 1, '', 'LOT-2025-C001', '', '', '', '3種混合ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 6, 3, 5, '2026-03-18', '2027-03-18', '1year', 2, '', 'LOT-2026-C001', '', '', '', '3種混合ワクチン追加接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 15, 13, 5, '2026-03-18', '2026-04-15', '4weeks', 1, '', 'LOT-2026-C003', '', '', '', '初回3種混合（猫）。4週後に2回目接種予定。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 20, 18, 5, '2026-03-18', '2027-03-18', '1year', 1, '', 'LOT-2026-C002', '', '', '', '3種混合ワクチン接種。体調良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 2, 22, 31, 13, '2025-12-30', '2026-12-30', '1year', 17, '', 'LOT-2025-J001', '', '', '', '3種混合ワクチン接種。体調良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 2, 26, 35, 13, '2025-11-15', '2026-11-15', '1year', 17, '', 'LOT-2025-J002', '', '', '', '3種混合ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 2, 21, 29, 11, '2026-01-20', '2027-01-20', '1year', 16, '', 'LOT-2025-J003', '', '', '', '5種混合ワクチン接種。体調良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 3, 29, 39, 16, '2025-12-15', '2026-12-15', '1year', 26, '', 'LOT-2025-S001', '', '', '', '5種混合ワクチン接種。体調良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 3, 30, 40, 18, '2026-02-27', '2027-02-27', '1year', 27, '', 'LOT-2025-S002', '', '', '', '3種混合ワクチン接種（猫）。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 3, 36, 47, 19, '2026-04-22', '2027-04-22', '1year', 26, '', 'LOT-2026-S001', '', '', '', '5種混合ワクチン（猫）接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 3, 29, 39, 20, '2025-12-15', '2026-12-15', '1year', 26, '', 'LOT-2025-S003', '', '', '', '狂犬病ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- checkups
-- -----------------------------------------------------------------------------
INSERT INTO checkups ("id", "medical_record_id", "clinic_id", "pet_id", "checkup_type_id", "date", "next_date", "doctor_id", "result", "created_at", "updated_at", "deleted_at") VALUES
(1, 63, 1, 10, 1, '2026-05-22', '2027-05-22', 1, '一般身体検査：良好。血液検査：一部高値。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 69, 1, 17, 2, '2026-05-22', '2026-11-22', 1, 'シニア健診。心雑音なし。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 81, 1, 53, 1, '2026-05-22', '2027-05-22', 1, '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 82, 1, 54, 3, '2026-05-22', '2027-05-22', 2, '画像診断：特記事項なし。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 1, 1, 1, '2025-10-10', '2026-10-10', 1, '去年の健診。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 2, 1, 1, 1, '2026-02-24', '2027-02-24', 1, '半年健診。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- exams
-- -----------------------------------------------------------------------------
INSERT INTO exams ("id", "medical_record_id", "clinic_id", "pet_id", "date", "exam_type_id", "doctor_id", "status", "result_summary", "machine", "created_at", "updated_at", "deleted_at") VALUES
(1, 2, 1, NULL, '2026-02-24', 1, 1, 'completed', 'CBC全項目正常範囲内。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 14, 1, NULL, '2025-08-25', 2, 2, 'completed', 'ALT軽度上昇（145 U/L）。他正常。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 13, 1, NULL, '2026-04-22', 1, 1, 'result_entered', 'WBC上昇（19.2）。脱水を反映。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 23, 2, NULL, '2026-02-24', 6, 16, 'completed', 'CBC全項目正常範囲内。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 25, 2, NULL, '2026-02-10', 8, 16, 'completed', '腹部正面：腸管ガス像あり。他異常なし。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 31, 3, NULL, '2026-03-20', 9, 26, 'completed', 'CBC全項目正常範囲内。加齢による軽度変化のみ。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 32, 3, NULL, '2026-04-27', 10, 27, 'completed', '尿比重低下（1.010）。腎機能低下疑い。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 33, 3, NULL, '2025-11-30', 11, 26, 'completed', '腹部エコー：異常なし。皮膚問題は別要因。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 63, 1, NULL, '2026-05-22', 1, 1, 'completed', 'WBC正常。炎症所見なし。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 63, 1, NULL, '2026-05-22', 2, 1, 'completed', '肝数値安定。腎数値にやや上昇あり。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 69, 1, NULL, '2026-05-22', 2, 1, 'completed', 'スクリーニング検査。全項目正常。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 64, 1, NULL, '2026-05-22', 1, 2, 'completed', '術前検査。異常なし。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 71, 1, NULL, '2026-05-22', 2, 1, 'completed', 'スケーリング前検査。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, 1, NULL, '2025-10-10', 1, 1, 'completed', '過去データ1。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 1, 1, NULL, '2025-12-20', 1, 1, 'completed', '過去データ2。', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- exam_results
-- -----------------------------------------------------------------------------
INSERT INTO exam_results ("id", "exam_id", "exam_type_field_id", "name", "inspection_value", "normal_value", "result", "unit", "reference_value", "ref_min", "ref_max", "is_abnormal", "status", "sort_order", "created_at", "updated_at") VALUES
(34, 1, 1, '', '9.8', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(35, 1, 2, '', '7.2', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(36, 1, 3, '', '45', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(37, 1, 4, '', '320', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(38, 2, 5, '', '145', '', '', '', '', NULL, NULL, 'f', 'high', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(39, 2, 6, '', '18', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(40, 2, 7, '', '1.2', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(41, 2, 8, '', '98', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(42, 3, 1, '', '19.2', '', '', '', '', NULL, NULL, 'f', 'high', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(43, 3, 2, '', '8.1', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(44, 3, 3, '', '52', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(45, 3, 4, '', '280', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(46, 4, 18, '', '8.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(47, 4, 19, '', '6.8', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(48, 4, 20, '', '46', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(49, 5, 24, '', '腸管ガス像あり', '', '', '', '', NULL, NULL, 'f', 'high', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(50, 5, 25, '', '異常なし', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(51, 6, 26, '', '9.2', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(52, 6, 27, '', '7.0', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(53, 6, 28, '', '320', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(54, 7, 29, '', '1.010', '', '', '', '', NULL, NULL, 'f', 'low', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(55, 7, 30, '', '6.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(56, 7, 31, '', '陰性', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(57, 8, 32, '', '異常なし', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(58, 8, 33, '', '異常なし', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(100, 10, 1, '', '12.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(101, 10, 2, '', '7.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(102, 10, 3, '', '42', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(103, 10, 4, '', '350', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(104, 11, 5, '', '65', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(105, 11, 6, '', '32', '', '', '', '', NULL, NULL, 'f', 'high', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(106, 11, 7, '', '2.1', '', '', '', '', NULL, NULL, 'f', 'high', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(107, 11, 8, '', '110', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(108, 12, 5, '', '45', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(109, 12, 6, '', '22', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(110, 12, 7, '', '1.1', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(111, 12, 8, '', '95', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(112, 13, 1, '', '10.2', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(113, 13, 2, '', '7.8', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(114, 13, 3, '', '40', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(115, 13, 4, '', '410', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(116, 14, 5, '', '55', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(117, 14, 6, '', '18', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(118, 14, 7, '', '0.9', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(119, 14, 8, '', '105', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(120, 15, 1, '', '8.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(121, 15, 2, '', '6.5', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(122, 15, 3, '', '38', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(123, 15, 4, '', '300', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(124, 16, 1, '', '11.0', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(125, 16, 2, '', '7.2', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(126, 16, 3, '', '44', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(127, 16, 4, '', '330', '', '', '', '', NULL, NULL, 'f', 'normal', 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- clinical_plans
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans ("id", "medical_record_id", "physical_exam", "diagnosis_type_id", "diagnosis_name_id", "diagnosis_2_type_id", "diagnosis_2_name_id", "diagnosis_details", "treatment_policy", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, '体温38.5℃。心肺音正常。', NULL, NULL, NULL, NULL, '健康状態良好。ワクチン接種可。', '5種混合ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 2, '体重増加あり。他異常なし。', NULL, NULL, NULL, NULL, '維持状態良好。', '定期検診継続。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 4, '右後肢跛行。パテラG2。', 8, 42, NULL, NULL, '膝蓋骨脱臼。', '消炎剤処方。体重管理指導。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 5, '異常なし。', NULL, NULL, NULL, NULL, '予防シーズン開始。', 'フィラリア予防薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 3, '良好。', NULL, NULL, NULL, NULL, '年次予防。', 'ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 6, '良好。', NULL, NULL, NULL, NULL, '年次予防。', 'ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 7, '良好。', NULL, NULL, NULL, NULL, '外部寄生虫予防。', 'スポットオン投与。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 8, '全身に発赤。搔痒感強。', 3, 6, NULL, NULL, 'アトピー性皮膚炎。', '抗ヒスタミン薬処方。薬用シャンプー推奨。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 9, '皮膚の一部に発赤。', 3, 7, NULL, NULL, '膿皮症初期。', '洗浄と消毒。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 10, '腹部軽度緊張。', 1, 1, NULL, NULL, '急性胃腸炎疑い。', '絶食・皮下補液実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 11, '耳道内に分泌物。', 3, 41, NULL, NULL, '外耳炎。', '耳道洗浄・点耳薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 12, 'シニア期に入る。', NULL, NULL, NULL, NULL, '健康診断実施。', '結果待ち。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 13, '脱水傾向あり。', 1, 1, NULL, NULL, '急性胃腸炎。', '対症療法と食事療法。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 14, '良好。', NULL, NULL, NULL, NULL, '経過観察。', '維持。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 15, '良好。', NULL, NULL, NULL, NULL, '幼若期検診。', '成長記録。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 16, '良好。', NULL, NULL, NULL, NULL, 'スクリーニング。', '異常なし。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 17, '重度の歯石。', NULL, NULL, NULL, NULL, '歯周病。', '抜歯を含めた歯科処置を計画。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 18, '良好。', NULL, NULL, NULL, NULL, '肥満気味。', 'ダイエットフード提案。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 19, '跛行消失。', 8, 42, NULL, NULL, '回復期。', '運動制限解除。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 20, '良好。', NULL, NULL, NULL, NULL, '年次予防。', 'ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 21, '体重28kg。体温38.4℃。心肺音正常。', NULL, NULL, NULL, NULL, 'フィラリア陰性。健康良好。', 'フィラリア検査・次回予防薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(22, 22, '体重3.8kg。良好。', NULL, NULL, NULL, NULL, 'ワクチン適応あり。', 'ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 23, 'シニア期。体重変化なし。', NULL, NULL, NULL, NULL, '経過良好。', '血液検査実施。結果説明。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(24, 24, '全身に搔痒。皮膚発赤あり。', 11, 47, NULL, NULL, 'アトピー性皮膚炎疑い。', '抗ヒスタミン薬処方。薬用シャンプー推奨。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(25, 25, '腹部軽度緊張。体温38.8℃。', 9, 21, NULL, NULL, '急性胃腸炎疑い。', '絶食指示・皮下補液実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(26, 26, '体重3.0kg。体調良好。', NULL, NULL, NULL, NULL, 'ワクチン適応あり。', 'ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(27, 27, '跛行改善傾向。', 20, 51, NULL, NULL, '前肢骨折（回復期）。', '運動制限継続。次回X線確認。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(28, 28, '腹部触診で軽度抵抗感。', 9, 21, NULL, NULL, '急性胃腸炎。', '補液・制吐剤投与。食事制限。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(29, 29, '体温38.3℃。体重24kg。心肺音正常。', NULL, NULL, NULL, NULL, '健康良好。ワクチン接種可。', '5種混合ワクチン接種実施。フィラリア予防薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(30, 30, '体重4.8kg。良好。', NULL, NULL, NULL, NULL, 'ワクチン適応あり。', 'ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(31, 31, 'シニア期。体重31kg。', NULL, NULL, NULL, NULL, '経過良好。', '血液化学検査実施。結果説明。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(32, 32, '脱水傾向あり。腹部緊張。', 14, 31, NULL, NULL, '急性胃腸炎。', '補液・絶食・制吐剤処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(33, 33, '体幹・四肢に紅斑。搔痒感強。', 15, 33, NULL, NULL, 'アトピー性皮膚炎。', 'ステロイド処方。薬用シャンプー推奨。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(34, 34, '一般状態良好。', NULL, NULL, NULL, NULL, '初診。特異所見なし。', '経過観察。次回再診。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(35, 35, '耳道内に暗褐色の耳垢。臭気あり。', NULL, NULL, NULL, NULL, '外耳炎。', '耳洗浄・点耳薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(36, 36, '体重4.0kg。体調良好。', NULL, NULL, NULL, NULL, 'ワクチン適応あり。', 'ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(37, 37, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 04:10:02.89686+00', '2026-05-07 04:10:02.89686+00', NULL),
(38, 38, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:44:37.891701+00', '2026-05-07 05:44:37.891701+00', NULL),
(39, 39, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:44:51.908098+00', '2026-05-07 05:44:51.908098+00', NULL),
(40, 40, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:45:37.465514+00', '2026-05-07 05:45:37.465514+00', NULL),
(41, 41, '', NULL, NULL, NULL, NULL, '', '', '2026-05-07 05:45:50.501602+00', '2026-05-07 05:45:50.501602+00', NULL),
(42, 42, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:48:12.303789+00', '2026-05-09 08:48:12.303789+00', NULL),
(43, 43, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:48:44.383347+00', '2026-05-09 08:48:44.383347+00', NULL),
(44, 44, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:51:27.802325+00', '2026-05-09 08:51:27.802325+00', NULL),
(45, 45, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:51:37.195867+00', '2026-05-09 08:51:37.195867+00', NULL),
(46, 46, '', NULL, NULL, NULL, NULL, '', '', '2026-05-09 08:52:06.842019+00', '2026-05-09 08:52:06.842019+00', NULL),
(47, 47, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:19:30.872507+00', '2026-05-11 08:19:30.872507+00', NULL),
(48, 48, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:30:32.843875+00', '2026-05-11 08:30:32.843875+00', NULL),
(49, 49, '', NULL, NULL, NULL, NULL, '', '', '2026-05-11 08:32:10.206396+00', '2026-05-11 08:32:10.206396+00', NULL),
(50, 50, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:08:01.563109+00', '2026-05-12 05:08:01.563109+00', NULL),
(51, 51, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:10:53.608981+00', '2026-05-12 05:10:53.608981+00', NULL),
(52, 52, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:01.764254+00', '2026-05-12 05:12:01.764254+00', NULL),
(53, 53, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:14.276389+00', '2026-05-12 05:12:14.276389+00', NULL),
(54, 54, '', NULL, NULL, NULL, NULL, '', '', '2026-05-12 05:12:27.858983+00', '2026-05-12 05:12:27.858983+00', NULL),
(55, 55, '', NULL, NULL, NULL, NULL, '', '', '2026-05-13 05:27:04.157177+00', '2026-05-13 05:27:04.157177+00', NULL),
(56, 56, '', NULL, NULL, NULL, NULL, '', '', '2026-05-13 05:29:38.41457+00', '2026-05-13 05:29:38.41457+00', NULL),
(57, 57, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:03.143892+00', '2026-05-20 02:16:03.143892+00', NULL),
(58, 58, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:14.103071+00', '2026-05-20 02:16:14.103071+00', NULL),
(59, 59, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:34.48364+00', '2026-05-20 02:16:34.48364+00', NULL),
(60, 60, '', NULL, NULL, NULL, NULL, '', '', '2026-05-20 02:16:53.008248+00', '2026-05-20 02:16:53.008248+00', NULL),
(100, 61, '体重26.4kg。体温38.5℃。心肺音正常。', NULL, NULL, NULL, NULL, '健康状態良好。', '5種混合ワクチン接種実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(101, 62, '左耳介に発赤、耳道内に黒色耳垢多量。', 3, 41, NULL, NULL, 'マラセチア性外耳炎。', '耳道洗浄、点耳薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(102, 63, '軽度脱水、腹部圧痛あり。', 1, 1, NULL, NULL, '急性胃腸炎。', '皮下補液、吐き気止め注射。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(103, 64, '術前再チェック、異常なし。', NULL, NULL, NULL, NULL, '避妊手術適応。', '全身麻酔下にて避妊手術実施。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(104, 65, '結膜充血、角膜に傷なし。', 3, 41, NULL, NULL, '結膜炎。', '抗菌点眼薬処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(105, 74, '意識朦朧、呼吸浅い。', 4, 33, NULL, NULL, 'てんかん重積の疑い。', '抗痙攣薬投与。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(106, 75, '右膝蓋骨脱臼(G2)を確認。', 8, 42, NULL, NULL, '膝蓋骨脱臼。', '運動制限、体重管理。消炎剤。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(107, 78, '口腔内歯石重度、歯肉炎。', 3, 41, NULL, NULL, '歯周病。', '抜歯を含む歯科処置を推奨。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(301, 301, '心雑音 Lev. 3/6。肺音正常。', NULL, NULL, NULL, NULL, '僧帽弁閉鎖不全症の維持。', '投薬継続。減塩食。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(302, 302, '微熱あり。腹部軟らかい。', NULL, NULL, NULL, NULL, '感冒の疑い。', '抗生剤・消炎剤。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(401, 401, '良好。', NULL, NULL, NULL, NULL, '年次予防。', 'ワクチン接種。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(402, 402, '良好。', NULL, NULL, NULL, NULL, '予防シーズン。', 'フィラリア薬。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(501, 501, '不正咬合の疑い。切歯が伸びている。', NULL, NULL, NULL, NULL, '過長歯（うサギ）', 'ニッパーによる切歯カット。牧草中心の食事指導。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(502, 502, '頬袋に炎症。', NULL, NULL, NULL, NULL, '頬袋炎', '内容物の除去。抗生剤シロップ処方。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(503, 503, '', NULL, NULL, NULL, NULL, '', '', '2026-05-31 04:40:33.978273+00', '2026-05-31 04:40:33.978273+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- inquiries
-- -----------------------------------------------------------------------------
INSERT INTO inquiries ("id", "medical_record_id", "chief_complaint_type_id", "chief_complaint", "history", "current_medications", "allergy_info", "last_meal", "last_defecation", "last_urination", "appetite", "water_intake", "owner_observations", "notes", "staff_id", "created_at", "updated_at") VALUES
(1, 1, NULL, '狂犬病ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', '体調良好。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 2, NULL, '定期健診', '', '', '', '', '', '', NULL, NULL, '', '特に異常なし。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 4, 6, '右足の跛行', '', '', '', '', '', '', NULL, NULL, '', '膝蓋骨脱臼を確認。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 5, NULL, 'フィラリア予防', '', '', '', '', '', '', NULL, NULL, '', '予防薬処方。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 3, NULL, '5種混合ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', '体調良好。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 6, NULL, '5種混合ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', '体調良好。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 7, NULL, 'ノミダニ予防薬', '', '', '', '', '', '', NULL, NULL, '', '予防薬処方。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 8, 3, '皮膚の痒み', '', '', '', '', '', '', NULL, NULL, '', 'アトピー性皮膚炎疑い。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 9, 3, 'トリミング後の皮膚チェック', '', '', '', '', '', '', NULL, NULL, '', '軽度の赤みあり。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 10, 1, '食欲不振', '', '', '', '', '', '', NULL, NULL, '', '2日前から食欲減退。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(11, 11, NULL, '耳を痒がる', '', '', '', '', '', '', NULL, NULL, '', '外耳炎疑い。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(12, 12, NULL, '定期健診・予防接種', '', '', '', '', '', '', NULL, NULL, '', '年次健診。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(13, 13, 2, '嘔吐・下痢', '', '', '', '', '', '', NULL, NULL, '', '昨日から嘔吐3回。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(14, 14, NULL, '生化学検査', '', '', '', '', '', '', NULL, NULL, '', 'シニア健診。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(15, 15, NULL, '3種混合ワクチン接種（猫）', '', '', '', '', '', '', NULL, NULL, '', '初回ワクチン。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(16, 16, NULL, '血液検査', '', '', '', '', '', '', NULL, NULL, '', '異常なし。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(17, 17, NULL, '歯石除去', '', '', '', '', '', '', NULL, NULL, '', '重度の歯石付着。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(18, 18, NULL, '定期検診', '', '', '', '', '', '', NULL, NULL, '', '体重管理継続。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(19, 19, 6, '再診（右足跛行）', '', '', '', '', '', '', NULL, NULL, '', '改善傾向。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(20, 20, NULL, '5種混合ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', '体調良好。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(25, 21, NULL, '定期健診・フィラリア検査', '', '', '', '', '', '', NULL, NULL, '', '体調良好。フィラリア陰性。', 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(26, 22, NULL, '3種混合ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', 'ワクチン接種。体調問題なし。', 17, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(27, 23, NULL, '血液検査・一般健診', '', '', '', '', '', '', NULL, NULL, '', '年次健診。', 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(28, 24, 9, '皮膚の痒み・発赤', '', '', '', '', '', '', NULL, NULL, '', '2週間前から痒がる。アトピー疑い。', 18, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(29, 25, 7, '食欲低下', '', '', '', '', '', '', NULL, NULL, '', '3日前から食欲落ちている。', 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(30, 26, NULL, '3種混合ワクチン接種（猫）', '', '', '', '', '', '', NULL, NULL, '', 'ワクチン接種。良好。', 17, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(31, 27, NULL, '再診（足の不調）', '', '', '', '', '', '', NULL, NULL, '', '前回から改善傾向。', 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(32, 28, 7, '食欲不振・嘔吐', '', '', '', '', '', '', NULL, NULL, '', '初診。昨日から嘔吐1回。', 18, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(33, 29, NULL, '年次健診・ワクチン接種', '', '', '', '', '', '', NULL, NULL, '', '体調良好。体重24kg。', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(34, 30, NULL, '3種混合ワクチン接種（猫）', '', '', '', '', '', '', NULL, NULL, '', 'ワクチン接種。体調良好。', 27, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(35, 31, NULL, '血液検査（年次健診）', '', '', '', '', '', '', NULL, NULL, '', 'シニア健診。', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(36, 32, 12, '嘔吐・下痢', '', '', '', '', '', '', NULL, NULL, '', '昨日から嘔吐2回。軟便あり。', 27, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(37, 33, 17, '皮膚の痒み・かさぶた', '', '', '', '', '', '', NULL, NULL, '', 'アレルギー疑い。フリーズドライ投与。', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(38, 34, NULL, '初診・一般診察', '', '', '', '', '', '', NULL, NULL, '', '初めての来院。', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(39, 35, NULL, '耳を痒がる', '', '', '', '', '', '', NULL, NULL, '', '外耳炎疑い。耳垢あり。', 27, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(40, 36, NULL, '3種混合ワクチン接種（猫）', '', '', '', '', '', '', NULL, NULL, '', 'ワクチン接種。体重4.0kg。', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(41, 37, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 04:10:02.879897+00', '2026-05-07 04:10:02.885494+00'),
(42, 38, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:44:37.883924+00', '2026-05-07 05:44:37.886174+00'),
(43, 39, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:44:51.901744+00', '2026-05-07 05:44:51.904033+00'),
(44, 40, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:45:37.461418+00', '2026-05-07 05:45:43.378725+00'),
(45, 41, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-07 05:45:50.497277+00', '2026-05-07 05:45:50.498997+00'),
(46, 42, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:48:12.294211+00', '2026-05-09 08:48:12.29652+00'),
(47, 43, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:48:44.376885+00', '2026-05-09 08:48:44.379088+00'),
(48, 44, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:51:27.795388+00', '2026-05-09 08:51:27.798253+00'),
(49, 45, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:51:37.1897+00', '2026-05-09 08:51:37.192864+00'),
(50, 46, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-09 08:52:06.837309+00', '2026-05-09 08:52:06.839046+00'),
(51, 47, 1, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:19:30.863395+00', '2026-05-11 08:22:59.398691+00'),
(52, 48, 1, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:30:32.836651+00', '2026-05-11 08:31:51.424634+00'),
(53, 49, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-11 08:32:10.20185+00', '2026-05-11 08:32:15.234448+00'),
(54, 50, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:08:01.550843+00', '2026-05-12 05:08:01.554051+00'),
(55, 51, 2, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:10:53.603986+00', '2026-05-12 05:11:45.172348+00'),
(56, 52, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:01.715127+00', '2026-05-12 05:12:01.751826+00'),
(57, 53, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:14.269626+00', '2026-05-12 05:12:14.272251+00'),
(58, 54, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-12 05:12:27.854086+00', '2026-05-12 05:12:27.855968+00'),
(59, 55, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-13 05:27:04.148505+00', '2026-05-13 05:27:35.8547+00'),
(60, 56, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-13 05:29:38.409629+00', '2026-05-13 05:29:38.41183+00'),
(61, 57, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:03.133988+00', '2026-05-20 02:16:03.13918+00'),
(62, 58, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:14.096188+00', '2026-05-20 02:16:14.098507+00'),
(63, 59, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:34.478667+00', '2026-05-20 02:16:34.480523+00'),
(64, 60, NULL, '', '', '', '', '', '', '', NULL, NULL, '', '', NULL, '2026-05-20 02:16:53.000132+00', '2026-05-20 02:16:53.002662+00'),
(100, 61, NULL, '定期健診とワクチン', '', '', '', '', '', '', NULL, NULL, '', '1年前の混合ワクチンから1年経過。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(101, 62, 3, '耳の赤みと痒み', '', '', '', '', '', '', NULL, NULL, '', '昨日からしきりに耳を振っている。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(102, 63, 2, '食欲不振と嘔吐', '', '', '', '', '', '', NULL, NULL, '', '今朝から3回、黄色い液を吐いた。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(103, 64, NULL, '避妊手術', '', '', '', '', '', '', NULL, NULL, '', '術前検査済み。絶食で来院。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(104, 65, 3, '目やに', '', '', '', '', '', '', NULL, NULL, '', '左目が開けにくそう。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(105, 74, NULL, '夜間救急：痙攣', '', '', '', '', '', '', NULL, NULL, '', '15分前に全身性の痙攣発作。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(106, 75, 6, '散歩中に足を引きずる', '', '', '', '', '', '', NULL, NULL, '', '右後ろ足を浮かせる。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(107, 78, 1, '元気がない', '', '', '', '', '', '', NULL, NULL, '', '1週間前から徐々に活動性低下。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(301, 301, NULL, '再診。心雑音のチェック。', '', '', '', '', '', '', NULL, NULL, '', '', 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(302, 302, NULL, '初診。昨日から元気がない。', '', '', '', '', '', '', NULL, NULL, '', '', 17, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(401, 401, NULL, '混合ワクチン。', '', '', '', '', '', '', NULL, NULL, '', '', 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(402, 402, NULL, 'フィラリア予防。', '', '', '', '', '', '', NULL, NULL, '', '', 27, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- treatments
-- -----------------------------------------------------------------------------
INSERT INTO treatments ("id", "medical_record_id", "item_type", "consultation_id", "procedure_id", "medicine_id", "is_selected", "status", "content", "memo", "admin_route", "is_insurance", "unit_price", "quantity", "discount_rate", "discount_amount", "inventory_id", "sort_order", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 'medicine', NULL, NULL, 1, 't', 'completed', 'アモキシシリン 50mg x 7日分', '', '', 'f', 500, 7.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 2, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 'procedure', NULL, 4, NULL, 't', 'completed', '耳道洗浄（左耳）', '', '', 'f', 2500, 1.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 2, 'consultation', 1, NULL, NULL, 't', 'completed', '初診料', '', '', 'f', 2000, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 2, 'medicine', NULL, NULL, 1, 't', 'completed', 'アモキシシリン 50mg x 5日分', '', '', 'f', 500, 5.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 3, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 21, 'consultation', 7, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 900, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 22, 'consultation', 6, NULL, NULL, 't', 'completed', '初診料', '', '', 'f', 2200, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 24, 'medicine', NULL, NULL, 104, 't', 'completed', 'プレドニゾロン 5mg x 7日分', '', '', 'f', 420, 7.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 25, 'consultation', 7, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 900, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 25, 'medicine', NULL, NULL, 101, 't', 'completed', 'アモキシシリン 50mg x 5日分', '', '', 'f', 520, 5.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 29, 'consultation', 9, NULL, NULL, 't', 'completed', '初診料', '', '', 'f', 1900, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 30, 'consultation', 10, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 32, 'consultation', 10, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 32, 'medicine', NULL, NULL, 201, 't', 'completed', 'アモキシシリン 50mg x 5日分', '', '', 'f', 510, 5.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 35, 'procedure', NULL, 27, NULL, 't', 'completed', '耳道洗浄・点耳薬処置', '', '', 'f', 2600, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(100, 61, 'consultation', 1, NULL, NULL, 't', 'completed', '初診料', '', '', 'f', 1500, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(101, 61, 'medicine', NULL, NULL, 5, 't', 'completed', '5種混合ワクチン', '', '', 'f', 3000, 1.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(102, 62, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(103, 62, 'procedure', NULL, 4, NULL, 't', 'completed', '耳道洗浄', '', '', 'f', 2500, 1.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(104, 63, 'consultation', 2, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(105, 63, 'procedure', NULL, 12, NULL, 't', 'completed', '皮下補液', '', '', 'f', 3500, 1.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(106, 74, 'consultation', 1, NULL, NULL, 't', 'completed', '夜間緊急診察料', '', '', 'f', 10000, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(107, 74, 'procedure', NULL, 5, NULL, 't', 'completed', '静脈確保・採血', '', '', 'f', 3000, 1.0, 0.00, 0, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(301, 301, 'consultation', 10, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(302, 302, 'consultation', 9, NULL, NULL, 't', 'completed', '初診料', '', '', 'f', 1900, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(401, 401, 'consultation', 10, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(402, 402, 'consultation', 10, NULL, NULL, 't', 'completed', '再診料', '', '', 'f', 800, 1.0, 0.00, 0, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- treatment_plans
-- -----------------------------------------------------------------------------
INSERT INTO treatment_plans ("id", "medical_record_id", "hospitalization_id", "treatment_content", "memo", "is_insurance", "unit_price", "quantity", "discount_rate", "discount_amount", "subtotal", "sort_order", "created_at", "updated_at", "deleted_at") VALUES
(1, 63, NULL, '腎臓ケア処方食の継続', '3ヶ月継続後に再評価', 't', 3000, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 63, NULL, '血液検査（腎パネル）', '次回再来時', 't', 6000, 1.0, 0.00, 0, 0, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 64, NULL, '避妊手術後の経過観察', '本日実施', 't', 0, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 64, NULL, '抜糸（1週間後）', '予約済み', 't', 1500, 1.0, 0.00, 0, 0, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 71, NULL, '歯科スケーリング', '完了', 'f', 15000, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 71, NULL, '抗生剤投与（7日間）', '術後感染予防', 't', 4200, 1.0, 0.00, 0, 0, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, NULL, '次年度混合ワクチン', '2026年12月頃', 't', 5000, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 82, NULL, '心エコー検査', '3ヶ月後', 't', 6000, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 88, NULL, '肝補助剤（ウルソ）', '長期投与', 't', 3500, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 90, NULL, '抜糸処置', '1週間後', 't', 1500, 1.0, 0.00, 0, 0, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- medical_record_images
-- -----------------------------------------------------------------------------
INSERT INTO medical_record_images ("id", "medical_record_id", "image_url", "thumbnail_url", "file_name", "file_size", "mime_type", "image_type", "description", "taken_at", "exam_id", "staff_id", "sort_order", "created_at", "updated_at") VALUES
(1, 61, 'files/clinic1/iris.jpg', 'files/clinic1/iris_t.jpg', 'iris_portrait.jpg', 102400, 'image/jpeg', 'photo', '初診時の写真', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 63, 'files/clinic1/mike_x.png', 'files/clinic1/mike_x_t.png', 'mike_xray.png', 204800, 'image/png', 'xray', '腹部レントゲン正面', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 65, 'files/clinic1/rocky_b.pdf', 'files/clinic1/rocky_b_t.jpg', 'rocky_blood_report.pdf', 512000, 'application/pdf', 'other', '血液検査結果PDF', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 64, 'files/clinic1/luna_c.pdf', 'files/clinic1/luna_c_t.jpg', 'luna_surgery_consent.pdf', 128000, 'application/pdf', 'other', '手術同意書（スキャン）', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 62, 'files/clinic1/choco_v.png', 'files/clinic1/choco_v_t.jpg', 'choco_vaccine_cert.png', 153600, 'image/png', 'photo', 'ワクチン証明書', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 7, 'files/clinic1/bone_xray.jpg', 'files/clinic1/bone_xray_t.jpg', 'fracture.jpg', 300000, 'image/jpeg', 'xray', '大腿骨骨折（術前）', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 10, 'files/clinic1/heart_echo.jpg', 'files/clinic1/heart_echo_t.jpg', 'echo.jpg', 250000, 'image/jpeg', 'echo', '心臓逆流所見', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 15, 'files/clinic1/skin_photo.jpg', 'files/clinic1/skin_photo_t.jpg', 'atopy.jpg', 150000, 'image/jpeg', 'photo', '腹部発赤の状態', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 17, 'files/clinic1/dental_x.jpg', 'files/clinic1/dental_x_t.jpg', 'dental.jpg', 200000, 'image/jpeg', 'xray', '臼歯部の歯石付着', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 75, 'files/clinic1/patella_x.jpg', 'files/clinic1/patella_x_t.jpg', 'patella.jpg', 350000, 'image/jpeg', 'xray', '膝蓋骨脱臼（右）', NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- billing_confirmations
-- -----------------------------------------------------------------------------
INSERT INTO billing_confirmations ("id", "medical_record_id", "status", "confirmed_by", "confirmed_at", "returned_by", "returned_at", "return_reason", "memo", "created_at", "updated_at") VALUES
(1, 61, 'confirmed', 1, '2026-05-31 02:33:17.574774+00', NULL, NULL, '', '確認済み。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 62, 'confirmed', 2, '2026-05-31 03:33:17.574774+00', NULL, NULL, '', '処置内容問題なし。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 63, 'pending', NULL, NULL, NULL, NULL, '', '検査結果待ち。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 64, 'confirmed', 1, '2026-05-31 04:03:17.574774+00', NULL, NULL, '', '手術記録と合致。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 66, 'returned', NULL, NULL, 1, '2026-05-31 04:13:17.574774+00', '処方薬の数量が不整合です。確認お願いします。', '差し戻し', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 68, 'confirmed', 2, '2026-05-31 04:23:17.574774+00', NULL, NULL, '', '保険適用確認済み。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 70, 'pending', NULL, NULL, NULL, NULL, '', '後ほど確認。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 71, 'confirmed', 1, '2026-05-31 04:28:17.574774+00', NULL, NULL, '', 'スケーリング費用適正。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 72, 'returned', NULL, NULL, 2, '2026-05-31 04:28:17.574774+00', '再診料の重複があります。', '要修正', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 74, 'pending', NULL, NULL, NULL, NULL, '', '救急加算の確認。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- billings
-- -----------------------------------------------------------------------------
INSERT INTO billings ("id", "clinic_id", "medical_record_id", "hospitalization_id", "owner_id", "pet_id", "subtotal", "tax_total", "total_amount", "has_insurance", "status", "scheduled_date", "completed_at", "memo", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 1, NULL, 1, 1, 4300, 430, 4730, 't', 'completed', '2026-04-27', '2026-04-27 01:30:00+00', 'アニコム保険適用', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 3, NULL, 1, 1, 6100, 554, 6654, 't', 'completed', '2026-05-10', '2026-05-10 02:00:00+00', 'アニコム保険適用（Iris 耳炎治療）+ フード販売', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 6, NULL, 2, 3, 800, 80, 880, 't', 'waiting', '2026-05-22', NULL, 'アニコム保険適用。会計待ち。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 61, NULL, 2, 3, 4500, 450, 4950, 'f', 'completed', '2026-05-22', '2026-05-22 00:30:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 62, NULL, 4, 6, 3500, 350, 3850, 'f', 'completed', '2026-05-22', '2026-05-22 01:30:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, NULL, NULL, 5, 7, 6200, 596, 6796, 'f', 'completed', '2026-05-22', '2026-05-22 02:30:00+00', 'トリミング来院', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 1, 63, NULL, 8, 10, 6800, 680, 7480, 't', 'completed', '2026-05-22', '2026-05-22 04:00:00+00', 'アニコム保険適用', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 1, 64, NULL, 9, 11, 25000, 2500, 27500, 't', 'completed', '2026-05-22', '2026-05-22 05:30:00+00', 'アニコム保険適用（避妊手術）', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 1, NULL, NULL, 10, 12, 5000, 500, 5500, 'f', 'completed', '2026-05-22', '2026-05-22 06:30:00+00', 'ペットホテル利用', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, 65, NULL, 11, 13, 1400, 140, 1540, 'f', 'completed', '2026-05-22', '2026-05-22 07:30:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, 66, NULL, 12, 14, 2500, 250, 2750, 'f', 'completed', '2026-05-22', '2026-05-22 00:45:00+00', '狂犬病ワクチン', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 1, 67, NULL, 13, 15, 1200, 120, 1320, 'f', 'completed', '2026-05-22', '2026-05-22 01:15:00+00', '再診・点耳薬のみ', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 1, NULL, NULL, 1, 1, 3000, 240, 3240, 'f', 'completed', '2026-05-22', '2026-05-22 02:00:00+00', 'フード購入のみ', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 1, 68, NULL, 14, 16, 5000, 500, 5500, 't', 'completed', '2026-05-22', '2026-05-22 03:30:00+00', 'アイペット 50%', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 1, 69, NULL, 15, 17, 8500, 850, 9350, 'f', 'completed', '2026-05-22', '2026-05-22 04:45:00+00', '血液検査フルセット', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 1, 70, NULL, 16, 18, 4200, 420, 4620, 'f', 'completed', '2026-05-22', '2026-05-22 05:15:00+00', '5種混合ワクチン', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 1, 71, NULL, 17, 19, 15000, 1500, 16500, 'f', 'completed', '2026-05-22', '2026-05-22 06:45:00+00', '歯科スケーリング', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 1, 72, NULL, 18, 20, 2800, 280, 3080, 'f', 'completed', '2026-05-22', '2026-05-22 08:15:00+00', '再診・爪切り・肛門腺', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 1, 73, NULL, 19, 21, 6000, 600, 6600, 't', 'completed', '2026-05-22', '2026-05-22 09:00:00+00', 'アニコム 70%', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 1, NULL, NULL, 20, 22, 10000, 1000, 11000, 'f', 'completed', '2026-05-22', '2026-05-22 09:15:00+00', '時間外・緊急', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(301, 2, 301, NULL, 23, 29, 800, 80, 880, 'f', 'completed', '2026-05-22', '2026-05-22 01:30:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(302, 2, 302, NULL, 24, 31, 1900, 190, 2090, 'f', 'completed', '2026-05-22', '2026-05-22 02:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(401, 3, 401, NULL, 31, 39, 800, 80, 880, 'f', 'completed', '2026-05-22', '2026-05-22 00:15:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(402, 3, 402, NULL, 32, 40, 800, 80, 880, 'f', 'completed', '2026-05-22', '2026-05-22 01:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(403, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-01', '2026-05-01 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(404, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-02', '2026-05-02 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(405, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-03', '2026-05-03 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(406, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-04', '2026-05-04 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(407, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-05', '2026-05-05 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(408, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-06', '2026-05-06 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(409, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-07', '2026-05-07 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(410, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-08', '2026-05-08 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(411, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-09', '2026-05-09 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(412, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-10', '2026-05-10 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(413, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-11', '2026-05-11 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(414, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-12', '2026-05-12 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(415, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-13', '2026-05-13 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(416, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-14', '2026-05-14 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(417, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-15', '2026-05-15 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(418, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-16', '2026-05-16 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(419, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-17', '2026-05-17 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(420, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-18', '2026-05-18 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(421, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-19', '2026-05-19 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(422, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-20', '2026-05-20 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(423, 1, NULL, NULL, 1, 1, 5000, 500, 5500, 'f', 'completed', '2026-05-21', '2026-05-21 11:00:00+00', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- billing_items
-- -----------------------------------------------------------------------------
INSERT INTO billing_items ("id", "billing_id", "category", "name", "unit_price", "quantity", "tax_type", "tax_rate", "is_insurance_applicable", "source", "merchandise_item_id", "sort_order", "created_at", "updated_at", "deleted_at", "treatment_id", "appointment_id", "trimming_course_id", "trimming_option_id") VALUES
(1, 1, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(2, 1, 'medicine', 'アモキシシリン 50mg x 7日分', 500, 7.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(3, 2, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(4, 2, 'procedure', '耳道洗浄', 2500, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(5, 1, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(6, 2, 'food', 'ロイヤルカナン 消化器サポート 1kg', 2800, 1.0, 'excluded', 0.08, 'f', 'manual', NULL, 3, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(7, 4, 'examination', '初診料', 1500, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(8, 4, 'vaccine', '猫3種混合ワクチン', 3000, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(9, 5, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(10, 5, 'procedure', 'アトピー皮膚炎処置', 1500, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(11, 5, 'medicine', 'ステロイド外用薬 30g', 1200, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 3, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(12, 6, 'trimming', 'フルトリミング（猫）', 5000, 1.0, 'excluded', 0.10, 'f', 'manual', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(13, 6, 'food', 'ヒルズ サイエンスダイエット 猫用 400g', 1200, 1.0, 'excluded', 0.08, 'f', 'manual', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(14, 7, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(15, 7, 'test', '血液検査（CBC + 生化学）', 6000, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(16, 8, 'surgery', '避妊手術（猫）', 25000, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(17, 9, 'hotel', 'ペットホテル（1泊）', 2500, 2.0, 'excluded', 0.10, 'f', 'manual', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(18, 10, 'other', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(19, 10, 'medicine', '抗菌点眼薬（5ml）', 600, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(20, 11, 'vaccine', '狂犬病ワクチン接種', 2500, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(21, 12, 'medicine', '点耳薬（外耳炎用）', 1200, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(22, 13, 'food', '療法食 消化器サポート 3kg', 3000, 1.0, 'excluded', 0.08, 'f', 'manual', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(23, 14, 'examination', '再診料', 800, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(24, 14, 'medicine', '抗生剤・消炎剤 7日分', 4200, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(25, 15, 'test', '生化学検査パネル12項目', 8500, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(26, 16, 'vaccine', '犬5種混合ワクチン', 4200, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(27, 17, 'surgery', '歯科スケーリング・ポリッシング', 15000, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(28, 18, 'other', '爪切り・耳掃除・肛門腺', 2800, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(29, 19, 'test', '心臓超音波検査（エコー）', 6000, 1.0, 'excluded', 0.10, 't', 'medical_record', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(30, 20, 'other', '夜間緊急対応・時間外加算', 10000, 1.0, 'excluded', 0.10, 'f', 'manual', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(31, 301, 'examination', '再診料', 800, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(32, 302, 'examination', '初診料', 1900, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(33, 401, 'examination', '再診料', 800, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL),
(34, 402, 'examination', '再診料', 800, 1.0, 'excluded', 0.10, 'f', 'medical_record', NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL, NULL, NULL, NULL, NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- payments
-- -----------------------------------------------------------------------------
INSERT INTO payments ("id", "billing_id", "subtotal", "tax_total", "total_amount", "insurance_name", "insurance_ratio", "insurance_amount", "discount_amount", "billing_amount", "received_amount", "change_amount", "method", "payment_method_id", "paid_by", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 2, 6100, 554, 6654, 'アニコム損保', 0.70, 2541, 0, 4113, 4200, 87, 'credit_card', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 4, 4500, 450, 4950, '', 0.00, 0, 0, 4950, 5000, 50, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 5, 3500, 350, 3850, '', 0.00, 0, 0, 3850, 3850, 0, 'credit_card', 2, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 6, 6200, 596, 6796, '', 0.00, 0, 0, 6796, 7000, 204, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 7, 6800, 680, 7480, 'アニコム損保', 0.70, 5236, 0, 2244, 2244, 0, 'credit_card', 2, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(7, 8, 25000, 2500, 27500, 'アニコム損保', 0.70, 19250, 0, 8250, 8250, 0, 'cash', 4, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(8, 9, 5000, 500, 5500, '', 0.00, 0, 0, 5500, 5500, 0, 'electronic_money', 3, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(9, 10, 1400, 140, 1540, '', 0.00, 0, 0, 1540, 2000, 460, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 11, 2500, 250, 2750, '', 0.00, 0, 0, 2750, 3000, 250, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 12, 1200, 120, 1320, '', 0.00, 0, 0, 1320, 1320, 0, 'credit_card', 2, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(12, 13, 3000, 240, 3240, '', 0.00, 0, 0, 3240, 3240, 0, 'electronic_money', 3, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 14, 5000, 500, 5500, 'アイペット', 0.50, 2750, 0, 2750, 3000, 250, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 15, 8500, 850, 9350, '', 0.00, 0, 0, 9350, 9350, 0, 'credit_card', 2, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(15, 16, 4200, 420, 4620, '', 0.00, 0, 0, 4620, 5000, 380, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 17, 15000, 1500, 16500, '', 0.00, 0, 0, 16500, 16500, 0, 'credit_card', 2, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 18, 2800, 280, 3080, '', 0.00, 0, 0, 3080, 3080, 0, 'electronic_money', 3, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(18, 19, 6000, 600, 6600, 'アニコム損保', 0.70, 4620, 0, 1980, 2000, 20, 'cash', NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 301, 800, 80, 880, '', 0.00, 0, 0, 880, 1000, 120, 'cash', NULL, 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 302, 1900, 190, 2090, '', 0.00, 0, 0, 2090, 2090, 0, 'credit_card', NULL, 16, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(22, 401, 800, 80, 880, '', 0.00, 0, 0, 880, 880, 0, 'cash', NULL, 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 402, 800, 80, 880, '', 0.00, 0, 0, 880, 1000, 120, 'cash', NULL, 26, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- payment_splits
-- -----------------------------------------------------------------------------
INSERT INTO payment_splits ("id", "clinic_id", "billing_id", "method", "payment_method_id", "amount", "received_amount", "change_amount", "paid_by", "created_at") VALUES
(1, 1, 20, 'cash', NULL, 5000, 5000, 0, 1, '2026-05-31 04:33:17.574774+00'),
(2, 1, 20, 'credit_card', 2, 6000, 6000, 0, 1, '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- billing_refunds
-- -----------------------------------------------------------------------------
INSERT INTO billing_refunds ("id", "clinic_id", "billing_id", "amount", "reason", "refunded_by", "refunded_at", "created_at") VALUES
(1, 1, 1, 919, '処置内容の変更に伴う部分返金', 1, '2026-04-28 01:00:00+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 1, 500, '薬剤変更による差額返金', 1, '2026-05-02 05:30:00+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 2, 500, '診察キャンセル分の返金', 1, '2026-05-11 00:00:00+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- estimates
-- -----------------------------------------------------------------------------
INSERT INTO estimates ("id", "clinic_id", "estimate_no", "medical_record_id", "title", "owner_id", "status", "subtotal", "tax_total", "total_amount", "insurance_amount", "discount_amount", "valid_until", "comment", "notes", "created_by", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 'EST-2026-001', 64, '避妊手術お見積り', 9, 'approved', 25000, 2500, 27500, 0, 0, '2026-06-22', '本日実施分。', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 'EST-2026-002', 71, '歯科処置お見積り', 17, 'draft', 15000, 1500, 16500, 0, 0, '2026-06-22', '概算です。', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 'EST-2026-003', NULL, '骨折手術概算', 19, 'draft', 120000, 12000, 132000, 0, 0, '2026-06-01', '大型犬のため麻酔量増。', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 1, 'EST-2026-004', NULL, 'MRI検査概算（外部委託）', 5, 'sent', 40000, 4000, 44000, 0, 0, '2026-06-15', '大学病院での検査費用です。', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 'EST-2026-005', NULL, '腫瘍摘出手術', 12, 'rejected', 50000, 5000, 55000, 0, 0, '2026-05-10', '高齢のため見送り。', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 1, 'EST-2026-008', 61, '避妊手術・術前検査パック', 3, 'approved', 45000, 4500, 49500, 0, 0, '2026-06-15', '', '', NULL, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- estimate_items
-- -----------------------------------------------------------------------------
INSERT INTO estimate_items ("id", "estimate_id", "name", "category", "unit_price", "quantity", "tax_type", "tax_rate", "discount_rate", "discount_amount", "is_insurance_applicable", "consultation_id", "procedure_id", "medicine_id", "merchandise_item_id", "sort_order", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, '避妊手術（猫）', 'surgery', 25000, 1.0, 'excluded', 0.10, 0.00, 0, 't', NULL, 1, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 2, '歯科スケーリング', 'surgery', 15000, 1.0, 'excluded', 0.10, 0.00, 0, 'f', NULL, NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 3, '骨折手術・プレート固定', 'surgery', 100000, 1.0, 'excluded', 0.10, 0.00, 0, 't', NULL, NULL, NULL, NULL, 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 3, '全身麻酔（大型）', 'procedure', 20000, 1.0, 'excluded', 0.10, 0.00, 0, 't', NULL, NULL, NULL, NULL, 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 6, '避妊手術一式', 'surgery', 35000, 1.0, 'excluded', 0.10, 0.00, 0, 'f', NULL, NULL, NULL, NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(6, 6, '術前血液検査', 'test', 10000, 1.0, 'excluded', 0.10, 0.00, 0, 'f', NULL, NULL, NULL, NULL, 0, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- vital_records
-- -----------------------------------------------------------------------------
INSERT INTO vital_records ("id", "pet_id", "medical_record_id", "daily_record_id", "recorded_at", "staff_id", "temperature", "heart_rate", "respiration_rate", "weight", "weight_unit", "notes", "created_at", "updated_at", "deleted_at") VALUES
(1, 1, 3, NULL, '2026-04-01 00:15:00+00', 1, 38.5, 80, 20, 26.5, 'Kg', '皮膚の搔痒感あり。体重良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(2, 1, 2, NULL, '2026-02-24 01:00:00+00', 2, 38.8, 82, 22, 26.0, 'Kg', '体重前回比-500g', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(3, 1, 3, NULL, '2026-04-01 00:30:00+00', 1, 38.3, 78, 20, 26.5, 'Kg', '定期検診。皮膚搔痒感 軽快傾向。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(4, 2, 4, NULL, '2026-01-15 02:00:00+00', 1, 39.1, 95, 24, 15.2, 'Kg', '軽度脱水。CRT 2秒。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(5, 1, 5, NULL, '2025-11-25 05:30:00+00', 2, 38.2, 160, 30, 4200, 'g', '粘膜色やや蒼白。食欲低下継続。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(10, 1, 1, NULL, '2025-10-10 01:00:00+00', 1, 38.4, 75, 18, 25.8, 'Kg', '健康。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(11, 1, 1, NULL, '2025-12-20 00:00:00+00', 1, 38.6, 80, 20, 26.2, 'Kg', 'ワクチン時。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(13, 3, 5, NULL, '2025-11-25 01:00:00+00', 2, 38.8, 110, 25, 4.0, 'Kg', '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(14, 3, 6, NULL, '2026-03-18 01:00:00+00', 2, 38.7, 115, 24, 4.2, 'Kg', 'ワクチン時。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(16, 6, 8, NULL, '2025-11-20 01:00:00+00', 1, 38.5, 90, 22, 3.7, 'Kg', '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(17, 6, 8, NULL, '2026-03-12 01:00:00+00', 1, 38.6, 92, 22, 3.8, 'Kg', '定期。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(19, 11, 13, NULL, '2025-02-28 01:00:00+00', 2, 39.0, 130, 28, 4.5, 'Kg', '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(20, 11, 13, NULL, '2025-10-15 01:00:00+00', 2, 38.8, 125, 26, 4.7, 'Kg', '定期。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(21, 11, 13, NULL, '2026-03-14 01:00:00+00', 2, 38.7, 120, 25, 4.8, 'Kg', '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(23, 13, 15, NULL, '2026-04-25 01:00:00+00', 1, 38.5, 110, 24, 3.0, 'Kg', '良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(25, 10, 12, NULL, '2024-09-08 01:00:00+00', 1, 38.2, 70, 16, 18.0, 'Kg', '去年の。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(26, 10, 12, NULL, '2025-11-20 01:00:00+00', 1, 38.3, 72, 18, 18.2, 'Kg', '冬。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(27, 10, 12, NULL, '2026-03-12 01:00:00+00', 1, 38.5, 75, 20, 18.5, 'Kg', '春。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(30, 1, 61, 3, '2026-05-22 00:00:00+00', 1, 38.5, 78, 19, 26.4, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(31, 3, 61, 5, '2026-05-22 00:30:00+00', 2, 38.9, 120, 26, 4.3, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(32, 6, 62, NULL, '2026-05-22 01:30:00+00', 1, 38.4, 88, 20, 3.9, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(33, 11, 64, NULL, '2026-05-22 05:30:00+00', 2, 38.9, 130, 30, 4.6, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(34, 13, 65, NULL, '2026-05-22 07:30:00+00', 1, 38.7, 115, 25, 3.1, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(35, 10, 63, NULL, '2026-05-22 04:00:00+00', 1, 38.4, 73, 18, 18.4, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(40, 29, 21, NULL, '2026-01-20 01:00:00+00', 16, 38.4, 80, 20, 28.0, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(41, 29, 21, NULL, '2026-05-22 01:00:00+00', 16, 38.6, 82, 22, 28.5, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(42, 39, 29, NULL, '2025-12-15 00:00:00+00', 26, 38.3, 110, 24, 24.0, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL),
(43, 39, 29, NULL, '2026-05-22 00:00:00+00', 26, 38.5, 115, 25, 24.2, 'Kg', '', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00', NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- care_plan_items
-- -----------------------------------------------------------------------------
INSERT INTO care_plan_items ("id", "hospitalization_id", "type", "name", "description", "timing", "status", "notes", "medicine_id", "procedure_id", "hospitalization_plan_id", "unit_price", "category", "sort_order", "created_at", "updated_at") VALUES
(1, 1, 'food', '療法食（消化器ケア）', '1日3回、少量ずつ与える', '{morning,noon,night}', 'active', '鶏肉不可。', NULL, NULL, NULL, 0, '食事', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 1, 'medicine', 'アモキシシリン', '1回1錠、朝夕食後', '{morning,night}', 'active', '抗生剤。', 1, NULL, NULL, 500, '投薬', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 1, 'instruction', 'バイタルチェック', '1日3回測定', '{morning,noon,night}', 'active', '異常値報告。', NULL, NULL, NULL, 0, '観察', 3, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 2, 'treatment', '耳道洗浄', '1日1回、朝に実施', '{morning}', 'completed', '左耳。', NULL, 4, NULL, 2500, '処置', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 2, 'item', '入院管理料', '小型犬1日分', '{morning}', 'completed', '', NULL, NULL, 1, 3000, '入院', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 8, 'instruction', '酸素濃度管理', '25%維持', '{morning,noon,night}', 'active', '', NULL, NULL, NULL, 0, '観察', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 8, 'instruction', '術後創傷チェック', '1日2回', '{morning,night}', 'active', '', NULL, NULL, NULL, 0, '処置', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 10, 'instruction', '酸素室管理', '40%高濃度酸素', '{morning,noon,night}', 'active', '', NULL, NULL, NULL, 0, '観察', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 10, 'treatment', 'ネブライザー', '1日3回', '{morning,noon,night}', 'active', '', NULL, 5, NULL, 1500, '処置', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 11, 'instruction', '血糖値測定', '食前30分', '{morning,night}', 'active', '', NULL, NULL, NULL, 1000, '検査', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(11, 11, 'medicine', 'インスリン投与', '指定単位を皮下注射', '{morning,night}', 'active', '', 1, NULL, NULL, 500, '投薬', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(12, 13, 'instruction', '尿量確認', '都度記録', '{morning,noon,night}', 'active', '', NULL, NULL, NULL, 0, '観察', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(13, 13, 'medicine', '強肝剤点滴', '24時間持続点滴', '{morning}', 'active', '', 1, NULL, NULL, 2000, '投薬', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(14, 14, 'treatment', '創部洗浄', '1日1回', '{morning}', 'active', '', NULL, 4, NULL, 1000, '処置', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(15, 14, 'medicine', '抗生剤点滴', '1日2回', '{morning,night}', 'active', '', 1, NULL, NULL, 1000, '投薬', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- care_logs
-- -----------------------------------------------------------------------------
INSERT INTO care_logs ("id", "daily_record_id", "time", "type", "status", "value", "staff_id", "notes", "created_at", "updated_at") VALUES
(1, 3, '08:30:00', 'food', 'completed', '完食', 1, '朝食：消化器サポート缶', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 3, '12:00:00', 'other', 'completed', '15分', 1, '院内歩行。軽快。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 3, '16:00:00', 'excretion', 'completed', '普通量', 2, '便：良好。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 5, '09:00:00', 'food', 'completed', '半分残す', 1, 'ドライは食べない。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 5, '13:00:00', 'excretion', 'completed', '多量', 1, '尿：色は薄い。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 6, '10:00:00', 'food', 'completed', '完食', 2, '食欲旺盛。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 6, '15:00:00', 'other', 'completed', '10分', 2, '少しふらつきあり（運動）。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(8, 4, '08:00:00', 'food', 'completed', '完食', 1, '元気あり。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(9, 4, '17:00:00', 'excretion', 'completed', '普通', 2, '尿', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(10, 10, '11:00:00', 'food', 'completed', '採食なし', 1, '強制給餌検討。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(11, 1, '09:00:00', 'medicine', 'completed', 'アモキシ', 1, '朝の投薬完了。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(12, 1, '18:00:00', 'medicine', 'completed', 'アモキシ', 1, '夕の投薬完了。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(13, 2, '09:00:00', 'treatment', 'completed', '傷口洗浄', 1, '浸出液なし。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(14, 3, '20:00:00', 'other', 'completed', '就寝', 2, '消灯。', '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- staff_notes
-- -----------------------------------------------------------------------------
INSERT INTO staff_notes ("id", "daily_record_id", "time", "content", "staff_id", "created_at", "updated_at") VALUES
(1, 3, '10:00:00', '自力で少量の飲水を確認。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(2, 3, '14:30:00', '排尿あり。性状正常。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(3, 5, '09:00:00', '元気あり。鳴いてアピールしている。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(4, 5, '12:00:00', '昼食完食。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(5, 6, '18:00:00', '呼吸状態安定。酸素室継続。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(6, 3, '09:00:00', '【注意】処置時に噛み付こうとする仕草あり。エリザベスカラーまたはマズル必須。', 1, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00'),
(7, 6, '15:00:00', '多頭飼育環境のため、感染症対策に留意。', 2, '2026-05-31 04:33:17.574774+00', '2026-05-31 04:33:17.574774+00')
ON CONFLICT DO NOTHING;


-- =============================================================================
-- シーケンスリセット
-- =============================================================================

SELECT setval(pg_get_serial_sequence('closing_special_periods', 'id'), (SELECT COALESCE(MAX(id), 1) FROM closing_special_periods));
SELECT setval(pg_get_serial_sequence('lstep_trigger_priorities', 'id'), (SELECT COALESCE(MAX(id), 1) FROM lstep_trigger_priorities));
SELECT setval(pg_get_serial_sequence('reservation_type_unavailable_times', 'id'), (SELECT COALESCE(MAX(id), 1) FROM reservation_type_unavailable_times));
SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT COALESCE(MAX(id), 1) FROM owners));
SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT COALESCE(MAX(id), 1) FROM pets));
SELECT setval(pg_get_serial_sequence('pet_chronic_conditions', 'id'), (SELECT COALESCE(MAX(id), 1) FROM pet_chronic_conditions));
SELECT setval(pg_get_serial_sequence('line_customers', 'id'), (SELECT COALESCE(MAX(id), 1) FROM line_customers));
SELECT setval(pg_get_serial_sequence('lstep_tag_cache', 'id'), (SELECT COALESCE(MAX(id), 1) FROM lstep_tag_cache));
SELECT setval(pg_get_serial_sequence('lstep_delivery_trigger_log', 'id'), (SELECT COALESCE(MAX(id), 1) FROM lstep_delivery_trigger_log));
-- lstep_csv_imports.id is UUID, no sequence to reset
SELECT setval(pg_get_serial_sequence('lstep_friend_attribute_snapshots', 'id'), (SELECT COALESCE(MAX(id), 1) FROM lstep_friend_attribute_snapshots));
SELECT setval(pg_get_serial_sequence('line_send_logs', 'id'), (SELECT COALESCE(MAX(id), 1) FROM line_send_logs));
SELECT setval(pg_get_serial_sequence('staff_reservation_exclusions', 'id'), (SELECT COALESCE(MAX(id), 1) FROM staff_reservation_exclusions));
SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT COALESCE(MAX(id), 1) FROM appointments));
SELECT setval(pg_get_serial_sequence('appointment_trimming_details', 'id'), (SELECT COALESCE(MAX(id), 1) FROM appointment_trimming_details));
SELECT setval(pg_get_serial_sequence('appointment_trimming_options', 'id'), (SELECT COALESCE(MAX(id), 1) FROM appointment_trimming_options));
SELECT setval(pg_get_serial_sequence('hospitalizations', 'id'), (SELECT COALESCE(MAX(id), 1) FROM hospitalizations));
SELECT setval(pg_get_serial_sequence('daily_records', 'id'), (SELECT COALESCE(MAX(id), 1) FROM daily_records));
SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT COALESCE(MAX(id), 1) FROM medical_records));
SELECT setval(pg_get_serial_sequence('prescriptions', 'id'), (SELECT COALESCE(MAX(id), 1) FROM prescriptions));
SELECT setval(pg_get_serial_sequence('medical_record_addenda', 'id'), (SELECT COALESCE(MAX(id), 1) FROM medical_record_addenda));
SELECT setval(pg_get_serial_sequence('vaccinations', 'id'), (SELECT COALESCE(MAX(id), 1) FROM vaccinations));
SELECT setval(pg_get_serial_sequence('checkups', 'id'), (SELECT COALESCE(MAX(id), 1) FROM checkups));
SELECT setval(pg_get_serial_sequence('exams', 'id'), (SELECT COALESCE(MAX(id), 1) FROM exams));
SELECT setval(pg_get_serial_sequence('exam_results', 'id'), (SELECT COALESCE(MAX(id), 1) FROM exam_results));
SELECT setval(pg_get_serial_sequence('clinical_plans', 'id'), (SELECT COALESCE(MAX(id), 1) FROM clinical_plans));
SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT COALESCE(MAX(id), 1) FROM inquiries));
SELECT setval(pg_get_serial_sequence('treatments', 'id'), (SELECT COALESCE(MAX(id), 1) FROM treatments));
SELECT setval(pg_get_serial_sequence('treatment_plans', 'id'), (SELECT COALESCE(MAX(id), 1) FROM treatment_plans));
SELECT setval(pg_get_serial_sequence('medical_record_images', 'id'), (SELECT COALESCE(MAX(id), 1) FROM medical_record_images));
SELECT setval(pg_get_serial_sequence('billing_confirmations', 'id'), (SELECT COALESCE(MAX(id), 1) FROM billing_confirmations));
SELECT setval(pg_get_serial_sequence('billings', 'id'), (SELECT COALESCE(MAX(id), 1) FROM billings));
SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT COALESCE(MAX(id), 1) FROM billing_items));
SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT COALESCE(MAX(id), 1) FROM payments));
SELECT setval(pg_get_serial_sequence('payment_splits', 'id'), (SELECT COALESCE(MAX(id), 1) FROM payment_splits));
SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT COALESCE(MAX(id), 1) FROM billing_refunds));
SELECT setval(pg_get_serial_sequence('estimates', 'id'), (SELECT COALESCE(MAX(id), 1) FROM estimates));
SELECT setval(pg_get_serial_sequence('estimate_items', 'id'), (SELECT COALESCE(MAX(id), 1) FROM estimate_items));
SELECT setval(pg_get_serial_sequence('vital_records', 'id'), (SELECT COALESCE(MAX(id), 1) FROM vital_records));
SELECT setval(pg_get_serial_sequence('care_plan_items', 'id'), (SELECT COALESCE(MAX(id), 1) FROM care_plan_items));
SELECT setval(pg_get_serial_sequence('care_logs', 'id'), (SELECT COALESCE(MAX(id), 1) FROM care_logs));
SELECT setval(pg_get_serial_sequence('staff_notes', 'id'), (SELECT COALESCE(MAX(id), 1) FROM staff_notes));
