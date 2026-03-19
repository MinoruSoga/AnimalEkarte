# ノア動物病院 電子カルテシステム ER図 (Entity Relationship Diagram)

バージョン: v22.0（migration との差分同期）
更新日: 2026-03-19
状態: Production Ready

本ドキュメントは、Animal Ekarteの全55テーブルとそのリレーションを定義します。
PostgreSQL 18 + Go/GORM（クリーンアーキテクチャ）で実装。

---

## 変更概要（v21.0 → v22.0）

| 変更内容 | 詳細 |
|---------|------|
| checkups・exams・daily_records に clinic_id FK 追加 | migration との差分同期。mermaid図・詳細テーブル・FK関係一覧を更新 |
| exam_type_items に unit・updated_at 追加 | migration との差分同期 |
| user_permissions に UNIQUE制約コメント追加 | (user_id, clinic_id, permission) UNIQUE 制約を明記 |
| vital_records に updated_at 追加、CHECK制約詳細を記載 | migration との差分同期 |
| care_log_records・staff_note_records・estimate_items に updated_at 追加 | migration との差分同期 |
| trimming_record_options に created_at・updated_at 追加 | migration との差分同期 |
| 金額型 numeric → integer に統一 | treatments, treatment_plans, estimates, estimate_items, billing_items, care_plan_items, マスタテーブルの price カラム |
| vaccinations・trimming_records・checkups・exams の pet_id FK: SET NULL → RESTRICT | migration との差分同期 |

## 変更概要（v20.0 → v21.0）

| 変更内容 | 詳細 |
|---------|------|
| v21.0 | Mermaidダイアグラムの全テーブルをGoモデル・SQLと完全一致させた（省略カラムを追記） |

## 変更概要（v19.0 → v20.0）

| 変更内容 | 詳細 |
|---------|------|
| v20.0 | medicines.drug_category 廃止 → parent_id (FK→medicines.id SET NULL) 追加。consultations / exam_types / procedures / vaccines / checkup_types に parent_id (自己参照 FK SET NULL) 追加。カテゴリ medicine レコードを parent_id=NULL・price=NULL で表現する親子構造に移行 |

## 変更概要（v18.0 → v19.0）

| 変更内容 | 詳細 |
|---------|------|
| v19.0 | estimates.clinic_id 追加（設計バグ修正）、billings.has_insurance追加、pg_trgm GINインデックス追加、論理削除考慮の部分インデックス追加、カルテ詳細lazy load方針を設計メモに追記 |

## 変更概要（v17.0 → v18.0）

| 変更内容 | 詳細 |
|---------|------|
| v18.0 | 型修正（time/numeric/integer/boolean）、billings 金額カラム追加、record_images updated_at 追加、payments スナップショット明記、非正規化 pet_id 明記、FK インデックス追加 |

## 変更概要（v16.0 → v17.0）

| 変更内容 | 詳細 |
|---------|------|
| v17.0 | 論理削除（deleted_at）全主要テーブルに追加。C-1〜C-6 Critical 修正、billings-payments 関係修正、record_no/estimate_no UNIQUE スコープ修正、medical_records に reservation_appointment_id FK 追加、treatment_plans CHECK 排他的 OR 強化 |

## 変更概要（v15.0 → v16.0）

| 変更内容 | 詳細 |
|---------|------|
| job_title ENUM 廃止 → job_titles マスタテーブル追加。staffs.job_title_id FK 化 | テーブル総数 54 → 55 |

## 変更概要（v14.0 → v15.0）

| 変更内容 | 詳細 |
|---------|------|
| スナップショット列を全廃 | owner_name / pet_name / pet_number / vaccine_name_snapshot / insurance_name / insurance_details をすべて削除。FK で JOIN して取得できるため不要 |
| テーブル総数 | 54（変更なし） |

## 変更概要（v13.0 → v14.0）

| 変更内容 | 詳細 |
|---------|------|
| `species` スナップショット列を削除 | `medical_records.species`, `hospitalizations.species`, `trimming_records.species`, `billings.pet_species` を削除。`pet_id` FK から JOIN で取得可能なため不要 |
| テーブル総数 | 54（変更なし） |

## 変更概要（v12.0 → v13.0）

| 変更内容 | 詳細 |
|---------|------|
| `animal_species` テーブル追加 | ペット種類マスタ。`pet_species` ENUM を廃止しマスタテーブル化 |
| `pets.species` FK化 | `pet_species ENUM` → `animal_species_id uuid FK → animal_species(id) RESTRICT` |
| `pet_species` ENUM 削除 | マスタ化により不要 |
| テーブル総数 | 53 → 54 |

## 変更概要（v11.0 → v12.0）

| 変更内容 | 詳細 |
|---------|------|
| `treatment_plans` 外来・入院共用化 | `medical_record_id` FK を追加。`hospitalization_id` を nullable に変更。CHECK制約で「どちらか一方必須」を保証 |
| テーブル総数 | 53（変更なし） |

## 変更概要（v10.0 → v11.0）

| 変更内容 | 詳細 |
|---------|------|
| `inquiry_templates` テーブル追加 | 問診タブ定型文マスタ。clinic_id + category（フィールド区分）+ title + content |
| `chief_complaint_categories` テーブル追加 | 問診タブ主訴区分マスタ。`inquiries.chief_complaint_category_id` FK を追加 |
| `diagnosis_names` → 診断病名マスタに改称 | 「診断1マスタ」の呼称を「診断病名マスタ」に統一 |
| `clinical_plans` 診断カラム整理 | `diagnosis1_*` → `diagnosis_*` にリネーム、`diagnosis2_*` を削除 |
| テーブル総数 | 51 → 53 |

## 変更概要（v9.0 → v10.0）

| 変更内容 | 詳細 |
|---------|------|
| マルチテナント対応 | 24テーブルに `clinic_id FK → clinics(id)` を追加 |
| カルテ直接所有エンティティ | medical_records, owners, pets, reservation_appointments, hospitalizations, trimming_records, shift_entries, billings |
| マスタデータ | staffs, inventory_items, cages, service_types, consultations, procedures, medicines, hospitalization_plans, trimming_courses, trimming_options, exam_types, vaccines, insurances, diagnosis_categories, diagnosis_names, checkup_types |
| 設計方針 | medical_records/hospitalizations の子テーブルは親経由でclinicを特定するため除外 |

## 変更概要（v8.0 → v9.0）

| 変更内容 | 詳細 |
|---------|------|
| `clinical_plans` テーブル追加 | 診察/治療タブ専用。`medical_records` から診察・診断フィールドを移動。1:1 |
| `medical_records` スリム化 | `physical_exam`, `treatment_policy`, `diagnosis_details`, `diagnosis1/2` FK を `clinical_plans` に移動 |
| テーブル総数 | 50 → 51 |

## 変更概要（v5.0 → v6.0）

| 変更内容 | 詳細 |
|---------|------|
| 法人テーブル追加 | `company`（ノア動物病院法人情報、シングルトン） |
| 医院テーブル追加 | `clinics`（八王子医院・城東医院・敷島医院等） |
| clinic_id追加予定 | `003_add_clinic_id.sql` にて以下テーブルに追加予定: owners, staffs, inventory_items, cages, service_types, consultations, procedures, hospitalization_plans, trimming_courses, trimming_options, exam_types, vaccines, medicines, insurances, diagnosis_categories, checkup_types |
| clinic_id保持済 | `user_clinic_memberships`, `user_permissions` のみ現時点で保持 |

## 変更概要（v7.0 → v8.0）

| 変更内容 | 詳細 |
|---------|------|
| 命名規則統一 | `_records`/`_entries`/`_items` サフィックスを排除し、短縮形に統一 |
| `examination_records` → `exams` | 検査記録 |
| `examination_record_items` → `exam_items` | 検査結果項目 |
| `examination_types` → `exam_types` | 検査種別マスタ |
| `examination_type_items` → `exam_type_items` | 検査項目定義マスタ |
| `treatment_items` → `treatments` | 治療項目 |
| `vital_entries` / `vitals` → `vital_records` | バイタル記録（外来・入院統合） |
| `vaccination_records` → `vaccinations` | 予防接種記録 |
| `checkup_records` → `checkups` | 健診記録 |
| `accountings` → `billings` | 会計 |
| `accounting_items` → `billing_items` | 会計明細 |
| `payment_infos` → `payments` | 支払情報 |
| `medical_inquiries` → `inquiries` | 問診（不要なプレフィックス削除） |
| `medical_images` → `record_images` | 診療画像 |
| `billing_confirmations` → `billing_reviews` | 会計医師確認 |
| テーブル総数 | 50テーブル（変更なし） |

---

## 設計方針

### 法人・医院の区別

| テーブル | 役割 | レコード数 |
|---------|------|-----------|
| `company` | 法人情報（ノア動物病院）。FK参照なし。設定画面で参照のみ | 1件固定 |
| `clinics` | 各医院（八王子医院等）。ユーザー・権限管理で参照 | 複数（増加あり） |

### clinic_id 戦略

v10.0 にて24テーブルへの `clinic_id` 追加完了（003_add_clinic_id.sql 実施済み）。
マスタ情報（staffs, cages, medicines等）および診療・業務データを医院ごとに独立管理できる。

**追加方針:**
- 「医院が直接所有するデータ」= clinic_id 追加
- medical_records の子テーブル（inquiries, clinical_plans, treatments 等）は medical_records 経由でクリニックが特定できるため追加しない
- hospitalizations の子テーブルも同様に追加しない

### カルテタブ ↔ テーブル対応表

| タブ | テーブル | 状態 |
|------|---------|------|
| 問診 | `inquiries` | v7.0追加 |
| 診察/治療 | `clinical_plans` | v9.0追加 |
| 治療 | `treatments` | 既存 |
| 予防接種 | `vaccinations` | 既存 |
| 定期健診 | `checkups` | 既存 |
| 検査 | `exams` + `exam_items` | 既存 |
| 画像 | `record_images` | v7.0追加 |
| 見積書 | `estimates` + `estimate_items` | v7.0追加 |
| 会計（医師確認） | `billing_reviews` | v7.0追加 |
| 会計情報 | `billings` + `billing_items` + `payments` | 既存 |
| 生体情報 | `vital_records` | 外来・入院統合 |
| ペット情報 | `pets`（参照） | 既存 |

---

## テーブル一覧（55テーブル）

