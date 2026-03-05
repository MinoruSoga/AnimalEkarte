-- 初期スキーマ（20テーブル完全版）
-- PostgreSQL 初回起動時に自動実行されます
-- FK定義をSQL側から削除。AutoMigrateで全FK管理する。

-- UUID拡張機能
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ====================================================================================
-- 1. Core Tables（独立テーブル）
-- ====================================================================================

-- 病院情報
CREATE TABLE IF NOT EXISTS clinics (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(100) NOT NULL,
  branch_name varchar(100),
  postal_code varchar(10),
  address text,
  phone_number varchar(20),
  fax_number varchar(20),
  registration_number varchar(50),
  director_name varchar(100),
  email varchar(255),
  website varchar(255),
  logo_url text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 在庫品目
CREATE TABLE IF NOT EXISTS inventory_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(200) NOT NULL,
  category varchar(30),  -- medicine, consumable, food, other
  quantity int,
  unit varchar(20),
  min_stock_level int,
  location varchar(100),
  expiry_date date,
  supplier varchar(200),
  last_restocked date,
  status varchar(20),  -- sufficient, low, out_of_stock
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ケージマスタ
CREATE TABLE IF NOT EXISTS cages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code varchar(20),
  name varchar(100) NOT NULL,
  size varchar(50),  -- S, M, L, XL
  type varchar(50),  -- 犬用, 猫用, 共用
  is_available boolean DEFAULT true,
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ====================================================================================
-- 2. Clinic-dependent Tables
-- ====================================================================================

-- スタッフ
CREATE TABLE IF NOT EXISTS staffs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  clinic_id uuid NOT NULL,
  name varchar(100) NOT NULL,
  role varchar(50),  -- veterinarian, nurse, groomer, admin
  email varchar(255),
  phone varchar(20),
  is_active boolean DEFAULT true,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_staffs_clinic_id ON staffs(clinic_id);

-- ====================================================================================
-- 3. Inventory-dependent Tables
-- ====================================================================================

-- 診療項目マスタ
CREATE TABLE IF NOT EXISTS master_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code varchar(20),
  name varchar(200) NOT NULL,
  category varchar(50),  -- examination, vaccine, medicine, cage, trimming_course, trimming_option, diagnosis_category, diagnosis_name, etc.
  price decimal(10,2),
  status varchar(20),  -- active, inactive
  description text,
  inventory_id uuid,
  default_quantity int,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_master_category ON master_items(category);

-- ====================================================================================
-- 4. Core Medical Tables（飼主・ペット）
-- ====================================================================================

-- 飼主
CREATE TABLE IF NOT EXISTS owners (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_number bigint,
  name varchar(100) NOT NULL,
  name_kana varchar(100),
  phone varchar(20),
  email varchar(255),
  address text,
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ペット
CREATE TABLE IF NOT EXISTS pets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id uuid NOT NULL,
  pet_number varchar(20),
  name varchar(100) NOT NULL,
  species varchar(50) NOT NULL,
  breed varchar(100),
  gender varchar(10),  -- オス, メス, 不明
  birth_date date,
  weight decimal(5,2),
  microchip_id varchar(50),
  environment varchar(100),
  status varchar(10),  -- 生存, 死亡
  insurance_name varchar(100),
  insurance_details text,
  last_visit date,
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pets_owner_id ON pets(owner_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pets_pet_number ON pets(pet_number);

-- ====================================================================================
-- 5. Medical Records & Related Tables
-- ====================================================================================

-- 電子カルテ（SOAPS形式）
CREATE TABLE IF NOT EXISTS medical_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  record_no varchar(20),
  pet_id uuid NOT NULL,
  owner_id uuid,
  doctor_id uuid,
  visit_date timestamp WITH TIME ZONE,
  visit_type varchar(20),  -- 初診, 再診
  chief_complaint text,
  subjective text,  -- S: 主観的情報
  objective text,  -- O: 客観的情報
  assessment text,  -- A: 評価
  plan text,  -- P: 計画
  surgery_notes text,  -- S: 手術・特記事項
  diagnosis text,
  treatment text,
  prescription text,
  notes text,
  status varchar(20),  -- 作成中, 確定済
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mr_pet_id ON medical_records(pet_id);
CREATE INDEX IF NOT EXISTS idx_mr_visit_date ON medical_records(visit_date);
CREATE UNIQUE INDEX IF NOT EXISTS idx_mr_record_no ON medical_records(record_no);

-- 検査記録
CREATE TABLE IF NOT EXISTS examinations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pet_id uuid NOT NULL,
  owner_id uuid,
  doctor_id uuid,
  medical_record_id uuid,
  examination_date timestamp WITH TIME ZONE,
  test_type varchar(100),
  machine varchar(100),
  status varchar(20),  -- 依頼中, 検査中, 完了
  result_summary text,
  items json,
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_exam_pet_id ON examinations(pet_id);
CREATE INDEX IF NOT EXISTS idx_exam_status ON examinations(status);

-- 予約
CREATE TABLE IF NOT EXISTS reservations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pet_id uuid NOT NULL,
  owner_id uuid,
  doctor_id uuid,
  start_time timestamp WITH TIME ZONE NOT NULL,
  end_time timestamp WITH TIME ZONE,
  visit_type varchar(20),  -- first, revisit
  service_type varchar(30),  -- 診療, 検診, 検査
  is_designated boolean,
  status varchar(30),  -- pending, confirmed, checked_in, etc.
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_res_pet_id ON reservations(pet_id);
CREATE INDEX IF NOT EXISTS idx_res_start_time ON reservations(start_time);
CREATE INDEX IF NOT EXISTS idx_res_doctor_id ON reservations(doctor_id);
CREATE INDEX IF NOT EXISTS idx_res_status ON reservations(status);

-- ワクチン接種
CREATE TABLE IF NOT EXISTS vaccinations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pet_id uuid NOT NULL,
  owner_id uuid,
  doctor_id uuid,
  vaccine_master_id uuid,
  vaccine_name varchar(100),
  vaccination_date date,
  next_date date,
  lot_number varchar(50),
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vac_pet_id ON vaccinations(pet_id);
CREATE INDEX IF NOT EXISTS idx_vac_next_date ON vaccinations(next_date);

-- トリミング
CREATE TABLE IF NOT EXISTS trimmings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pet_id uuid NOT NULL,
  owner_id uuid,
  staff_id uuid,
  appointment_date timestamp WITH TIME ZONE,
  course varchar(100),
  options json,
  style_request text,
  status varchar(20),  -- 予約, 進行中, 完了
  total_price decimal(10,2),
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ====================================================================================
-- 6. Hospitalization Tables
-- ====================================================================================

-- 入院/ホテル
CREATE TABLE IF NOT EXISTS hospitalizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_no varchar(20),
  pet_id uuid NOT NULL,
  owner_id uuid,
  cage_id uuid,
  type varchar(20),  -- 入院, ホテル
  start_date date NOT NULL,
  end_date date,
  status varchar(20),  -- 入院中, 退院済, 予約, 一時帰宅
  owner_request text,
  staff_notes text,
  memo text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hosp_pet_id ON hospitalizations(pet_id);
CREATE INDEX IF NOT EXISTS idx_hosp_status ON hospitalizations(status);

-- ケアプラン
CREATE TABLE IF NOT EXISTS care_plan_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id uuid NOT NULL,
  master_id uuid,
  type varchar(30),  -- food, medicine, treatment, instruction, item
  name varchar(100) NOT NULL,
  description text,
  timing json,
  status varchar(20),  -- active, completed, discontinued
  unit_price decimal(10,2),
  category varchar(50),
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- デイリーレコード
CREATE TABLE IF NOT EXISTS daily_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id uuid NOT NULL,
  record_date date NOT NULL,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- バイタル
CREATE TABLE IF NOT EXISTS vitals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id uuid NOT NULL,
  staff_id uuid,
  recorded_time time,
  temperature decimal(4,1),
  heart_rate int,
  respiration_rate int,
  weight decimal(5,2),
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ケアログ
CREATE TABLE IF NOT EXISTS care_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id uuid NOT NULL,
  staff_id uuid,
  recorded_time time,
  type varchar(30),  -- food, excretion, medicine, treatment, other
  status varchar(20),  -- completed, partial, skipped
  value varchar(100),
  notes text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- スタッフメモ
CREATE TABLE IF NOT EXISTS staff_notes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id uuid NOT NULL,
  staff_id uuid,
  recorded_time time,
  content text NOT NULL,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ====================================================================================
-- 7. Accounting Tables
-- ====================================================================================

-- 会計
CREATE TABLE IF NOT EXISTS accountings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id uuid,
  pet_id uuid NOT NULL,
  owner_id uuid,
  scheduled_date date,
  completed_at timestamp WITH TIME ZONE,
  status varchar(20),  -- 未収, 保留, 回収済, キャンセル
  subtotal decimal(10,2),
  tax_total decimal(10,2),
  total_amount decimal(10,2),
  insurance_name varchar(100),
  insurance_ratio decimal(3,2),
  insurance_amount decimal(10,2),
  discount_amount decimal(10,2),
  billing_amount decimal(10,2),
  received_amount decimal(10,2),
  change_amount decimal(10,2),
  payment_method varchar(30),  -- 現金, クレジットカード, 電子マネー
  memo text,
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_acc_pet_id ON accountings(pet_id);
CREATE INDEX IF NOT EXISTS idx_acc_status ON accountings(status);

-- 会計明細
CREATE TABLE IF NOT EXISTS accounting_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  accounting_id uuid NOT NULL,
  master_id uuid,
  code varchar(20),
  category varchar(50),
  name varchar(200) NOT NULL,
  unit_price decimal(10,2),
  quantity int,
  tax_rate decimal(3,2),
  is_insurance_applicable boolean,
  source varchar(20),  -- medical_record, manual
  created_at timestamp WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
