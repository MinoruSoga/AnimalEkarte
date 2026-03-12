-- =============================================================================
-- Animal Ekarte - 初期スキーマ定義
-- PostgreSQL 18
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 拡張機能
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -----------------------------------------------------------------------------
-- 2. ENUM型定義
-- -----------------------------------------------------------------------------

-- ペット関連
CREATE TYPE pet_species AS ENUM ('犬', '猫', '鳥', 'その他');
CREATE TYPE pet_status AS ENUM ('生存', '死亡');
CREATE TYPE pet_gender AS ENUM ('雄', '雌', '不明');
CREATE TYPE acquisition_type AS ENUM ('購入', '譲渡', '保護', 'その他');
CREATE TYPE danger_level AS ENUM ('低', '中', '高');
CREATE TYPE membership_type AS ENUM ('非会員', '会員', '退亡者', '他診/準');

-- 電子カルテ関連
CREATE TYPE medical_record_status AS ENUM ('作成中', '確定済');
CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other');
CREATE TYPE treatment_status AS ENUM ('未完了', '完了', '-');
CREATE TYPE examination_status AS ENUM ('依頼中', '検査中', '完了');
CREATE TYPE examination_result_status AS ENUM ('normal', 'high', 'low');
CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other');

-- 入院関連
CREATE TYPE hospitalization_type AS ENUM ('入院', 'ホテル');
CREATE TYPE hospitalization_status AS ENUM ('入院中', '退院済', '予約');
CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item');
CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued');
CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other');
CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped');
CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night');

-- 予約・トリミング関連
CREATE TYPE reservation_status AS ENUM (
    'confirmed', 'pending', 'cancelled', 'checked_in',
    'in_consultation', 'accounting', 'completed'
);
CREATE TYPE visit_type AS ENUM ('first', 'revisit');
CREATE TYPE trimming_status AS ENUM ('完了', '予約', '進行中');
CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g');

-- 会計関連
CREATE TYPE accounting_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money');
CREATE TYPE item_category AS ENUM (
    'examination', 'test', 'procedure', 'surgery',
    'medicine', 'food', 'goods', 'other'
);
CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization');

-- マスタ共通
CREATE TYPE master_status AS ENUM ('active', 'inactive');

-- マスタ固有
CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both');
CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder');
CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram');
CREATE TYPE staff_role AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager');
CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general');
CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE coverage_rate AS ENUM ('50', '70', '80', '100');
CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat');
CREATE TYPE combinable AS ENUM ('yes', 'no');
CREATE TYPE body_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night');
CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'general');

-- 在庫関連
CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other');
CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock');

-- シフト関連
CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave');

-- 認証関連
CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff');
CREATE TYPE job_title AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'general_staff');
CREATE TYPE permission_type AS ENUM (
    'account_admin', 'medical', 'medical_read', 'trimming', 'billing',
    'reception', 'hospitalization', 'master_admin', 'shift_admin', 'inventory'
);
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');

-- -----------------------------------------------------------------------------
-- 3. テーブル作成
-- -----------------------------------------------------------------------------

-- ============================================================
-- 3-1. 独立テーブル（FK参照なし）
-- ============================================================

-- 在庫アイテム
CREATE TABLE IF NOT EXISTS inventory_items (
    id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name              TEXT        NOT NULL,
    category          inventory_category NOT NULL,
    quantity          INTEGER     DEFAULT 0,
    unit              TEXT        NOT NULL DEFAULT '',
    min_stock_level   INTEGER     DEFAULT 0,
    location          TEXT        DEFAULT '',
    expiry_date       DATE,
    supplier          TEXT        DEFAULT '',
    last_restocked    DATE,
    status            inventory_status DEFAULT 'sufficient',
    created_at        TIMESTAMPTZ DEFAULT now(),
    updated_at        TIMESTAMPTZ DEFAULT now()
);