| # | テーブル名 | 区分 | 説明 |
|---|-----------|------|------|
| 1 | `company` | 法人情報 | 法人（ノア動物病院）情報シングルトン |
| 2 | `clinics` | 医院情報 | 各医院（八王子・城東・敷島） |
| 3 | `user_accounts` | 認証 | ログインアカウント |
| 4 | `user_clinic_memberships` | 認証 | ユーザーの医院所属 |
| 5 | `user_permissions` | 認証 | ユーザーの権限 |
| 6 | `owners` | コア | 飼い主 |
| 7 | `pets` | コア | ペット |
| 8 | `animal_species` | マスタ | ペット種類マスタ（犬・猫・鳥等） |
| 9 | `staffs` | マスタ | スタッフ（獣医師・看護師等） |
| 10 | `inventory_items` | 在庫 | 在庫アイテム |
| 11 | `exam_types` | マスタ | 検査種別 |
| 12 | `exam_type_items` | マスタ | 検査種別の検査項目定義 |
| 13 | `vaccines` | マスタ | ワクチン |
| 14 | `medicines` | マスタ | 薬剤 |
| 15 | `insurances` | マスタ | 保険 |
| 16 | `cages` | マスタ | ケージ |
| 17 | `service_types` | マスタ | サービス種別 |
| 18 | `consultations` | マスタ | 診察項目 |
| 19 | `procedures` | マスタ | 処置項目 |
| 20 | `hospitalization_plans` | マスタ | 入院プラン |
| 21 | `trimming_courses` | マスタ | トリミングコース |
| 22 | `trimming_options` | マスタ | トリミングオプション |
| 23 | `diagnosis_categories` | マスタ | 診断カテゴリ |
| 24 | `diagnosis_names` | マスタ | 診断病名 |
| 25 | `checkup_types` | マスタ | 健診種別 |
| 26 | `chief_complaint_categories` | マスタ | 主訴区分マスタ |
| 27 | `inquiry_templates` | マスタ | 問診定型文マスタ |
| 28 | `job_titles` | マスタ | 職種マスタ（clinic単位） |
| 29 | `medical_records` | 診療 | カルテ（診療記録） |
| 30 | `clinical_plans` | 診療 | 診察所見・診断・治療方針（診察/治療タブ） |
| 31 | `inquiries` | 診療 | 問診情報（カルテ問診タブ） |
| 32 | `treatments` | 診療 | 処置・診察・薬剤明細 |
| 33 | `vital_records` | 診療・入院 | バイタル記録（外来・入院統合） |
| 34 | `exams` | 診療 | 検査記録 |
| 35 | `exam_items` | 診療 | 検査記録の検査結果項目 |
| 36 | `vaccinations` | 診療 | ワクチン接種記録 |
| 37 | `checkups` | 診療 | 健診記録 |
| 38 | `record_images` | 診療 | 診療画像（レントゲン・エコー等） |
| 39 | `estimates` | 診療 | 見積書 |
| 40 | `estimate_items` | 診療 | 見積書明細 |
| 41 | `billing_reviews` | 診療 | 会計医師確認 |
| 42 | `reservation_appointments` | 予約 | 予約 |
| 43 | `hospitalizations` | 入院 | 入院・ホテル |
| 44 | `daily_records` | 入院 | 入院日次記録 |
| 45 | `care_plan_items` | 入院 | ケアプラン項目 |
| 46 | `care_log_records` | 入院 | ケアログ |
| 47 | （`vitals` と統合 → `vital_records`） | - | - |
| 48 | `staff_note_records` | 入院 | スタッフノート |
| 49 | `treatment_plans` | 診療 | 治療プラン（外来・入院共用） |
| 50 | `trimming_records` | トリミング | トリミング記録 |
| 51 | `trimming_record_options` | トリミング | トリミング記録のオプション選択 |
| 52 | `billings` | 会計 | 会計 |
| 53 | `billing_items` | 会計 | 会計明細 |
| 54 | `payments` | 会計 | 支払い情報 |
| 55 | `shift_entries` | シフト | スタッフシフト |

---

## システム全体 ER図

