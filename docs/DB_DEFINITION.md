# 動物病院管理システム DB定義書

**更新日:** 2026-03-12
**バージョン:** 2.0（31テーブル完全版）
**データベース:** PostgreSQL 18
**ORM:** GORM（Go）

---

## 概要

| 項目 | 値 |
|------|-----|
| テーブル数 | 31 |
| ENUM型数 | 30+ |
| UUID拡張 | `uuid-ossp` |
| タイムゾーン | TIMESTAMPTZ（JST想定） |
| 文字コード | UTF-8 |
| 接続先 | `db:5432/ekarte_db`（Docker内部） |

---

## テーブル一覧

| # | テーブル名 | 区分 | 説明 | 主要FK数 |
|---|------------|------|------|---------|
| 1 | `inventory_items` | 独立 | 在庫品目 | 0 |
| 2 | `master_items` | マスタ | 診療項目マスタ（15カテゴリ統合） | 2（自己参照+inventory） |
| 3 | `clinic_info` | マスタ | 病院情報（シングルトン） | 0 |
| 4 | `clinics` | 認証 | クリニック（マルチクリニック対応） | 0 |
| 5 | `user_accounts` | 認証 | ユーザーアカウント | 1 |
| 6 | `user_clinic_memberships` | 認証 | ユーザー・クリニック所属 | 2 |
| 7 | `user_permissions` | 認証 | ユーザー権限 | 3 |
| 8 | `owners` | コア | 飼主 | 0 |
| 9 | `pets` | コア | ペット | 1 |
| 10 | `medical_records` | カルテ | 電子カルテ | 6 |
| 11 | `treatment_items` | カルテ | 治療項目 | 2 |
| 12 | `vital_entries` | カルテ | バイタルサイン（カルテ内） | 1 |
| 13 | `examination_records` | カルテ | 検査記録 | 2 |
| 14 | `examination_record_items` | カルテ | 検査結果項目 | 1 |
| 15 | `vaccination_records` | カルテ | 予防接種記録 | 2 |
| 16 | `checkup_records` | カルテ | 定期健診記録 | 2 |
| 17 | `hospitalizations` | 入院 | 入院/ホテル記録 | 3 |
| 18 | `care_plan_items` | 入院 | ケアプラン項目 | 2 |
| 19 | `daily_records` | 入院 | 日次記録コンテナ | 1 |
| 20 | `vital_records` | 入院 | バイタルサイン（入院内） | 1 |
| 21 | `care_log_records` | 入院 | ケアログ | 1 |
| 22 | `staff_note_records` | 入院 | スタッフメモ | 1 |
| 23 | `treatment_plans` | 入院 | 入院治療プラン | 1 |
| 24 | `reservation_appointments` | 予約 | 予約 | 1 |
| 25 | `trimming_records` | トリミング | トリミング記録 | 2 |
| 26 | `trimming_record_options` | トリミング | オプション中間テーブル（N:M） | 2 |
| 27 | `accountings` | 会計 | 会計レコード | 4 |
| 28 | `accounting_items` | 会計 | 会計明細行 | 1 |
| 29 | `payment_infos` | 会計 | 支払情報（1:1） | 1 |
| 30 | `shift_entries` | シフト | シフトエントリ | 1 |
| 31 | `master_item_inspections` | マスタ | 検査マスタ項目定義 | 1 |

---

## ENUM型定義

### グローバルENUM型

```sql
-- UUID拡張
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ペット関連
CREATE TYPE pet_species AS ENUM ('犬', '猫', '鳥', 'その他');
CREATE TYPE pet_status  AS ENUM ('生存', '死亡');

-- 電子カルテ
CREATE TYPE medical_record_status AS ENUM ('作成中', '確定済');

-- 入院
CREATE TYPE hospitalization_type   AS ENUM ('入院', 'ホテル');
CREATE TYPE hospitalization_status AS ENUM ('入院中', '退院済', '予約');
CREATE TYPE care_plan_type         AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item');
CREATE TYPE care_plan_status       AS ENUM ('active', 'completed', 'discontinued');
CREATE TYPE care_log_type          AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other');
CREATE TYPE care_log_status        AS ENUM ('completed', 'partial', 'skipped');
CREATE TYPE plan_timing            AS ENUM ('morning', 'noon', 'night');

-- 予約
CREATE TYPE reservation_status AS ENUM (
  'confirmed', 'pending', 'cancelled',
  'checked_in', 'in_consultation', 'accounting', 'completed'
);
CREATE TYPE visit_type AS ENUM ('first', 'revisit');

-- その他
CREATE TYPE trimming_status          AS ENUM ('完了', '予約', '進行中');
CREATE TYPE examination_status       AS ENUM ('依頼中', '検査中', '完了');
CREATE TYPE master_item_status       AS ENUM ('active', 'inactive');
CREATE TYPE master_category          AS ENUM (
  'examination', 'vaccine', 'medicine', 'staff', 'insurance', 'cage',
  'serviceType', 'consultation', 'procedure', 'hospitalization',
  'trimming_course', 'trimming_option', 'diagnosis_category', 'diagnosis_name', 'checkup'
);
CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other');
CREATE TYPE inventory_status   AS ENUM ('sufficient', 'low', 'out_of_stock');
```