-- クリニック基本情報（シングルトン）
CREATE TABLE IF NOT EXISTS clinic_info (
    id                  UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                TEXT    NOT NULL,
    branch_name         TEXT    DEFAULT '',
    postal_code         TEXT    DEFAULT '',
    address             TEXT    DEFAULT '',
    phone_number        TEXT    DEFAULT '',
    fax_number          TEXT    DEFAULT '',
    registration_number TEXT    DEFAULT '',
    director_name       TEXT    DEFAULT '',
    email               TEXT    DEFAULT '',
    website             TEXT    DEFAULT '',
    logo_url            TEXT    DEFAULT '',
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

-- クリニック一覧
CREATE TABLE IF NOT EXISTS clinics (
    id                  UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                TEXT    NOT NULL,
    branch_name         TEXT    DEFAULT '',
    postal_code         TEXT    DEFAULT '',
    address             TEXT    DEFAULT '',
    phone_number        TEXT    DEFAULT '',
    fax_number          TEXT    DEFAULT '',
    registration_number TEXT    DEFAULT '',
    director_name       TEXT    DEFAULT '',
    email               TEXT    DEFAULT '',
    website             TEXT    DEFAULT '',
    logo_url            TEXT    DEFAULT '',
    is_active           BOOLEAN DEFAULT true,
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

-- オーナー（飼い主）
CREATE TABLE IF NOT EXISTS owners (
    id               UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_name       TEXT    NOT NULL,
    owner_name_kana  TEXT    DEFAULT '',
    company          TEXT    DEFAULT '',
    postal_code      TEXT    DEFAULT '',
    address1         TEXT    DEFAULT '',
    address2         TEXT    DEFAULT '',
    home_postal_code TEXT    DEFAULT '',
    home_address1    TEXT    DEFAULT '',
    home_address2    TEXT    DEFAULT '',
    phone            TEXT    DEFAULT '',
    company_phone    TEXT    DEFAULT '',
    email            TEXT    DEFAULT '',
    remarks          TEXT    DEFAULT '',
    is_dangerous     BOOLEAN DEFAULT false,
    discount_rate    NUMERIC(5,2) DEFAULT 0,
    membership_type  membership_type DEFAULT '非会員',
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-2. マスタテーブル（FKなし）
-- ============================================================

-- 検査種別マスタ
CREATE TABLE IF NOT EXISTS examination_types (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- ワクチンマスタ
CREATE TABLE IF NOT EXISTS vaccines (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    species     vaccine_species,
    interval    TEXT    DEFAULT '',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- 薬剤マスタ
CREATE TABLE IF NOT EXISTS medicines (
    id               UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code             TEXT    DEFAULT '',
    name             TEXT    NOT NULL,
    price            NUMERIC(10,2),
    status           master_status DEFAULT 'active',
    description      TEXT    DEFAULT '',
    dosage_form      dosage_form,
    medicine_unit    medicine_unit,
    inventory_id     UUID    REFERENCES inventory_items(id) ON DELETE SET NULL,
    default_quantity INTEGER DEFAULT 1,
    sort_order       INTEGER DEFAULT 0,
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

-- スタッフマスタ
CREATE TABLE IF NOT EXISTS staffs (
    id             UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code           TEXT    DEFAULT '',
    name           TEXT    NOT NULL,
    status         master_status DEFAULT 'active',
    staff_role     staff_role NOT NULL,
    license_number TEXT    DEFAULT '',
    sort_order     INTEGER DEFAULT 0,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

-- 保険マスタ
CREATE TABLE IF NOT EXISTS insurances (
    id            UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code          TEXT    DEFAULT '',
    name          TEXT    NOT NULL,
    status        master_status DEFAULT 'active',
    description   TEXT    DEFAULT '',
    coverage_rate coverage_rate,
    contact_phone TEXT    DEFAULT '',
    sort_order    INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

-- ケージマスタ
CREATE TABLE IF NOT EXISTS cages (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    cage_type   cage_type NOT NULL,
    cage_size   cage_size NOT NULL,
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- サービス種別マスタ
CREATE TABLE IF NOT EXISTS service_types (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    color       TEXT    DEFAULT '#3B82F6',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- 診察項目マスタ
CREATE TABLE IF NOT EXISTS consultations (
    id             UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code           TEXT    DEFAULT '',
    name           TEXT    NOT NULL,
    price          NUMERIC(10,2),
    status         master_status DEFAULT 'active',
    description    TEXT    DEFAULT '',
    time_condition TEXT    DEFAULT '',
    duration       TEXT    DEFAULT '',
    sort_order     INTEGER DEFAULT 0,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

-- 処置項目マスタ
CREATE TABLE IF NOT EXISTS procedures (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    duration    TEXT    DEFAULT '',
    anesthesia  anesthesia_type DEFAULT 'none',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- 入院プランマスタ
CREATE TABLE IF NOT EXISTS hospitalization_plans (
    id           UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code         TEXT    DEFAULT '',
    name         TEXT    NOT NULL,
    price        NUMERIC(10,2),
    status       master_status DEFAULT 'active',
    description  TEXT    DEFAULT '',
    body_size    body_size,
    billing_unit billing_unit DEFAULT 'per_day',
    sort_order   INTEGER DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT now(),
    updated_at   TIMESTAMPTZ DEFAULT now()
);

-- トリミングコースマスタ
CREATE TABLE IF NOT EXISTS trimming_courses (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    target_size target_size,
    duration    TEXT    DEFAULT '',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- トリミングオプションマスタ
CREATE TABLE IF NOT EXISTS trimming_options (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    duration    TEXT    DEFAULT '',
    combinable  combinable DEFAULT 'yes',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- 診断カテゴリマスタ
CREATE TABLE IF NOT EXISTS diagnosis_categories (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- 健診種別マスタ
CREATE TABLE IF NOT EXISTS checkup_types (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        TEXT    DEFAULT '',
    name        TEXT    NOT NULL,
    price       NUMERIC(10,2),
    status      master_status DEFAULT 'active',
    description TEXT    DEFAULT '',
    interval    TEXT    DEFAULT '',
    target_age  TEXT    DEFAULT '',
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-3. diagnosis_categoriesに依存するテーブル
-- ============================================================

-- 診断名マスタ
CREATE TABLE IF NOT EXISTS diagnosis_names (
    id                    UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    code                  TEXT    DEFAULT '',
    name                  TEXT    NOT NULL,
    status                master_status DEFAULT 'active',
    description           TEXT    DEFAULT '',
    diagnosis_category_id UUID    NOT NULL REFERENCES diagnosis_categories(id) ON DELETE CASCADE,
    sort_order            INTEGER DEFAULT 0,
    created_at            TIMESTAMPTZ DEFAULT now(),
    updated_at            TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-4. examination_typesに依存するテーブル
-- ============================================================

-- 検査項目定義
CREATE TABLE IF NOT EXISTS examination_type_items (
    id                  UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    examination_type_id UUID    NOT NULL REFERENCES examination_types(id) ON DELETE CASCADE,
    name                TEXT    NOT NULL,
    inspection_value    TEXT    DEFAULT '',
    normal_value        TEXT    DEFAULT '',
    sort_order          INTEGER DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-5. staffsに依存するテーブル
-- ============================================================

-- ユーザーアカウント
CREATE TABLE IF NOT EXISTS user_accounts (
    id                UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    email             TEXT    NOT NULL UNIQUE,
    display_name      TEXT    NOT NULL,
    display_name_kana TEXT    DEFAULT '',
    user_type         user_type NOT NULL DEFAULT 'staff',
    job_title         job_title,
    status            account_status DEFAULT 'active',
    avatar_url        TEXT    DEFAULT '',
    staff_id          UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ DEFAULT now(),
    updated_at        TIMESTAMPTZ DEFAULT now()
);

-- シフトエントリ
CREATE TABLE IF NOT EXISTS shift_entries (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    staff_id   UUID    NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    date       DATE    NOT NULL,
    shift_type shift_type NOT NULL,
    start_time TEXT    DEFAULT '',
    end_time   TEXT    DEFAULT '',
    note       TEXT    DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (staff_id, date)
);

-- ============================================================
-- 3-6. clinicsとuser_accountsに依存するテーブル
-- ============================================================

-- ユーザークリニック所属
CREATE TABLE IF NOT EXISTS user_clinic_memberships (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID    NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    clinic_id  UUID    NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    is_main    BOOLEAN DEFAULT false,
    joined_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, clinic_id)
);

-- ユーザー権限
CREATE TABLE IF NOT EXISTS user_permissions (
    id         UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID    NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    clinic_id  UUID    NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    permission permission_type NOT NULL,
    granted_by UUID    REFERENCES user_accounts(id) ON DELETE SET NULL,
    granted_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-7. ownersに依存するテーブル
-- ============================================================

-- ペット
CREATE TABLE IF NOT EXISTS pets (
    id               UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id         UUID    NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    pet_number       TEXT    DEFAULT '',
    name             TEXT    NOT NULL,
    pet_name_kana    TEXT    DEFAULT '',
    species          pet_species NOT NULL,
    gender           pet_gender DEFAULT '不明',
    status           pet_status DEFAULT '生存',
    birth_date       DATE,
    breed            TEXT    DEFAULT '',
    color            TEXT    DEFAULT '',
    weight           TEXT    DEFAULT '',
    neutered_date    DATE,
    acquisition_type acquisition_type,
    danger_level     danger_level DEFAULT '低',
    food             TEXT    DEFAULT '',
    environment      TEXT    DEFAULT '',
    phone            TEXT    DEFAULT '',
    last_visit       DATE,
    insurance_id     UUID    REFERENCES insurances(id) ON DELETE SET NULL,
    insurance_name   TEXT    DEFAULT '',
    insurance_details TEXT   DEFAULT '',
    remarks          TEXT    DEFAULT '',
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-8. petsとdiagnosis_categories/diagnosis_namesとstaffsに依存するテーブル
-- ============================================================

-- 電子カルテ
CREATE TABLE IF NOT EXISTS medical_records (
    id                        UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    record_no                 TEXT    NOT NULL UNIQUE,
    date                      DATE    NOT NULL,
    owner_id                  UUID    REFERENCES owners(id) ON DELETE SET NULL,
    owner_name                TEXT    NOT NULL DEFAULT '',
    pet_id                    UUID    REFERENCES pets(id) ON DELETE SET NULL,
    pet_name                  TEXT    NOT NULL DEFAULT '',
    species                   pet_species NOT NULL,
    chief_complaint           TEXT    DEFAULT '',
    treatment_policy          TEXT    DEFAULT '',
    physical_exam             TEXT    DEFAULT '',
    diagnosis_details         TEXT    DEFAULT '',
    diagnosis1_category_id    UUID    REFERENCES diagnosis_categories(id) ON DELETE SET NULL,
    diagnosis1_name_id        UUID    REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    diagnosis2_category_id    UUID    REFERENCES diagnosis_categories(id) ON DELETE SET NULL,
    diagnosis2_name_id        UUID    REFERENCES diagnosis_names(id) ON DELETE SET NULL,
    doctor_id                 UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    status                    medical_record_status DEFAULT '作成中',
    created_at                TIMESTAMPTZ DEFAULT now(),
    updated_at                TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-9. medical_recordsとinventory_items/consultations/procedures/medicinesに依存するテーブル
-- ============================================================

-- 治療項目
CREATE TABLE IF NOT EXISTS treatment_items (
    id                UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id UUID    NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    item_type         treatment_item_type NOT NULL DEFAULT 'other',
    consultation_id   UUID    REFERENCES consultations(id) ON DELETE SET NULL,
    procedure_id      UUID    REFERENCES procedures(id) ON DELETE SET NULL,
    medicine_id       UUID    REFERENCES medicines(id) ON DELETE SET NULL,
    selected          BOOLEAN DEFAULT false,
    status            treatment_status DEFAULT '未完了',
    content           TEXT    NOT NULL DEFAULT '',
    memo              TEXT    DEFAULT '',
    insurance         BOOLEAN DEFAULT false,
    unit_price        NUMERIC(10,2) DEFAULT 0,
    quantity          INTEGER DEFAULT 1,
    discount_rate     NUMERIC(5,2) DEFAULT 0,
    discount_amount   NUMERIC(10,2) DEFAULT 0,
    inventory_id      UUID    REFERENCES inventory_items(id) ON DELETE SET NULL,
    sort_order        INTEGER DEFAULT 0,
    created_at        TIMESTAMPTZ DEFAULT now(),
    updated_at        TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-10. medical_recordsとstaffsに依存するテーブル
-- ============================================================

-- バイタルエントリ（外来）
CREATE TABLE IF NOT EXISTS vital_entries (
    id                UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id UUID    NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    recorded_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    staff_id          UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    temperature       NUMERIC(4,1),
    heart_rate        INTEGER,
    respiration_rate  INTEGER,
    weight            NUMERIC(6,2),
    notes             TEXT    DEFAULT '',
    created_at        TIMESTAMPTZ DEFAULT now()
);

-- 検査記録
CREATE TABLE IF NOT EXISTS examination_records (
    id                  UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id   UUID    REFERENCES medical_records(id) ON DELETE CASCADE,
    pet_id              UUID    REFERENCES pets(id) ON DELETE CASCADE,
    date                DATE    NOT NULL,
    owner_name          TEXT    NOT NULL DEFAULT '',
    pet_name            TEXT    NOT NULL DEFAULT '',
    examination_type_id UUID    NOT NULL REFERENCES examination_types(id) ON DELETE RESTRICT,
    doctor_id           UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    status              examination_status DEFAULT '依頼中',
    result_summary      TEXT    DEFAULT '',
    machine             TEXT    DEFAULT '',
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

-- 検査記録項目
CREATE TABLE IF NOT EXISTS examination_record_items (
    id                    UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    examination_record_id UUID    NOT NULL REFERENCES examination_records(id) ON DELETE CASCADE,
    name                  TEXT    NOT NULL DEFAULT '',
    inspection_value      TEXT    DEFAULT '',
    normal_value          TEXT    DEFAULT '',
    result                TEXT    DEFAULT '',
    unit                  TEXT    DEFAULT '',
    ref                   TEXT    DEFAULT '',
    status                examination_result_status DEFAULT 'normal',
    sort_order            INTEGER DEFAULT 0,
    created_at            TIMESTAMPTZ DEFAULT now()
);

-- ワクチン接種記録
CREATE TABLE IF NOT EXISTS vaccination_records (
    id                     UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id      UUID    REFERENCES medical_records(id) ON DELETE CASCADE,
    pet_id                 UUID    REFERENCES pets(id) ON DELETE CASCADE,
    owner_name             TEXT    NOT NULL DEFAULT '',
    pet_name               TEXT    NOT NULL DEFAULT '',
    vaccine_id             UUID    NOT NULL REFERENCES vaccines(id) ON DELETE RESTRICT,
    vaccine_name_snapshot  TEXT    NOT NULL DEFAULT '',
    date                   DATE    NOT NULL,
    next_date              DATE,
    next_schedule_type     next_schedule_type,
    doctor_id              UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    supplemental           TEXT    DEFAULT '',
    lot1                   TEXT    DEFAULT '',
    lot2                   TEXT    DEFAULT '',
    lot3                   TEXT    DEFAULT '',
    lot4                   TEXT    DEFAULT '',
    remarks                TEXT    DEFAULT '',
    created_at             TIMESTAMPTZ DEFAULT now(),
    updated_at             TIMESTAMPTZ DEFAULT now()
);

-- 健診記録
CREATE TABLE IF NOT EXISTS checkup_records (
    id                UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id UUID    REFERENCES medical_records(id) ON DELETE CASCADE,
    pet_id            UUID    REFERENCES pets(id) ON DELETE CASCADE,
    owner_name        TEXT    NOT NULL DEFAULT '',
    pet_name          TEXT    NOT NULL DEFAULT '',
    checkup_type_id   UUID    NOT NULL REFERENCES checkup_types(id) ON DELETE RESTRICT,
    date              DATE    NOT NULL,
    next_date         DATE,
    doctor_id         UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    result            TEXT    DEFAULT '',
    created_at        TIMESTAMPTZ DEFAULT now(),
    updated_at        TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-11. petsとcagesとstaffsに依存するテーブル
-- ============================================================

-- 入院記録
CREATE TABLE IF NOT EXISTS hospitalizations (
    id                   UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id             UUID    REFERENCES owners(id) ON DELETE SET NULL,
    owner_name           TEXT    NOT NULL DEFAULT '',
    pet_id               UUID    REFERENCES pets(id) ON DELETE SET NULL,
    pet_name             TEXT    NOT NULL DEFAULT '',
    species              pet_species NOT NULL,
    hospitalization_type hospitalization_type NOT NULL,
    start_date           DATE    NOT NULL,
    end_date             DATE    NOT NULL,
    status               hospitalization_status DEFAULT '予約',
    cage_id              UUID    REFERENCES cages(id) ON DELETE SET NULL,
    doctor_id            UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    memo                 TEXT    DEFAULT '',
    owner_request        TEXT    DEFAULT '',
    staff_notes          TEXT    DEFAULT '',
    created_at           TIMESTAMPTZ DEFAULT now(),
    updated_at           TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-12. hospitalizationsとmedicines/procedures/hospitalization_plansに依存するテーブル
-- ============================================================

-- ケアプラン項目
CREATE TABLE IF NOT EXISTS care_plan_items (
    id                    UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospitalization_id    UUID    NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
    type                  care_plan_type NOT NULL,
    name                  TEXT    NOT NULL DEFAULT '',
    description           TEXT    DEFAULT '',
    timing                plan_timing[] DEFAULT '{}',
    status                care_plan_status DEFAULT 'active',
    notes                 TEXT    DEFAULT '',
    medicine_id           UUID    REFERENCES medicines(id) ON DELETE SET NULL,
    procedure_id          UUID    REFERENCES procedures(id) ON DELETE SET NULL,
    hospitalization_plan_id UUID  REFERENCES hospitalization_plans(id) ON DELETE SET NULL,
    unit_price            NUMERIC(10,2) DEFAULT 0,
    category              TEXT    DEFAULT '',
    sort_order            INTEGER DEFAULT 0,
    created_at            TIMESTAMPTZ DEFAULT now(),
    updated_at            TIMESTAMPTZ DEFAULT now()
);

-- 治療プラン（入院）
CREATE TABLE IF NOT EXISTS treatment_plans (
    id                 UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospitalization_id UUID    NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
    treatment_content  TEXT    NOT NULL DEFAULT '',
    memo               TEXT    DEFAULT '',
    insurance          BOOLEAN DEFAULT false,
    unit_price         NUMERIC(10,2) DEFAULT 0,
    quantity           INTEGER DEFAULT 1,
    discount_rate      NUMERIC(5,2) DEFAULT 0,
    discount_amount    NUMERIC(10,2) DEFAULT 0,
    subtotal           NUMERIC(10,2) DEFAULT 0,
    sort_order         INTEGER DEFAULT 0,
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now()
);

-- 日次記録
CREATE TABLE IF NOT EXISTS daily_records (
    id                 UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    hospitalization_id UUID    NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
    date               DATE    NOT NULL,
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now(),
    UNIQUE (hospitalization_id, date)
);

-- ============================================================
-- 3-13. daily_recordsとstaffsに依存するテーブル
-- ============================================================

-- バイタル記録（入院）
CREATE TABLE IF NOT EXISTS vital_records (
    id               UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    daily_record_id  UUID    NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time             TEXT    NOT NULL DEFAULT '',
    temperature      NUMERIC(4,1),
    heart_rate       INTEGER,
    respiration_rate INTEGER,
    weight           NUMERIC(6,2),
    notes            TEXT    DEFAULT '',
    staff_id         UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ DEFAULT now()
);

-- ケアログ記録
CREATE TABLE IF NOT EXISTS care_log_records (
    id              UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    daily_record_id UUID    NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time            TEXT    NOT NULL DEFAULT '',
    type            care_log_type NOT NULL,
    status          care_log_status NOT NULL DEFAULT 'completed',
    value           TEXT    DEFAULT '',
    staff_id        UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    notes           TEXT    DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- スタッフノート記録
CREATE TABLE IF NOT EXISTS staff_note_records (
    id              UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    daily_record_id UUID    NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
    time            TEXT    NOT NULL DEFAULT '',
    content         TEXT    NOT NULL DEFAULT '',
    staff_id        UUID    REFERENCES staffs(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-14. petsとservice_typesとstaffsに依存するテーブル
-- ============================================================

-- 予約アポイントメント
CREATE TABLE IF NOT EXISTS reservation_appointments (
    id             UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_time     TIMESTAMPTZ NOT NULL,
    end_time       TIMESTAMPTZ NOT NULL,
    owner_name     TEXT    NOT NULL DEFAULT '',
    pet_name       TEXT    NOT NULL DEFAULT '',
    pet_id         UUID    REFERENCES pets(id) ON DELETE SET NULL,
    visit_type     visit_type NOT NULL DEFAULT 'revisit',
    service_type_id UUID   REFERENCES service_types(id) ON DELETE SET NULL,
    doctor_id      UUID    NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    is_designated  BOOLEAN DEFAULT false,
    status         reservation_status DEFAULT 'pending',
    notes          TEXT    DEFAULT '',
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-15. petsとtrimming_courses/staffsに依存するテーブル
-- ============================================================

-- トリミング記録
CREATE TABLE IF NOT EXISTS trimming_records (
    id             UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    date           DATE    NOT NULL,
    pet_id         UUID    NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    pet_number     TEXT    NOT NULL DEFAULT '',
    pet_name       TEXT    NOT NULL DEFAULT '',
    owner_name     TEXT    NOT NULL DEFAULT '',
    species        pet_species NOT NULL,
    weight         TEXT    DEFAULT '',
    style_request  TEXT    DEFAULT '',
    staff_id       UUID    NOT NULL REFERENCES staffs(id) ON DELETE RESTRICT,
    status         trimming_status DEFAULT '予約',
    course_id      UUID    REFERENCES trimming_courses(id) ON DELETE SET NULL,
    bw             TEXT    DEFAULT '',
    bw_unit        body_weight_unit DEFAULT 'Kg',
    bt             TEXT    DEFAULT '',
    used_shampoo   TEXT    DEFAULT '',
    used_ribbon    TEXT    DEFAULT '',
    remarks        TEXT    DEFAULT '',
    style_image    TEXT    DEFAULT '',
    completed_image TEXT   DEFAULT '',
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now()
);

-- ============================================================
-- 3-16. trimming_recordsとtrimming_optionsに依存するテーブル
-- ============================================================

-- トリミング記録オプション（中間テーブル）
CREATE TABLE IF NOT EXISTS trimming_record_options (
    id                 UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    trimming_record_id UUID    NOT NULL REFERENCES trimming_records(id) ON DELETE CASCADE,
    option_id          UUID    NOT NULL REFERENCES trimming_options(id) ON DELETE CASCADE,
    sort_order         INTEGER DEFAULT 0,
    UNIQUE (trimming_record_id, option_id)
);

-- ============================================================
-- 3-17. ownersとpetsに依存するテーブル
-- ============================================================

-- 会計
CREATE TABLE IF NOT EXISTS accountings (
    id                 UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    medical_record_id  UUID    UNIQUE REFERENCES medical_records(id) ON DELETE SET NULL,
    hospitalization_id UUID    REFERENCES hospitalizations(id) ON DELETE SET NULL,
    owner_id           UUID    NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    owner_name         TEXT    NOT NULL DEFAULT '',
    pet_id             UUID    NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    pet_name           TEXT    NOT NULL DEFAULT '',
    pet_species        pet_species,
    status             accounting_status DEFAULT 'waiting',
    scheduled_date     DATE    NOT NULL,
    completed_at       TIMESTAMPTZ,
    memo               TEXT    DEFAULT '',
    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now()
);

-- 会計項目
CREATE TABLE IF NOT EXISTS accounting_items (
    id                    UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    accounting_id         UUID    NOT NULL REFERENCES accountings(id) ON DELETE CASCADE,
    code                  TEXT    DEFAULT '',
    category              item_category NOT NULL,
    name                  TEXT    NOT NULL DEFAULT '',
    unit_price            NUMERIC(10,2) NOT NULL DEFAULT 0,
    quantity              INTEGER NOT NULL DEFAULT 1,
    tax_rate              NUMERIC(3,2) DEFAULT 0.10,
    is_insurance_applicable BOOLEAN DEFAULT false,
    source                item_source DEFAULT 'manual',
    sort_order            INTEGER DEFAULT 0,
    created_at            TIMESTAMPTZ DEFAULT now()
);

-- 支払情報
CREATE TABLE IF NOT EXISTS payment_infos (
    id               UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    accounting_id    UUID    NOT NULL UNIQUE REFERENCES accountings(id) ON DELETE CASCADE,
    subtotal         NUMERIC(10,2) NOT NULL DEFAULT 0,
    tax_total        NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_amount     NUMERIC(10,2) NOT NULL DEFAULT 0,
    insurance_name   TEXT    DEFAULT '',
    insurance_ratio  NUMERIC(3,2) DEFAULT 0,
    insurance_amount NUMERIC(10,2) DEFAULT 0,
    discount_amount  NUMERIC(10,2) DEFAULT 0,
    billing_amount   NUMERIC(10,2) NOT NULL DEFAULT 0,
    received_amount  NUMERIC(10,2) DEFAULT 0,
    change_amount    NUMERIC(10,2) DEFAULT 0,
    method           payment_method DEFAULT 'cash',
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now()
);

-- -----------------------------------------------------------------------------
-- 4. インデックス
-- -----------------------------------------------------------------------------

-- コア
CREATE INDEX IF NOT EXISTS idx_pets_owner_id       ON pets (owner_id);
CREATE INDEX IF NOT EXISTS idx_pets_species        ON pets (species);
CREATE INDEX IF NOT EXISTS idx_pets_name           ON pets (name);
CREATE INDEX IF NOT EXISTS idx_pets_status         ON pets (status);
CREATE INDEX IF NOT EXISTS idx_owners_membership   ON owners (membership_type);
CREATE INDEX IF NOT EXISTS idx_owners_name         ON owners (owner_name);

-- 電子カルテ
CREATE INDEX IF NOT EXISTS idx_medical_records_pet_id    ON medical_records (pet_id);
CREATE INDEX IF NOT EXISTS idx_medical_records_owner_id  ON medical_records (owner_id);
CREATE INDEX IF NOT EXISTS idx_medical_records_date      ON medical_records (date DESC);
CREATE INDEX IF NOT EXISTS idx_medical_records_status    ON medical_records (status);
CREATE INDEX IF NOT EXISTS idx_medical_records_doctor_id ON medical_records (doctor_id);

CREATE INDEX IF NOT EXISTS idx_treatment_items_medical_record_id ON treatment_items (medical_record_id);
CREATE INDEX IF NOT EXISTS idx_treatment_items_item_type         ON treatment_items (item_type);
CREATE INDEX IF NOT EXISTS idx_treatment_items_medicine_id       ON treatment_items (medicine_id);
CREATE INDEX IF NOT EXISTS idx_treatment_items_consultation_id   ON treatment_items (consultation_id);
CREATE INDEX IF NOT EXISTS idx_treatment_items_procedure_id      ON treatment_items (procedure_id);

CREATE INDEX IF NOT EXISTS idx_vital_entries_medical_record_id ON vital_entries (medical_record_id);
CREATE INDEX IF NOT EXISTS idx_vital_entries_recorded_at       ON vital_entries (recorded_at);

CREATE INDEX IF NOT EXISTS idx_examination_records_medical_record_id   ON examination_records (medical_record_id);
CREATE INDEX IF NOT EXISTS idx_examination_records_pet_id              ON examination_records (pet_id);
CREATE INDEX IF NOT EXISTS idx_examination_records_examination_type_id ON examination_records (examination_type_id);
CREATE INDEX IF NOT EXISTS idx_examination_records_status              ON examination_records (status);

CREATE INDEX IF NOT EXISTS idx_examination_record_items_record_id ON examination_record_items (examination_record_id);

CREATE INDEX IF NOT EXISTS idx_vaccination_records_medical_record_id ON vaccination_records (medical_record_id);
CREATE INDEX IF NOT EXISTS idx_vaccination_records_pet_id            ON vaccination_records (pet_id);
CREATE INDEX IF NOT EXISTS idx_vaccination_records_vaccine_id        ON vaccination_records (vaccine_id);

CREATE INDEX IF NOT EXISTS idx_checkup_records_medical_record_id ON checkup_records (medical_record_id);
CREATE INDEX IF NOT EXISTS idx_checkup_records_pet_id            ON checkup_records (pet_id);
CREATE INDEX IF NOT EXISTS idx_checkup_records_checkup_type_id   ON checkup_records (checkup_type_id);

-- 入院
CREATE INDEX IF NOT EXISTS idx_hospitalizations_pet_id     ON hospitalizations (pet_id);
CREATE INDEX IF NOT EXISTS idx_hospitalizations_status     ON hospitalizations (status);
CREATE INDEX IF NOT EXISTS idx_hospitalizations_start_date ON hospitalizations (start_date DESC);
CREATE INDEX IF NOT EXISTS idx_hospitalizations_cage_id    ON hospitalizations (cage_id);

CREATE INDEX IF NOT EXISTS idx_care_plan_items_hospitalization_id ON care_plan_items (hospitalization_id);
CREATE INDEX IF NOT EXISTS idx_care_plan_items_type               ON care_plan_items (type);
CREATE INDEX IF NOT EXISTS idx_care_plan_items_status             ON care_plan_items (status);

CREATE INDEX IF NOT EXISTS idx_daily_records_hospitalization_id ON daily_records (hospitalization_id);
CREATE INDEX IF NOT EXISTS idx_daily_records_date               ON daily_records (date);

CREATE INDEX IF NOT EXISTS idx_vital_records_daily_record_id ON vital_records (daily_record_id);

CREATE INDEX IF NOT EXISTS idx_care_log_records_daily_record_id ON care_log_records (daily_record_id);
CREATE INDEX IF NOT EXISTS idx_care_log_records_type            ON care_log_records (type);

CREATE INDEX IF NOT EXISTS idx_staff_note_records_daily_record_id ON staff_note_records (daily_record_id);

CREATE INDEX IF NOT EXISTS idx_treatment_plans_hospitalization_id ON treatment_plans (hospitalization_id);

-- 予約・トリミング
CREATE INDEX IF NOT EXISTS idx_reservation_appointments_start_time    ON reservation_appointments (start_time);
CREATE INDEX IF NOT EXISTS idx_reservation_appointments_status        ON reservation_appointments (status);
CREATE INDEX IF NOT EXISTS idx_reservation_appointments_pet_id        ON reservation_appointments (pet_id);
CREATE INDEX IF NOT EXISTS idx_reservation_appointments_doctor_id     ON reservation_appointments (doctor_id);
CREATE INDEX IF NOT EXISTS idx_reservation_appointments_service_type  ON reservation_appointments (service_type_id);

CREATE INDEX IF NOT EXISTS idx_trimming_records_pet_id   ON trimming_records (pet_id);
CREATE INDEX IF NOT EXISTS idx_trimming_records_date     ON trimming_records (date DESC);
CREATE INDEX IF NOT EXISTS idx_trimming_records_status   ON trimming_records (status);
CREATE INDEX IF NOT EXISTS idx_trimming_records_staff_id ON trimming_records (staff_id);

-- 会計
CREATE INDEX IF NOT EXISTS idx_accountings_owner_id       ON accountings (owner_id);
CREATE INDEX IF NOT EXISTS idx_accountings_pet_id         ON accountings (pet_id);
CREATE INDEX IF NOT EXISTS idx_accountings_status         ON accountings (status);
CREATE INDEX IF NOT EXISTS idx_accountings_scheduled_date ON accountings (scheduled_date DESC);

CREATE INDEX IF NOT EXISTS idx_accounting_items_accounting_id ON accounting_items (accounting_id);
CREATE INDEX IF NOT EXISTS idx_accounting_items_category      ON accounting_items (category);

-- マスタ
CREATE INDEX IF NOT EXISTS idx_examination_types_status     ON examination_types (status);
CREATE INDEX IF NOT EXISTS idx_examination_types_sort_order ON examination_types (sort_order);

CREATE INDEX IF NOT EXISTS idx_examination_type_items_type_id ON examination_type_items (examination_type_id);

CREATE INDEX IF NOT EXISTS idx_vaccines_status  ON vaccines (status);
CREATE INDEX IF NOT EXISTS idx_vaccines_species ON vaccines (species);

CREATE INDEX IF NOT EXISTS idx_medicines_status      ON medicines (status);
CREATE INDEX IF NOT EXISTS idx_medicines_dosage_form ON medicines (dosage_form);
CREATE INDEX IF NOT EXISTS idx_medicines_inventory_id ON medicines (inventory_id);

CREATE INDEX IF NOT EXISTS idx_staffs_status     ON staffs (status);
CREATE INDEX IF NOT EXISTS idx_staffs_staff_role ON staffs (staff_role);

CREATE INDEX IF NOT EXISTS idx_insurances_status ON insurances (status);

CREATE INDEX IF NOT EXISTS idx_cages_status    ON cages (status);
CREATE INDEX IF NOT EXISTS idx_cages_cage_type ON cages (cage_type);

CREATE INDEX IF NOT EXISTS idx_service_types_status     ON service_types (status);
CREATE INDEX IF NOT EXISTS idx_service_types_sort_order ON service_types (sort_order);

CREATE INDEX IF NOT EXISTS idx_consultations_status     ON consultations (status);
CREATE INDEX IF NOT EXISTS idx_consultations_sort_order ON consultations (sort_order);

CREATE INDEX IF NOT EXISTS idx_procedures_status     ON procedures (status);
CREATE INDEX IF NOT EXISTS idx_procedures_sort_order ON procedures (sort_order);

CREATE INDEX IF NOT EXISTS idx_hospitalization_plans_status ON hospitalization_plans (status);

CREATE INDEX IF NOT EXISTS idx_trimming_courses_status      ON trimming_courses (status);
CREATE INDEX IF NOT EXISTS idx_trimming_courses_target_size ON trimming_courses (target_size);

CREATE INDEX IF NOT EXISTS idx_trimming_options_status ON trimming_options (status);

CREATE INDEX IF NOT EXISTS idx_diagnosis_categories_status     ON diagnosis_categories (status);
CREATE INDEX IF NOT EXISTS idx_diagnosis_categories_sort_order ON diagnosis_categories (sort_order);

CREATE INDEX IF NOT EXISTS idx_diagnosis_names_status              ON diagnosis_names (status);
CREATE INDEX IF NOT EXISTS idx_diagnosis_names_category_id         ON diagnosis_names (diagnosis_category_id);

CREATE INDEX IF NOT EXISTS idx_checkup_types_status ON checkup_types (status);

-- シフト・認証
CREATE INDEX IF NOT EXISTS idx_shift_entries_date     ON shift_entries (date);
CREATE INDEX IF NOT EXISTS idx_shift_entries_staff_id ON shift_entries (staff_id);

CREATE INDEX IF NOT EXISTS idx_user_accounts_email    ON user_accounts (email);
CREATE INDEX IF NOT EXISTS idx_user_accounts_type     ON user_accounts (user_type);
CREATE INDEX IF NOT EXISTS idx_user_accounts_status   ON user_accounts (status);
CREATE INDEX IF NOT EXISTS idx_user_accounts_staff_id ON user_accounts (staff_id);

CREATE INDEX IF NOT EXISTS idx_user_clinic_memberships_clinic_id ON user_clinic_memberships (clinic_id);

CREATE INDEX IF NOT EXISTS idx_user_permissions_clinic_id ON user_permissions (clinic_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id   ON user_permissions (user_id);

-- -----------------------------------------------------------------------------
-- 5. CHECK制約
-- -----------------------------------------------------------------------------

-- treatment_items: item_typeとFK列の整合性
ALTER TABLE treatment_items ADD CONSTRAINT chk_treatment_item_ref CHECK (
    CASE item_type
        WHEN 'consultation' THEN consultation_id IS NOT NULL
        WHEN 'procedure'    THEN procedure_id IS NOT NULL
        WHEN 'medicine'     THEN medicine_id IS NOT NULL
        ELSE TRUE
    END
);

-- care_plan_items: typeとFK列の整合性
ALTER TABLE care_plan_items ADD CONSTRAINT chk_care_plan_item_ref CHECK (
    CASE type
        WHEN 'medicine'  THEN medicine_id IS NOT NULL
        WHEN 'treatment' THEN procedure_id IS NOT NULL
        ELSE TRUE
    END
);

-- -----------------------------------------------------------------------------
-- 6. 部分一意インデックス
-- -----------------------------------------------------------------------------

-- is_main=trueは1ユーザーにつき1件のみ
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_clinic_memberships_main
    ON user_clinic_memberships (user_id)
    WHERE is_main = true;

-- -----------------------------------------------------------------------------
-- 7. マスタテーブル codeカラム 部分一意インデックス
--    code='' の空文字は除外（デフォルト値として複数レコードが持てる）
-- -----------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS idx_staffs_code               ON staffs               (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_vaccines_code             ON vaccines             (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_medicines_code            ON medicines            (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_insurances_code           ON insurances           (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_cages_code                ON cages                (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_types_code        ON service_types        (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_consultations_code        ON consultations        (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_procedures_code           ON procedures           (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_hospitalization_plans_code ON hospitalization_plans (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_trimming_courses_code     ON trimming_courses     (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_trimming_options_code     ON trimming_options     (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_examination_types_code    ON examination_types    (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_diagnosis_categories_code ON diagnosis_categories (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_diagnosis_names_code      ON diagnosis_names      (code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_checkup_types_code        ON checkup_types        (code) WHERE code != '';