```mermaid
erDiagram
    %% ===== 法人・医院 =====
    company {
        integer id PK
        text name
        text postal_code
        text address
        text phone_number
        text fax_number
        text registration_number
        text director_name
        text email
        text website
        text logo_url
        timestamptz created_at
        timestamptz updated_at
    }

    clinics {
        uuid id PK
        bigint company_id FK
        text name
        boolean is_active
        text postal_code
        text address
        text phone_number
        text fax_number
        text registration_number
        text director_name
        text email
        text website
        text logo_url
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 認証 =====
    user_accounts {
        uuid id PK
        text email
        text display_name
        user_type user_type
        uuid job_title_id FK
        account_status status
        uuid staff_id FK
        text display_name_kana
        text avatar_url
        text password_hash
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    user_clinic_memberships {
        uuid id PK
        uuid user_id FK
        uuid clinic_id FK
        boolean is_main
        timestamptz joined_at
    }

    user_permissions {
        uuid id PK
        uuid user_id FK
        uuid clinic_id FK
        permission_type permission "UK: user_id+clinic_id+permission"
        uuid granted_by FK
        timestamptz granted_at
    }

    %% ===== コア =====
    owners {
        uuid id PK
        uuid clinic_id FK
        text owner_name
        text phone
        text email
        date birth_date
        membership_type membership_type
        text owner_name_kana
        text company
        text postal_code
        text address1
        text address2
        text home_postal_code
        text home_address1
        text home_address2
        text company_phone
        boolean is_dangerous
        numeric discount_rate
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    pets {
        uuid id PK
        uuid clinic_id FK
        uuid owner_id FK
        uuid animal_species_id FK
        text name
        pet_gender gender
        pet_status status
        uuid insurance_id FK
        text pet_number
        text pet_name_kana
        date birth_date
        text breed
        text color
        numeric weight
        date neutered_date
        acquisition_type acquisition_type
        danger_level danger_level
        text food
        text environment
        text phone
        date last_visit
        text remarks
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    %% ===== マスタ =====
    animal_species {
        uuid id PK
        text name
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    staffs {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        staff_role staff_role
        uuid job_title_id FK
        text license_number
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    job_titles {
        uuid id PK
        uuid clinic_id FK
        text name
        text description
        int sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }

    inventory_items {
        uuid id PK
        uuid clinic_id FK
        text name
        inventory_category category
        integer quantity
        inventory_status status
        text unit
        integer min_stock_level
        text location
        date expiry_date
        text supplier
        date last_restocked
        timestamptz created_at
        timestamptz updated_at
    }

    exam_types {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        integer price
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    exam_type_items {
        uuid id PK
        uuid exam_type_id FK
        text name
        integer sort_order
        text inspection_value
        text normal_value
        text unit
        timestamptz created_at
        timestamptz updated_at
    }

    vaccines {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        vaccine_species species
        uuid inventory_id FK
        integer price
        text description
        text interval
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    medicines {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        dosage_form dosage_form
        uuid inventory_id FK
        integer price
        text description
        medicine_unit medicine_unit
        integer default_quantity
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    insurances {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        integer coverage_rate
        text description
        text contact_phone
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    cages {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        cage_type cage_type
        cage_size cage_size
        integer price
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    service_types {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        text color
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    consultations {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        integer price
        text description
        text time_condition
        integer duration
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    procedures {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        anesthesia_type anesthesia
        integer price
        text description
        integer duration
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    hospitalization_plans {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        body_size body_size
        billing_unit billing_unit
        integer price
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    trimming_courses {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        target_size target_size
        integer price
        text description
        integer duration
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    trimming_options {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        boolean combinable
        integer price
        text description
        integer duration
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    diagnosis_categories {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    diagnosis_names {
        uuid id PK
        uuid clinic_id FK
        text name
        boolean is_active
        text description
        uuid diagnosis_category_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    checkup_types {
        uuid id PK
        uuid clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        text interval
        integer price
        text description
        text target_age
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    chief_complaint_categories {
        uuid id PK
        uuid clinic_id FK
        text name
        text description
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    inquiry_templates {
        uuid id PK
        uuid clinic_id FK
        text category
        text title
        text content
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 診療 =====
    medical_records {
        uuid id PK
        uuid clinic_id FK
        text record_no
        date date
        uuid owner_id FK
        uuid pet_id FK
        uuid doctor_id FK
        uuid reservation_appointment_id FK
        medical_record_status status
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    clinical_plans {
        uuid id PK
        uuid medical_record_id FK
        text physical_exam
        uuid diagnosis_category_id FK
        uuid diagnosis_name_id FK
        text diagnosis_details
        text treatment_policy
        timestamptz created_at
        timestamptz updated_at
    }

    treatments {
        uuid id PK
        uuid medical_record_id FK
        treatment_item_type item_type
        uuid consultation_id FK
        uuid procedure_id FK
        uuid medicine_id FK
        uuid inventory_id FK
        integer unit_price
        integer quantity
        boolean selected
        treatment_status status
        text content
        text memo
        boolean insurance
        numeric discount_rate
        integer discount_amount
        integer sort_order
        timestamptz deleted_at
    }

    vital_records {
        uuid id PK
        uuid pet_id FK
        uuid medical_record_id FK "外来時"
        uuid daily_record_id FK "入院時"
        timestamptz recorded_at
        uuid staff_id FK
        numeric temperature
        integer heart_rate
        integer respiration_rate
        numeric weight
        body_weight_unit weight_unit
        text notes
        timestamptz created_at
        timestamptz updated_at
    }

    exams {
        uuid id PK
        uuid medical_record_id FK
        uuid clinic_id FK
        uuid pet_id FK
        uuid exam_type_id FK
        uuid doctor_id FK
        examination_status status
        date date
        text result_summary
        text machine
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    exam_items {
        uuid id PK
        uuid exam_id FK
        uuid exam_type_item_id FK
        text name
        text inspection_value
        examination_result_status status
        text normal_value
        text result
        text unit
        text ref
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    vaccinations {
        uuid id PK
        uuid medical_record_id FK
        uuid pet_id FK
        uuid vaccine_id FK
        date date
        uuid doctor_id FK
        uuid clinic_id FK
        date next_date
        text next_schedule_type
        text supplemental
        text lot1
        text lot2
        text lot3
        text lot4
        text remarks
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    checkups {
        uuid id PK
        uuid medical_record_id FK
        uuid clinic_id FK
        uuid pet_id FK
        uuid checkup_type_id FK
        date date
        uuid doctor_id FK
        text result
        date next_date
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    inquiries {
        uuid id PK
        uuid medical_record_id FK
        uuid chief_complaint_category_id FK
        text chief_complaint
        text history
        text current_medications
        text allergy_info
        text last_meal
        text last_defecation
        text last_urination
        appetite_level appetite
        water_intake_level water_intake
        text owner_observations
        text notes
        uuid staff_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    record_images {
        uuid id PK
        uuid medical_record_id FK
        text image_url
        text thumbnail_url
        text file_name
        bigint file_size
        text mime_type
        medical_image_type image_type
        text description
        timestamptz taken_at
        uuid exam_id FK
        uuid staff_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    estimates {
        uuid id PK
        uuid clinic_id FK
        text estimate_no
        uuid medical_record_id FK
        text title
        uuid owner_id FK
        estimate_status status
        integer subtotal
        integer tax_total
        integer total_amount
        integer insurance_amount
        integer discount_amount
        date valid_until
        text comment
        text notes
        uuid created_by FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    estimate_items {
        uuid id PK
        uuid estimate_id FK
        text name
        item_category category
        integer unit_price
        integer quantity
        numeric tax_rate
        numeric discount_rate
        integer discount_amount
        boolean is_insurance_applicable
        uuid consultation_id FK
        uuid procedure_id FK
        uuid medicine_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    billing_reviews {
        uuid id PK
        uuid medical_record_id FK
        billing_review_status status
        uuid confirmed_by FK
        timestamptz confirmed_at
        uuid returned_by FK
        timestamptz returned_at
        text return_reason
        text memo
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 予約 =====
    reservation_appointments {
        uuid id PK
        uuid clinic_id FK
        timestamptz start_time
        timestamptz end_time
        uuid pet_id FK
        uuid service_type_id FK
        uuid doctor_id FK
        reservation_status status
        uuid owner_id FK
        visit_type visit_type
        boolean is_designated
        text notes
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    %% ===== 入院 =====
    hospitalizations {
        uuid id PK
        uuid clinic_id FK
        uuid owner_id FK
        uuid pet_id FK
        hospitalization_type hospitalization_type
        date start_date
        date end_date
        uuid cage_id FK
        uuid doctor_id FK
        hospitalization_status status
        text memo
        text owner_request
        text staff_notes
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    daily_records {
        uuid id PK
        uuid hospitalization_id FK
        uuid clinic_id FK
        date date
        timestamptz created_at
        timestamptz updated_at
    }

    care_plan_items {
        uuid id PK
        uuid hospitalization_id FK
        care_plan_type type
        plan_timing timing
        uuid medicine_id FK
        uuid procedure_id FK
        uuid hospitalization_plan_id FK
        care_plan_status status
        text name
        text description
        text notes
        integer unit_price
        item_category category
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    care_log_records {
        uuid id PK
        uuid daily_record_id FK
        care_log_type type
        uuid staff_id FK
        time time
        care_log_status status
        text value
        text notes
        timestamptz created_at
        timestamptz updated_at
    }

    staff_note_records {
        uuid id PK
        uuid daily_record_id FK
        time time
        uuid staff_id FK
        text content
        timestamptz created_at
        timestamptz updated_at
    }

    treatment_plans {
        uuid id PK
        uuid medical_record_id FK
        uuid hospitalization_id FK
        text treatment_content
        integer unit_price
        integer quantity
        text memo
        boolean insurance
        numeric discount_rate
        integer discount_amount
        integer subtotal
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    %% ===== トリミング =====
    trimming_records {
        uuid id PK
        uuid clinic_id FK
        date date
        uuid pet_id FK
        uuid staff_id FK
        uuid course_id FK
        trimming_status status
        text style_request
        numeric bw "体重"
        body_weight_unit bw_unit
        numeric bt "体温(℃)"
        text used_shampoo
        text used_ribbon
        text remarks
        text style_image
        text completed_image
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    trimming_record_options {
        uuid id PK
        uuid trimming_record_id FK
        uuid option_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 会計 =====
    billings {
        uuid id PK
        uuid clinic_id FK
        uuid medical_record_id FK
        uuid hospitalization_id FK
        uuid owner_id FK
        uuid pet_id FK
        integer subtotal
        integer tax_total
        integer total_amount
        boolean has_insurance
        billing_status status
        date scheduled_date
        timestamptz completed_at
        text memo
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    billing_items {
        uuid id PK
        uuid billing_id FK
        item_category category
        text name
        integer unit_price
        integer quantity
        numeric tax_rate
        boolean is_insurance_applicable
        item_source source
        integer sort_order
        timestamptz created_at
        timestamptz deleted_at
    }

    payments {
        uuid id PK
        uuid billing_id FK
        integer total_amount
        integer billing_amount
        payment_method method
        integer subtotal
        integer tax_total
        text insurance_name
        numeric insurance_ratio "小数(0.7等)"
        integer insurance_amount
        integer discount_amount
        integer received_amount
        integer change_amount
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== シフト =====
    shift_entries {
        uuid id PK
        uuid clinic_id FK
        uuid staff_id FK
        date date
        shift_type shift_type
        time start_time
        time end_time
        text note
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== リレーション =====

    %% 認証
    user_accounts ||--o{ user_clinic_memberships : "user_id"
    user_accounts ||--o{ user_permissions : "user_id"
    user_accounts }o--|| staffs : "staff_id"
    user_permissions }o--|| user_accounts : "granted_by"
    clinics ||--o{ user_clinic_memberships : "clinic_id"
    clinics ||--o{ user_permissions : "clinic_id"

    %% コア
    owners ||--o{ pets : "owner_id"
    insurances ||--o{ pets : "insurance_id"
    animal_species ||--o{ pets : "animal_species_id"

    %% マスタ
    clinics ||--o{ job_titles : "clinic_id"
    job_titles ||--o{ staffs : "job_title_id"
    job_titles ||--o{ user_accounts : "job_title_id"
    exam_types ||--o{ exam_type_items : "exam_type_id"
    diagnosis_categories ||--o{ diagnosis_names : "diagnosis_category_id"
    inventory_items ||--o{ medicines : "inventory_id"
    inventory_items ||--o{ vaccines : "inventory_id"
    clinics ||--o{ chief_complaint_categories : "clinic_id"
    clinics ||--o{ inquiry_templates : "clinic_id"

    %% 診療
    owners ||--o{ medical_records : "owner_id"
    pets ||--o{ medical_records : "pet_id"
    staffs ||--o{ medical_records : "doctor_id"
    medical_records ||--o| clinical_plans : "medical_record_id"
    clinical_plans }o--|| diagnosis_categories : "diagnosis_category_id"
    clinical_plans }o--|| diagnosis_names : "diagnosis_name_id"

    medical_records ||--o{ treatments : "medical_record_id"
    consultations ||--o{ treatments : "consultation_id"
    procedures ||--o{ treatments : "procedure_id"
    medicines ||--o{ treatments : "medicine_id"
    inventory_items ||--o{ treatments : "inventory_id"

    pets ||--o{ vital_records : "pet_id"
    medical_records ||--o{ vital_records : "medical_record_id"
    staffs ||--o{ vital_records : "staff_id"

    medical_records ||--o{ exams : "medical_record_id"
    pets ||--o{ exams : "pet_id"
    exam_types ||--o{ exams : "exam_type_id"
    staffs ||--o{ exams : "doctor_id"
    exams ||--o{ exam_items : "exam_id"
    exam_type_items ||--o{ exam_items : "exam_type_item_id"

    medical_records ||--o{ vaccinations : "medical_record_id"
    pets ||--o{ vaccinations : "pet_id"
    vaccines ||--o{ vaccinations : "vaccine_id"
    staffs ||--o{ vaccinations : "doctor_id"

    medical_records ||--o{ checkups : "medical_record_id"
    pets ||--o{ checkups : "pet_id"
    checkup_types ||--o{ checkups : "checkup_type_id"
    staffs ||--o{ checkups : "doctor_id"

    medical_records ||--o| inquiries : "medical_record_id"
    chief_complaint_categories ||--o{ inquiries : "chief_complaint_category_id"
    staffs ||--o{ inquiries : "staff_id"
    medical_records ||--o{ record_images : "medical_record_id"
    exams ||--o{ record_images : "exam_id"
    staffs ||--o{ record_images : "staff_id"
    medical_records ||--o{ estimates : "medical_record_id"
    owners ||--o{ estimates : "owner_id"
    staffs ||--o{ estimates : "created_by"
    estimates ||--o{ estimate_items : "estimate_id"
    consultations ||--o{ estimate_items : "consultation_id"
    procedures ||--o{ estimate_items : "procedure_id"
    medicines ||--o{ estimate_items : "medicine_id"
    medical_records ||--o| billing_reviews : "medical_record_id"
    staffs ||--o{ billing_reviews : "confirmed_by"
    staffs ||--o{ billing_reviews : "returned_by"

    %% 予約
    reservation_appointments ||--o{ medical_records : "reservation_appointment_id"
    pets ||--o{ reservation_appointments : "pet_id"
    service_types ||--o{ reservation_appointments : "service_type_id"
    staffs ||--o{ reservation_appointments : "doctor_id"
    owners ||--o{ reservation_appointments : "owner_id"

    %% 入院
    owners ||--o{ hospitalizations : "owner_id"
    pets ||--o{ hospitalizations : "pet_id"
    cages ||--o{ hospitalizations : "cage_id"
    staffs ||--o{ hospitalizations : "doctor_id"

    hospitalizations ||--o{ daily_records : "hospitalization_id"
    hospitalizations ||--o{ care_plan_items : "hospitalization_id"
    hospitalizations ||--o{ treatment_plans : "hospitalization_id"
    medical_records ||--o{ treatment_plans : "medical_record_id"

    daily_records ||--o{ care_log_records : "daily_record_id"
    daily_records ||--o{ vital_records : "daily_record_id"
    daily_records ||--o{ staff_note_records : "daily_record_id"

    staffs ||--o{ care_log_records : "staff_id"
    staffs ||--o{ staff_note_records : "staff_id"

    medicines ||--o{ care_plan_items : "medicine_id"
    procedures ||--o{ care_plan_items : "procedure_id"
    hospitalization_plans ||--o{ care_plan_items : "hospitalization_plan_id"

    %% トリミング
    pets ||--o{ trimming_records : "pet_id"
    staffs ||--o{ trimming_records : "staff_id"
    trimming_courses ||--o{ trimming_records : "course_id"
    trimming_records ||--o{ trimming_record_options : "trimming_record_id"
    trimming_options ||--o{ trimming_record_options : "option_id"

    %% 会計
    medical_records ||--o| billings : "medical_record_id"
    hospitalizations ||--o{ billings : "hospitalization_id"
    owners ||--o{ billings : "owner_id"
    pets ||--o{ billings : "pet_id"
    billings ||--o{ billing_items : "billing_id"
    billings ||--o| payments : "billing_id"

    %% シフト
    staffs ||--o{ shift_entries : "staff_id"

    %% clinic_id マルチテナント（v10.0 + v22.0追加分）
    clinics ||--o{ checkups : "clinic_id"
    clinics ||--o{ exams : "clinic_id"
    clinics ||--o{ daily_records : "clinic_id"
    clinics ||--o{ medical_records : "clinic_id"
    clinics ||--o{ owners : "clinic_id"
    clinics ||--o{ pets : "clinic_id"
    clinics ||--o{ reservation_appointments : "clinic_id"
    clinics ||--o{ hospitalizations : "clinic_id"
    clinics ||--o{ trimming_records : "clinic_id"
    clinics ||--o{ shift_entries : "clinic_id"
    clinics ||--o{ billings : "clinic_id"
    clinics ||--o{ estimates : "clinic_id"
    clinics ||--o{ staffs : "clinic_id"
    clinics ||--o{ inventory_items : "clinic_id"
    clinics ||--o{ cages : "clinic_id"
    clinics ||--o{ service_types : "clinic_id"
    clinics ||--o{ consultations : "clinic_id"
    clinics ||--o{ procedures : "clinic_id"
    clinics ||--o{ medicines : "clinic_id"
    clinics ||--o{ hospitalization_plans : "clinic_id"
    clinics ||--o{ trimming_courses : "clinic_id"
    clinics ||--o{ trimming_options : "clinic_id"
    clinics ||--o{ exam_types : "clinic_id"
    clinics ||--o{ vaccines : "clinic_id"
    clinics ||--o{ insurances : "clinic_id"
    clinics ||--o{ diagnosis_categories : "clinic_id"
    clinics ||--o{ diagnosis_names : "clinic_id"
    clinics ||--o{ checkup_types : "clinic_id"
    clinics ||--o{ chief_complaint_categories : "clinic_id"
    clinics ||--o{ inquiry_templates : "clinic_id"
```

---

## ENUM型定義

| ENUM名 | 値 |
| ------- | ---- |
| `account_status` | active, inactive, locked |
| `appetite_level` | normal, increased, decreased, none |
| `billing_review_status` | pending, confirmed, returned |
| `billing_status` | waiting, completed, cancelled, pending |
| `acquisition_type` | purchased, transferred, rescued, other |
| `anesthesia_type` | none, local, sedation, general |
| `billing_unit` | per_day, per_night |
| `body_size` | small, medium, large |
| `body_weight_unit` | Kg, g |
| `cage_size` | small, medium, large |
| `cage_type` | icu, dog, cat, general |
| `care_log_status` | completed, partial, skipped |
| `care_log_type` | food, excretion, medicine, treatment, other |
| `care_plan_status` | active, completed, discontinued |
| `care_plan_type` | food, medicine, treatment, instruction, item |
| `danger_level` | low, medium, high |
| `dosage_form` | tablet, liquid, injection, topical, powder |
| `estimate_status` | draft, sent, approved, rejected |
| `examination_result_status` | normal, high, low |
| `examination_status` | pending, in_progress, completed |
| `hospitalization_status` | admitted, discharged, reserved |
| `hospitalization_type` | hospitalization, hotel |
| `inventory_category` | medicine, consumable, food, other |
| `inventory_status` | sufficient, low, out_of_stock |
| `item_category` | examination, test, procedure, surgery, medicine, food, goods, other |
| `item_source` | medical_record, manual, hospitalization |
| `medical_image_type` | xray, echo, photo, endoscope, ct, mri, microscope, other |
| `medical_record_status` | draft, finalized |
| `medicine_unit` | per_tablet, per_ml, per_dose, per_gram |
| `membership_type` | non_member, member, deceased, transferred |
| `next_schedule_type` | 3weeks, 4weeks, 1year, other |
| `payment_method` | cash, credit_card, electronic_money |
| `permission_type` | account_admin, medical, medical_read, trimming, billing, reception, hospitalization, master_admin, shift_admin, inventory |
| `pet_gender` | male, female, unknown |
| `pet_status` | alive, deceased |
| `plan_timing` | morning, noon, night |
| `reservation_status` | confirmed, pending, cancelled, checked_in, in_consultation, accounting, completed |
| `shift_type` | full, morning, afternoon, off, paid_leave |
| `staff_role` | veterinarian, nurse, trimmer, reception, manager |
| `target_size` | small, medium, large, cat |
| `treatment_item_type` | consultation, procedure, medicine, other |
| `treatment_status` | pending, completed, not_applicable |
| `trimming_status` | completed, reserved, in_progress |
| `user_type` | system_admin, clinic_admin, staff |
| `vaccine_species` | dog, cat, both |
| `visit_type` | first, revisit |
| `water_intake_level` | normal, increased, decreased, none |