### Feature固有ENUM型

```sql
-- owners / pets
CREATE TYPE pet_gender       AS ENUM ('雄', '雌', '不明');
CREATE TYPE acquisition_type AS ENUM ('購入', '譲渡', '保護', 'その他');
CREATE TYPE danger_level     AS ENUM ('低', '中', '高');
CREATE TYPE membership_type  AS ENUM ('非会員', '会員', '退亡者', '他診/準');

-- accounting
CREATE TYPE accounting_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending');
CREATE TYPE payment_method    AS ENUM ('cash', 'credit_card', 'electronic_money');
CREATE TYPE item_category     AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other');
CREATE TYPE item_source       AS ENUM ('medical_record', 'manual', 'hospitalization');

-- medical-records
CREATE TYPE treatment_status        AS ENUM ('未完了', '完了', '-');
CREATE TYPE next_schedule_type      AS ENUM ('3weeks', '4weeks', '1year', 'other');
CREATE TYPE examination_result_status AS ENUM ('normal', 'high', 'low');

-- master
CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both');
CREATE TYPE dosage_form     AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder');
CREATE TYPE medicine_unit   AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram');
CREATE TYPE staff_role      AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager');
CREATE TYPE cage_type       AS ENUM ('icu', 'dog', 'cat', 'general');
CREATE TYPE cage_size       AS ENUM ('small', 'medium', 'large');
CREATE TYPE coverage_rate   AS ENUM ('50', '70', '80', '100');
CREATE TYPE target_size     AS ENUM ('small', 'medium', 'large', 'cat');
CREATE TYPE combinable      AS ENUM ('yes', 'no');
CREATE TYPE body_size       AS ENUM ('small', 'medium', 'large');
CREATE TYPE billing_unit    AS ENUM ('per_day', 'per_night');

-- trimming
CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g');

-- shifts
CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave');
```

### 認証・認可ENUM型

```sql
CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff');
CREATE TYPE job_title AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'general_staff');
CREATE TYPE permission_type AS ENUM (
  'account_admin', 'medical', 'medical_read', 'trimming', 'billing',
  'reception', 'hospitalization', 'master_admin', 'shift_admin', 'inventory'
);
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');
```

---

## テーブル定義（詳細）

### 0. 独立テーブル

#### `inventory_items` — 在庫品目

```sql
CREATE TABLE inventory_items (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT        NOT NULL,
  category        inventory_category NOT NULL,
  quantity        INTEGER     NOT NULL DEFAULT 0,
  unit            TEXT        NOT NULL,
  min_stock_level INTEGER     NOT NULL DEFAULT 0,
  location        TEXT        DEFAULT '',
  expiry_date     DATE,
  supplier        TEXT        DEFAULT '',
  last_restocked  DATE,
  status          inventory_status DEFAULT 'sufficient',
  created_at      TIMESTAMPTZ DEFAULT now(),
  updated_at      TIMESTAMPTZ DEFAULT now()
);
```

| カラム | 型 | 説明 |
|--------|-----|------|
| `id` | UUID PK | 自動生成 |
| `name` | TEXT | 品目名 |
| `category` | inventory_category | 医薬品/消耗品/フード/その他 |
| `quantity` | INTEGER | 現在在庫数 |
| `unit` | TEXT | 単位（本、袋、錠等） |
| `min_stock_level` | INTEGER | 発注点（これを下回るとlow/out_of_stock） |
| `status` | inventory_status | sufficient/low/out_of_stock |

---

### 1. マスタ・クリニックテーブル

#### `master_items` — 診療項目マスタ（15カテゴリ統合・STIパターン）

