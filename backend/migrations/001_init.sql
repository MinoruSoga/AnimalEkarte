-- ====================================================================================
-- 動物病院管理システム 初期スキーマ（31テーブル完全版）
-- DB_DEFINITION.md / ERD.md / AUTH.md 準拠
-- PostgreSQL 初回起動時に自動実行されます
-- ====================================================================================

-- UUID拡張機能
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ====================================================================================
-- ENUM型定義（グローバル列挙型）
-- ====================================================================================

DO $$ BEGIN CREATE TYPE pet_species AS ENUM ('犬', '猫', '鳥', 'その他'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE pet_status AS ENUM ('生存', '死亡'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE medical_record_status AS ENUM ('作成中', '確定済'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE hospitalization_type AS ENUM ('入院', 'ホテル'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE hospitalization_status AS ENUM ('入院中', '退院済', '予約'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled', 'checked_in', 'in_consultation', 'accounting', 'completed'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE visit_type AS ENUM ('first', 'revisit'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE trimming_status AS ENUM ('完了', '予約', '進行中'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE examination_status AS ENUM ('依頼中', '検査中', '完了'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE master_item_status AS ENUM ('active', 'inactive'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE master_category AS ENUM (
  'examination', 'vaccine', 'medicine', 'staff', 'insurance', 'cage',
  'serviceType', 'consultation', 'procedure', 'hospitalization',
  'trimming_course', 'trimming_option', 'diagnosis_category', 'diagnosis_name',
  'checkup'
); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ====================================================================================
-- ENUM型定義（Feature固有列挙型）
-- ====================================================================================

-- owners feature
DO $$ BEGIN CREATE TYPE pet_gender AS ENUM ('雄', '雌', '不明'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE acquisition_type AS ENUM ('購入', '譲渡', '保護', 'その他'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE danger_level AS ENUM ('低', '中', '高'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE membership_type AS ENUM ('非会員', '会員', '退亡者', '他診/準'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- accounting feature
DO $$ BEGIN CREATE TYPE accounting_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- medical-records feature
DO $$ BEGIN CREATE TYPE treatment_status AS ENUM ('未完了', '完了', '-'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE examination_result_status AS ENUM ('normal', 'high', 'low'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- master feature
DO $$ BEGIN CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE staff_role AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE coverage_rate AS ENUM ('50', '70', '80', '100'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE combinable AS ENUM ('yes', 'no'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE body_size AS ENUM ('small', 'medium', 'large'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- trimming feature
DO $$ BEGIN CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- shifts feature
DO $$ BEGIN CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ====================================================================================
-- ENUM型定義（テーブル未使用型）
-- ====================================================================================

DO $$ BEGIN CREATE TYPE visit_reason AS ENUM ('injury', 'vomiting', 'diarrhea', 'skin', 'eye', 'ear', 'dental', 'checkup', 'vaccination', 'other'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE appointment_visit_type AS ENUM ('初診', '再診'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ====================================================================================
-- ENUM型定義（認証認可）
-- ====================================================================================

DO $$ BEGIN CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE job_title AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'general_staff'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE permission_type AS ENUM (
  'account_admin', 'medical', 'medical_read', 'trimming', 'billing',
  'reception', 'hospitalization', 'master_admin', 'shift_admin', 'inventory'
); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ====================================================================================
-- 0. 独立テーブル
-- ====================================================================================

-- 在庫品目
CREATE TABLE IF NOT EXISTS inventory_items (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT NOT NULL,
  category        inventory_category NOT NULL,
  quantity        INTEGER NOT NULL DEFAULT 0,
  unit            TEXT NOT NULL,
  min_stock_level INTEGER NOT NULL DEFAULT 0,
  location        TEXT DEFAULT '',
  expiry_date     DATE,
  supplier        TEXT DEFAULT '',
  last_restocked  DATE,
  status          inventory_status DEFAULT 'sufficient',
  created_at      TIMESTAMPTZ DEFAULT now(),
  updated_at      TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 1. マスタ・クリニックテーブル
-- ====================================================================================

-- 診療項目マスタ（15カテゴリ統合 - STIパターン）
-- ※ parent_id は自己参照FK（ON DELETE CASCADE）
CREATE TABLE IF NOT EXISTS master_items (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code              TEXT NOT NULL,
  name              TEXT NOT NULL,
  category          master_category NOT NULL,
  price             NUMERIC(10,2),
  status            master_item_status DEFAULT 'active',
  description       TEXT DEFAULT '',
  inventory_id      UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
  default_quantity  INTEGER,
  species           vaccine_species,
  "interval"        TEXT,
  parent_id         UUID REFERENCES master_items(id) ON DELETE CASCADE,
  sort_order        INTEGER DEFAULT 0,
  -- カテゴリ固有拡張フィールド（nullable）
  color             TEXT,
  time_condition    TEXT,
  anesthesia        TEXT,
  target_age        TEXT,
  dosage_form       dosage_form,
  medicine_unit     medicine_unit,
  staff_role        staff_role,
  license_number    TEXT,
  email             TEXT,
  password_hash     TEXT,
  user_type         TEXT,
  clinics           TEXT[],
  last_login_at     TIMESTAMPTZ,
  cage_type         cage_type,
  cage_size         cage_size,
  coverage_rate     coverage_rate,
  contact_phone     TEXT,
  target_size       target_size,
  duration          TEXT,
  combinable        combinable,
  body_size         body_size,
  billing_unit      billing_unit,
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- 病院情報（シングルトン）
CREATE TABLE IF NOT EXISTS clinic_info (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                TEXT NOT NULL,
  branch_name         TEXT DEFAULT '',
  postal_code         TEXT DEFAULT '',
  address             TEXT DEFAULT '',
  phone_number        TEXT DEFAULT '',
  fax_number          TEXT DEFAULT '',
  registration_number TEXT DEFAULT '',
  director_name       TEXT DEFAULT '',
  email               TEXT DEFAULT '',
  website             TEXT DEFAULT '',
  logo_url            TEXT,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 2. 認証テーブル
-- ====================================================================================

-- クリニック（複数院対応）
CREATE TABLE IF NOT EXISTS clinics (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                TEXT NOT NULL,
  branch_name         TEXT DEFAULT '',
  postal_code         TEXT DEFAULT '',
  address             TEXT DEFAULT '',
  phone_number        TEXT DEFAULT '',
  fax_number          TEXT DEFAULT '',
  registration_number TEXT DEFAULT '',
  director_name       TEXT DEFAULT '',
  email               TEXT DEFAULT '',
  website             TEXT DEFAULT '',
  logo_url            TEXT,
  is_active           BOOLEAN DEFAULT true,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);

-- ユーザーアカウント
CREATE TABLE IF NOT EXISTS user_accounts (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             TEXT NOT NULL UNIQUE,
  display_name      TEXT NOT NULL,
  display_name_kana TEXT,
  user_type         user_type NOT NULL DEFAULT 'staff',
  job_title         job_title,
  status            account_status NOT NULL DEFAULT 'active',
  avatar_url        TEXT,
  staff_master_id   UUID REFERENCES master_items(id) ON DELETE SET NULL,
  last_login_at     TIMESTAMPTZ,
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- ユーザー・クリニック所属
CREATE TABLE IF NOT EXISTS user_clinic_memberships (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id  UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  is_main    BOOLEAN NOT NULL DEFAULT false,
  joined_at  TIMESTAMPTZ DEFAULT now(),
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, clinic_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_clinic_main
  ON user_clinic_memberships(user_id)
  WHERE is_main = true;

-- ユーザー権限（クリニックスコープ）
CREATE TABLE IF NOT EXISTS user_permissions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id   UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  permission  permission_type NOT NULL,
  granted_by  UUID REFERENCES user_accounts(id) ON DELETE SET NULL,
  granted_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, clinic_id, permission)
);

-- ====================================================================================
-- 3. コア医療テーブル
-- ====================================================================================

-- 飼主（顧客）
CREATE TABLE IF NOT EXISTS owners (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_name       TEXT NOT NULL,
  owner_name_kana  TEXT,
  company          TEXT DEFAULT '',
  postal_code      TEXT DEFAULT '',
  address1         TEXT DEFAULT '',
  address2         TEXT DEFAULT '',
  home_postal_code TEXT DEFAULT '',
  home_address1    TEXT DEFAULT '',
  home_address2    TEXT DEFAULT '',
  birth_date       DATE,
  phone            TEXT DEFAULT '',
  company_phone    TEXT DEFAULT '',
  email            TEXT DEFAULT '',
  remarks          TEXT DEFAULT '',
  is_dangerous     BOOLEAN DEFAULT false,
  discount_rate    NUMERIC(5,2) DEFAULT 0,
  membership_type  membership_type DEFAULT '非会員',
  created_at       TIMESTAMPTZ DEFAULT now(),
  updated_at       TIMESTAMPTZ DEFAULT now()
);

-- ペット（患者）
CREATE TABLE IF NOT EXISTS pets (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id          UUID NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  owner_name        TEXT NOT NULL,
  pet_number        TEXT,
  name              TEXT NOT NULL,
  pet_name_kana     TEXT,
  species           pet_species NOT NULL,
  gender            pet_gender,
  status            pet_status DEFAULT '生存',
  birth_date        DATE,
  breed             TEXT,
  color             TEXT,
  weight            TEXT,
  neutered_date     DATE,
  acquisition_type  acquisition_type,
  danger_level      danger_level,
  food              TEXT,
  environment       TEXT,
  phone             TEXT,
  last_visit        DATE,
  insurance_name    TEXT,
  insurance_details TEXT,
  remarks           TEXT,
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 4. 電子カルテ
-- ====================================================================================

-- 電子カルテ（SOAP形式）
CREATE TABLE IF NOT EXISTS medical_records (
  id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  record_no                TEXT NOT NULL UNIQUE,
  date                     DATE NOT NULL,
  owner_id                 UUID REFERENCES owners(id) ON DELETE SET NULL,
  owner_name               TEXT NOT NULL,
  pet_id                   UUID REFERENCES pets(id) ON DELETE SET NULL,
  pet_name                 TEXT NOT NULL,
  species                  pet_species NOT NULL,
  chief_complaint          TEXT DEFAULT '',
  treatment_policy         TEXT DEFAULT '',
  physical_exam            TEXT DEFAULT '',
  diagnosis_details        TEXT DEFAULT '',
  diagnosis1_category_id   UUID REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis1_name_id       UUID REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis2_category_id   UUID REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis2_name_id       UUID REFERENCES master_items(id) ON DELETE SET NULL,
  doctor                   TEXT NOT NULL,
  status                   medical_record_status DEFAULT '作成中',
  created_at               TIMESTAMPTZ DEFAULT now(),
  updated_at               TIMESTAMPTZ DEFAULT now()
);

-- 治療/処置項目
CREATE TABLE IF NOT EXISTS treatment_items (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  selected          BOOLEAN DEFAULT false,
  status            treatment_status DEFAULT '-',
  content           TEXT NOT NULL,
  memo              TEXT DEFAULT '',
  insurance         BOOLEAN DEFAULT false,
  unit_price        NUMERIC(10,2) DEFAULT 0,
  quantity          INTEGER DEFAULT 1,
  discount_rate     NUMERIC(5,2) DEFAULT 0,
  discount_amount   NUMERIC(10,2) DEFAULT 0,
  inventory_id      UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
  sort_order        INTEGER DEFAULT 0,
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- カルテ内バイタル
CREATE TABLE IF NOT EXISTS vital_entries (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  recorded_at       TIMESTAMPTZ NOT NULL,
  staff             TEXT NOT NULL,
  temperature       NUMERIC(4,1),
  heart_rate        INTEGER,
  respiration_rate  INTEGER,
  weight            NUMERIC(6,2),
  notes             TEXT DEFAULT '',
  created_at        TIMESTAMPTZ DEFAULT now()
);

-- 検査記録
CREATE TABLE IF NOT EXISTS examination_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id            UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  date              DATE NOT NULL,
  owner_name        TEXT NOT NULL,
  pet_name          TEXT NOT NULL,
  test_type         TEXT NOT NULL,
  doctor            TEXT NOT NULL,
  status            examination_status DEFAULT '依頼中',
  result_summary    TEXT DEFAULT '',
  machine           TEXT DEFAULT '',
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- 検査結果項目
CREATE TABLE IF NOT EXISTS examination_record_items (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  examination_record_id  UUID NOT NULL REFERENCES examination_records(id) ON DELETE CASCADE,
  name                   TEXT NOT NULL,
  inspection_value       TEXT DEFAULT '',
  normal_value           TEXT DEFAULT '',
  result                 TEXT DEFAULT '',
  unit                   TEXT DEFAULT '',
  ref                    TEXT DEFAULT '',
  status                 examination_result_status,
  sort_order             INTEGER DEFAULT 0,
  created_at             TIMESTAMPTZ DEFAULT now()
);

-- 予防接種記録
CREATE TABLE IF NOT EXISTS vaccination_records (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id   UUID NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id              UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  owner_name          TEXT NOT NULL,
  pet_name            TEXT NOT NULL,
  vaccine_name        TEXT NOT NULL,
  date                DATE NOT NULL,
  next_date           DATE,
  next_schedule_type  next_schedule_type,
  doctor              TEXT DEFAULT '',
  supplemental        TEXT DEFAULT '',
  lot1                TEXT DEFAULT '',
  lot2                TEXT DEFAULT '',
  lot3                TEXT DEFAULT '',
  lot4                TEXT DEFAULT '',
  remarks             TEXT DEFAULT '',
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);

-- 定期健診記録
CREATE TABLE IF NOT EXISTS checkup_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id            UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  owner_name        TEXT NOT NULL,
  pet_name          TEXT NOT NULL,
  checkup_type      TEXT NOT NULL,
  date              DATE NOT NULL,
  next_date         DATE,
  doctor            TEXT DEFAULT '',
  result            TEXT DEFAULT '',
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 5. 入院テーブル
-- ====================================================================================

-- 入院/ホテル記録
CREATE TABLE IF NOT EXISTS hospitalizations (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id              UUID REFERENCES owners(id) ON DELETE SET NULL,
  owner_name            TEXT NOT NULL,
  pet_id                UUID REFERENCES pets(id) ON DELETE SET NULL,
  pet_name              TEXT NOT NULL,
  species               pet_species NOT NULL,
  hospitalization_type  hospitalization_type NOT NULL,
  start_date            DATE NOT NULL,
  end_date              DATE NOT NULL,
  status                hospitalization_status DEFAULT '予約',
  cage_id               UUID REFERENCES master_items(id) ON DELETE SET NULL,
  doctor_name           TEXT,
  memo                  TEXT DEFAULT '',
  owner_request         TEXT DEFAULT '',
  staff_notes           TEXT DEFAULT '',
  created_at            TIMESTAMPTZ DEFAULT now(),
  updated_at            TIMESTAMPTZ DEFAULT now()
);

-- ケアプラン項目
CREATE TABLE IF NOT EXISTS care_plan_items (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  type                care_plan_type NOT NULL,
  name                TEXT NOT NULL,
  description         TEXT DEFAULT '',
  timing              plan_timing[] DEFAULT '{}',
  status              care_plan_status DEFAULT 'active',
  notes               TEXT DEFAULT '',
  master_id           UUID REFERENCES master_items(id) ON DELETE SET NULL,
  unit_price          NUMERIC(10,2),
  category            TEXT,
  sort_order          INTEGER DEFAULT 0,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);

-- 日次記録コンテナ
CREATE TABLE IF NOT EXISTS daily_records (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  date                DATE NOT NULL,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now(),
  UNIQUE (hospitalization_id, date)
);

-- 入院バイタルサイン
CREATE TABLE IF NOT EXISTS vital_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id   UUID NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time              TEXT NOT NULL,
  temperature       NUMERIC(4,1),
  heart_rate        INTEGER,
  respiration_rate  INTEGER,
  weight            NUMERIC(6,2),
  notes             TEXT DEFAULT '',
  staff             TEXT NOT NULL,
  created_at        TIMESTAMPTZ DEFAULT now()
);

-- ケアログ（食事/排泄/投薬等）
CREATE TABLE IF NOT EXISTS care_log_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id   UUID NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time              TEXT NOT NULL,
  type              care_log_type NOT NULL,
  status            care_log_status NOT NULL,
  value             TEXT DEFAULT '',
  staff             TEXT NOT NULL,
  notes             TEXT DEFAULT '',
  created_at        TIMESTAMPTZ DEFAULT now()
);

-- スタッフメモ
CREATE TABLE IF NOT EXISTS staff_note_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id   UUID NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time              TEXT NOT NULL,
  content           TEXT NOT NULL,
  staff             TEXT NOT NULL,
  created_at        TIMESTAMPTZ DEFAULT now()
);

-- 入院治療プラン
CREATE TABLE IF NOT EXISTS treatment_plans (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  treatment_content   TEXT NOT NULL,
  memo                TEXT DEFAULT '',
  insurance           BOOLEAN DEFAULT false,
  unit_price          NUMERIC(10,2) DEFAULT 0,
  quantity            INTEGER DEFAULT 1,
  discount_rate       NUMERIC(5,2) DEFAULT 0,
  discount_amount     NUMERIC(10,2) DEFAULT 0,
  subtotal            NUMERIC(10,2) DEFAULT 0,
  sort_order          INTEGER DEFAULT 0,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 6. 予約・トリミング・会計テーブル
-- ====================================================================================

-- 予約
CREATE TABLE IF NOT EXISTS reservation_appointments (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  start_time     TIMESTAMPTZ NOT NULL,
  end_time       TIMESTAMPTZ NOT NULL,
  owner_name     TEXT NOT NULL,
  pet_name       TEXT NOT NULL,
  pet_id         UUID REFERENCES pets(id) ON DELETE SET NULL,
  visit_type     visit_type NOT NULL,
  type           VARCHAR NOT NULL,
  doctor         TEXT NOT NULL,
  is_designated  BOOLEAN DEFAULT false,
  status         reservation_status DEFAULT 'pending',
  notes          TEXT DEFAULT '',
  created_at     TIMESTAMPTZ DEFAULT now(),
  updated_at     TIMESTAMPTZ DEFAULT now()
);

-- トリミング記録
CREATE TABLE IF NOT EXISTS trimming_records (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  date            DATE NOT NULL,
  pet_id          UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  pet_number      TEXT NOT NULL,
  pet_name        TEXT NOT NULL,
  owner_name      TEXT NOT NULL,
  species         pet_species NOT NULL,
  weight          TEXT DEFAULT '',
  style_request   TEXT DEFAULT '',
  staff           TEXT NOT NULL,
  status          trimming_status DEFAULT '予約',
  course_id       UUID REFERENCES master_items(id) ON DELETE SET NULL,
  bw              TEXT DEFAULT '',
  bw_unit         body_weight_unit DEFAULT 'Kg',
  bt              TEXT DEFAULT '',
  used_shampoo    TEXT DEFAULT '',
  used_ribbon     TEXT DEFAULT '',
  treatment       TEXT DEFAULT '',
  medicine        TEXT DEFAULT '',
  charge          TEXT DEFAULT '',
  remarks         TEXT DEFAULT '',
  final_check     TEXT DEFAULT '',
  style_image     TEXT,
  completed_image TEXT,
  created_at      TIMESTAMPTZ DEFAULT now(),
  updated_at      TIMESTAMPTZ DEFAULT now()
);

-- トリミング記録 ↔ オプション（N:M 中間テーブル）
CREATE TABLE IF NOT EXISTS trimming_record_options (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  trimming_record_id  UUID NOT NULL REFERENCES trimming_records(id) ON DELETE CASCADE,
  option_id           UUID NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  sort_order          INTEGER DEFAULT 0,
  UNIQUE (trimming_record_id, option_id)
);

-- 会計レコード
CREATE TABLE IF NOT EXISTS accountings (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id  UUID UNIQUE REFERENCES medical_records(id) ON DELETE SET NULL,
  hospitalization_id UUID REFERENCES hospitalizations(id) ON DELETE SET NULL,
  owner_id           UUID NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  owner_name         TEXT NOT NULL,
  pet_id             UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  pet_name           TEXT NOT NULL,
  pet_species        pet_species,
  status             accounting_status DEFAULT 'waiting',
  scheduled_date     DATE NOT NULL,
  completed_at       TIMESTAMPTZ,
  memo               TEXT DEFAULT '',
  created_at         TIMESTAMPTZ DEFAULT now(),
  updated_at         TIMESTAMPTZ DEFAULT now()
);

-- 会計明細行
CREATE TABLE IF NOT EXISTS accounting_items (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  accounting_id           UUID NOT NULL REFERENCES accountings(id) ON DELETE CASCADE,
  code                    TEXT,
  category                item_category NOT NULL,
  name                    TEXT NOT NULL,
  unit_price              NUMERIC(10,2) NOT NULL,
  quantity                INTEGER NOT NULL DEFAULT 1,
  tax_rate                NUMERIC(3,2) NOT NULL DEFAULT 0.10,
  is_insurance_applicable BOOLEAN DEFAULT false,
  source                  item_source DEFAULT 'manual',
  sort_order              INTEGER DEFAULT 0,
  created_at              TIMESTAMPTZ DEFAULT now()
);

-- 支払情報（1:1）
CREATE TABLE IF NOT EXISTS payment_infos (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  accounting_id    UUID NOT NULL UNIQUE REFERENCES accountings(id) ON DELETE CASCADE,
  subtotal         NUMERIC(10,2) NOT NULL DEFAULT 0,
  tax_total        NUMERIC(10,2) NOT NULL DEFAULT 0,
  total_amount     NUMERIC(10,2) NOT NULL DEFAULT 0,
  insurance_name   TEXT,
  insurance_ratio  NUMERIC(3,2),
  insurance_amount NUMERIC(10,2) DEFAULT 0,
  discount_amount  NUMERIC(10,2) DEFAULT 0,
  billing_amount   NUMERIC(10,2) NOT NULL DEFAULT 0,
  received_amount  NUMERIC(10,2) DEFAULT 0,
  change_amount    NUMERIC(10,2) DEFAULT 0,
  method           payment_method DEFAULT 'cash',
  created_at       TIMESTAMPTZ DEFAULT now(),
  updated_at       TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- 7. シフト管理テーブル
-- ====================================================================================

-- シフトエントリ（staff_id → master_items.category='staff'）
CREATE TABLE IF NOT EXISTS shift_entries (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id    UUID NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  date        DATE NOT NULL,
  shift_type  shift_type NOT NULL,
  start_time  TEXT DEFAULT '',
  end_time    TEXT DEFAULT '',
  note        TEXT DEFAULT '',
  created_at  TIMESTAMPTZ DEFAULT now(),
  updated_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE (staff_id, date)
);

-- ====================================================================================
-- 8. マスタサブテーブル
-- ====================================================================================

-- 検査マスタの検査項目定義
CREATE TABLE IF NOT EXISTS master_item_inspections (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  master_item_id   UUID NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  name             TEXT NOT NULL,
  inspection_value TEXT DEFAULT '',
  normal_value     TEXT DEFAULT '',
  sort_order       INTEGER DEFAULT 0,
  created_at       TIMESTAMPTZ DEFAULT now()
);

-- ====================================================================================
-- インデックス定義
-- ====================================================================================

-- pets
CREATE INDEX IF NOT EXISTS idx_pets_owner_id ON pets(owner_id);
CREATE INDEX IF NOT EXISTS idx_pets_species ON pets(species);
CREATE INDEX IF NOT EXISTS idx_pets_name ON pets(name);

-- medical_records
CREATE INDEX IF NOT EXISTS idx_medical_records_pet_id ON medical_records(pet_id);
CREATE INDEX IF NOT EXISTS idx_medical_records_owner_id ON medical_records(owner_id);
CREATE INDEX IF NOT EXISTS idx_medical_records_date ON medical_records(date DESC);
CREATE INDEX IF NOT EXISTS idx_medical_records_status ON medical_records(status);

-- hospitalizations
CREATE INDEX IF NOT EXISTS idx_hospitalizations_pet_id ON hospitalizations(pet_id);
CREATE INDEX IF NOT EXISTS idx_hospitalizations_status ON hospitalizations(status);
CREATE INDEX IF NOT EXISTS idx_hospitalizations_start_date ON hospitalizations(start_date DESC);

-- reservation_appointments
CREATE INDEX IF NOT EXISTS idx_reservations_start_time ON reservation_appointments(start_time);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservation_appointments(status);
CREATE INDEX IF NOT EXISTS idx_reservations_pet_id ON reservation_appointments(pet_id);

-- trimming_records
CREATE INDEX IF NOT EXISTS idx_trimming_records_pet_id ON trimming_records(pet_id);
CREATE INDEX IF NOT EXISTS idx_trimming_records_date ON trimming_records(date DESC);
CREATE INDEX IF NOT EXISTS idx_trimming_records_status ON trimming_records(status);

-- accountings
CREATE INDEX IF NOT EXISTS idx_accountings_owner_id ON accountings(owner_id);
CREATE INDEX IF NOT EXISTS idx_accountings_pet_id ON accountings(pet_id);
CREATE INDEX IF NOT EXISTS idx_accountings_status ON accountings(status);
CREATE INDEX IF NOT EXISTS idx_accountings_scheduled_date ON accountings(scheduled_date DESC);

-- treatment_items
CREATE INDEX IF NOT EXISTS idx_treatment_items_record_id ON treatment_items(medical_record_id);

-- vital_entries
CREATE INDEX IF NOT EXISTS idx_vital_entries_record_id ON vital_entries(medical_record_id);

-- examination_records
CREATE INDEX IF NOT EXISTS idx_examination_records_record_id ON examination_records(medical_record_id);
CREATE INDEX IF NOT EXISTS idx_examination_records_pet_id ON examination_records(pet_id);

-- vaccination_records
CREATE INDEX IF NOT EXISTS idx_vaccination_records_record_id ON vaccination_records(medical_record_id);
CREATE INDEX IF NOT EXISTS idx_vaccination_records_pet_id ON vaccination_records(pet_id);

-- checkup_records
CREATE INDEX IF NOT EXISTS idx_checkup_records_record_id ON checkup_records(medical_record_id);
CREATE INDEX IF NOT EXISTS idx_checkup_records_pet_id ON checkup_records(pet_id);

-- care_plan_items
CREATE INDEX IF NOT EXISTS idx_care_plan_items_hosp_id ON care_plan_items(hospitalization_id);

-- vital_records, care_log_records, staff_note_records
CREATE INDEX IF NOT EXISTS idx_vital_records_daily_id ON vital_records(daily_record_id);
CREATE INDEX IF NOT EXISTS idx_care_log_records_daily_id ON care_log_records(daily_record_id);
CREATE INDEX IF NOT EXISTS idx_staff_note_records_daily_id ON staff_note_records(daily_record_id);

-- treatment_plans
CREATE INDEX IF NOT EXISTS idx_treatment_plans_hosp_id ON treatment_plans(hospitalization_id);

-- accounting_items
CREATE INDEX IF NOT EXISTS idx_accounting_items_acct_id ON accounting_items(accounting_id);

-- master_items
CREATE INDEX IF NOT EXISTS idx_master_items_category ON master_items(category);
CREATE INDEX IF NOT EXISTS idx_master_items_parent_id ON master_items(parent_id);
CREATE INDEX IF NOT EXISTS idx_master_items_status ON master_items(status);
CREATE INDEX IF NOT EXISTS idx_master_items_category_sort ON master_items(category, parent_id, sort_order);

-- inventory_items
CREATE INDEX IF NOT EXISTS idx_inventory_items_category ON inventory_items(category);
CREATE INDEX IF NOT EXISTS idx_inventory_items_status ON inventory_items(status);

-- shift_entries
CREATE INDEX IF NOT EXISTS idx_shift_entries_date ON shift_entries(date);

-- 認証認可テーブル
CREATE INDEX IF NOT EXISTS idx_user_accounts_email ON user_accounts(email);
CREATE INDEX IF NOT EXISTS idx_user_accounts_user_type ON user_accounts(user_type);
CREATE INDEX IF NOT EXISTS idx_user_accounts_status ON user_accounts(status);
CREATE INDEX IF NOT EXISTS idx_user_accounts_staff_master ON user_accounts(staff_master_id);
CREATE INDEX IF NOT EXISTS idx_user_clinic_memberships_clinic ON user_clinic_memberships(clinic_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_clinic ON user_permissions(clinic_id);