---

## テーブル詳細

### 法人・医院

---

#### `company`

用途: ノア動物病院の法人情報。システム全体で1件のみ存在するシングルトン。FK参照なし。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| name | text | NO | | 法人名 |
| postal_code | text | YES | '' | 郵便番号 |
| address | text | YES | '' | 住所 |
| phone_number | text | YES | '' | 電話番号 |
| fax_number | text | YES | '' | FAX番号 |
| registration_number | text | YES | '' | 登録番号 |
| director_name | text | YES | '' | 院長名 |
| email | text | YES | '' | メールアドレス |
| website | text | YES | '' | ウェブサイトURL |
| logo_url | text | YES | '' | ロゴ画像URL |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

---

#### `clinics`

用途: 各医院（八王子医院・城東医院・敷島医院等）の情報。ユーザーの所属・権限管理で参照される。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| company_id | bigint | NO | | company.id FK |
| name | text | NO | | 医院名 |
| postal_code | text | YES | '' | 郵便番号 |
| address | text | YES | '' | 住所 |
| phone_number | text | YES | '' | 電話番号 |
| fax_number | text | YES | '' | FAX番号 |
| registration_number | text | YES | '' | 登録番号 |
| director_name | text | YES | '' | 院長名 |
| email | text | YES | '' | メールアドレス |
| website | text | YES | '' | ウェブサイトURL |
| logo_url | text | YES | '' | ロゴ画像URL |
| is_active | boolean | YES | true | 有効フラグ |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

---

### 認証

---

#### `user_accounts`

用途: システムへのログインアカウント。staffsと1対1で紐づく（任意）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| email | text | NO | | メールアドレス（UNIQUE） |
| password_hash | text | NO | | bcryptハッシュ化パスワード（JSON非公開） |
| display_name | text | NO | | 表示名 |
| display_name_kana | text | YES | '' | 表示名カナ |
| user_type | user_type | NO | 'staff' | ユーザー種別 |
| job_title_id | uuid | YES | NULL | job_titles.id FK（職種） |
| status | account_status | YES | 'active' | アカウント状態 |
| avatar_url | text | YES | '' | アバター画像URL |
| staff_id | uuid | YES | | staffs.id FK |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `staff_id` → `staffs.id` (SET NULL)
- `job_title_id` → `job_titles.id` (SET NULL)

---

#### `user_clinic_memberships`

用途: ユーザーの医院所属。1ユーザーが複数医院に所属可能。is_main=trueは各ユーザーにつき1件のみ。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| user_id | uuid | NO | | user_accounts.id FK |
| clinic_id | uuid | NO | | clinics.id FK |
| is_main | boolean | YES | false | 主所属医院フラグ |
| joined_at | timestamptz | YES | now() | 所属開始日時 |

**FK:**
- `user_id` → `user_accounts.id` (CASCADE)
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:**
- `(user_id, clinic_id)` UNIQUE
- `(user_id) WHERE is_main = true` UNIQUE（部分インデックス）

---

#### `user_permissions`

用途: ユーザーの医院別権限。医院ごとに複数の権限を付与可能。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| user_id | uuid | NO | | user_accounts.id FK |
| clinic_id | uuid | NO | | clinics.id FK |
| permission | permission_type | NO | | 権限種別 |
| granted_by | uuid | YES | | 付与者 user_accounts.id FK |
| granted_at | timestamptz | YES | now() | 付与日時 |

**FK:**
- `user_id` → `user_accounts.id` (CASCADE)
- `clinic_id` → `clinics.id` (RESTRICT)
- `granted_by` → `user_accounts.id` (SET NULL)

**UNIQUE制約:** `(user_id, clinic_id, permission)` — 同一ユーザー・医院・権限の重複登録防止

---

### コア

---

#### `owners`

用途: ペットの飼い主情報。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| owner_name | text | NO | | 飼い主名 |
| owner_name_kana | text | YES | '' | 飼い主名カナ |
| birth_date | date | YES | | 生年月日 |
| company | text | YES | '' | 会社名 |
| postal_code | text | YES | '' | 郵便番号（会社） |
| address1 | text | YES | '' | 住所1（会社） |
| address2 | text | YES | '' | 住所2（会社） |
| home_postal_code | text | YES | '' | 郵便番号（自宅） |
| home_address1 | text | YES | '' | 住所1（自宅） |
| home_address2 | text | YES | '' | 住所2（自宅） |
| phone | text | YES | '' | 電話番号 |
| company_phone | text | YES | '' | 会社電話番号 |
| email | text | YES | '' | メールアドレス |
| remarks | text | YES | '' | 備考 |
| is_dangerous | boolean | YES | false | 危険フラグ |
| discount_rate | numeric | YES | 0 | 割引率 |
| membership_type | membership_type | YES | '非会員' | 会員種別 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `pets`

用途: ペット情報。飼い主（owners）に属する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| owner_id | uuid | NO | | owners.id FK |
| pet_number | text | YES | '' | ペット番号 |
| name | text | NO | | ペット名 |
| pet_name_kana | text | YES | '' | ペット名カナ |
| animal_species_id | uuid | NO | | animal_species.id FK（種類） |
| gender | pet_gender | YES | '不明' | 性別 |
| status | pet_status | YES | '生存' | 生存状態 |
| birth_date | date | YES | | 誕生日 |
| breed | text | YES | '' | 品種 |
| color | text | YES | '' | 毛色 |
| weight | numeric | YES | | 体重(kg) |
| neutered_date | date | YES | | 去勢・避妊手術日 |
| acquisition_type | acquisition_type | YES | | 取得区分 |
| danger_level | danger_level | YES | '低' | 危険度 |
| food | text | YES | '' | 食事内容 |
| environment | text | YES | '' | 飼育環境 |
| phone | text | YES | '' | ペット専用電話 |
| last_visit | date | YES | | 最終来院日 |
| insurance_id | uuid | YES | | insurances.id FK |
| remarks | text | YES | '' | 備考 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `owner_id` → `owners.id` (RESTRICT)
- `animal_species_id` → `animal_species.id` (RESTRICT)
- `insurance_id` → `insurances.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

### マスタ

---

#### `animal_species`

用途: ペット種類マスタ（犬・猫・鳥・その他等）。システム共通マスタのため clinic_id なし。`pets.animal_species_id` から参照される。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| name | text | NO | | 種類名（犬・猫・鳥・その他 等） |
| is_active | boolean | YES | true | 状態 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:** なし（システム共通マスタ）


---

#### `staffs`

用途: スタッフ（獣医師・看護師・トリマー等）のマスタ。認証情報は持たず user_accounts と別管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | スタッフ名 |
| is_active | boolean | YES | true | 状態 |
| staff_role | staff_role | NO | | 役割 |
| license_number | text | YES | '' | 免許番号 |
| job_title_id | uuid | YES | - | job_titles.id FK（職種） |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `job_title_id` → `job_titles.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

#### `job_titles`

用途: 職種マスタ。clinic 単位で管理（スタッフマスタと同じスコープ）。`staffs.job_title_id` から参照される。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | gen_random_uuid() | PK |
| clinic_id | uuid | NO | - | clinics.id FK（所属医院） |
| name | text | NO | '' | 職種名（例: 獣医師, 看護師） |
| description | text | YES | '' | 説明 |
| sort_order | int | NO | 0 | 表示順 |
| is_active | boolean | NO | true | 有効フラグ |
| created_at | timestamptz | NO | now() | 作成日時 |
| updated_at | timestamptz | NO | now() | 更新日時 |

**FK**: clinic_id → clinics(id) RESTRICT


---

#### `inventory_items`

用途: 在庫アイテム。薬剤マスタ（medicines）から参照される。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 品目名 |
| category | inventory_category | NO | | カテゴリ |
| quantity | integer | YES | 0 | 在庫数量 |
| unit | text | NO | '' | 単位 |
| min_stock_level | integer | YES | 0 | 最低在庫数 |
| location | text | YES | '' | 保管場所 |
| expiry_date | date | YES | | 有効期限 |
| supplier | text | YES | '' | 仕入先 |
| last_restocked | date | YES | | 最終補充日 |
| status | inventory_status | YES | 'sufficient' | 在庫状態 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `exam_types`