```sql
CREATE TABLE master_items (
  id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
  code             TEXT           NOT NULL DEFAULT '',
  name             TEXT           NOT NULL,
  category         master_category NOT NULL,
  price            NUMERIC(10,2),
  status           master_item_status DEFAULT 'active',
  description      TEXT           DEFAULT '',
  inventory_id     UUID           REFERENCES inventory_items(id) ON DELETE SET NULL,
  default_quantity INTEGER,
  species          vaccine_species,
  "interval"       TEXT,
  parent_id        UUID           REFERENCES master_items(id) ON DELETE CASCADE,
  sort_order       INTEGER        DEFAULT 0,
  -- カテゴリ固有拡張フィールド（nullable）
  color            TEXT,                         -- serviceType: 表示カラー
  time_condition   TEXT,                         -- consultation: 適用区分
  anesthesia       TEXT,                         -- procedure: 麻酔区分
  target_age       TEXT,                         -- checkup: 対象年齢
  dosage_form      dosage_form,                  -- medicine: 剤形
  medicine_unit    medicine_unit,                -- medicine: 単位
  staff_role       staff_role,                   -- staff: ロール
  license_number   TEXT,                         -- staff: 資格番号
  email            TEXT,                         -- staff: メールアドレス
  password_hash    TEXT,                         -- staff: パスワードハッシュ
  user_type        TEXT,                         -- staff: ユーザー種別
  clinics          TEXT[],                       -- staff: 所属医院ID配列
  last_login_at    TIMESTAMPTZ,                  -- staff: 最終ログイン
  cage_type        cage_type,                    -- cage: ケージタイプ
  cage_size        cage_size,                    -- cage: ケージサイズ
  coverage_rate    coverage_rate,               -- insurance: 補償割合
  contact_phone    TEXT,                         -- insurance: 請求先電話
  target_size      target_size,                 -- trimming_course: 対象サイズ
  duration         TEXT,                         -- trimming_course/procedure: 所要時間
  combinable       combinable,                   -- trimming_option: 併用可否
  body_size        body_size,                    -- hospitalization: 体格区分
  billing_unit     billing_unit,                -- hospitalization: 課金単位
  created_at       TIMESTAMPTZ    DEFAULT now(),
  updated_at       TIMESTAMPTZ    DEFAULT now()
);
```

**マスタカテゴリ一覧（15種）:**

| カテゴリキー | 用途 | 固有フィールド |
|------------|------|--------------|
| `examination` | 検査マスタ | master_item_inspections でサブ定義 |
| `vaccine` | 予防接種マスタ | species, interval |
| `medicine` | 薬剤マスタ | dosage_form, medicine_unit |
| `staff` | スタッフマスタ | staff_role, license_number, email, clinics, user_type |
| `insurance` | 保険会社マスタ | coverage_rate, contact_phone |
| `cage` | ケージマスタ | cage_type, cage_size |
| `serviceType` | 予約区分マスタ | color |
| `consultation` | 診察マスタ | time_condition, duration |
| `procedure` | 処置マスタ | duration, anesthesia |
| `hospitalization` | 入院マスタ | body_size, billing_unit |
| `trimming_course` | トリミングコース | target_size, duration |
| `trimming_option` | トリミングオプション | duration, combinable |
| `diagnosis_category` | 診断カテゴリマスタ | （共通フィールドのみ） |
| `diagnosis_name` | 診断名マスタ | parent_id → diagnosis_category |
| `checkup` | 定期健診マスタ | interval, target_age |

#### `clinic_info` — 病院情報（シングルトン）

```sql
CREATE TABLE clinic_info (
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
```

---

### 2. 認証テーブル

#### `clinics` — クリニック（マルチクリニック対応）

```sql
CREATE TABLE clinics (
  id                  UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
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
  logo_url            TEXT,
  is_active           BOOLEAN DEFAULT true,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);
```

#### `user_accounts` — ユーザーアカウント

```sql
CREATE TABLE user_accounts (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  email             TEXT         NOT NULL UNIQUE,
  display_name      TEXT         NOT NULL,
  display_name_kana TEXT         DEFAULT '',
  user_type         user_type    NOT NULL DEFAULT 'staff',
  job_title         job_title,
  status            account_status DEFAULT 'active',
  avatar_url        TEXT,
  staff_master_id   UUID         REFERENCES master_items(id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ  DEFAULT now(),
  updated_at        TIMESTAMPTZ  DEFAULT now()
);
```

#### `user_clinic_memberships` — ユーザー・クリニック所属（N:M）

```sql
CREATE TABLE user_clinic_memberships (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id  UUID        NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  is_main    BOOLEAN     DEFAULT false,
  joined_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE (user_id, clinic_id)
);
```

#### `user_permissions` — ユーザー権限

```sql
CREATE TABLE user_permissions (
  id          UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID             NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id   UUID             NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  permission  permission_type  NOT NULL,
  granted_by  UUID             REFERENCES user_accounts(id) ON DELETE SET NULL,
  granted_at  TIMESTAMPTZ      DEFAULT now()
);
```

---

### 3. コアテーブル

#### `owners` — 飼主

```sql
CREATE TABLE owners (
  id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_name       TEXT           NOT NULL,
  owner_name_kana  TEXT           DEFAULT '',
  company          TEXT           DEFAULT '',
  postal_code      TEXT           DEFAULT '',
  address1         TEXT           DEFAULT '',
  address2         TEXT           DEFAULT '',
  home_postal_code TEXT           DEFAULT '',
  home_address1    TEXT           DEFAULT '',
  home_address2    TEXT           DEFAULT '',
  phone            TEXT           DEFAULT '',
  company_phone    TEXT           DEFAULT '',
  email            TEXT           DEFAULT '',
  remarks          TEXT           DEFAULT '',
  is_dangerous     BOOLEAN        DEFAULT false,
  discount_rate    NUMERIC(5,2)   DEFAULT 0,
  membership_type  membership_type DEFAULT '非会員',
  created_at       TIMESTAMPTZ    DEFAULT now(),
  updated_at       TIMESTAMPTZ    DEFAULT now()
);
```

