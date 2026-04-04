-- =============================================================================
-- Animal Ekarte - 初期スキーマ定義 v19.0
-- PostgreSQL 18
-- テーブル数: 56
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 拡張機能
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- -----------------------------------------------------------------------------
-- 2. ENUM型定義（全56テーブル対応）
-- -----------------------------------------------------------------------------

-- 認証関連

-- ペット関連
CREATE TYPE pet_status AS ENUM ('alive', 'deceased');
CREATE TYPE pet_gender AS ENUM ('male', 'female', 'unknown');
CREATE TYPE acquisition_type AS ENUM ('purchased', 'transferred', 'rescued', 'other');
CREATE TYPE danger_level AS ENUM ('low', 'medium', 'high');
CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred');

-- マスタ共通
CREATE TYPE staff_role AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager');
CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other');
CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock');
CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder');
CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram');
CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general');
CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE body_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night');
CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat');
CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'sedation', 'general');
CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both');

-- 電子カルテ関連
CREATE TYPE medical_record_status AS ENUM ('draft', 'finalized');
CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other');
CREATE TYPE treatment_status AS ENUM ('pending', 'completed', 'not_applicable');
CREATE TYPE examination_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed');
CREATE TYPE examination_result_status AS ENUM ('normal', 'high', 'low');
CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other');
CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none');
CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other');
CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected');
CREATE TYPE billing_review_status AS ENUM ('pending', 'confirmed', 'returned');
CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other');
CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization');

-- 予約・会計・入院関連
CREATE TYPE visit_type AS ENUM ('first', 'revisit');
CREATE TYPE reservation_status AS ENUM (
    'confirmed', 'pending', 'cancelled', 'checked_in',
    'in_consultation', 'accounting', 'completed'
);
CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending');
CREATE TYPE hospitalization_type AS ENUM ('hospitalization', 'hotel');
CREATE TYPE hospitalization_status AS ENUM ('admitted', 'discharged', 'reserved');
CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item');
CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued');
CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other');
CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped');
CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night');
CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g');

-- トリミング・シフト関連
CREATE TYPE trimming_status AS ENUM ('completed', 'reserved', 'in_progress');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money');
CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave');
CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt'); -- 内税, 外税, 非課税

-- -----------------------------------------------------------------------------
-- 3. テーブル定義（依存関係順）
-- -----------------------------------------------------------------------------

-- ==========================================================================
-- レイヤー1: 依存なし
-- ==========================================================================