用途: 検査種別マスタ（血液検査・尿検査等）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 検査種別名 |
| parent_id | bigint | YES | NULL | 親検査種別ID FK → exam_types(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `exam_types.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`

---

#### `exam_type_items`

用途: 検査種別に属する検査項目定義（検査結果のテンプレート）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| exam_type_id | uuid | NO | | exam_types.id FK |
| name | text | NO | | 検査項目名 |
| inspection_value | text | YES | '' | 検査値（テンプレート） |
| normal_value | text | YES | '' | 正常値 |
| unit | text | YES | '' | 単位 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:** `exam_type_id` → `exam_types.id` (CASCADE)

---

#### `vaccines`

用途: ワクチンマスタ。動物種別・接種間隔を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | ワクチン名 |
| parent_id | bigint | YES | NULL | 親ワクチンID FK → vaccines(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| species | vaccine_species | YES | | 対象動物種 |
| interval | text | YES | '' | 接種間隔 |
| inventory_id | uuid | YES | | inventory_items.id FK |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `vaccines.id` (SET NULL)
- `inventory_id` → `inventory_items.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`

---

#### `medicines`

用途: 薬剤マスタ。在庫アイテム（inventory_items）と紐づく。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 薬剤名 |
| parent_id | bigint | YES | NULL | 親薬剤ID（カテゴリ medicine を参照）FK → medicines(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| dosage_form | dosage_form | YES | | 剤形 |
| medicine_unit | medicine_unit | YES | | 単位 |
| inventory_id | uuid | YES | | inventory_items.id FK |
| default_quantity | integer | YES | 1 | デフォルト数量 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `medicines.id` (SET NULL)
- `inventory_id` → `inventory_items.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`

---

#### `insurances`

用途: 保険マスタ。保険種別・補償率を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 保険名 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| coverage_rate | integer | NO | - | 補償率(%) CHECK (0 <= coverage_rate AND coverage_rate <= 100) |
| contact_phone | text | YES | '' | 問い合わせ電話番号 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `cages`

用途: ケージマスタ。入院・ホテルで使用するケージの種別・サイズを管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | ケージ名 |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| cage_type | cage_type | NO | | ケージ種別 |
| cage_size | cage_size | NO | | ケージサイズ |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `service_types`

用途: サービス種別マスタ（予約に使用）。表示色を保持。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | サービス種別名 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| color | text | YES | '#3B82F6' | 表示色（HEX） |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `consultations`

用途: 診察項目マスタ（初診・再診・往診等）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 診察項目名 |
| parent_id | bigint | YES | NULL | 親診察項目ID FK → consultations(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| time_condition | text | YES | '' | 時間条件 |
| duration | integer | YES | | 標準診察時間(分) |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `consultations.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`

---

#### `procedures`

用途: 処置項目マスタ（手術・注射・処置等）。麻酔種別を保持。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 処置項目名 |
| parent_id | bigint | YES | NULL | 親処置項目ID FK → procedures(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| duration | integer | YES | | 所要時間目安(分) |
| anesthesia | anesthesia_type | YES | 'none' | 麻酔種別 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `procedures.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`

---

#### `hospitalization_plans`

用途: 入院プランマスタ。体格・課金単位（1泊/1日）を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | プラン名 |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| body_size | body_size | YES | | 体格区分 |
| billing_unit | billing_unit | YES | 'per_day' | 課金単位 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `trimming_courses`

用途: トリミングコースマスタ。対象サイズ・所要時間を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | コース名 |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| target_size | target_size | YES | | 対象サイズ |
| duration | integer | YES | | 所要時間(分) |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `trimming_options`

用途: トリミングオプションマスタ（シャンプー・カット等の追加オプション）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | オプション名 |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| duration | integer | YES | | 追加所要時間(分) |
| combinable | boolean | NO | true | 他オプションと組み合わせ可能か |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `diagnosis_categories`

用途: 診断カテゴリマスタ（消化器・循環器等）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | カテゴリ名 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `diagnosis_names`

用途: 診断病名マスタ。診断カテゴリに属する具体的な診断病名。self-referencing廃止・明示的FK採用。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 診断名 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| diagnosis_category_id | uuid | NO | | diagnosis_categories.id FK |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `diagnosis_category_id` → `diagnosis_categories.id` (CASCADE)

**インデックス:** `(clinic_id)`

---

#### `checkup_types`

用途: 健診種別マスタ（定期健診・シニア健診等）。間隔・対象年齢を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 健診種別名 |
| parent_id | bigint | YES | NULL | 親健診種別ID FK → checkup_types(id) SET NULL |
| price | integer | YES | | 価格 |
| is_active | boolean | YES | true | 状態 |
| description | text | YES | '' | 説明 |
| interval | text | YES | '' | 推奨間隔 |
| target_age | text | YES | '' | 対象年齢 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `parent_id` → `checkup_types.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(parent_id)`, `(deleted_at)`

---

#### `chief_complaint_categories`

用途: 問診タブの主訴区分マスタ（呼吸器系・消化器系・外傷等）。問診記録で主訴の区分を選択する際に使用。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO | - | FK → clinics(id) RESTRICT（所属医院） |
| name | text | NO | | 区分名 |
| description | text | YES | '' | 説明 |
| is_active | boolean | YES | true | 状態 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`

---

#### `inquiry_templates`

用途: 問診タブの定型文マスタ。フィールド区分（category）ごとに定型文を登録し、問診入力時に選択して使用する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO | - | FK → clinics(id) RESTRICT（所属医院） |
| category | text | NO | '' | 使用フィールド区分（chief_complaint / history / current_medications / allergy_info / notes 等） |
| title | text | NO | | 定型文タイトル |
| content | text | NO | '' | 定型文本文 |
| is_active | boolean | YES | true | 状態 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(clinic_id)`, `(clinic_id, category)`

---

### 診療

---

#### `medical_records`

用途: カルテ（診療記録）。1回の来院に対し1件作成。record_noは clinic_id スコープで UNIQUE。

> ⚠️ v7.0: `chief_complaint` は `inquiries.chief_complaint` に移動。
> ⚠️ v9.0: `physical_exam`, `treatment_policy`, `diagnosis_details`, `diagnosis1/2` FK は `clinical_plans` に移動。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| record_no | text | NO | | カルテ番号（clinic_id + record_no UNIQUE） |
| date | date | NO | | 診療日 |
| owner_id | uuid | YES | | FK → owners(id) RESTRICT |
| pet_id | uuid | YES | | FK → pets(id) RESTRICT |
| doctor_id | uuid | YES | | FK → staffs(id) SET NULL（担当医師） |
| reservation_appointment_id | uuid | YES | NULL | reservation_appointments.id FK SET NULL（紐づく予約） |
| status | medical_record_status | YES | '作成中' | カルテ状態 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `owner_id` → `owners.id` (RESTRICT)
- `pet_id` → `pets.id` (RESTRICT)
- `doctor_id` → `staffs.id` (SET NULL)
- `reservation_appointment_id` → `reservation_appointments.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

#### `clinical_plans`

**用途**: 診察/治療タブ。医師による身体検査所見・診断・治療方針を記録。1カルテに1件（1:1）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO | | FK → medical_records(id) CASCADE（UNIQUE） |
| physical_exam | text | YES | '' | 身体検査所見（O: Objective） |
| diagnosis_category_id | uuid | YES | | FK → diagnosis_categories(id) SET NULL（診断カテゴリ） |
| diagnosis_name_id | uuid | YES | | FK → diagnosis_names(id) SET NULL（診断病名） |
| diagnosis_details | text | YES | '' | 診断詳細（A: Assessment） |
| treatment_policy | text | YES | '' | 治療方針（P: Plan） |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE) UNIQUE
- `diagnosis_category_id` → `diagnosis_categories.id` (SET NULL)
- `diagnosis_name_id` → `diagnosis_names.id` (SET NULL)

**インデックス:** `medical_record_id` UNIQUE（1:1保証）

---

#### `inquiries`

**用途**: カルテ問診タブ。飼主からの問診情報を記録。1カルテに1件（1:1）。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO | - | FK → medical_records(id) CASCADE, UNIQUE |
| chief_complaint_category_id | uuid | YES | - | FK → chief_complaint_categories(id) SET NULL（主訴区分） |
| chief_complaint | text | YES | '' | 主訴 |
| history | text | YES | '' | 既往歴・現病歴 |
| current_medications | text | YES | '' | 現在の投薬状況 |
| allergy_info | text | YES | '' | アレルギー情報 |
| last_meal | text | YES | '' | 最終食事 |
| last_defecation | text | YES | '' | 最終排便 |
| last_urination | text | YES | '' | 最終排尿 |
| appetite | appetite_level | YES | - | 食欲レベル |
| water_intake | water_intake_level | YES | - | 飲水量レベル |
| owner_observations | text | YES | '' | 飼主の気になる点 |
| notes | text | YES | '' | その他メモ |
| staff_id | uuid | YES | - | FK → staffs(id) SET NULL（問診担当） |
| created_at | timestamptz | YES | now() | |
| updated_at | timestamptz | YES | now() | |

**FK**: medical_record_id → medical_records(id) CASCADE, chief_complaint_category_id → chief_complaint_categories(id) SET NULL, staff_id → staffs(id) SET NULL

**インデックス**: medical_record_id UNIQUE（1:1保証）

---

#### `treatments`

用途: カルテに紐づく処置・診察・薬剤の明細。item_typeで種別を区別し、対応するFKが設定される。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO | | medical_records.id FK |
| item_type | treatment_item_type | NO | 'other' | 明細種別 |
| consultation_id | uuid | YES | | consultations.id FK |
| procedure_id | uuid | YES | | procedures.id FK |
| medicine_id | uuid | YES | | medicines.id FK |
| selected | boolean | YES | false | 選択フラグ |
| status | treatment_status | YES | 'pending' | 処置状態 |
| content | text | NO | '' | 内容 |
| memo | text | YES | '' | メモ |
| insurance | boolean | YES | false | 保険適用フラグ |
| unit_price | integer | YES | 0 | 単価 |
| quantity | integer | YES | 1 | 数量 |
| discount_rate | numeric | YES | 0 | 割引率 |
| discount_amount | integer | YES | 0 | 割引額 |
| inventory_id | uuid | YES | | inventory_items.id FK |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE)
- `consultation_id` → `consultations.id` (SET NULL)
- `procedure_id` → `procedures.id` (SET NULL)
- `medicine_id` → `medicines.id` (SET NULL)
- `inventory_id` → `inventory_items.id` (SET NULL)

**CHECK制約:** `chk_treatment_item_ref` — item_typeとFK列の整合性

---

#### `vital_records`

用途: バイタル記録（外来・入院統合）。`pet_id` を主軸に、外来時は `medical_record_id`、入院時は `daily_record_id` を持つ。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | bigint | NO | BIGSERIAL | PK |
| pet_id | bigint | NO | | pets.id FK |
| medical_record_id | bigint | YES | | medical_records.id FK（外来時） |
| daily_record_id | bigint | YES | | daily_records.id FK（入院時） |
| recorded_at | timestamptz | NO | now() | 測定日時 |
| staff_id | bigint | YES | | staffs.id FK |
| temperature | numeric | YES | | 体温（℃） |
| heart_rate | integer | YES | | 心拍数（bpm） |
| respiration_rate | integer | YES | | 呼吸数（回/分） |
| weight | numeric | YES | | 体重（kg） |
| notes | text | NO | '' | 備考 |
| created_at | timestamptz | NO | now() | 作成日時 |
| updated_at | timestamptz | NO | now() | 更新日時 |

**CHECK制約:**
- `(medical_record_id IS NOT NULL) OR (daily_record_id IS NOT NULL)` — 外来か入院どちらか必須
- `chk_vital_temperature` — temperature IS NULL OR (temperature >= 30.0 AND temperature <= 50.0)
- `chk_vital_heart_rate` — heart_rate IS NULL OR (heart_rate > 0 AND heart_rate < 500)
- `chk_vital_respiration` — respiration_rate IS NULL OR (respiration_rate > 0 AND respiration_rate < 200)
- `chk_vital_weight` — weight IS NULL OR weight > 0

**FK:**
- `pet_id` → `pets.id` (CASCADE)
- `medical_record_id` → `medical_records.id` (CASCADE)
- `daily_record_id` → `daily_records.id` (CASCADE)
- `staff_id` → `staffs.id` (SET NULL)

---

#### `exams`

用途: 検査記録。カルテ・ペットに紐づき、検査種別マスタを参照する。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO  | | medical_records.id FK |
| clinic_id | uuid | NO | | clinics.id FK（所属医院） |
| pet_id | uuid | YES | | pets.id FK（ペット単位検索用。medical_record_id 経由でも辿れるが、JOIN 削減のため意図的に保持） |
| date | date | NO | | 検査日 |
| exam_type_id | uuid | NO | | exam_types.id FK |
| doctor_id | uuid | YES | | staffs.id FK |
| status | examination_status | YES | 'pending' | 検査状態 |
| result_summary | text | YES | '' | 検査結果サマリ |
| machine | text | YES | '' | 使用機器 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE)
- `clinic_id` → `clinics.id` (RESTRICT)
- `pet_id` → `pets.id` (RESTRICT)
- `exam_type_id` → `exam_types.id` (RESTRICT)
- `doctor_id` → `staffs.id` (SET NULL)

---

#### `exam_items`

用途: 検査記録の各検査項目結果。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| exam_id | uuid | NO | | exams.id FK |
| exam_type_item_id | uuid | YES | | exam_type_items.id FK（検査項目定義への参照） |
| name | text | NO | '' | 検査項目名 |
| inspection_value | text | YES | '' | 検査値 |
| normal_value | text | YES | '' | 正常値 |
| result | text | YES | '' | 結果コメント |
| unit | text | YES | '' | 単位 |
| ref | text | YES | '' | 参考値 |
| status | examination_result_status | YES | 'normal' | 結果状態 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `exam_id` → `exams.id` (CASCADE)
- `exam_type_item_id` → `exam_type_items.id` (SET NULL)

---

#### `vaccinations`

用途: ワクチン接種記録。vaccine_id FK にてワクチンマスタを参照。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | YES | | medical_records.id FK（NULL = カルテなしの単独接種） |
| pet_id | uuid | YES | | pets.id FK（ペット単位検索用。medical_record_id 経由でも辿れるが、JOIN 削減のため意図的に保持） |
| vaccine_id | uuid | NO | | vaccines.id FK |
| clinic_id | uuid | NO | | clinics.id FK（接種医院） |
| date | date | NO | | 接種日 |
| next_date | date | YES | | 次回接種予定日 |
| next_schedule_type | next_schedule_type | YES | | 次回スケジュール種別 |
| doctor_id | uuid | YES | | staffs.id FK |
| supplemental | text | YES | '' | 補足情報 |
| lot1 | text | YES | '' | ロット番号1 |
| lot2 | text | YES | '' | ロット番号2 |
| lot3 | text | YES | '' | ロット番号3 |
| lot4 | text | YES | '' | ロット番号4 |
| remarks | text | YES | '' | 備考 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE)
- `pet_id` → `pets.id` (RESTRICT)
- `vaccine_id` → `vaccines.id` (RESTRICT)
- `doctor_id` → `staffs.id` (SET NULL)

---

#### `checkups`

用途: 健診記録。健診種別マスタを参照し、次回健診日を管理。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO  | | medical_records.id FK |
| clinic_id | uuid | NO | | clinics.id FK（所属医院） |
| pet_id | uuid | YES | | pets.id FK（ペット単位検索用。medical_record_id 経由でも辿れるが、JOIN 削減のため意図的に保持） |
| checkup_type_id | uuid | NO | | checkup_types.id FK |
| date | date | NO | | 健診日 |
| next_date | date | YES | | 次回健診予定日 |
| doctor_id | uuid | YES | | staffs.id FK |
| result | text | YES | '' | 健診結果 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE)
- `clinic_id` → `clinics.id` (RESTRICT)
- `pet_id` → `pets.id` (RESTRICT)
- `checkup_type_id` → `checkup_types.id` (RESTRICT)
- `doctor_id` → `staffs.id` (SET NULL)

---

#### `record_images`

**用途**: カルテ画像タブ。レントゲン・エコー・写真等の診療画像を管理。1カルテに複数件。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO | - | FK → medical_records(id) CASCADE |
| image_url | text | NO | '' | 画像URL（オブジェクトストレージ） |
| thumbnail_url | text | YES | '' | サムネイルURL |
| file_name | text | NO | '' | 元ファイル名 |
| file_size | bigint | YES | 0 | ファイルサイズ（bytes） |
| mime_type | text | YES | '' | MIMEタイプ |
| image_type | medical_image_type | NO | 'other' | 画像種別 |
| description | text | YES | '' | 説明・所見メモ |
| taken_at | timestamptz | YES | - | 撮影日時 |
| exam_id | uuid | YES | - | FK → exams(id) SET NULL |
| staff_id | uuid | YES | - | FK → staffs(id) SET NULL（撮影者） |
| sort_order | integer | YES | 0 | 表示順 |
| created_at | timestamptz | YES | now() | |
| updated_at | timestamptz | NO | now() | 更新日時 |

**FK**: medical_record_id → medical_records(id) CASCADE, exam_id → exams(id) SET NULL, staff_id → staffs(id) SET NULL

**インデックス**: (medical_record_id), (image_type), (taken_at DESC), (exam_id) WHERE NOT NULL

---

#### `estimates`

**用途**: カルテ見積書タブ。診察前後の費用見積書。1カルテに複数件作成可。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO | - | clinics.id FK（所属医院、RESTRICT） |
| estimate_no | text | NO | - | 見積書番号（clinic_id + estimate_no UNIQUE） |
| medical_record_id | uuid | YES | - | FK → medical_records(id) SET NULL |
| title | text | YES | '' | 件名 |
| owner_id | uuid | YES | - | FK → owners(id) SET NULL |
| status | estimate_status | YES | 'draft' | draft/sent/approved/rejected |
| subtotal | integer | NO | 0 | 小計 |
| tax_total | integer | NO | 0 | 税合計 |
| total_amount | integer | NO | 0 | 合計金額 |
| insurance_amount | integer | YES | 0 | 保険適用額 |
| discount_amount | integer | YES | 0 | 値引き額 |
| valid_until | date | YES | - | 有効期限 |
| comment | text | YES | '' | コメント |
| notes | text | YES | '' | 備考 |
| created_by | uuid | YES | - | FK → staffs(id) SET NULL（作成者） |
| created_at | timestamptz | YES | now() | |
| updated_at | timestamptz | YES | now() | |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK**: clinic_id → clinics(id) RESTRICT, medical_record_id → medical_records(id) RESTRICT, owner_id → owners(id) SET NULL, created_by → staffs(id) SET NULL

**インデックス**: (clinic_id, estimate_no) UNIQUE, (medical_record_id), (status), (owner_id)

---

#### `estimate_items`

**用途**: 見積書の明細行。診察・処置・薬剤等を行単位で管理。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| estimate_id | uuid | NO | - | FK → estimates(id) CASCADE |
| name | text | NO | '' | 項目名 |
| category | item_category | NO | - | 区分 |
| unit_price | integer | NO | 0 | 単価 |
| quantity | integer | NO | 1 | 数量 |
| tax_rate | numeric | YES | 0.10 | 税率 |
| discount_rate | numeric | YES | 0 | 割引率 |
| discount_amount | integer | YES | 0 | 値引額 |
| is_insurance_applicable | boolean | YES | false | 保険適用可否 |
| consultation_id | uuid | YES | - | FK → consultations(id) SET NULL |
| procedure_id | uuid | YES | - | FK → procedures(id) SET NULL |
| medicine_id | uuid | YES | - | FK → medicines(id) SET NULL |
| sort_order | integer | YES | 0 | 表示順 |
| created_at | timestamptz | YES | now() | |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK**: estimate_id → estimates(id) CASCADE, consultation_id → consultations(id) SET NULL, procedure_id → procedures(id) SET NULL, medicine_id → medicines(id) SET NULL

**インデックス**: (estimate_id)

---

#### `billing_reviews`

**用途**: カルテ会計（医師確認）タブ。医師が会計内容を確認・承認するレコード。1カルテに1件（1:1）。

| カラム | 型 | NULL | デフォルト | 説明 |
|--------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | NO | - | FK → medical_records(id) CASCADE, UNIQUE |
| status | billing_review_status | YES | 'pending' | pending/confirmed/returned |
| confirmed_by | uuid | YES | - | FK → staffs(id) SET NULL（確認医師） |
| confirmed_at | timestamptz | YES | - | 確認日時 |
| returned_by | uuid | YES | - | FK → staffs(id) SET NULL（差戻し者） |
| returned_at | timestamptz | YES | - | 差戻し日時 |
| return_reason | text | YES | '' | 差戻し理由 |
| memo | text | YES | '' | メモ |
| created_at | timestamptz | YES | now() | |
| updated_at | timestamptz | YES | now() | |

**FK**: medical_record_id → medical_records(id) CASCADE, confirmed_by → staffs(id) SET NULL, returned_by → staffs(id) SET NULL

**インデックス**: (medical_record_id) UNIQUE, (status)

---

### 予約

---

#### `reservation_appointments`

用途: 予約情報。ペット・サービス種別・担当医に紐づく。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| start_time | timestamptz | NO | | 開始日時 |
| end_time | timestamptz | NO | | 終了日時 |
| owner_id | uuid | YES | | owners.id FK |
| pet_id | uuid | YES | | pets.id FK |
| visit_type | visit_type | NO | 'revisit' | 来院種別 |
| service_type_id | uuid | NO  | | service_types.id FK |
| doctor_id | uuid | YES | | staffs.id FK |
| is_designated | boolean | YES | false | 担当医指名フラグ |
| status | reservation_status | YES | 'pending' | 予約状態 |
| notes | text | YES | '' | 備考 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `owner_id` → `owners.id` (SET NULL)
- `pet_id` → `pets.id` (SET NULL)
- `service_type_id` → `service_types.id` (RESTRICT)
- `doctor_id` → `staffs.id` (SET NULL)

**インデックス:** `(clinic_id)`, `(owner_id)`

---

### 入院

---

#### `hospitalizations`

用途: 入院・ホテル管理。ペット・ケージ・担当医に紐づく。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| owner_id | uuid | NO  | | owners.id FK |
| pet_id | uuid | NO  | | pets.id FK |
| hospitalization_type | hospitalization_type | NO | | 入院種別 |
| start_date | date | NO | | 入院開始日 |
| end_date | date | NO | | 入院終了日 |
| status | hospitalization_status | YES | '予約' | 入院状態 |
| cage_id | uuid | YES | | cages.id FK |
| doctor_id | uuid | YES | | staffs.id FK |
| memo | text | YES | '' | メモ |
| owner_request | text | YES | '' | 飼い主要望 |
| staff_notes | text | YES | '' | スタッフメモ |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `owner_id` → `owners.id` (RESTRICT)
- `pet_id` → `pets.id` (RESTRICT)
- `cage_id` → `cages.id` (SET NULL)
- `doctor_id` → `staffs.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

#### `daily_records`

用途: 入院の日次記録ヘッダ。1入院・1日につき1件（UNIQUE制約）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| hospitalization_id | uuid | NO | | hospitalizations.id FK |
| clinic_id | uuid | NO | | clinics.id FK（所属医院） |
| date | date | NO | | 日付 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `hospitalization_id` → `hospitalizations.id` (CASCADE)
- `clinic_id` → `clinics.id` (RESTRICT)

**インデックス:** `(hospitalization_id, date)` UNIQUE

---

#### `care_plan_items`

用途: 入院のケアプラン項目。食事・投薬・処置・指示・物品を管理。ポリモーフィック参照廃止・3専用FK採用。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| hospitalization_id | uuid | NO | | hospitalizations.id FK |
| type | care_plan_type | NO | | ケアプラン種別 |
| name | text | NO | '' | 項目名 |
| description | text | YES | '' | 説明 |
| timing | plan_timing[] | YES | '{}' | 実施タイミング（配列） |
| status | care_plan_status | YES | 'active' | 状態 |
| notes | text | YES | '' | 備考 |
| medicine_id | uuid | YES | | medicines.id FK |
| procedure_id | uuid | YES | | procedures.id FK |
| hospitalization_plan_id | uuid | YES | | hospitalization_plans.id FK |
| unit_price | integer | YES | 0 | 単価 |
| category | text | YES | '' | カテゴリ |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `hospitalization_id` → `hospitalizations.id` (CASCADE)
- `medicine_id` → `medicines.id` (SET NULL)
- `procedure_id` → `procedures.id` (SET NULL)
- `hospitalization_plan_id` → `hospitalization_plans.id` (SET NULL)

**CHECK制約:** `chk_care_plan_item_ref` — typeとFK列の整合性

---

#### `care_log_records`

用途: 日次記録に紐づくケアログ（実際の処置・食事・排泄等の記録）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| daily_record_id | uuid | NO | | daily_records.id FK |
| time | time | NO | | 実施時刻 |
| type | care_log_type | NO | | ケアログ種別 |
| status | care_log_status | NO | 'completed' | 実施状態 |
| value | text | YES | '' | 値（量・回数等） |
| staff_id | uuid | YES | | staffs.id FK |
| notes | text | YES | '' | 備考 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `daily_record_id` → `daily_records.id` (CASCADE)
- `staff_id` → `staffs.id` (SET NULL)

---

#### `staff_note_records`

用途: 入院中のスタッフノート（日次記録に紐づく自由記述）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| daily_record_id | uuid | NO | | daily_records.id FK |
| time | text | NO | '' | 記録時刻 |
| content | text | NO | '' | 内容 |
| staff_id | uuid | YES | | staffs.id FK |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `daily_record_id` → `daily_records.id` (CASCADE)
- `staff_id` → `staffs.id` (SET NULL)

---

#### `treatment_plans`

用途: 治療プラン・費用明細。外来（`medical_record_id`）と入院（`hospitalization_id`）の両方で使用。どちらか一方が必ず設定される（CHECK制約）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| medical_record_id | uuid | YES | | medical_records.id FK（外来カルテ） |
| hospitalization_id | uuid | YES | | hospitalizations.id FK（入院） |
| treatment_content | text | NO | '' | 治療内容 |
| memo | text | YES | '' | メモ |
| insurance | boolean | YES | false | 保険適用フラグ |
| unit_price | integer | YES | 0 | 単価 |
| quantity | integer | YES | 1 | 数量 |
| discount_rate | numeric | YES | 0 | 割引率 |
| discount_amount | integer | YES | 0 | 割引額 |
| subtotal | integer | YES | 0 | 小計 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `medical_record_id` → `medical_records.id` (CASCADE)
- `hospitalization_id` → `hospitalizations.id` (CASCADE)

**CHECK制約:** `chk_treatment_plans_ref` — `((medical_record_id IS NOT NULL AND hospitalization_id IS NULL) OR (medical_record_id IS NULL AND hospitalization_id IS NOT NULL))`

---

### トリミング

---

#### `trimming_records`

用途: トリミング実施記録。ペット・担当スタッフ・コースに紐づく。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| date | date | NO | | トリミング日 |
| pet_id | uuid | YES | | pets.id FK |
| weight | text | YES | '' | 体重 |
| style_request | text | YES | '' | スタイルリクエスト |
| staff_id | uuid | YES | | staffs.id FK |
| status | trimming_status | YES | 'reserved' | 状態 |
| course_id | uuid | YES | | trimming_courses.id FK |
| bw | text | YES | '' | 体重測定値 |
| bw_unit | body_weight_unit | YES | 'Kg' | 体重単位 |
| bt | text | YES | '' | 体温 |
| used_shampoo | text | YES | '' | 使用シャンプー |
| used_ribbon | text | YES | '' | 使用リボン |
| remarks | text | YES | '' | 備考 |
| style_image | text | YES | '' | スタイル見本画像URL |
| completed_image | text | YES | '' | 完成画像URL |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `pet_id` → `pets.id` (RESTRICT)
- `staff_id` → `staffs.id` (SET NULL)
- `course_id` → `trimming_courses.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

#### `trimming_record_options`

用途: トリミング記録に紐づく選択オプション（多対多中間テーブル）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| trimming_record_id | uuid | NO | | trimming_records.id FK |
| option_id | uuid | NO | | trimming_options.id FK |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `trimming_record_id` → `trimming_records.id` (CASCADE)
- `option_id` → `trimming_options.id` (RESTRICT)

**インデックス:** `(trimming_record_id, option_id)` UNIQUE

---

### 会計

---

#### `billings`

用途: 会計情報。カルテまたは入院に1件対応（medical_record_idはUNIQUE）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| medical_record_id | uuid | YES | | medical_records.id FK（UNIQUE） |
| hospitalization_id | uuid | YES | | hospitalizations.id FK |
| owner_id | uuid | YES | | owners.id FK（飼主単位検索用。意図的な非正規化） |
| pet_id | uuid | YES | | pets.id FK（ペット単位検索用。意図的な非正規化） |
| subtotal | integer | NO | 0 | 小計（税抜） |
| tax_total | integer | NO | 0 | 消費税額 |
| total_amount | integer | NO | 0 | 合計（税込） |
| has_insurance | boolean | NO | false | 保険適用フラグ（会計一覧での保険フィルタ用。payments.insurance_name の有無と連動） |
| status | billing_status | YES | 'waiting' | 会計状態 |
| scheduled_date | date | NO | | 会計予定日 |
| completed_at | timestamptz | YES | | 会計完了日時 |
| memo | text | YES | '' | メモ |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `medical_record_id` → `medical_records.id` (SET NULL)
- `hospitalization_id` → `hospitalizations.id` (SET NULL)
- `owner_id` → `owners.id` (SET NULL)
- `pet_id` → `pets.id` (SET NULL)

**インデックス:** `(clinic_id)`

---

#### `billing_items`

用途: 会計明細。会計に紐づく各請求項目。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| billing_id | uuid | NO | | billings.id FK |
| category | item_category | NO | | 明細カテゴリ |
| name | text | NO | '' | 項目名 |
| unit_price | integer | NO | 0 | 単価 |
| quantity | integer | NO | 1 | 数量 |
| tax_rate | numeric | YES | 0.10 | 税率 |
| is_insurance_applicable | boolean | YES | false | 保険適用フラグ |
| source | item_source | YES | 'manual' | 明細元 |
| sort_order | integer | YES | 0 | 並び順 |
| created_at | timestamptz | YES | now() | 作成日時 |
| deleted_at | timestamptz | YES | NULL | 論理削除日時（NULL = 有効） |

**FK:** `billing_id` → `billings.id` (CASCADE)

---

#### `payments`

用途: 支払い情報。会計に1対1で紐づく（billing_idはUNIQUE）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| billing_id | uuid | NO | | billings.id FK（UNIQUE） |
| subtotal | numeric | NO | 0 | 小計 |
| tax_total | numeric | NO | 0 | 消費税合計 |
| total_amount | numeric | NO | 0 | 合計金額 |
| insurance_name | text | YES | '' | 保険名（会計確定時スナップショット、変更後も履歴保持のため意図的に保持） |
| insurance_ratio | numeric | YES | 0 | 保険補償率スナップショット（同上） |
| insurance_amount | numeric | YES | 0 | 保険補填額 |
| discount_amount | numeric | YES | 0 | 割引額 |
| billing_amount | numeric | NO | 0 | 請求額 |
| received_amount | numeric | YES | 0 | 受取金額 |
| change_amount | numeric | YES | 0 | お釣り |
| method | payment_method | YES | 'cash' | 支払い方法 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:** `billing_id` → `billings.id` (CASCADE)

---

### シフト

---

#### `shift_entries`

用途: スタッフのシフト情報。1スタッフ・1日につき1件（UNIQUE制約）。

| カラム名 | 型 | NULL | デフォルト | 説明 |
|---------|-----|------|-----------|------|
| id | uuid | NO | uuid_generate_v4() | PK |
| clinic_id | uuid | NO  | - | FK → clinics(id) RESTRICT（所属医院） |
| staff_id | uuid | NO | | staffs.id FK |
| date | date | NO | | 日付 |
| shift_type | shift_type | NO | | シフト種別 |
| start_time | time | YES | | 開始時刻 |
| end_time | time | YES | | 終了時刻 |
| note | text | YES | '' | 備考 |
| created_at | timestamptz | YES | now() | 作成日時 |
| updated_at | timestamptz | YES | now() | 更新日時 |

**FK:**
- `clinic_id` → `clinics.id` (RESTRICT)
- `staff_id` → `staffs.id` (RESTRICT)

**インデックス:**
- `(staff_id, date)` UNIQUE
- `(clinic_id)`

---

## FK関係一覧

### animal_species

FK なし（システム共通マスタ）

### billings

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| medical_record_id | medical_records.id | SET NULL |
| hospitalization_id | hospitalizations.id | SET NULL |
| owner_id | owners.id | SET NULL |
| pet_id | pets.id | SET NULL |

### billing_reviews

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| medical_record_id | medical_records.id | CASCADE |
| confirmed_by | staffs.id | SET NULL |
| returned_by | staffs.id | SET NULL |

### estimate_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| estimate_id | estimates.id | CASCADE |
| consultation_id | consultations.id | SET NULL |
| procedure_id | procedures.id | SET NULL |
| medicine_id | medicines.id | SET NULL |

### estimates

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| medical_record_id | medical_records.id | RESTRICT |
| owner_id | owners.id | SET NULL |
| created_by | staffs.id | SET NULL |

### record_images

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| medical_record_id | medical_records.id | CASCADE |
| exam_id | exams.id | SET NULL |
| staff_id | staffs.id | SET NULL |

### inquiries

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| medical_record_id | medical_records.id | CASCADE |
| chief_complaint_category_id | chief_complaint_categories.id | SET NULL |
| staff_id | staffs.id | SET NULL |

### billing_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| billing_id | billings.id | CASCADE |

### care_log_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| daily_record_id | daily_records.id | CASCADE |
| staff_id | staffs.id | SET NULL |

### care_plan_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| hospitalization_id | hospitalizations.id | CASCADE |
| hospitalization_plan_id | hospitalization_plans.id | SET NULL |
| medicine_id | medicines.id | SET NULL |
| procedure_id | procedures.id | SET NULL |

### checkups

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| checkup_type_id | checkup_types.id | RESTRICT |
| clinic_id | clinics.id | RESTRICT |
| doctor_id | staffs.id | SET NULL |
| medical_record_id | medical_records.id | CASCADE |
| pet_id | pets.id | RESTRICT |

### daily_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| hospitalization_id | hospitalizations.id | CASCADE |

### diagnosis_names

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| diagnosis_category_id | diagnosis_categories.id | CASCADE |

### exam_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| exam_id | exams.id | CASCADE |
| exam_type_item_id | exam_type_items.id | SET NULL |

### exams

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| doctor_id | staffs.id | SET NULL |
| exam_type_id | exam_types.id | RESTRICT |
| medical_record_id | medical_records.id | CASCADE |
| pet_id | pets.id | RESTRICT |

### exam_type_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| exam_type_id | exam_types.id | CASCADE |

### hospitalizations

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| cage_id | cages.id | SET NULL |
| doctor_id | staffs.id | SET NULL |
| owner_id | owners.id | RESTRICT |
| pet_id | pets.id | RESTRICT |

### medical_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| doctor_id | staffs.id | SET NULL |
| owner_id | owners.id | RESTRICT |
| pet_id | pets.id | RESTRICT |
| reservation_appointment_id | reservation_appointments.id | SET NULL |

### clinical_plans

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| medical_record_id | medical_records.id | CASCADE |
| diagnosis_category_id | diagnosis_categories.id | SET NULL |
| diagnosis_name_id | diagnosis_names.id | SET NULL |

### medicines

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| parent_id | medicines.id | SET NULL |
| inventory_id | inventory_items.id | SET NULL |

### payments

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| billing_id | billings.id | CASCADE |

### pets

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| animal_species_id | animal_species.id | RESTRICT |
| clinic_id | clinics.id | RESTRICT |
| insurance_id | insurances.id | SET NULL |
| owner_id | owners.id | RESTRICT |

### reservation_appointments

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| owner_id | owners.id | SET NULL |
| doctor_id | staffs.id | SET NULL |
| pet_id | pets.id | SET NULL |
| service_type_id | service_types.id | RESTRICT |

### shift_entries

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| staff_id | staffs.id | RESTRICT |

### staff_note_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| daily_record_id | daily_records.id | CASCADE |
| staff_id | staffs.id | SET NULL |

### treatments

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| consultation_id | consultations.id | SET NULL |
| inventory_id | inventory_items.id | SET NULL |
| medical_record_id | medical_records.id | CASCADE |
| medicine_id | medicines.id | SET NULL |
| procedure_id | procedures.id | SET NULL |

### treatment_plans

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| medical_record_id | medical_records.id | CASCADE |
| hospitalization_id | hospitalizations.id | CASCADE |

### trimming_record_options

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| option_id | trimming_options.id | RESTRICT |
| trimming_record_id | trimming_records.id | CASCADE |

### trimming_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| course_id | trimming_courses.id | RESTRICT |
| pet_id | pets.id | RESTRICT |
| staff_id | staffs.id | RESTRICT |

### cages

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### checkup_types

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### chief_complaint_categories

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### consultations

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### diagnosis_categories

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### exam_types

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### hospitalization_plans

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### insurances

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### inquiry_templates

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### inventory_items

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### owners

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### procedures

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### service_types

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### staffs

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| job_title_id | job_titles.id | SET NULL |

### job_titles

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### trimming_courses

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### trimming_options

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |

### vaccines

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| inventory_id | inventory_items.id | SET NULL |

### user_accounts

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| staff_id | staffs.id | SET NULL |
| job_title_id | job_titles.id | SET NULL |

### user_clinic_memberships

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| user_id | user_accounts.id | CASCADE |

### user_permissions

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| granted_by | user_accounts.id | SET NULL |
| user_id | user_accounts.id | CASCADE |

### vaccinations

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| clinic_id | clinics.id | RESTRICT |
| doctor_id | staffs.id | SET NULL |
| medical_record_id | medical_records.id | CASCADE |
| pet_id | pets.id | RESTRICT |
| vaccine_id | vaccines.id | RESTRICT |

### vital_records

| FK元カラム | 参照先 | 削除時 |
| ----------- | ------- | -------- |
| pet_id | pets.id | CASCADE |
| medical_record_id | medical_records.id | CASCADE |
| daily_record_id | daily_records.id | CASCADE |
| staff_id | staffs.id | SET NULL |

---

## インデックス一覧

### UNIQUE制約・インデックス

| テーブル | カラム | 備考 |
| --------- | ------- | ------ |
| medical_records | (clinic_id, record_no) | カルテ番号の一意性（医院スコープ） |
| clinical_plans | medical_record_id | 1カルテ1診察記録（1:1保証） |
| inquiries | medical_record_id | 1カルテ1問診（1:1保証） |
| billing_reviews | medical_record_id | 1カルテ1医師確認（1:1保証） |
| estimates | (clinic_id, estimate_no) | 見積書番号の一意性（医院スコープ） |
| billings | medical_record_id | 1カルテ1会計（1対1） |
| payments | billing_id | 1会計1支払情報（1対1） |
| user_accounts | email | メールアドレスの一意性 |
| daily_records | (hospitalization_id, date) | 1入院1日1件 |
| shift_entries | (staff_id, date) | 1スタッフ1日1シフト |
| user_clinic_memberships | (user_id, clinic_id) | 重複所属防止 |
| user_clinic_memberships | (user_id) WHERE is_main = true | 主所属医院は1件のみ（部分インデックス） |
| trimming_record_options | (trimming_record_id, option_id) | 重複オプション防止 |

### v7.0 追加テーブルのインデックス

| テーブル | カラム | 備考 |
| --------- | ------- | ------ |
| inquiries | (medical_record_id) UNIQUE | 1:1保証 |
| record_images | (medical_record_id) | カルテ別画像検索 |
| record_images | (image_type) | 種別フィルタ |
| record_images | (taken_at DESC) | 撮影日時ソート |
| record_images | (exam_id) WHERE NOT NULL | 検査別画像検索 |
| estimates | (clinic_id, estimate_no) UNIQUE | 見積書番号の一意性（医院スコープ） |
| estimates | (medical_record_id) | カルテ別見積書検索 |
| estimates | (status) | ステータスフィルタ |
| estimates | (owner_id) | 飼主別見積書検索 |
| estimate_items | (estimate_id) | 見積書明細検索 |
| billing_reviews | (medical_record_id) UNIQUE | 1:1保証 |
| billing_reviews | (status) | ステータスフィルタ |

### 主要 FK 列のインデックス（v18.0 W-14）

```sql
-- medical_records 子テーブル FK インデックス
CREATE INDEX idx_treatments_medical_record_id ON treatments(medical_record_id);
CREATE INDEX idx_vital_records_pet_id ON vital_records(pet_id);
CREATE INDEX idx_vital_records_medical_record_id ON vital_records(medical_record_id);
CREATE INDEX idx_vital_records_daily_record_id ON vital_records(daily_record_id);
CREATE INDEX idx_exams_medical_record_id ON exams(medical_record_id);
CREATE INDEX idx_exams_pet_id ON exams(pet_id);
CREATE INDEX idx_vaccinations_medical_record_id ON vaccinations(medical_record_id);
CREATE INDEX idx_vaccinations_pet_id ON vaccinations(pet_id);
CREATE INDEX idx_checkups_medical_record_id ON checkups(medical_record_id);
CREATE INDEX idx_checkups_pet_id ON checkups(pet_id);
CREATE INDEX idx_clinical_plans_medical_record_id ON clinical_plans(medical_record_id);
CREATE INDEX idx_inquiries_medical_record_id ON inquiries(medical_record_id);
CREATE INDEX idx_record_images_medical_record_id ON record_images(medical_record_id);
CREATE INDEX idx_treatment_plans_medical_record_id ON treatment_plans(medical_record_id);
CREATE INDEX idx_treatment_plans_hospitalization_id ON treatment_plans(hospitalization_id);

-- hospitalization 子テーブル FK インデックス
CREATE INDEX idx_hospitalizations_pet_id ON hospitalizations(pet_id);
CREATE INDEX idx_hospitalizations_owner_id ON hospitalizations(owner_id);
CREATE INDEX idx_hospitalizations_cage_id ON hospitalizations(cage_id);
CREATE INDEX idx_care_plan_items_hospitalization_id ON care_plan_items(hospitalization_id);
CREATE INDEX idx_daily_records_hospitalization_id ON daily_records(hospitalization_id);

-- billing 子テーブル FK インデックス
CREATE INDEX idx_billing_items_billing_id ON billing_items(billing_id);
CREATE INDEX idx_billings_pet_id ON billings(pet_id);
CREATE INDEX idx_billings_owner_id ON billings(owner_id);
CREATE INDEX idx_billings_medical_record_id ON billings(medical_record_id);

-- reservation FK インデックス
CREATE INDEX idx_reservation_appointments_pet_id ON reservation_appointments(pet_id);
CREATE INDEX idx_reservation_appointments_service_type_id ON reservation_appointments(service_type_id);
CREATE INDEX idx_reservation_appointments_doctor_id ON reservation_appointments(doctor_id);

-- 担当医 FK インデックス（staffs）
CREATE INDEX idx_treatments_doctor_id ON treatments(doctor_id);
CREATE INDEX idx_vital_records_staff_id ON vital_records(staff_id);
CREATE INDEX idx_trimming_records_staff_id ON trimming_records(staff_id);
```

### 全文検索インデックス（pg_trgm）

```sql
-- 全文検索インデックス（中間一致検索対応）
-- 事前に: CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_owners_name_trgm ON owners USING gin (owner_name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_name_kana_trgm ON owners USING gin (owner_name_kana gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_pets_name_trgm ON pets USING gin (name gin_trgm_ops) WHERE deleted_at IS NULL;
-- record_no は UNIQUE インデックス (clinic_id, record_no) で前方一致検索に対応済み
```

### パフォーマンス最適化インデックス（論理削除考慮）

```sql
-- ダッシュボード・カレンダー（最高頻度）
CREATE INDEX idx_reservation_appointments_clinic_date
  ON reservation_appointments(clinic_id, start_time)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_reservation_appointments_clinic_status
  ON reservation_appointments(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_reservation_appointments_pet_date
  ON reservation_appointments(pet_id, start_time)
  WHERE deleted_at IS NULL;

-- カルテ一覧・検索
CREATE INDEX idx_medical_records_clinic_date
  ON medical_records(clinic_id, date DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_pet
  ON medical_records(clinic_id, pet_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_owner
  ON medical_records(clinic_id, owner_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_medical_records_clinic_status
  ON medical_records(clinic_id, status)
  WHERE deleted_at IS NULL;

-- ペット一覧（飼主別）
CREATE INDEX idx_pets_owner_id
  ON pets(owner_id)
  WHERE deleted_at IS NULL;

-- 会計一覧
CREATE INDEX idx_billings_clinic_date
  ON billings(clinic_id, scheduled_date)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_billings_clinic_status
  ON billings(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_billings_has_insurance
  ON billings(clinic_id, has_insurance)
  WHERE deleted_at IS NULL;

-- 入院管理
CREATE INDEX idx_hospitalizations_clinic_status
  ON hospitalizations(clinic_id, status)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_hospitalizations_clinic_doctor
  ON hospitalizations(clinic_id, doctor_id)
  WHERE deleted_at IS NULL;

-- トリミング一覧
CREATE INDEX idx_trimming_records_clinic_date
  ON trimming_records(clinic_id, date DESC)
  WHERE deleted_at IS NULL;
```

---

## カルテ詳細 API 設計方針（lazy load）

カルテ詳細画面は11テーブル超のJOINが発生するため、タブ単位のlazy load方式を採用する。

| API エンドポイント | 取得テーブル | タイミング |
|------------------|------------|-----------|
| GET /medical-records/:id | medical_records + inquiries + clinical_plans | 初期表示（必須） |
| GET /medical-records/:id/treatments | treatment_plans + treatments | Tab2/3 開時 |
| GET /medical-records/:id/vitals | vital_records | Tab1 詳細展開時 |
| GET /medical-records/:id/exams | exams + exam_items | Tab4 開時 |
| GET /medical-records/:id/billing | estimates + billing_reviews | Tab7/8 開時 |

---

## 未対応事項・今後の予定

### clinic_id 追加（003_add_clinic_id.sql 実施済み・v10.0）

v10.0 にて24テーブル、v19.0 にて estimates を追加し計25テーブルへの `clinic_id` 追加完了。医院ごとにデータを独立管理できる。

| テーブル | 用途 |
| --------- | ------ |
| owners | 医院別の飼い主管理 |
| pets | 医院別のペット管理 |
| medical_records | 医院別のカルテ管理 |
| reservation_appointments | 医院別の予約管理 |
| hospitalizations | 医院別の入院管理 |
| trimming_records | 医院別のトリミング記録 |
| shift_entries | 医院別のシフト管理 |
| billings | 医院別の会計管理 |
| estimates | 医院別の見積書管理（v19.0追加） |
| staffs | 医院別のスタッフ管理 |
| inventory_items | 医院別の在庫管理 |
| cages | 医院別のケージ管理 |
| service_types | 医院別のサービス種別 |
| consultations | 医院別の診察項目 |
| procedures | 医院別の処置項目 |
| hospitalization_plans | 医院別の入院プラン |
| trimming_courses | 医院別のトリミングコース |
| trimming_options | 医院別のトリミングオプション |
| exam_types | 医院別の検査種別 |
| vaccines | 医院別のワクチン |
| medicines | 医院別の薬剤 |
| insurances | 医院別の保険 |
| diagnosis_categories | 医院別の診断カテゴリ |
| diagnosis_names | 医院別の診断名 |
| checkup_types | 医院別の健診種別 |

**clinic_id を保持するテーブル（全体）:** `user_clinic_memberships`, `user_permissions` + 上記25テーブル（計27テーブル）。

### chief_complaint 移行（004_migrate_chief_complaint.sql 予定）

`medical_records.chief_complaint` を `inquiries.chief_complaint` に移行するマイグレーション。

| 作業 | 内容 |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 004_migrate_chief_complaint.sql | `inquiries` レコードを既存 `medical_records` 分だけ生成し、`chief_complaint` の値をコピー後、`medical_records.chief_complaint` カラムを削除 |

> v7.0 時点では `medical_records.chief_complaint` は削除済み。新規カルテ作成時は `inquiries` に書き込むこと。