#### `pets` — ペット

```sql
CREATE TABLE pets (
  id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id         UUID             NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  pet_number       TEXT             DEFAULT '',
  name             TEXT             NOT NULL,
  pet_name_kana    TEXT             DEFAULT '',
  species          pet_species      NOT NULL,
  gender           pet_gender,
  status           pet_status       DEFAULT '生存',
  birth_date       DATE,
  breed            TEXT             DEFAULT '',
  color            TEXT             DEFAULT '',
  weight           TEXT             DEFAULT '',
  neutered_date    DATE,
  acquisition_type acquisition_type DEFAULT 'その他',
  danger_level     danger_level     DEFAULT '低',
  food             TEXT             DEFAULT '',
  environment      TEXT             DEFAULT '',
  phone            TEXT             DEFAULT '',
  last_visit       DATE,
  insurance_name   TEXT             DEFAULT '',
  insurance_details TEXT            DEFAULT '',
  remarks          TEXT             DEFAULT '',
  created_at       TIMESTAMPTZ      DEFAULT now(),
  updated_at       TIMESTAMPTZ      DEFAULT now()
);
```

---

### 4. 電子カルテテーブル

#### `medical_records` — 電子カルテ

```sql
CREATE TABLE medical_records (
  id                     UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
  record_no              TEXT                 NOT NULL UNIQUE,
  date                   DATE                 NOT NULL,
  owner_id               UUID                 REFERENCES owners(id) ON DELETE SET NULL,
  owner_name             TEXT                 NOT NULL,
  pet_id                 UUID                 REFERENCES pets(id) ON DELETE SET NULL,
  pet_name               TEXT                 NOT NULL,
  species                pet_species          NOT NULL,
  chief_complaint        TEXT                 DEFAULT '',
  treatment_policy       TEXT                 DEFAULT '',
  physical_exam          TEXT                 DEFAULT '',
  diagnosis_details      TEXT                 DEFAULT '',
  diagnosis1_category_id UUID                 REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis1_name_id     UUID                 REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis2_category_id UUID                 REFERENCES master_items(id) ON DELETE SET NULL,
  diagnosis2_name_id     UUID                 REFERENCES master_items(id) ON DELETE SET NULL,
  doctor                 TEXT                 DEFAULT '',
  status                 medical_record_status DEFAULT '作成中',
  created_at             TIMESTAMPTZ          DEFAULT now(),
  updated_at             TIMESTAMPTZ          DEFAULT now()
);
```

#### `treatment_items` — 治療項目

```sql
CREATE TABLE treatment_items (
  id                UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID             NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  selected          BOOLEAN          DEFAULT false,
  status            treatment_status DEFAULT '未完了',
  content           TEXT             NOT NULL,
  memo              TEXT             DEFAULT '',
  insurance         BOOLEAN          DEFAULT false,
  unit_price        NUMERIC(10,2)    DEFAULT 0,
  quantity          INTEGER          DEFAULT 1,
  discount_rate     NUMERIC(5,2)     DEFAULT 0,
  discount_amount   NUMERIC(10,2)    DEFAULT 0,
  inventory_id      UUID             REFERENCES inventory_items(id) ON DELETE SET NULL,
  sort_order        INTEGER          DEFAULT 0,
  created_at        TIMESTAMPTZ      DEFAULT now(),
  updated_at        TIMESTAMPTZ      DEFAULT now()
);
```

#### `vital_entries` — バイタル（カルテ内）

```sql
CREATE TABLE vital_entries (
  id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID          NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  recorded_at       TIMESTAMPTZ   NOT NULL,
  staff             TEXT          DEFAULT '',
  temperature       NUMERIC(4,1),   -- 体温 (℃)
  heart_rate        INTEGER,        -- 心拍数 (bpm)
  respiration_rate  INTEGER,        -- 呼吸数 (/min)
  weight            NUMERIC(6,2),   -- 体重 (kg)
  notes             TEXT          DEFAULT '',
  created_at        TIMESTAMPTZ   DEFAULT now()
);
```

#### `examination_records` — 検査記録

```sql
CREATE TABLE examination_records (
  id                UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID               NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id            UUID               NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  date              DATE               NOT NULL,
  owner_name        TEXT               NOT NULL,
  pet_name          TEXT               NOT NULL,
  test_type         TEXT               NOT NULL,
  doctor            TEXT               DEFAULT '',
  status            examination_status DEFAULT '依頼中',
  result_summary    TEXT               DEFAULT '',
  machine           TEXT               DEFAULT '',
  created_at        TIMESTAMPTZ        DEFAULT now(),
  updated_at        TIMESTAMPTZ        DEFAULT now()
);
```

#### `examination_record_items` — 検査結果項目