-- ------------------------------------
-- 1. company（シングルトン: 本部情報）
-- ------------------------------------
CREATE TABLE company (
    id                  BIGSERIAL   PRIMARY KEY,
    name                text        NOT NULL,
    postal_code         text        NOT NULL DEFAULT '',
    address             text        NOT NULL DEFAULT '',
    phone_number        text        NOT NULL DEFAULT '',
    fax_number          text        NOT NULL DEFAULT '',
    registration_number text        NOT NULL DEFAULT '',
    invoice_registration_number text NOT NULL DEFAULT '',
    director_name       text        NOT NULL DEFAULT '',
    email               text        NOT NULL DEFAULT '',
    website             text        NOT NULL DEFAULT '',
    logo_url            text        NOT NULL DEFAULT '',
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 2. clinics（クリニック情報）
-- ------------------------------------
CREATE TABLE clinics (
    id                  BIGSERIAL   PRIMARY KEY,
    company_id          bigint      NOT NULL REFERENCES company(id),
    name                text        NOT NULL,
    postal_code         text        NOT NULL DEFAULT '',
    address             text        NOT NULL DEFAULT '',
    phone_number        text        NOT NULL DEFAULT '',
    fax_number          text        NOT NULL DEFAULT '',
    registration_number text        NOT NULL DEFAULT '',
    director_name       text        NOT NULL DEFAULT '',
    email               text        NOT NULL DEFAULT '',
    website             text        NOT NULL DEFAULT '',
    logo_url            text        NOT NULL DEFAULT '',
    is_active           boolean     NOT NULL DEFAULT true,
    standard_tax_rate   numeric     NOT NULL DEFAULT 0.10,
    reduced_tax_rate    numeric     NOT NULL DEFAULT 0.08,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

-- ==========================================================================
-- レイヤー2: clinics依存
-- ==========================================================================

-- ------------------------------------
-- 3. animal_species（ペット種類マスタ: システム共通）
-- ------------------------------------
CREATE TABLE animal_species (
    id         BIGSERIAL     PRIMARY KEY,
    name       text          NOT NULL,
    is_active  boolean       NOT NULL DEFAULT true,
    sort_order integer                DEFAULT 0,
    created_at timestamptz   NOT NULL DEFAULT now(),
    updated_at timestamptz   NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 4. job_titles（職種マスタ）
-- ------------------------------------
CREATE TABLE job_titles (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    sort_order  integer     NOT NULL DEFAULT 0,
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ------------------------------------
-- 5. accounts（認証用アカウント）
-- ------------------------------------
CREATE TABLE accounts (
    id             BIGSERIAL   PRIMARY KEY,
    email          text        NOT NULL UNIQUE,
    password_hash  text        NOT NULL,
    is_active      boolean     NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

CREATE INDEX idx_accounts_email ON accounts(email) WHERE deleted_at IS NULL;

-- ------------------------------------
-- 6. staffs（スタッフマスタ）
-- ------------------------------------
CREATE TABLE staffs (
    id             BIGSERIAL   PRIMARY KEY,
    account_id     bigint               REFERENCES accounts(id) ON DELETE SET NULL,
    name           text        NOT NULL,
    is_active      boolean     NOT NULL DEFAULT true,
    staff_role     staff_role  NOT NULL,
    license_number text        NOT NULL DEFAULT '',
    job_title_id   bigint               REFERENCES job_titles(id) ON DELETE SET NULL,
    sort_order     integer              DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz
);

CREATE INDEX idx_staffs_account ON staffs(account_id);

-- ------------------------------------
-- 7. staff_clinic_assignments（スタッフ-クリニック中間テーブル）
-- ------------------------------------
CREATE TABLE staff_clinic_assignments (
    id             BIGSERIAL   PRIMARY KEY,
    staff_id       bigint      NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    is_main        boolean     NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_staff_clinic UNIQUE (staff_id, clinic_id)
);

CREATE INDEX idx_staff_clinic_staff ON staff_clinic_assignments(staff_id);
CREATE INDEX idx_staff_clinic_clinic ON staff_clinic_assignments(clinic_id);
CREATE INDEX idx_staff_clinic_main ON staff_clinic_assignments(staff_id, is_main);

CREATE INDEX idx_audit_logs_clinic   ON audit_logs(clinic_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor    ON audit_logs(actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource, resource_id, created_at DESC);

COMMENT ON TABLE audit_logs IS '権限変更・認証操作の監査ログ（削除禁止）';


-- =============================================================================
-- Animal Ekarte - 統合シード v20.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING
-- 内容: マスタデータ + デモアカウント
-- 依存: 001_init.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. company（本部情報: 1件）
-- -----------------------------------------------------------------------------
INSERT INTO company (name) VALUES
    ('ノア動物病院')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 2. clinics（クリニック: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO clinics (id, company_id, name) VALUES
    (3, 1, '八王子院'),
    (4, 1, '城東医院'),
    (5, 1, '敷島医院')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('clinics', 'id'), (SELECT MAX(id) FROM clinics));

-- -----------------------------------------------------------------------------
-- 3. animal_species（ペット種類: 6件、システム共通・clinic_idなし）
-- -----------------------------------------------------------------------------
INSERT INTO animal_species (id, name, is_active, sort_order) VALUES
    (1, '犬',         true, 1),
    (2, '猫',         true, 2),
    (3, '鳥',         true, 3),
    (4, 'うさぎ',     true, 4),
    (5, 'ハムスター', true, 5),
    (6, 'その他',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('animal_species', 'id'), (SELECT MAX(id) FROM animal_species));

-- -----------------------------------------------------------------------------
-- 4. job_titles（職種: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO job_titles (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 3, '獣医師',     true, 1),
    (2, 3, '動物看護師', true, 2),
    (3, 3, 'トリマー',   true, 3),
    (4, 3, '受付',       true, 4),
    (5, 3, '管理者',     true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('job_titles', 'id'), (SELECT MAX(id) FROM job_titles));

-- -----------------------------------------------------------------------------
-- 5. staffs（スタッフ: 12件）
-- -----------------------------------------------------------------------------
INSERT INTO staffs (id, clinic_id, name, is_active, staff_role, license_number, job_title_id, sort_order) VALUES
    (1,  3, '山田 太郎',   true, 'veterinarian', 'V-10001', 1, 1),
    (2,  3, '高橋 健一',   true, 'veterinarian', 'V-10002', 1, 2),
    (3,  3, '渡辺 博',     true, 'manager',      '',        5, 3),
    (4,  3, '佐藤 花子',   true, 'nurse',        '',        2, 4),
    (5,  3, '伊藤 さくら', true, 'nurse',        '',        2, 5),
    (6,  3, '木村 健太',   true, 'trimmer',      '',        3, 6),
    (7,  3, '田中 美咲',   true, 'reception',    '',        4, 7),
    -- デモアカウント用スタッフ（八王子院）
    (8,  3, '田中 太郎',   true, 'veterinarian', 'V-20001', 1, 1),
    (9,  3, '山田 花子',   true, 'veterinarian', 'V-20002', 1, 2),
    (10, 3, '佐藤 美咲',   true, 'nurse',        '',        2, 3),
    (11, 3, '鈴木 一郎',   true, 'reception',    '',        4, 4),
    (12, 3, '高橋 さくら', true, 'trimmer',      '',        3, 5),
    -- 管理者・執行グループ デモアカウント用スタッフ
    (13, 3, '渡辺 院長',   true, 'manager',      '',        5, 6),
    (14, 3, '小林 部長',   true, 'manager',      '',        5, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- -----------------------------------------------------------------------------
-- 6. user_accounts（ユーザーアカウント: 9件）
-- password_hash: bcrypt("password", cost=10)
-- -----------------------------------------------------------------------------
INSERT INTO user_accounts (id, email, display_name, display_name_kana, user_type, job_title_id, status, staff_id, password_hash) VALUES
    -- 渋谷院・新宿院スタッフ
    (1, 'admin@noavet.jp',   'システム管理者', 'システムカンリシャ',   'system_admin', 5, 'active', 3,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (2, 'clinic1@noavet.jp', '渋谷院管理者',   'シブヤインカンリシャ', 'clinic_admin', 5, 'active', NULL, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (3, 'yamada@noavet.jp',  '山田 太郎',      'ヤマダ タロウ',        'staff',        1, 'active', 1,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    -- デモアカウント（八王子院・frontend mock-data.ts 対応）
    (4, 'admin@example.com',     '田中 太郎',  'タナカ タロウ',    'clinic_admin', 1, 'active', 8,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (5, 'vet@example.com',       '山田 花子',  'ヤマダ ハナコ',    'staff',        1, 'active', 9,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (6, 'nurse@example.com',     '佐藤 美咲',  'サトウ ミサキ',    'staff',        2, 'active', 10,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (7, 'reception@example.com', '鈴木 一郎',  'スズキ イチロウ',  'staff',        4, 'active', 11,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (8, 'trimmer@example.com',   '高橋 さくら','タカハシ サクラ',  'staff',        3, 'active', 12,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (9,  'system@example.com',   '本部 管理者', 'ホンブ カンリシャ','system_admin', 5, 'active', NULL, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    -- 管理者・執行グループ デモアカウント
    (10, 'manager@example.com',  '渡辺 院長',  'ワタナベ インチョウ','staff',        5, 'active', 13, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (11, 'exec@example.com',     '小林 部長',  'コバヤシ ブチョウ', 'staff',        5, 'active', 14, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('user_accounts', 'id'), (SELECT MAX(id) FROM user_accounts));

-- -----------------------------------------------------------------------------
-- 7. user_clinic_memberships（ユーザー所属クリニック: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO user_clinic_memberships (id, user_id, clinic_id, is_main) VALUES
    -- デモアカウント（system=本部管理者: 全3院、他: 八王子院のみ）
    (5,  4, 3, true),
    (6,  5, 3, true),
    (7,  6, 3, true),
    (8,  7, 3, true),
    (9,  8, 3, true),
    (10, 9, 3, true),
    (11, 9, 4, false),
    (12, 9, 5, false),
    -- 管理者・執行グループ ユーザー（八王子院）
    (13, 10, 3, true),
    (14, 11, 3, true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('user_clinic_memberships', 'id'), (SELECT MAX(id) FROM user_clinic_memberships));

-- -----------------------------------------------------------------------------
-- 7b. permission_groups（権限グループ）& user_permission_groups（割当）
-- -----------------------------------------------------------------------------
-- ノア動物病院 (company_id=1): 管理者・執行・一般 の3グループ
INSERT INTO permission_groups (id, company_id, name, description, color) VALUES
    (1, 1, '管理者', '全機能フルアクセス・権限設定管理', '#EF4444'),
    (2, 1, '執行',   '業務全般閲覧・権限設定変更',       '#6366F1'),
    (3, 1, '一般',   '基本的な業務操作',                 '#10B981')
ON CONFLICT DO NOTHING;

-- グループルール（管理者: 全リソースフルアクセス）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (1, 'dashboard',        true, false, false, false),
    (1, 'owners',           true, true,  true,  true),
    (1, 'reservations',     true, true,  true,  true),
    (1, 'medical-records',  true, true,  true,  true),
    (1, 'hospitalization',  true, true,  true,  true),
    (1, 'trimming',         true, true,  true,  true),
    (1, 'examinations',     true, true,  true,  true),
    (1, 'accounting',       true, true,  true,  true),
    (1, 'vaccinations',     true, true,  true,  true),
    (1, 'checkups',         true, true,  true,  true),
    (1, 'inventory',        true, true,  true,  true),
    (1, 'estimates',        true, true,  true,  true),
    (1, 'shifts',           true, true,  true,  true),
    (1, 'master',           true, true,  true,  true),
    (1, 'hospital-settings',true, true,  true,  true)
ON CONFLICT DO NOTHING;

-- グループルール（執行: 業務全般閲覧＋権限設定変更）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (2, 'dashboard',        true, false, false, false),
    (2, 'owners',           true, true,  true,  false),
    (2, 'reservations',     true, true,  true,  false),
    (2, 'medical-records',  true, false, false, false),
    (2, 'hospitalization',  true, true,  true,  false),
    (2, 'trimming',         true, false, false, false),
    (2, 'examinations',     true, false, false, false),
    (2, 'accounting',       true, true,  true,  false),
    (2, 'vaccinations',     true, false, false, false),
    (2, 'checkups',         true, false, false, false),
    (2, 'inventory',        true, true,  true,  false),
    (2, 'estimates',        true, true,  true,  false),
    (2, 'shifts',           true, true,  true,  false),
    (2, 'master',           true, true,  true,  false),
    (2, 'hospital-settings',true, false, false, false)
ON CONFLICT DO NOTHING;

-- グループルール（一般: 基本的な業務操作）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (3, 'dashboard',        true, false, false, false),
    (3, 'owners',           true, true,  true,  false),
    (3, 'reservations',     true, true,  true,  false),
    (3, 'medical-records',  true, true,  true,  false),
    (3, 'hospitalization',  true, true,  true,  false),
    (3, 'trimming',         true, true,  true,  false),
    (3, 'examinations',     true, true,  true,  false),
    (3, 'accounting',       true, false, false, false),
    (3, 'vaccinations',     true, true,  true,  false),
    (3, 'checkups',         true, false, false, false),
    (3, 'inventory',        true, false, false, false),
    (3, 'estimates',        true, false, false, false),
    (3, 'shifts',           true, true,  true,  false),
    (3, 'master',           true, false, false, false),
    (3, 'hospital-settings',true, false, false, false)
ON CONFLICT DO NOTHING;

-- ユーザーへのグループ割当
-- manager@example.com (user_id=10) → 管理者
INSERT INTO user_permission_groups (user_id, group_id) VALUES (10, 1) ON CONFLICT DO NOTHING;
-- exec@example.com (user_id=11) → 執行
INSERT INTO user_permission_groups (user_id, group_id) VALUES (11, 2) ON CONFLICT DO NOTHING;
-- vet@example.com (user_id=5) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (5, 3) ON CONFLICT DO NOTHING;
-- nurse@example.com (user_id=6) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (6, 3) ON CONFLICT DO NOTHING;
-- reception@example.com (user_id=7) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (7, 3) ON CONFLICT DO NOTHING;
-- trimmer@example.com (user_id=8) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (8, 3) ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), (SELECT MAX(id) FROM permission_groups));

-- -----------------------------------------------------------------------------
-- 8. service_types（サービス種別: 7件）
-- 「再診」は visit_type（予約区分: 初診/再診）で管理するため service_types には含めない
-- -----------------------------------------------------------------------------
INSERT INTO service_types (id, clinic_id, name, is_active, description, color, sort_order) VALUES
    (1, 3, '一般診療',     true, '内科・外科・皮膚科などの一般的な診療', '#3B82F6', 1),
    (2, 3, 'ワクチン接種', true, '各種ワクチン接種（予防接種）',         '#10B981', 2),
    (3, 3, '健康診断',     true, '定期健康診断・フィラリア検査など',     '#8B5CF6', 3),
    (4, 3, '手術・処置',   true, '去勢・避妊・その他外科手術',           '#EF4444', 4),
    (5, 3, 'トリミング',   true, 'グルーミング・爪切り・耳掃除など',     '#F59E0B', 5),
    (6, 3, '入院',         true, '入院・ホテル管理',                     '#6B7280', 6),
    (7, 3, '検査',         true, '血液検査・尿検査・画像診断など',       '#EC4899', 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('service_types', 'id'), (SELECT MAX(id) FROM service_types));

-- -----------------------------------------------------------------------------
-- 9. cages（ケージ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (1, 3, 'ICUケージA',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 1),
    (2, 3, 'ICUケージB',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 2),
    (3, 3, '犬用ケージ（小）', 3000, true, '小型犬・ホテル利用可',    'dog',     'small',  3),
    (4, 3, '犬用ケージ（中）', 3500, true, '中型犬・一般入院用',      'dog',     'medium', 4),
    (5, 3, '犬用ケージ（大）', 4000, true, '大型犬・術後管理用',      'dog',     'large',  5),
    (6, 3, '猫用ケージ（小）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'small',  6),
    (7, 3, '猫用ケージ（中）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'medium', 7),
    (8, 3, '汎用ケージA',     2500, true, '小動物・鳥類等対応',      'general', 'small',  8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('cages', 'id'), (SELECT MAX(id) FROM cages));

-- -----------------------------------------------------------------------------
-- 10. insurances（保険: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (1, 3, 'アニコム損保',         true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034',  1),
    (2, 3, 'アイペット損保',       true, 'うちの子シリーズ',                     70, '0120-956-099',  2),
    (3, 3, 'ペット&ファミリー',     true, 'げんきナンバーワンシリーズ',           80, '0120-81-8505',  3),
    (4, 3, '楽天ペット保険',       true, '楽天が提供するペット保険',             70, '0120-600-810',  4),
    (5, 3, 'その他（自費）',       true, '保険未加入・全額自費',                100, '',              5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('insurances', 'id'), (SELECT MAX(id) FROM insurances));

-- -----------------------------------------------------------------------------
-- 11. exam_types（検査種別: 5件）+ exam_type_items（検査項目）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1, 3, '血液検査（CBC）',     3000, true, '全血球計算（Complete Blood Count）',         1),
    (2, 3, '血液化学検査',         5000, true, '肝機能・腎機能・血糖値など生化学的検査',     2),
    (3, 3, '尿検査',               1500, true, '尿試験紙・尿沈渣検査',                       3),
    (4, 3, 'レントゲン検査',       3000, true, 'X線撮影（胸部・腹部・四肢）',                4),
    (5, 3, '超音波検査（エコー）', 5000, true, '腹部エコー・心臓エコー',                     5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- exam_type_items: 血液検査（CBC）
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (1, 1, 'WBC（白血球数）',      '', '6.0-17.0 x10^3/uL', 1),
    (2, 1, 'RBC（赤血球数）',      '', '5.5-8.5 x10^6/uL',  2),
    (3, 1, 'HCT（ヘマトクリット）', '', '37-55%',            3),
    (4, 1, 'PLT（血小板数）',      '', '175-500 x10^3/uL',  4)
ON CONFLICT DO NOTHING;

-- exam_type_items: 血液化学検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (5, 2, 'ALT（GPT）',        '', '10-125 U/L',    1),
    (6, 2, 'BUN（尿素窒素）',   '', '7-27 mg/dL',    2),
    (7, 2, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL', 3),
    (8, 2, 'GLU（血糖値）',     '', '74-143 mg/dL',   4)
ON CONFLICT DO NOTHING;

-- exam_type_items: 尿検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (9,  3, '尿比重',   '', '1.015-1.045', 1),
    (10, 3, '尿pH',     '', '5.5-7.5',     2),
    (11, 3, '尿タンパク', '', '陰性',       3),
    (12, 3, '尿潜血',   '', '陰性',        4)
ON CONFLICT DO NOTHING;

-- exam_type_items: レントゲン検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (13, 4, '胸部正面', '', '異常なし', 1),
    (14, 4, '腹部正面', '', '異常なし', 2),
    (15, 4, '四肢',     '', '異常なし', 3)
ON CONFLICT DO NOTHING;

-- exam_type_items: 超音波検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (16, 5, '腹部エコー', '', '異常なし', 1),
    (17, 5, '心臓エコー', '', '異常なし', 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_items', 'id'), (SELECT MAX(id) FROM exam_type_items));

-- -----------------------------------------------------------------------------
-- 12. vaccines（ワクチン: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (1,  3, '混合ワクチン5種（犬）',      4500, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ',         'dog', '1年',   1),
    (2,  3, '混合ワクチン6種（犬）',      5500, true, '5種＋コロナウイルス',                                                    'dog', '1年',   2),
    (3,  3, '混合ワクチン8種（犬）',      6500, true, '5種＋レプトスピラ3種',                                                   'dog', '1年',   3),
    (4,  3, '混合ワクチン10種（犬）',     8000, true, '5種＋レプトスピラ5種',                                                   'dog', '1年',   4),
    (5,  3, '混合ワクチン3種（猫）',      4000, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',                     'cat', '1年',   5),
    (6,  3, '混合ワクチン5種（猫）',      5500, true, '3種＋猫白血病・猫クラミジア',                                             'cat', '1年',   6),
    (7,  3, '狂犬病ワクチン',             3000, true, '狂犬病予防法に基づく接種',                                               'dog', '1年',   7),
    (8,  3, 'フィラリア予防薬（小型犬）',  900, true, '体重10kg以下犬用フィラリア予防',                                          'dog', '1ヶ月', 8),
    (9,  3, 'フィラリア予防薬（中型犬）', 1100, true, '体重11-25kg犬用フィラリア予防',                                           'dog', '1ヶ月', 9),
    (10, 3, 'フィラリア予防薬（大型犬）', 1500, true, '体重26kg以上犬用フィラリア予防',                                          'dog', '1ヶ月', 10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccines', 'id'), (SELECT MAX(id) FROM vaccines));

-- -----------------------------------------------------------------------------
-- 13. medicines（薬剤カテゴリ: 9件 + 薬剤: 15件）
-- カテゴリレコードは id 1001〜1009、price=NULL、parent_id=NULL
-- 薬剤レコードは parent_id でカテゴリを参照
-- -----------------------------------------------------------------------------

-- カテゴリレコード（id 1001〜1009）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1001, 3, '抗生剤',       NULL, true, '抗生物質カテゴリ',   1),
    (1002, 3, 'ステロイド',   NULL, true, 'ステロイド剤カテゴリ', 2),
    (1003, 3, '利尿剤',       NULL, true, '利尿剤カテゴリ',     3),
    (1004, 3, '消炎剤',       NULL, true, '消炎鎮痛剤カテゴリ', 4),
    (1005, 3, '神経系薬',     NULL, true, '神経系薬カテゴリ',   5),
    (1006, 3, '制吐剤',       NULL, true, '制吐剤カテゴリ',     6),
    (1007, 3, '消化器用薬',   NULL, true, '消化器用薬カテゴリ', 7),
    (1008, 3, '駆虫剤',       NULL, true, '駆虫剤カテゴリ',     8),
    (1009, 3, '輸液',         NULL, true, '輸液カテゴリ',       9)
ON CONFLICT DO NOTHING;

-- 薬剤レコード（parent_id でカテゴリ参照）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (1,  3, 'アモキシシリン 50mg',         500,  true, '広域スペクトラム抗生物質',               'tablet',    'per_tablet', 1,   1, 1001),
    (2,  3, 'メトロニダゾール 250mg',       600,  true, '嫌気性菌・原虫感染症治療薬',             'tablet',    'per_tablet', 1,   2, 1001),
    (3,  3, 'プレドニゾロン 5mg',           400,  true, 'ステロイド系抗炎症・免疫抑制剤',         'tablet',    'per_tablet', 1,   3, 1002),
    (4,  3, 'フロセミド注射液 20mg/2ml',    800,  true, '利尿剤（心臓・腎臓病の浮腫治療）',       'injection', 'per_ml',     2,   4, 1003),
    (5,  3, 'メロキシカム経口液',           700,  true, 'NSAIDs・痛み・炎症の緩和',               'liquid',    'per_ml',     1,   5, 1004),
    (6,  3, 'ガバペンチン 100mg',           550,  true, '神経因性疼痛・てんかん補助療法',         'tablet',    'per_tablet', 1,   6, 1005),
    (7,  3, 'マロピタント 16mg',            800,  true, '制吐剤（乗り物酔い・嘔吐治療）',         'tablet',    'per_tablet', 1,   7, 1006),
    (8,  3, 'ラクツロース液',               500,  true, '便秘・肝性脳症の治療',                   'liquid',    'per_ml',     5,   8, 1007),
    (9,  3, 'ノミ・ダニ駆除薬（犬用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,   9, 1008),
    (10, 3, 'ノミ・ダニ駆除薬（猫用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,  10, 1008),
    (11, 3, '抗生剤点眼薬',                 600,  true, '眼科用抗菌点眼剤',                       'liquid',    'per_ml',     1,  11, 1001),
    (12, 3, 'デキサメタゾン注射液',         700,  true, '強力ステロイド・アレルギー緊急治療',     'injection', 'per_ml',     1,  12, 1002),
    (13, 3, '生理食塩水 500ml',             400,  true, '点滴・洗浄用生理食塩水',                 'liquid',    'per_ml',     500, 13, 1009),
    (14, 3, 'セファレキシン 250mg',         450,  true, '第1世代セフェム系抗生物質',              'tablet',    'per_tablet', 1,  14, 1001),
    (15, 3, 'オメプラゾール 10mg',          350,  true, 'プロトンポンプ阻害薬（胃酸抑制）',       'tablet',    'per_tablet', 1,  15, 1007)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('medicines', 'id'), (SELECT MAX(id) FROM medicines));

-- -----------------------------------------------------------------------------
-- 14. consultations（診察項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (1, 3, '初診料',       2000, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit',  30, 1),
    (2, 3, '再診料',        800, true, '継続通院の診察料',                         'revisit',      15, 2),
    (3, 3, '往診料',       5000, true, '自宅への往診料（基本料金）',               'anytime',      60, 3),
    (4, 3, '時間外診療料', 3000, true, '診療時間外・休日の緊急診察',               'after_hours',  30, 4),
    (5, 3, '電話相談料',    500, true, '電話による診察相談',                       'anytime',      15, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('consultations', 'id'), (SELECT MAX(id) FROM consultations));

-- -----------------------------------------------------------------------------
-- 15. procedures（処置項目: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (1,  3, '去勢手術（犬）',   25000, true, '雄犬の去勢手術',                   60,  'general', 1),
    (2,  3, '避妊手術（猫）',   25000, true, '雌猫の避妊手術',                   90,  'general', 2),
    (3,  3, '歯石除去',         15000, true, '全身麻酔下での歯石除去・歯周治療', 45,  'general', 3),
    (4,  3, '耳洗浄',            2500, true, '外耳炎治療・耳道内の洗浄処置',     15,  'none',    4),
    (5,  3, '爪切り',             500, true, '爪のカット・やすりがけ',           10,  'none',    5),
    (6,  3, '皮膚縫合',          5000, true, '裂傷・切傷の縫合処置',             30,  'local',   6),
    (7,  3, '骨折整復',         80000, true, '骨折の外科的整復・固定',          120,  'general', 7),
    (8,  3, '腫瘍摘出',         20000, true, '皮膚腫瘍の外科的摘出',             60,  'local',   8),
    (9,  3, '胃洗浄',           10000, true, '異物誤飲時の胃洗浄処置',           30,  'general', 9),
    (10, 3, '点滴処置',          3000, true, '静脈内点滴（1時間）',              60,  'none',   10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('procedures', 'id'), (SELECT MAX(id) FROM procedures));

-- -----------------------------------------------------------------------------
-- 16. hospitalization_plans（入院プラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (1, 3, '一般入院（小型）', 3000, true, '体重10kg以下の入院管理料（1日）',  'small',  'per_day',   1),
    (2, 3, '一般入院（中型）', 3500, true, '体重10-25kgの入院管理料（1日）',   'medium', 'per_day',   2),
    (3, 3, '一般入院（大型）', 4500, true, '体重25kg以上の入院管理料（1日）',  'large',  'per_day',   3),
    (4, 3, 'ICU入院',          8000, true, '集中治療室管理料（1日）',          'small',  'per_day',   4),
    (5, 3, 'ホテル（小型）',   2500, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('hospitalization_plans', 'id'), (SELECT MAX(id) FROM hospitalization_plans));

-- -----------------------------------------------------------------------------
-- 17. trimming_courses（トリミングコース: 5件）
-- ※ duration は integer (分)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (1, 3, 'シャンプー&ブロー（小型）', 4000,  true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (2, 3, 'シャンプー&ブロー（中型）', 5500,  true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  2),
    (3, 3, 'フルコース（小型）',        7000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 3),
    (4, 3, 'フルコース（中型）',        9000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 4),
    (5, 3, 'フルコース（大型）',        12000, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'large',  180, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_courses', 'id'), (SELECT MAX(id) FROM trimming_courses));

-- -----------------------------------------------------------------------------
-- 18. trimming_options（トリミングオプション: 5件）
-- ※ combinable は boolean, duration は integer（分単位）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, combinable, sort_order) VALUES
    (1, 3, '爪切り',     300, true, '爪のカット・やすりがけ',       10, true, 1),
    (2, 3, '耳掃除',     500, true, '外耳道の洗浄・清掃',           10, true, 2),
    (3, 3, '歯磨き',     500, true, '歯ブラシによるデンタルケア',   15, true, 3),
    (4, 3, '肛門腺絞り', 300, true, '肛門嚢の分泌液除去',            5, true, 4),
    (5, 3, 'リボン装着', 200, true, '仕上げのアクセサリー装着',      5, true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- 19. diagnosis_categories（診断カテゴリ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_categories (id, clinic_id, name, is_active, description, sort_order) VALUES
    (1, 3, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患',   1),
    (2, 3, '呼吸器系',       true, '肺・気管・鼻腔などの呼吸器系疾患',     2),
    (3, 3, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',     3),
    (4, 3, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患',   4),
    (5, 3, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   5),
    (6, 3, '感染症・寄生虫', true, '細菌・ウイルス・寄生虫感染症',         6),
    (7, 3, '腫瘍',           true, '良性・悪性腫瘍（がん）',               7),
    (8, 3, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',             8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_categories', 'id'), (SELECT MAX(id) FROM diagnosis_categories));

-- -----------------------------------------------------------------------------
-- 20. diagnosis_names（診断名: 各カテゴリ2-3件、計20件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_category_id, sort_order) VALUES
    -- 消化器系
    (1,  3, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         1, 1),
    (2,  3, '膵炎',               true, '膵臓の炎症',                         1, 2),
    (3,  3, '肝疾患',             true, '肝炎・肝不全・脂肪肝など',           1, 3),
    -- 呼吸器系
    (4,  3, '気管支炎',           true, '気管支の炎症',                       2, 1),
    (5,  3, '肺炎',               true, '肺の感染性・非感染性炎症',           2, 2),
    -- 皮膚・被毛
    (6,  3, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 3, 1),
    (7,  3, '膿皮症',             true, '細菌性の皮膚感染症',                 3, 2),
    (8,  3, '真菌症',             true, '皮膚糸状菌による感染症',             3, 3),
    -- 泌尿器系
    (9,  3, '膀胱炎',             true, '細菌性・特発性膀胱炎',               4, 1),
    (10, 3, '腎不全',             true, '急性・慢性腎不全',                   4, 2),
    (11, 3, '尿路結石',           true, '腎結石・膀胱結石・尿道結石',         4, 3),
    -- 神経系
    (12, 3, 'てんかん',           true, '反復性の痙攣発作',                   5, 1),
    (13, 3, '椎間板ヘルニア',     true, '頸椎・腰椎の椎間板突出',             5, 2),
    -- 感染症・寄生虫
    (14, 3, 'パルボウイルス感染症', true, '犬パルボウイルスによる感染症',       6, 1),
    (15, 3, 'フィラリア症',       true, '犬糸状虫による心肺疾患',             6, 2),
    (16, 3, '猫風邪（FVR）',      true, '猫ウイルス性鼻気管炎',               6, 3),
    -- 腫瘍
    (17, 3, '肥満細胞腫',         true, '皮膚または内臓の肥満細胞腫瘍',       7, 1),
    (18, 3, 'リンパ腫',           true, '悪性リンパ腫',                       7, 2),
    -- 外傷・骨格
    (19, 3, '骨折',               true, '各部位の骨折',                       8, 1),
    (20, 3, '咬傷',               true, '他動物による咬傷・咬傷感染',         8, 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_names', 'id'), (SELECT MAX(id) FROM diagnosis_names));

-- -----------------------------------------------------------------------------
-- 21. checkup_types（健診種別: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (1, 3, '一般健診',       5000,  true, '身体検査・体重測定・問診',                     '1年',   '全年齢', 1),
    (2, 3, '老齢検診',       15000, true, '身体検査＋血液検査＋レントゲン＋超音波',         '6ヶ月', '7歳以上', 2),
    (3, 3, 'フィラリア検査', 2500,  true, 'フィラリア抗原検査（予防シーズン前）',           '1年',   '成犬',   3),
    (4, 3, '歯科検診',       3000,  true, '歯周病チェック・歯石付着度の確認',             '1年',   '成犬',   4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('checkup_types', 'id'), (SELECT MAX(id) FROM checkup_types));

-- -----------------------------------------------------------------------------
-- 22. chief_complaint_categories（主訴区分: 6件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_categories (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 3, '食欲不振',       true, 1),
    (2, 3, '嘔吐・下痢',     true, 2),
    (3, 3, '皮膚・被毛異常', true, 3),
    (4, 3, '呼吸困難',       true, 4),
    (5, 3, '排尿・排泄異常', true, 5),
    (6, 3, '外傷・骨折',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_categories', 'id'), (SELECT MAX(id) FROM chief_complaint_categories));

-- -----------------------------------------------------------------------------
-- 23. inquiry_templates（問診定型文: 10件）
-- ※ category は text 型: chief_complaint / history / current_medications / notes
-- -----------------------------------------------------------------------------
INSERT INTO inquiry_templates (id, clinic_id, category, title, content, is_active, sort_order) VALUES
    (1,  3, 'chief_complaint',    '食欲不振（急性）',           'いつ頃から食欲が落ちましたか？完全に食べないのか、減っているだけか確認してください。', true, 1),
    (2,  3, 'chief_complaint',    '嘔吐（回数・内容物）',       '嘔吐の回数、内容物（食物・胆汁・血液など）、嘔吐のタイミング（食後すぐ・空腹時）を確認してください。', true, 2),
    (3,  3, 'chief_complaint',    '下痢（性状・頻度）',         '便の性状（軟便・水様便・血便・粘液便）、排便頻度、いつから続いているか確認してください。', true, 3),
    (4,  3, 'chief_complaint',    '皮膚の痒み・発赤',           '痒がる部位、発症時期、季節性の有無、ノミ・ダニ予防の状況を確認してください。', true, 4),
    (5,  3, 'chief_complaint',    '排尿異常',                   '排尿回数の変化、尿の色・量、排尿時の痛みの有無、血尿の有無を確認してください。', true, 5),
    (6,  3, 'history',            '既往歴確認（手術歴）',       '過去の手術歴（去勢・避妊含む）、入院歴、重大な疾患の既往を確認してください。', true, 6),
    (7,  3, 'history',            '予防接種歴確認',             '最終ワクチン接種日、狂犬病予防接種の有無、フィラリア予防の状況を確認してください。', true, 7),
    (8,  3, 'current_medications', '現在の投薬状況',            '現在服用中の薬剤名、用量、投与期間、処方元の病院を確認してください。', true, 8),
    (9,  3, 'current_medications', 'サプリメント・フード',      '現在与えているサプリメント、療法食、おやつの種類を確認してください。', true, 9),
    (10, 3, 'notes',              '生活環境確認',               '室内飼い/外飼い、同居動物の有無、散歩の頻度・時間を確認してください。', true, 10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inquiry_templates', 'id'), (SELECT MAX(id) FROM inquiry_templates));

-- -----------------------------------------------------------------------------
-- 24. inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (1, 3, 'フロントライン プラス（犬用）',     'medicine',   50, '本',    10, '薬品棚A',   'メリアルジャパン',         'sufficient'),
    (2, 3, 'ネクスガード チュアブル',            'medicine',   30, '錠',    10, '薬品棚A',   'ベーリンガーインゲルハイム', 'sufficient'),
    (3, 3, 'ヒルズ i/d（犬用・消化器ケア）',     'food',       20, '袋',     5, '食品棚B',   'ヒルズ・コルゲート',       'sufficient'),
    (4, 3, 'ロイヤルカナン 消化器サポート',      'food',       15, '袋',     5, '食品棚B',   'ロイヤルカナン',           'sufficient'),
    (5, 3, '包帯・ガーゼセット',                 'consumable', 100, 'セット', 20, '消耗品棚C', '白十字',               'sufficient')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 25. merchandise_items（物販・フード・その他: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (1, 3, 'ロイヤルカナン 消化器サポート 1kg', 'food', 2800, 0.10, 1),
    (2, 3, 'ヒルズ k/d 2kg', 'food', 3500, 0.10, 2),
    (3, 3, 'ペット用歯ブラシセット', 'goods', 1200, 0.10, 3),
    (4, 3, 'エリザベスカラー（S）', 'goods', 800, 0.10, 4),
    (5, 3, 'ノミ・ダニ予防首輪', 'goods', 1500, 0.10, 5),
    (6, 3, '文書料', 'other', 3000, 0.10, 6),
    (7, 3, '時間外診療費', 'other', 5000, 0.10, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('merchandise_items', 'id'), (SELECT MAX(id) FROM merchandise_items));

-- =============================================================================
-- マスタ設定完了
-- =============================================================================


-- =============================================================================
-- デモデータ投入（飼主・ペット一覧ページ対応）
-- 内容: 飼主・ペット・取引記録（カルテ・予約・会計・入院・在庫・監査ログ等）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. owners（飼主: 22件）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, owner_name, owner_name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (1,  3, '林 文明', 'ハヤシ フミアキ', '1980-05-15', 'サンプル株式会社', '150-0001', '東京都渋谷区神宮前1-2-3', '', '090-1111-2222', '03-1234-5678', 'hayashi@example.com', '定期検診を希望', false, 10, 'member'),
    (2,  3, '田中 花子', 'タナカ ハナコ', '1985-03-20', '', '160-0022', '東京都新宿区新宿1-1-1', '', '080-3333-4444', '', 'tanaka@example.com', '', false, 0, 'non_member'),
    (3,  3, '鈴木 一郎', 'スズキ イチロウ', '1978-11-03', '', '170-0001', '東京都豊島区西巣鴨1-3-5', '', '070-5555-6666', '', 'suzuki@example.com', '', false, 0, 'member'),
    (4,  3, '田中 美咲', 'タナカ ミサキ', '1990-07-22', '', '153-0044', '東京都目黒区大橋2-4-6', '', '090-9999-8888', '', 'misaki.tanaka@example.com', '', false, 0, 'non_member'),
    (5,  3, '佐藤 花子', 'サトウ ハナコ', '1975-02-14', '', '140-0001', '東京都品川区北品川3-5-7', '', '080-2222-3333', '', 'hanako.sato@example.com', '', false, 5, 'member'),
    (6,  3, '伊藤 次郎', 'イトウ ジロウ', '1983-09-30', '', '166-0013', '東京都杉並区堀ノ内1-7-9', '', '090-1234-5678', '', 'jiro.ito@example.com', '', false, 0, 'non_member'),
    (7,  3, '小林 さくら', 'コバヤシ サクラ', '1992-04-05', '', '176-0012', '東京都練馬区豊玉北4-2-8', '', '080-9876-5432', '', 'sakura.kobayashi@example.com', '', false, 0, 'member'),
    (8,  3, '中村 勇気', 'ナカムラ ユウキ', '1987-12-18', '', '174-0041', '東京都板橋区舟渡2-6-10', '', '090-1122-3344', '', 'yuuki.nakamura@example.com', '', false, 0, 'non_member'),
    (9,  3, '加藤 恵', 'カトウ メグミ', '1995-06-25', '', '134-0083', '東京都江戸川区中葛西5-3-2', '', '080-5566-7788', '', 'megumi.kato@example.com', '', false, 10, 'member'),
    (10, 3, '山田 太郎', 'ヤマダ タロウ', '1970-01-10', '', '144-0051', '東京都大田区西蒲田6-8-4', '', '090-2233-4455', '', 'taro.yamada@example.com', '', false, 0, 'non_member'),
    (11, 3, '高橋 由美', 'タカハシ ユミ', '1988-08-15', '', '110-0005', '東京都台東区上野5-1-3', '', '080-6677-8899', '', 'yumi.takahashi@example.com', '', false, 0, 'member'),
    (12, 3, '松本 隆', 'マツモト タカシ', '1965-03-28', '', '125-0061', '東京都葛飾区亀有3-9-7', '', '090-3344-5566', '', 'takashi.matsumoto@example.com', '', false, 0, 'non_member'),
    (13, 3, '吉田 誠', 'ヨシダ マコト', '1982-11-05', '', '123-0845', '東京都足立区西新井7-4-6', '', '080-7788-9900', '', 'makoto.yoshida@example.com', '', false, 0, 'non_member'),
    (14, 3, '井上 京子', 'イノウエ キョウコ', '1973-05-19', '', '189-0023', '東京都東村山市美住町1-5-2', '', '090-4455-6677', '', 'kyoko.inoue@example.com', '', false, 5, 'member'),
    (15, 3, '木村 拓也', 'キムラ タクヤ', '1991-07-14', '', '179-0081', '東京都練馬区北町3-6-9', '', '080-8899-0011', '', 'takuya.kimura@example.com', '', false, 0, 'non_member'),
    (16, 3, '佐々木 亮', 'ササキ リョウ', '1986-02-23', '', '207-0013', '東京都東大和市清水2-4-8', '', '090-5566-7788', '', 'ryo.sasaki@example.com', '', false, 0, 'non_member'),
    (17, 3, '山本 健太', 'ヤマモト ケンタ', '1998-09-12', '', '206-0802', '東京都稲城市東長沼2-8-3', '', '090-1234-9876', '', 'kenta.yamamoto@example.com', '', false, 0, 'non_member'),
    (18, 3, '青木 麻衣', 'アオキ マイ', '1993-03-10', '', '150-0002', '東京都渋谷区渋谷2-1-1', '', '090-1111-1111', '', 'mai.aoki@example.com', '', false, 0, 'non_member'),
    (19, 3, '橋本 俊介', 'ハシモト シュンスケ', '1980-07-25', '', '130-0001', '東京都墨田区吾妻橋1-3-5', '', '080-2222-2222', '', 'shunsuke.h@example.com', '', false, 0, 'member'),
    (20, 3, '福田 裕子', 'フクダ ユウコ', '1977-11-14', '', '145-0062', '東京都大田区北千束2-5-8', '', '090-3333-3333', '', 'yuko.fukuda@example.com', '', false, 5, 'member'),
    (21, 3, '石川 大輔', 'イシカワ ダイスケ', '1989-04-02', '', '167-0041', '東京都杉並区善福寺3-2-6', '', '080-4444-4444', '', 'daisuke.ishikawa@example.com', '', false, 0, 'non_member'),
    (22, 3, '村田 奈々', 'ムラタ ナナ', '1996-09-19', '', '182-0021', '東京都調布市調布ヶ丘1-4-7', '', '090-5555-5555', '', 'nana.murata@example.com', '', false, 0, 'non_member')
ON CONFLICT (id) DO UPDATE SET
    owner_name      = EXCLUDED.owner_name,
    owner_name_kana = EXCLUDED.owner_name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 2. pets（ペット: 28件）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, pet_name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (1,  3, 1,  '1-1', 'Iris(イリス)', 'イリス', 1, 'male',   'alive', '2015-04-14', 'ゴールデンレトリーバー',     '茶色',           26.5,  1, '2015-08-28'),
    (2,  3, 1,  '1-2', 'Max(マックス)', 'マックス', 1, 'male', 'alive', '2018-06-20', 'ラブラドール',               'ゴールデン',     15.2,  NULL, '2024-11-15'),
    (3,  3, 2,  '2-1', 'ミケ',         'ミケ',     2, 'female','alive', '2020-03-10', '三毛猫',                     '三毛',            4.20, 2, '2024-11-18'),
    (4,  3, 3,  '3-1', 'タロウ',       'タロウ',   1, 'male',  'alive', '2019-05-15', '柴犬',                       'レッド',          8.3,  NULL, NULL),
    (5,  3, 3,  '3-2', 'ジロウ',       'ジロウ',   1, 'male',  'alive', '2021-08-10', '柴犬',                       'ブラック',        7.1,  NULL, NULL),
    (6,  3, 4,  '4-1', 'チョコ',       'チョコ',   1, 'female','alive', '2017-11-20', 'トイプードル',               'チョコ',          3.80, 1, NULL),
    (7,  3, 5,  '5-1', 'レオ',         'レオ',     2, 'male',  'alive', '2016-07-04', 'スコティッシュフォールド',   'グレー',          5.5,  NULL, NULL),
    (8,  3, 6,  '6-1', 'ハチ',         'ハチ',     1, 'male',  'alive', '2018-03-25', '秋田犬',                     'ホワイト',       22.0,  NULL, NULL),
    (9,  3, 7,  '7-1', 'モモ',         'モモ',     2, 'female','alive', '2022-01-15', 'マンチカン',                 'キャリコ',        3.2,  2, NULL),
    (10, 3, 8,  '8-1', 'ロッキー',     'ロッキー', 1, 'male',  'alive', '2014-09-08', 'ボーダーコリー',             'ブラックホワイト',18.5,  NULL, NULL),
    (11, 3, 9,  '9-1', 'ルナ',         'ルナ',     2, 'female','alive', '2021-02-28', 'ペルシャ',                   'シルバー',        4.80, 1, NULL),
    (12, 3, 10, '10-1', 'ケン',        'ケン',     1, 'male',  'alive', '2013-06-18', 'ジャーマンシェパード',       'ブラックタン',   32.0,  NULL, NULL),
    (13, 3, 11, '11-1', 'ソラ',        'ソラ',     2, 'male',  'alive', '2023-04-01', 'アメリカンショートヘア',     'タビー',          3.0,  NULL, NULL),
    (14, 3, 12, '12-1', 'ゴン',        'ゴン',     1, 'male',  'alive', '2016-12-05', '紀州犬',                     'ホワイト',       19.5,  NULL, NULL),
    (15, 3, 13, '13-1', 'シロ',        'シロ',     1, 'male',  'alive', '2020-08-10', 'ミックス犬',                 'ホワイト',        6.2,  NULL, NULL),
    (16, 3, 14, '14-1', 'トラ',        'トラ',     2, 'male',  'alive', '2019-10-22', 'トラ猫',                     'トラ',            5.1,  NULL, NULL),
    (17, 3, 15, '15-1', 'ベロ',        'ベロ',     1, 'male',  'alive', '2018-05-03', 'ビーグル',                   'トライカラー',   13.2,  NULL, NULL),
    (18, 3, 16, '16-1', 'チビ',        'チビ',     2, 'female','alive', '2022-06-20', 'ミックス猫',                 'サビ',            3.50, NULL, NULL),
    (19, 3, 17, '17-1', 'ポチ',        'ポチ',     1, 'male',  'alive', '2017-02-14', 'ダックスフンド',             'チョコ',          7.8,  NULL, NULL),
    (20, 3, 18, '18-1', 'モカ',        'モカ',     2, 'female','alive', '2022-05-10', 'ミックス猫',                 'ホワイト',        4.1,  NULL, NULL),
    (21, 3, 18, '18-2', 'クルミ',      'クルミ',   1, 'male',  'alive', '2020-08-20', 'ミックス犬',                 'ベージュ',        8.3,  NULL, NULL),
    (22, 3, 19, '19-1', 'ハル',        'ハル',     1, 'male',  'alive', '2019-03-15', 'ミックス犬',                 'ブラック',       12.5,  NULL, NULL),
    (23, 3, 19, '19-2', 'ユキ',        'ユキ',     2, 'female','alive', '2021-12-01', 'ミックス猫',                 'ホワイト',        3.80, NULL, NULL),
    (24, 3, 20, '20-1', 'ピーチ',      'ピーチ',   2, 'female','alive', '2023-01-07', 'ミックス猫',                 'オレンジ',        3.2,  NULL, NULL),
    (25, 3, 21, '21-1', 'コタ',        'コタ',     1, 'male',  'alive', '2018-09-23', 'ミックス犬',                 'ブラウン',       22.0,  NULL, NULL),
    (26, 3, 21, '21-2', 'アン',        'アン',     2, 'female','alive', '2020-04-11', 'ミックス猫',                 'キャリコ',        4.5,  NULL, NULL),
    (27, 3, 22, '22-1', 'ゴマ',        'ゴマ',     2, 'male',  'alive', '2022-11-30', 'ミックス猫',                 'グレー',          5.0,  NULL, NULL),
    (28, 3, 22, '22-2', 'マル',        'マル',     1, 'female','alive', '2021-06-18', 'ミックス犬',                 'ゴールデン',      9.7,  NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- 3. reservation_appointments（予約: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO reservation_appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, service_type_id, doctor_id, is_designated, status, notes) VALUES
    (1,  3, '2026-03-12 09:00:00+09', '2026-03-12 09:30:00+09', 1,  1,  'revisit', 1, 1, true,  'completed',       '皮膚の経過観察'),
    (2,  3, '2026-03-12 09:30:00+09', '2026-03-12 10:00:00+09', 2,  3,  'revisit', 3, 2, false, 'accounting',      '猫の定期検診'),
    (3,  3, '2026-03-12 10:00:00+09', '2026-03-12 10:30:00+09', 3,  4,  'revisit', 1, 1, true,  'in_consultation',  '足を引きずっている'),
    (4,  3, '2026-03-12 10:30:00+09', '2026-03-12 11:00:00+09', 4,  6,  'first',   2, 2, false, 'checked_in',      'ワクチン接種希望'),
    (5,  3, '2026-03-12 14:00:00+09', '2026-03-12 14:30:00+09', 6,  8,  'revisit', 1, 1, false, 'confirmed',       '食欲低下が続いている'),
    (6,  3, '2026-03-13 09:00:00+09', '2026-03-13 09:30:00+09', 7,  9,  'revisit', 1, 2, true,  'confirmed',       '耳の治療経過確認'),
    (7,  3, '2026-03-13 10:00:00+09', '2026-03-13 10:30:00+09', 8,  10, 'first',   1, 1, false, 'confirmed',       '嘔吐が続いている'),
    (8,  3, '2026-03-14 09:30:00+09', '2026-03-14 10:00:00+09', 9,  11, 'revisit', 1, 2, false, 'confirmed',       'ルナの経過観察'),
    (9,  3, '2026-03-15 11:00:00+09', '2026-03-15 11:30:00+09', 10, 12, 'first',   2, 1, false, 'confirmed',       '初回ワクチン接種'),
    (10, 3, '2026-03-16 14:00:00+09', '2026-03-16 14:30:00+09', 11, 13, 'revisit', 1, 2, true,  'confirmed',       '腎臓値の経過観察')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('reservation_appointments', 'id'), (SELECT MAX(id) FROM reservation_appointments));

-- -----------------------------------------------------------------------------
-- 4. medical_records（カルテ: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (1,  3, 'R-2025-001', '2025-10-10', 1,  1,  1, 'finalized'),
    (2,  3, 'R-2025-002', '2025-12-15', 1,  1,  1, 'finalized'),
    (3,  3, 'R-2026-001', '2026-01-20', 1,  1,  2, 'finalized'),
    (4,  3, 'R-2025-003', '2025-11-05', 1,  2,  2, 'finalized'),
    (5,  3, 'R-2025-004', '2025-09-15', 2,  3,  2, 'finalized'),
    (6,  3, 'R-2026-002', '2026-01-06', 2,  3,  1, 'finalized'),
    (7,  3, 'R-2025-005', '2025-08-22', 3,  4,  2, 'finalized'),
    (8,  3, 'R-2025-006', '2025-10-18', 4,  6,  1, 'finalized'),
    (9,  3, 'R-2025-007', '2025-07-30', 5,  7,  2, 'finalized'),
    (10, 3, 'R-2026-003', '2026-01-15', 6,  8,  1, 'draft'),
    (11, 3, 'R-2025-008', '2025-12-01', 7,  9,  2, 'finalized'),
    (12, 3, 'R-2025-009', '2025-11-20', 8,  10, 1, 'finalized'),
    (13, 3, 'R-2026-004', '2026-02-10', 9,  11, 2, 'draft'),
    (14, 3, 'R-2025-010', '2025-06-15', 10, 12, 1, 'finalized'),
    (15, 3, 'R-2026-005', '2026-01-06', 11, 13, 2, 'finalized'),
    (16, 3, 'R-2025-011', '2025-09-08', 12, 14, 1, 'finalized'),
    (17, 3, 'R-2026-006', '2026-02-28', 13, 15, 2, 'draft'),
    (18, 3, 'R-2025-012', '2025-08-20', 14, 16, 1, 'finalized'),
    (19, 3, 'R-2026-007', '2026-01-03', 15, 17, 2, 'finalized'),
    (20, 3, 'R-2026-008', '2026-01-06', 16, 18, 1, 'finalized')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- -----------------------------------------------------------------------------
-- 5. inquiries（問診: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_category_id, chief_complaint, notes, staff_id) VALUES
    (1,  1,  NULL, '狂犬病ワクチン接種',         '体調良好。', 1),
    (2,  2,  NULL, '定期健診',                   '特に異常なし。', 2),
    (3,  3,  6,    '右足の跛行',                 '膝蓋骨脱臼を確認。', 1),
    (4,  4,  NULL, 'フィラリア予防',             '予防薬処方。', 2),
    (5,  5,  NULL, '5種混合ワクチン接種',        '体調良好。', 1),
    (6,  6,  NULL, '5種混合ワクチン接種',        '体調良好。', 2),
    (7,  7,  NULL, 'ノミダニ予防薬',             '予防薬処方。', 1),
    (8,  8,  3,    '皮膚の痒み',                 'アトピー性皮膚炎疑い。', 2),
    (9,  9,  3,    'トリミング後の皮膚チェック', '軽度の赤みあり。', 1),
    (10, 10, 1,    '食欲不振',                   '2日前から食欲減退。', 2),
    (11, 11, NULL, '耳を痒がる',                 '外耳炎疑い。', 1),
    (12, 12, NULL, '定期健診・予防接種',         '年次健診。', 2),
    (13, 13, 2,    '嘔吐・下痢',                 '昨日から嘔吐3回。', 1),
    (14, 14, NULL, '生化学検査',                 'シニア健診。', 2),
    (15, 15, NULL, 'ジステンパーワクチン接種',   '初回ワクチン。', 1),
    (16, 16, NULL, '血液検査',                   '異常なし。', 2),
    (17, 17, NULL, '歯石除去',                   '重度の歯石付着。', 1),
    (18, 18, NULL, '定期検診',                   '体重管理継続。', 2),
    (19, 19, 6,    '再診（右足跛行）',           '改善傾向。', 1),
    (20, 20, NULL, '5種混合ワクチン接種',        '体調良好。', 2)
ON CONFLICT (id) DO UPDATE SET
    medical_record_id = EXCLUDED.medical_record_id,
    updated_at        = now();

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));

-- -----------------------------------------------------------------------------
-- 5b. clinical_plans（診察/治療プラン: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (id, medical_record_id, physical_exam, diagnosis_category_id, diagnosis_name_id, diagnosis_details, treatment_policy) VALUES
    (1,  1, '体温38.5℃。心肺音正常。', NULL, NULL, '健康状態良好。ワクチン接種可。', '5種混合ワクチン接種実施。'),
    (2,  2, '体重増加あり。他異常なし。', NULL, NULL, '維持状態良好。', '定期検診継続。'),
    (3,  3, '右後肢跛行。パテラG2。', 8, 19, '膝蓋骨脱臼。', '消炎剤処方。体重管理指導。'),
    (4,  4, '異常なし。', NULL, NULL, '予防シーズン開始。', 'フィラリア予防薬処方。'),
    (5,  5, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (6,  6, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (7,  7, '良好。', NULL, NULL, '外部寄生虫予防。', 'スポットオン投与。'),
    (8,  8, '全身に発赤。搔痒感強。', 3, 6, 'アトピー性皮膚炎。', '抗ヒスタミン薬処方。薬用シャンプー推奨。'),
    (9,  9, '皮膚の一部に発赤。', 3, 7, '膿皮症初期。', '洗浄と消毒。'),
    (10, 10, '腹部軽度緊張。', 1, 1, '急性胃腸炎疑い。', '絶食・皮下補液実施。'),
    (11, 11, '耳道内に分泌物。', 3, 7, '外耳炎。', '耳道洗浄・点耳薬処方。'),
    (12, 12, 'シニア期に入る。', NULL, NULL, '健康診断実施。', '結果待ち。'),
    (13, 13, '脱水傾向あり。', 1, 1, '急性胃腸炎。', '対症療法と食事療法。'),
    (14, 14, '良好。', NULL, NULL, '経過観察。', '維持。'),
    (15, 15, '良好。', NULL, NULL, '幼若期検診。', '成長記録。'),
    (16, 16, '良好。', NULL, NULL, 'スクリーニング。', '異常なし。'),
    (17, 17, '重度の歯石。', NULL, NULL, '歯周病。', '抜歯を含めた歯科処置を計画。'),
    (18, 18, '良好。', NULL, NULL, '肥満気味。', 'ダイエットフード提案。'),
    (19, 19, '跛行消失。', 8, 19, '回復期。', '運動制限解除。'),
    (20, 20, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('clinical_plans', 'id'), (SELECT MAX(id) FROM clinical_plans));

-- -----------------------------------------------------------------------------
-- 6. vital_records（バイタル: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vital_records (id, pet_id, medical_record_id, recorded_at, staff_id, temperature, heart_rate, respiration_rate, weight, weight_unit, notes) VALUES
    (1, 1,  3, '2026-01-20 09:15:00+09', 1, 38.5, 80,  20, 26.5, 'Kg', '皮膚の搔痒感あり。体重良好。'),
    (2, 1,  2, '2025-12-15 10:00:00+09', 2, 38.8, 82,  22, 26.0, 'Kg', '体重前回比-500g'),
    (3, 1,  3, '2026-01-20 09:30:00+09', 1, 38.3, 78,  20, 26.5, 'Kg', '定期検診。皮膚搔痒感 軽快傾向。'),
    (4, 2,  4, '2025-11-05 11:00:00+09', 1, 39.1, 95,  24, 15.2, 'Kg', '軽度脱水。CRT 2秒。'),
    (5, 3,  5, '2025-09-15 14:30:00+09', 2, 38.2, 160, 30, 4200, 'g',  '粘膜色やや蒼白。食欲低下継続。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('vital_records', 'id'), (SELECT MAX(id) FROM vital_records));

-- -----------------------------------------------------------------------------
-- 7. treatments（治療明細: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, selected, status, content, unit_price, quantity, sort_order) VALUES
    (1, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (2, 1, 'medicine',     NULL, NULL, 1,    1,    true, 'completed', 'アモキシシリン 50mg x 7日分', 500,  7, 2),
    (3, 2, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (4, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (5, 3, 'procedure',    NULL, 4,    NULL, NULL, true, 'completed', '耳道洗浄（左耳）',          2500, 1, 2),
    (6, 4, 'consultation', 1,    NULL, NULL, NULL, true, 'completed', '初診料',                    2000, 1, 1),
    (7, 4, 'medicine',     NULL, NULL, 1,    1,    true, 'completed', 'アモキシシリン 50mg x 5日分', 500,  5, 2),
    (8, 5, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('treatments', 'id'), (SELECT MAX(id) FROM treatments));

-- -----------------------------------------------------------------------------
-- 8. trimming_records（トリミング: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_records (id, clinic_id, date, pet_id, bw, bw_unit, style_request, staff_id, status, course_id) VALUES
    (1, 3, '2025-10-10', 1,  26.5,  'Kg', 'サマーカット希望',        6,  'completed',   3),
    (2, 3, '2025-10-15', 2,  15.2,  'Kg', 'ふんわりカット',          12, 'reserved',    4),
    (3, 3, '2025-10-12', 3,  4.2,   'Kg', '毛玉カット',              6,  'in_progress', 1),
    (4, 3, '2026-01-06', 6,  3800,  'g',  'シャンプーコース',        6,  'completed',   1),
    (5, 3, '2026-01-06', 17, 12.0,  'Kg', '全体カット',              12, 'completed',   4),
    (6, 3, '2026-01-06', 10, 8.0,   'Kg', '爪切り・ブラッシング',   12, 'reserved',    2),
    (7, 3, '2026-01-06', 15, 5.0,   'Kg', 'シャンプー',              6,  'completed',   1),
    (8, 3, '2026-01-06', 6,  3800,  'g',  'トリミング',              6,  'reserved',    3)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('trimming_records', 'id'), (SELECT MAX(id) FROM trimming_records));

-- -----------------------------------------------------------------------------
-- 9. hospitalizations（入院: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalizations (id, clinic_id, owner_id, pet_id, hospitalization_type, start_date, end_date, status, cage_id, doctor_id, memo, owner_request, staff_notes) VALUES
    (1, 3, 3, 5,  'hospitalization', '2026-03-10', '2026-03-14', 'admitted',   5,    1, '急性胃腸炎による脱水治療。点滴管理中。',  '食事のアレルギーに注意してほしい（鶏肉不可）', '3/10入院開始。静脈点滴開始。3/11嘔吐1回。3/12状態改善傾向。'),
    (2, 3, 6, 8,  'hospitalization', '2026-02-25', '2026-02-28', 'discharged', 4,    1, '外耳炎重症化に伴う入院治療。',             '怖がりなので優しく接してほしい',               '耳道洗浄を毎日実施。2/28退院時、症状改善。点耳薬処方。'),
    (3, 3, 17, 19, 'hospitalization', '2026-02-10', '2026-02-20', 'discharged', NULL, 1, '骨折治療による入院。手術後経過観察。', '', '2/10手術実施。2/15抜糸。2/20退院。'),
    (4, 3, 4,  6,  'hotel',           '2026-03-15', '2026-03-18', 'reserved',   NULL, 1, '旅行中のホテル預かり。', 'フードはロイヤルカナンのみ', ''),
    (5, 3, 1,  1,  'hospitalization', '2026-03-20', '2026-03-25', 'reserved',   NULL, 1, '膝蓋骨脱臼手術予定。術前検査済み。', '怖がりなので静かな環境を希望', ''),
    (6, 3, 9,  11, 'hospitalization', '2026-03-05', '2026-03-12', 'admitted',   6,    2, '慢性腎臓病の集中治療。点滴管理中。', 'ペルシャ猫のため温度管理に注意', '3/5入院。毎日皮下補液実施。3/12現在状態安定。'),
    (7, 3, 3,  4,  'hospitalization', '2026-01-03', '2026-01-06', 'discharged', NULL, 1, '急性胃腸炎による脱水治療。', 'チキンアレルギーあり', '1/3入院。点滴開始。1/6状態改善し退院。')
ON CONFLICT (id) DO UPDATE SET
    updated_at            = now();

SELECT setval(pg_get_serial_sequence('hospitalizations', 'id'), (SELECT MAX(id) FROM hospitalizations));

-- -----------------------------------------------------------------------------
-- 10. care_plan_items（ケアプラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO care_plan_items (id, hospitalization_id, type, name, description, timing, status, notes, medicine_id, procedure_id, hospitalization_plan_id, unit_price, category, sort_order) VALUES
    (1, 1, 'food',        '療法食（消化器ケア）', '1日3回、少量ずつ与える', ARRAY['morning','noon','night']::plan_timing[], 'active', '鶏肉不可。', NULL, NULL, NULL, 0, '食事', 1),
    (2, 1, 'medicine',    'アモキシシリン',       '1回1錠、朝夕食後',       ARRAY['morning','night']::plan_timing[],       'active', '抗生剤。', 1,    NULL, NULL, 500, '投薬', 2),
    (3, 1, 'instruction', 'バイタルチェック',     '1日3回測定',             ARRAY['morning','noon','night']::plan_timing[], 'active', '異常値報告。', NULL, NULL, NULL, 0, '観察', 3),
    (4, 2, 'treatment',   '耳道洗浄',             '1日1回、朝に実施',       ARRAY['morning']::plan_timing[],               'completed', '左耳。', NULL, 4,    NULL, 2500, '処置', 1),
    (5, 2, 'item',        '入院管理料',           '小型犬1日分',            ARRAY['morning']::plan_timing[],               'completed', '', NULL, NULL, 1,    3000, '入院', 2)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('care_plan_items', 'id'), (SELECT MAX(id) FROM care_plan_items));

-- -----------------------------------------------------------------------------
-- 11. inventory_items（在庫管理: 9件追加）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (6,  3, '5種混合ワクチン',               'medicine',   25,  'バイアル', 15, '冷蔵庫 1',    '共立製薬',                'sufficient'),
    (7,  3, '留置針 22G',                    'consumable',  0,   '本',       50, '処置室 棚D',  'テルモ',                  'out_of_stock'),
    (8,  3, 'シリンジ 5mL',                  'consumable', 300,  '本',      100, '処置室 棚D',  'テルモ',                  'sufficient'),
    (9,  3, 'メトクロプラミド注 10mg',        'medicine',    8,   'アンプル', 10, '薬品棚 A-3', '日本全薬工業',            'low'),
    (10, 3, '療法食 消化器サポート（猫用）',  'food',       10,   '袋',        5, 'フード棚 C-1','ヒルズ',                 'sufficient'),
    (11, 3, 'エリザベスカラー（S）',          'other',      15,   '個',        5, '倉庫 A',     'ペットメディカルサプライ', 'sufficient'),
    (12, 3, 'ガーゼ 滅菌 7.5cm',            'consumable',  45,   '枚',       50, '処置室 棚E',  '白十字',                  'low'),
    (13, 3, 'フィラリア予防薬（S）',          'medicine',   60,   '錠',       30, '薬品棚 B-1', 'メリアル・ジャパン',       'sufficient'),
    (14, 3, 'ノミダニ駆除薬 スポット',        'medicine',   40,   'ピペット',  20, '薬品棚 B-2', 'エランコジャパン',         'sufficient')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 12. billings / billing_items / payments
-- -----------------------------------------------------------------------------
INSERT INTO billings (id, clinic_id, medical_record_id, hospitalization_id, owner_id, pet_id, subtotal, tax_total, total_amount, has_insurance, status, scheduled_date, completed_at, memo) VALUES
    (1, 3, 1,    NULL, 1,  1,  4300, 430, 4730, true, 'completed', '2026-02-15', '2026-02-15 10:30:00+09', 'アニコム保険適用'),
    (2, 3, 3,    NULL, 1,  1,  3300, 330, 3630, true, 'completed', '2026-02-28', '2026-02-28 11:00:00+09', 'アニコム保険適用（Iris 耳炎治療）'),
    (3, 3, 6,    NULL, 2,  3,  800,  80,  880,  true, 'waiting',   '2026-03-12', NULL,                     'アニコム保険適用。会計待ち。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('billings', 'id'), (SELECT MAX(id) FROM billings));

INSERT INTO billing_items (id, billing_id, category, name, unit_price, quantity, tax_rate, is_insurance_applicable, source, sort_order) VALUES
    (1, 1, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (2, 1, 'medicine', 'アモキシシリン 50mg x 7日分', 500,  7, 0.10, true, 'medical_record', 2),
    (3, 2, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (4, 2, 'procedure','耳道洗浄',                  2500, 1, 0.10, true, 'medical_record', 2),
    (5, 3, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT MAX(id) FROM billing_items));

INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method) VALUES
    (1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash'),
    (2, 2, 3300, 330, 3630, 'アニコム損保', 0.70, 2541, 0, 1089, 1100, 11, 'credit_card')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT MAX(id) FROM payments));

-- -----------------------------------------------------------------------------
-- 13. billing_refunds（返金デモデータ）
-- -----------------------------------------------------------------------------
INSERT INTO billing_refunds (id, clinic_id, billing_id, amount, reason, refunded_at) VALUES
    (1, 3, 1, 919,  '処置内容の変更に伴う部分返金',   '2026-02-16 10:00:00+09'),
    (2, 3, 1, 500,  '薬剤変更による差額返金',         '2026-02-20 14:30:00+09'),
    (3, 3, 2, 500,  '診察キャンセル分の返金',          '2026-03-01 09:00:00+09')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT MAX(id) FROM billing_refunds));

-- -----------------------------------------------------------------------------
-- 14. audit_logs（監査ログ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO audit_logs (clinic_id, actor_id, actor_type, action, resource, resource_id, old_value, new_value, ip_address, user_agent) VALUES
    (3, 10, 'staff', 'permission_rules.update', 'permission_groups', 1, '{"can_delete": false}', '{"can_delete": true}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'auth.login.success', 'user_accounts', 10, NULL, NULL, '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'owner.update', 'owners', 1, '{"phone": "old"}', '{"phone": "090-1234-5678"}', '192.168.1.2', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'pet.create', 'pets', 28, NULL, '{"name": "マル"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'medical_record.finalize', 'medical_records', 1, '{"status": "draft"}', '{"status": "finalized"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'auth.login.failure', 'user_accounts', NULL, '{"email": "exec@example.com"}', NULL, '192.168.1.5', 'PostmanRuntime/7.26.8'),
    (3, 10, 'staff', 'inventory.decrease', 'inventory_items', 1, '{"quantity": 50}', '{"quantity": 43}', '192.168.1.10', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'treatment.create', 'treatments', 2, NULL, '{"content": "アモキシシリン"}', '192.168.1.10', 'Mozilla/5.0...')
ON CONFLICT DO NOTHING;
