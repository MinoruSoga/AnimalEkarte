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
    (112, 1, '2026-05-11 16:45:00+00', '2026-05-11 17:45:00+00', 1, 1, 'revisit', 2, 1, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-11 08:33:53.134392+00', '2026-05-12 01:00:00.734922+00', NULL, NULL),
    (113, 1, '2026-05-11 16:30:00+00', '2026-05-11 17:30:00+00', 1, 1, 'revisit', 2, NULL, false, 'no_show', '', 'manual', 4, false, '{}', '2026-05-11 08:34:46.519909+00', '2026-05-12 01:00:00.785954+00', NULL, NULL),
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