```sql
CREATE TABLE examination_record_items (
  id                     UUID                       PRIMARY KEY DEFAULT gen_random_uuid(),
  examination_record_id  UUID                       NOT NULL REFERENCES examination_records(id) ON DELETE CASCADE,
  name                   TEXT                       NOT NULL,
  inspection_value       TEXT                       DEFAULT '',
  normal_value           TEXT                       DEFAULT '',
  result                 TEXT                       DEFAULT '',
  unit                   TEXT                       DEFAULT '',
  ref                    TEXT                       DEFAULT '',
  status                 examination_result_status,
  sort_order             INTEGER                    DEFAULT 0,
  created_at             TIMESTAMPTZ                DEFAULT now()
);
```

#### `vaccination_records` — 予防接種記録

```sql
CREATE TABLE vaccination_records (
  id                  UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id   UUID               NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id              UUID               NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  owner_name          TEXT               NOT NULL,
  pet_name            TEXT               NOT NULL,
  vaccine_name        TEXT               NOT NULL,
  date                DATE               NOT NULL,
  next_date           DATE,
  next_schedule_type  next_schedule_type,
  doctor              TEXT               DEFAULT '',
  supplemental        TEXT               DEFAULT '',
  lot1                TEXT               DEFAULT '',
  lot2                TEXT               DEFAULT '',
  lot3                TEXT               DEFAULT '',
  lot4                TEXT               DEFAULT '',
  remarks             TEXT               DEFAULT '',
  created_at          TIMESTAMPTZ        DEFAULT now(),
  updated_at          TIMESTAMPTZ        DEFAULT now()
);
```

#### `checkup_records` — 定期健診記録

```sql
CREATE TABLE checkup_records (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id UUID        NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id            UUID        NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  owner_name        TEXT        NOT NULL,
  pet_name          TEXT        NOT NULL,
  checkup_type      TEXT        NOT NULL,
  date              DATE        NOT NULL,
  next_date         DATE,
  doctor            TEXT        DEFAULT '',
  result            TEXT        DEFAULT '',
  created_at        TIMESTAMPTZ DEFAULT now(),
  updated_at        TIMESTAMPTZ DEFAULT now()
);
```

---

### 5. 入院テーブル

#### `hospitalizations` — 入院/ホテル記録

```sql
CREATE TABLE hospitalizations (
  id                    UUID                   PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id              UUID                   REFERENCES owners(id) ON DELETE SET NULL,
  owner_name            TEXT                   NOT NULL,
  pet_id                UUID                   REFERENCES pets(id) ON DELETE SET NULL,
  pet_name              TEXT                   NOT NULL,
  species               pet_species            NOT NULL,
  hospitalization_type  hospitalization_type   NOT NULL,
  start_date            DATE                   NOT NULL,
  end_date              DATE                   NOT NULL,
  status                hospitalization_status DEFAULT '予約',
  cage_id               UUID                   REFERENCES master_items(id) ON DELETE SET NULL,
  doctor_name           TEXT,
  memo                  TEXT                   DEFAULT '',
  owner_request         TEXT                   DEFAULT '',
  staff_notes           TEXT                   DEFAULT '',
  created_at            TIMESTAMPTZ            DEFAULT now(),
  updated_at            TIMESTAMPTZ            DEFAULT now()
);
```

#### `care_plan_items` — ケアプラン項目

```sql
CREATE TABLE care_plan_items (
  id                  UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID              NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  type                care_plan_type    NOT NULL,
  name                TEXT              NOT NULL,
  description         TEXT              DEFAULT '',
  timing              plan_timing[]     DEFAULT '{}',
  status              care_plan_status  DEFAULT 'active',
  notes               TEXT              DEFAULT '',
  master_id           UUID              REFERENCES master_items(id) ON DELETE SET NULL,
  unit_price          NUMERIC(10,2),
  category            TEXT,
  sort_order          INTEGER           DEFAULT 0,
  created_at          TIMESTAMPTZ       DEFAULT now(),
  updated_at          TIMESTAMPTZ       DEFAULT now()
);
```

#### `daily_records` — 日次記録コンテナ

```sql
CREATE TABLE daily_records (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID        NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  date                DATE        NOT NULL,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now(),
  UNIQUE (hospitalization_id, date)
);
```

#### `vital_records` — 入院バイタル

```sql
CREATE TABLE vital_records (
  id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id  UUID          NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time             TEXT          NOT NULL,
  temperature      NUMERIC(4,1),
  heart_rate       INTEGER,
  respiration_rate INTEGER,
  weight           NUMERIC(6,2),
  notes            TEXT          DEFAULT '',
  staff            TEXT          NOT NULL,
  created_at       TIMESTAMPTZ   DEFAULT now()
);
```

#### `care_log_records` — ケアログ

```sql
CREATE TABLE care_log_records (
  id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id  UUID             NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time             TEXT             NOT NULL,
  type             care_log_type    NOT NULL,
  status           care_log_status  NOT NULL,
  value            TEXT             DEFAULT '',
  staff            TEXT             NOT NULL,
  notes            TEXT             DEFAULT '',
  created_at       TIMESTAMPTZ      DEFAULT now()
);
```

#### `staff_note_records` — スタッフメモ

```sql
CREATE TABLE staff_note_records (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  daily_record_id  UUID        NOT NULL REFERENCES daily_records(id) ON DELETE CASCADE,
  time             TEXT        NOT NULL,
  content          TEXT        NOT NULL,
  staff            TEXT        NOT NULL,
  created_at       TIMESTAMPTZ DEFAULT now()
);
```

#### `treatment_plans` — 入院治療プラン

```sql
CREATE TABLE treatment_plans (
  id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  hospitalization_id  UUID          NOT NULL REFERENCES hospitalizations(id) ON DELETE CASCADE,
  treatment_content   TEXT          NOT NULL,
  memo                TEXT          DEFAULT '',
  insurance           BOOLEAN       DEFAULT false,
  unit_price          NUMERIC(10,2) DEFAULT 0,
  quantity            INTEGER       DEFAULT 1,
  discount_rate       NUMERIC(5,2)  DEFAULT 0,
  discount_amount     NUMERIC(10,2) DEFAULT 0,
  subtotal            NUMERIC(10,2) DEFAULT 0,
  sort_order          INTEGER       DEFAULT 0,
  created_at          TIMESTAMPTZ   DEFAULT now(),
  updated_at          TIMESTAMPTZ   DEFAULT now()
);
```

---

### 6. 予約・トリミング・会計テーブル

#### `reservation_appointments` — 予約

```sql
CREATE TABLE reservation_appointments (
  id             UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
  start_time     TIMESTAMPTZ        NOT NULL,
  end_time       TIMESTAMPTZ        NOT NULL,
  owner_name     TEXT               NOT NULL,
  pet_name       TEXT               NOT NULL,
  pet_id         UUID               REFERENCES pets(id) ON DELETE SET NULL,
  visit_type     visit_type         NOT NULL,
  type           VARCHAR            NOT NULL,  -- serviceType マスタ値（文字列）
  doctor         TEXT               NOT NULL,
  is_designated  BOOLEAN            DEFAULT false,
  status         reservation_status DEFAULT 'pending',
  notes          TEXT               DEFAULT '',
  created_at     TIMESTAMPTZ        DEFAULT now(),
  updated_at     TIMESTAMPTZ        DEFAULT now()
);
```

#### `trimming_records` — トリミング記録

```sql
CREATE TABLE trimming_records (
  id              UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  date            DATE             NOT NULL,
  pet_id          UUID             NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  pet_number      TEXT             NOT NULL,
  pet_name        TEXT             NOT NULL,
  owner_name      TEXT             NOT NULL,
  species         pet_species      NOT NULL,
  weight          TEXT             DEFAULT '',
  style_request   TEXT             DEFAULT '',
  staff           TEXT             NOT NULL,
  status          trimming_status  DEFAULT '予約',
  course_id       UUID             REFERENCES master_items(id) ON DELETE SET NULL,
  bw              TEXT             DEFAULT '',
  bw_unit         body_weight_unit DEFAULT 'Kg',
  bt              TEXT             DEFAULT '',
  used_shampoo    TEXT             DEFAULT '',
  used_ribbon     TEXT             DEFAULT '',
  treatment       TEXT             DEFAULT '',
  medicine        TEXT             DEFAULT '',
  charge          TEXT             DEFAULT '',
  remarks         TEXT             DEFAULT '',
  final_check     TEXT             DEFAULT '',
  style_image     TEXT,
  completed_image TEXT,
  created_at      TIMESTAMPTZ      DEFAULT now(),
  updated_at      TIMESTAMPTZ      DEFAULT now()
);
```

#### `trimming_record_options` — トリミングオプション中間テーブル

```sql
CREATE TABLE trimming_record_options (
  id                  UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  trimming_record_id  UUID    NOT NULL REFERENCES trimming_records(id) ON DELETE CASCADE,
  option_id           UUID    NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  sort_order          INTEGER DEFAULT 0,
  UNIQUE (trimming_record_id, option_id)
);
```

#### `accountings` — 会計レコード

```sql
CREATE TABLE accountings (
  id                 UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_id  UUID              UNIQUE REFERENCES medical_records(id) ON DELETE SET NULL,
  hospitalization_id UUID              REFERENCES hospitalizations(id) ON DELETE SET NULL,
  owner_id           UUID              NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
  owner_name         TEXT              NOT NULL,
  pet_id             UUID              NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
  pet_name           TEXT              NOT NULL,
  pet_species        pet_species,
  status             accounting_status DEFAULT 'waiting',
  scheduled_date     DATE              NOT NULL,
  completed_at       TIMESTAMPTZ,
  memo               TEXT              DEFAULT '',
  created_at         TIMESTAMPTZ       DEFAULT now(),
  updated_at         TIMESTAMPTZ       DEFAULT now()
);
```

#### `accounting_items` — 会計明細行

```sql
CREATE TABLE accounting_items (
  id                      UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  accounting_id           UUID          NOT NULL REFERENCES accountings(id) ON DELETE CASCADE,
  code                    TEXT,
  category                item_category NOT NULL,
  name                    TEXT          NOT NULL,
  unit_price              NUMERIC(10,2) NOT NULL,
  quantity                INTEGER       NOT NULL DEFAULT 1,
  tax_rate                NUMERIC(3,2)  NOT NULL DEFAULT 0.10,
  is_insurance_applicable BOOLEAN       DEFAULT false,
  source                  item_source   DEFAULT 'manual',
  sort_order              INTEGER       DEFAULT 0,
  created_at              TIMESTAMPTZ   DEFAULT now()
);
```

#### `payment_infos` — 支払情報（1:1）

```sql
CREATE TABLE payment_infos (
  id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
  accounting_id    UUID           NOT NULL UNIQUE REFERENCES accountings(id) ON DELETE CASCADE,
  subtotal         NUMERIC(10,2)  NOT NULL DEFAULT 0,
  tax_total        NUMERIC(10,2)  NOT NULL DEFAULT 0,
  total_amount     NUMERIC(10,2)  NOT NULL DEFAULT 0,
  insurance_name   TEXT,
  insurance_ratio  NUMERIC(3,2),
  insurance_amount NUMERIC(10,2)  DEFAULT 0,
  discount_amount  NUMERIC(10,2)  DEFAULT 0,
  billing_amount   NUMERIC(10,2)  NOT NULL DEFAULT 0,
  received_amount  NUMERIC(10,2)  DEFAULT 0,
  change_amount    NUMERIC(10,2)  DEFAULT 0,
  method           payment_method DEFAULT 'cash',
  created_at       TIMESTAMPTZ    DEFAULT now(),
  updated_at       TIMESTAMPTZ    DEFAULT now()
);
```

---

### 7. シフト管理テーブル

#### `shift_entries` — シフトエントリ

```sql
CREATE TABLE shift_entries (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id    UUID        NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  date        DATE        NOT NULL,
  shift_type  shift_type  NOT NULL,
  start_time  TEXT        DEFAULT '',
  end_time    TEXT        DEFAULT '',
  note        TEXT        DEFAULT '',
  created_at  TIMESTAMPTZ DEFAULT now(),
  updated_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE (staff_id, date)
);
```

---

### 8. マスタサブテーブル

#### `master_item_inspections` — 検査マスタ項目定義

```sql
CREATE TABLE master_item_inspections (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  master_item_id   UUID        NOT NULL REFERENCES master_items(id) ON DELETE CASCADE,
  name             TEXT        NOT NULL,
  inspection_value TEXT        DEFAULT '',
  normal_value     TEXT        DEFAULT '',
  sort_order       INTEGER     DEFAULT 0,
  created_at       TIMESTAMPTZ DEFAULT now()
);
```

---

## インデックス設計

```sql
-- pets
CREATE INDEX idx_pets_owner_id ON pets(owner_id);
CREATE INDEX idx_pets_species   ON pets(species);
CREATE INDEX idx_pets_name      ON pets(name);

-- medical_records
CREATE INDEX idx_medical_records_pet_id   ON medical_records(pet_id);
CREATE INDEX idx_medical_records_owner_id ON medical_records(owner_id);
CREATE INDEX idx_medical_records_date     ON medical_records(date DESC);
CREATE INDEX idx_medical_records_status   ON medical_records(status);

-- hospitalizations
CREATE INDEX idx_hospitalizations_pet_id     ON hospitalizations(pet_id);
CREATE INDEX idx_hospitalizations_status     ON hospitalizations(status);
CREATE INDEX idx_hospitalizations_start_date ON hospitalizations(start_date DESC);

-- reservation_appointments
CREATE INDEX idx_reservations_start_time ON reservation_appointments(start_time);
CREATE INDEX idx_reservations_status     ON reservation_appointments(status);
CREATE INDEX idx_reservations_pet_id     ON reservation_appointments(pet_id);

-- trimming_records
CREATE INDEX idx_trimming_records_pet_id ON trimming_records(pet_id);
CREATE INDEX idx_trimming_records_date   ON trimming_records(date DESC);
CREATE INDEX idx_trimming_records_status ON trimming_records(status);

-- accountings
CREATE INDEX idx_accountings_owner_id       ON accountings(owner_id);
CREATE INDEX idx_accountings_pet_id         ON accountings(pet_id);
CREATE INDEX idx_accountings_status         ON accountings(status);
CREATE INDEX idx_accountings_scheduled_date ON accountings(scheduled_date DESC);

-- カルテサブテーブル
CREATE INDEX idx_treatment_items_record_id      ON treatment_items(medical_record_id);
CREATE INDEX idx_vital_entries_record_id         ON vital_entries(medical_record_id);
CREATE INDEX idx_examination_records_record_id   ON examination_records(medical_record_id);
CREATE INDEX idx_examination_records_pet_id      ON examination_records(pet_id);
CREATE INDEX idx_vaccination_records_record_id   ON vaccination_records(medical_record_id);
CREATE INDEX idx_vaccination_records_pet_id      ON vaccination_records(pet_id);
CREATE INDEX idx_checkup_records_record_id       ON checkup_records(medical_record_id);
CREATE INDEX idx_checkup_records_pet_id          ON checkup_records(pet_id);

-- 入院サブテーブル
CREATE INDEX idx_care_plan_items_hosp_id    ON care_plan_items(hospitalization_id);
CREATE INDEX idx_vital_records_daily_id     ON vital_records(daily_record_id);
CREATE INDEX idx_care_log_records_daily_id  ON care_log_records(daily_record_id);
CREATE INDEX idx_staff_note_records_daily_id ON staff_note_records(daily_record_id);
CREATE INDEX idx_treatment_plans_hosp_id    ON treatment_plans(hospitalization_id);

-- accounting
CREATE INDEX idx_accounting_items_acct_id ON accounting_items(accounting_id);

-- master_items（複合インデックス重要）
CREATE INDEX idx_master_items_category      ON master_items(category);
CREATE INDEX idx_master_items_parent_id     ON master_items(parent_id);
CREATE INDEX idx_master_items_status        ON master_items(status);
CREATE INDEX idx_master_items_category_sort ON master_items(category, parent_id, sort_order);

-- inventory_items
CREATE INDEX idx_inventory_items_category ON inventory_items(category);
CREATE INDEX idx_inventory_items_status   ON inventory_items(status);

-- shift_entries
CREATE INDEX idx_shift_entries_date ON shift_entries(date);

-- 認証
CREATE INDEX idx_user_accounts_email       ON user_accounts(email);
CREATE INDEX idx_user_accounts_user_type   ON user_accounts(user_type);
CREATE INDEX idx_user_accounts_status      ON user_accounts(status);
CREATE INDEX idx_user_accounts_staff_master ON user_accounts(staff_master_id);
CREATE INDEX idx_user_clinic_memberships_clinic ON user_clinic_memberships(clinic_id);
CREATE INDEX idx_user_permissions_clinic   ON user_permissions(clinic_id);
```

---

## フロントエンド型マッピング

| PostgreSQL型 | Go型（GORM） | TypeScript型 |
|-------------|-------------|-------------|
| `UUID` | `uuid.UUID` | `string` |
| `TEXT` | `string` | `string` |
| `TEXT NOT NULL` | `string` | `string` |
| `TEXT` (nullable) | `*string` | `string \| undefined` |
| `NUMERIC(10,2)` | `*float64` | `number \| undefined` |
| `INTEGER` | `int` | `number` |
| `BOOLEAN` | `bool` | `boolean` |
| `DATE` | `time.Time` / `*time.Time` | `string` (ISO date) |
| `TIMESTAMPTZ` | `time.Time` | `string` (ISO datetime) |
| `TEXT[]` | `pq.StringArray` | `string[]` |
| `ENUM` | `string` | `"val1" \| "val2" \| ...` |

---

## RLSポリシー（将来実装）

マルチクリニック対応時に全データテーブルに `clinic_id UUID NOT NULL REFERENCES clinics(id)` を追加し、
以下のRLSポリシーを適用する計画（AUTH.md §6.3 参照）。

```sql
-- 例: pets テーブルへのRLS
ALTER TABLE pets ENABLE ROW LEVEL SECURITY;

CREATE POLICY pets_clinic_isolation ON pets
  USING (clinic_id = current_setting('app.current_clinic_id')::uuid);
```

**適用対象（27テーブル）**: owners, pets, medical_records, treatment_items, vital_entries,
examination_records, vaccination_records, checkup_records, hospitalizations, care_plan_items,
daily_records, vital_records, care_log_records, staff_note_records, treatment_plans,
reservation_appointments, trimming_records, accountings, accounting_items, payment_infos,
shift_entries, master_items, inventory_items 等

**グローバルマスタ**: `clinic_id = NULL` をグローバルマスタとして扱う設計も検討中

---

## マイグレーション

マイグレーションファイルは `backend/migrations/` に配置。
Docker起動時（`entrypoint.sh`）に自動実行される。

| ファイル | 説明 |
|---------|------|
| `001_init.sql` | 初期スキーマ（31テーブル、ENUM型、インデックス） |
| `002_seed_master.sql` | マスタデータ投入（92件） |

### シードデータ（002_seed_master.sql）

| カテゴリ | 件数 | 内容 |
|---------|------|------|
| `staff` | 4 | 獣医師2名、看護師1名、トリマー1名 |
| `cage` | 6 | 犬用S/M/L、猫用2種、ホテル個室 |
| `medicine` | 8 | 抗菌薬、ステロイド、駆虫薬等 |
| `insurance` | 3 | アニコム、アイペット、PS保険 |
| `trimming_course` | 5 | 小型犬〜猫、シャンプーのみ |
| `trimming_option` | 5 | 歯磨き、耳掃除、爪切り等 |
| `examination` | 10 | CBC、生化学、尿検査等 |
| その他 | 51+ | consultation, procedure, vaccine, diagnosis等 |
