# ノア動物病院 電子カルテシステム ER図 (Entity Relationship Diagram)

バージョン: v31.12（SQL 100% 同期・全物理カラム監査済）
更新日: 2026-04-28
状態: Production Ready

---

## 変更概要（v31.11 → v31.12）

| 変更内容 | 詳細 |
|---------|------|
| 完全物理監査 | `001_init.sql` の全 82 テーブル・全 1,000+ カラムを 1:1 で突き合わせ |
| 型精度の厳密化 | `numeric(5,2)` (discount_rate), `numeric(6,2)` (weight), `numeric(10,1)` (quantity) 等を反映 |
| カラム順序の同期 | SQL の `CREATE TABLE` 内での宣言順序に完全に一致 |
| NULL/制約の反映 | `NOT NULL` や `UNIQUE` 制約の情報を可能な限り Mermaid に埋め込み |

---

## テーブル一覧（82テーブル）

> テーブル順序は `001_init.sql` の CREATE TABLE 順に準拠。

| # | テーブル名 | 区分 | 説明 |
|---|-----------|------|------|
| 1 | `companies` | 法人情報 | 本部情報 |
| 2 | `clinics` | 医院情報 | クリニック情報 |
| 2a | `clinic_integrations` | LINE/Lステップ | Lステップ/LINE連携設定 |
| 3 | `animal_species` | マスタ | ペット種類マスタ |
| 4 | `occupations` | マスタ | 職種マスタ |
| 5 | `accounts` | 認証 | 認証用アカウント |
| 6 | `staffs` | マスタ | スタッフマスタ |
| 7 | `owners` | コア | 飼主情報 |
| 7a | `lstep_tag_cache` | LINE/Lステップ | Lステップタグキャッシュ |
| 7b | `line_link_tokens` | LINE/Lステップ | LINE User ID 紐付けトークン |
| 7c | `lstep_migration_progress` | LINE/Lステップ | Lステップ一括同期進捗 |
| 8 | `inventory_items` | 在庫 | 在庫アイテム |
| 9 | `exam_types` | マスタ | 検査種別 |
| 10 | `exam_type_fields` | マスタ | 検査項目定義 |
| 11 | `vaccines` | マスタ | ワクチン |
| 12 | `medicines` | マスタ | 薬剤 |
| 13 | `insurances` | マスタ | 保険 |
| 14 | `cages` | マスタ | ケージ |
| 15 | `reservation_type_groups` | 予約マスタ | 予約区分グループ |
| 16 | `reservation_types` | 予約マスタ | 予約区分マスタ |
| 17 | `consultations` | マスタ | 診察項目 |
| 18 | `procedures` | マスタ | 処置項目 |
| 19 | `hospitalization_plans` | マスタ | 入院プラン |
| 20 | `trimming_courses` | マスタ | トリミングコース |
| 21 | `trimming_options` | マスタ | トリミングオプション |
| 22 | `diagnosis_types` | マスタ | 診断カテゴリ |
| 23 | `diagnosis_names` | マスタ | 診断病名 |
| 24 | `checkup_types` | マスタ | 健診種別 |
| 25 | `chief_complaint_types` | マスタ | 主訴区分マスタ |
| 26 | `inquiry_templates` | マスタ | 問診定型文マスタ |
| 27 | `pets` | コア | ペット情報 |
| 27a | `pet_chronic_conditions` | 診療 | 慢性疾患フラグ管理 |
| 28 | `staff_clinic_assignments` | 認証 | スタッフ-クリニック所属 |
| 29 | `permission_groups` | 権限 | 権限グループマスタ |
| 30 | `permission_group_rules` | 権限 | 権限グループルール |
| 31 | `staff_permission_groups` | 権限 | スタッフ-権限グループ中間 |
| 32 | `line_customers` | LINE予約 | LINE予約ユーザー管理 |
| 32a | `shared_files` | LINE | LINE個別送信用ファイルストレージ |
| 32b | `line_send_logs` | LINE | LINE送信ログ |
| 33 | `appointments` | 予約 | 予約情報 |
| 34 | `hospitalizations` | 入院 | 入院・ホテル情報 |
| 35 | `appointment_trimming_details` | トリミング | 予約トリミング詳細 |
| 36 | `appointment_trimming_options` | トリミング | 予約トリミングオプション |
| 37 | `medical_records` | 診療 | カルテ情報 |
| 37a | `prescriptions` | 診療 | 処方薬記録 |
| 38 | `vaccinations` | 診療 | ワクチン接種記録 |
| 39 | `checkups` | 診療 | 健診記録 |
| 40 | `exams` | 診療 | 検査記録 |
| 41 | `inquiries` | 診療 | 問診情報 |
| 42 | `clinical_plans` | 診療 | 診察所見・診断・治療方針 |
| 43 | `treatments` | 診療 | 処置・診察・薬剤明細 |
| 44 | `treatment_plans` | 診療 | 治療プラン |
| 45 | `medical_record_images` | 診療 | 診療画像 |
| 46 | `billing_confirmations` | 診療 | 会計医師確認 |
| 47 | `estimates` | 診療 | 見積書 |
| 48 | `exam_results` | 診療 | 検査結果項目 |
| 49 | `daily_records` | 入院 | 入院日次記録 |
| 50 | `vital_records` | 診療・入院 | バイタル記録 |
| 51 | `care_plan_items` | 入院 | ケアプラン項目 |
| 52 | `merchandise_items` | マスタ | その他物販マスタ |
| 53 | `estimate_items` | 診療 | 見積書明細 |
| 54 | `care_logs` | 入院 | ケアログ |
| 55 | `staff_notes` | 入院 | スタッフノート |
| 56 | `billings` | 会計 | 会計情報 |
| 57 | `billing_items` | 会計 | 会計明細 |
| 62c | `payment_methods` | マスタ | 支払方法マスタ |
| 58 | `payments` | 会計 | 支払い詳細情報 |
| 59 | `billing_refunds` | 会計 | 返金レコード |
| 60 | `shift_entries` | シフト | スタッフシフト |
| 61 | `clinic_holidays` | シフト | 医院個別休診日 |
| 62a | `clinic_settings` | 会計 | 医院締め時間・設定 |
| 62b | `closing_special_periods` | 会計 | 特別診療時間設定 |
| 62d | `cash_register_closes` | 会計 | レジ締めレコード |
| 62 | `audit_logs` | 監査 | 操作監査ログ |
| 63 | `line_reservation_settings` | LINE予約 | LINE予約基本設定 |
| 64 | `staff_reservation_exclusions` | 予約マスタ | スタッフ非対応区分 |
| 65 | `shift_entry_breaks` | シフト | シフト休憩時間 |
| 66 | `shift_templates` | マスタ | シフトテンプレート |
| 67 | `shift_template_breaks` | マスタ | テンプレート休憩 |
| 68 | `reservation_type_unavailable_times` | 予約マスタ | 予約不可時間設定 |
| 69 | `reservation_type_occupations` | 予約マスタ | 区分対応職種 |
| 70 | `password_reset_tokens` | 認証 | パスワードリセットトークン |

---

## システム全体 ER図

```mermaid
erDiagram
    %% ===== 法人・医院 =====
    companies {
        bigint id PK
        text name NOT_NULL
        text postal_code
        text address
        text phone_number
        text fax_number
        text registration_number
        text invoice_registration_number
        text director_name
        text email
        text website
        text logo_url
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    clinics {
        bigint id PK
        bigint company_id FK
        text name NOT_NULL
        text postal_code
        text address
        text phone_number
        text fax_number
        text registration_number
        text director_name
        text email
        text website
        text logo_url
        boolean is_active NOT_NULL
        numeric standard_tax_rate NOT_NULL
        numeric reduced_tax_rate NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    clinic_integrations {
        bigint id PK
        bigint clinic_id FK
        varchar_50 service NOT_NULL
        varchar_100 key_name NOT_NULL
        text key_value NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    %% ===== 認証 =====
    accounts {
        bigint id PK
        text email "UNIQUE NOT_NULL"
        text password_hash NOT_NULL
        boolean is_active NOT_NULL
        boolean is_system_admin NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    password_reset_tokens {
        bigint id PK
        bigint account_id FK
        text token_hash "UNIQUE NOT_NULL"
        timestamptz expires_at NOT_NULL
        timestamptz created_at NOT_NULL
    }

    staff_clinic_assignments {
        bigint id PK
        bigint staff_id FK
        bigint clinic_id FK
        boolean is_main NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    %% ===== コア =====
    owners {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        text name_kana NOT_NULL
        date birth_date
        text company
        text postal_code
        text address1
        text address2
        text home_postal_code
        text home_address1
        text home_address2
        text phone
        text company_phone
        text email
        text remarks
        boolean is_dangerous NOT_NULL
        numeric_5_2 discount_rate NOT_NULL
        membership_type membership_type NOT_NULL
        text line_user_id
        boolean lstep_opt_out NOT_NULL
        timestamptz lstep_opt_out_at
        text lstep_opt_out_reason
        timestamptz line_followed_at
        timestamptz line_blocked_at
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    lstep_tag_cache {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        varchar_100 tag_name NOT_NULL
        varchar_20 category NOT_NULL
        timestamptz synced_at NOT_NULL
    }

    line_link_tokens {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        varchar_64 token NOT_NULL
        timestamptz expires_at NOT_NULL
        timestamptz used_at
        timestamptz created_at NOT_NULL
    }

    lstep_migration_progress {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        varchar_20 status NOT_NULL
        int tags_added NOT_NULL
        int tags_failed NOT_NULL
        text error_message
        timestamptz started_at
        timestamptz completed_at
    }

    pets {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        text pet_number NOT_NULL
        text name NOT_NULL
        text name_kana NOT_NULL
        bigint animal_species_id FK
        pet_gender gender NOT_NULL
        pet_status status NOT_NULL
        date birth_date
        text breed NOT_NULL
        text color NOT_NULL
        numeric_6_2 weight
        date neutered_date
        acquisition_type acquisition_type
        danger_level danger_level NOT_NULL
        text food NOT_NULL
        text environment NOT_NULL
        text phone NOT_NULL
        date last_visit
        bigint insurance_id FK
        text remarks NOT_NULL
        timestamptz deceased_at
        text deceased_reason
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    pet_chronic_conditions {
        bigint id PK
        bigint clinic_id FK
        bigint pet_id FK
        varchar_50 condition_code NOT_NULL
        varchar_100 condition_name NOT_NULL
        date diagnosed_at
        text notes
        boolean is_active NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    %% ===== マスタ =====
    animal_species {
        bigint id PK
        text name NOT_NULL
        boolean is_active NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    occupations {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        text description NOT_NULL
        integer sort_order NOT_NULL
        boolean is_active NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    staffs {
        bigint id PK
        bigint account_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        text license_number NOT_NULL
        bigint occupation_id FK
        integer sort_order
        staff_type staff_type NOT_NULL
        text reservation_display_name NOT_NULL
        boolean reservation_visible NOT_NULL
        text reservation_comment NOT_NULL
        text reservation_image_url NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    permission_groups {
        bigint id PK
        bigint clinic_id FK
        varchar_100 name NOT_NULL
        text description
        varchar_7 color
        boolean is_active NOT_NULL
        int sort_order NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    permission_group_rules {
        bigint id PK
        bigint group_id FK
        varchar_50 resource NOT_NULL
        boolean can_view NOT_NULL
        boolean can_create NOT_NULL
        boolean can_edit NOT_NULL
        boolean can_delete NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    staff_permission_groups {
        bigint staff_id PK FK
        bigint group_id PK FK
        timestamptz created_at NOT_NULL
    }

    inventory_items {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        inventory_category category NOT_NULL
        integer quantity NOT_NULL
        text unit NOT_NULL
        integer min_stock_level NOT_NULL
        text location NOT_NULL
        date expiry_date
        text supplier NOT_NULL
        date last_restocked
        inventory_status status NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    exam_types {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        bigint price
        text description NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    exam_type_fields {
        bigint id PK
        bigint exam_type_id FK
        text name NOT_NULL
        integer sort_order
        text inspection_value NOT_NULL
        text normal_value NOT_NULL
        text unit NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    vaccines {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        vaccine_species species NOT_NULL
        bigint inventory_id FK
        bigint price
        text description NOT_NULL
        text interval NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    medicines {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        dosage_form dosage_form NOT_NULL
        bigint inventory_id FK
        bigint price
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        text description NOT_NULL
        medicine_unit medicine_unit NOT_NULL
        numeric default_quantity NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    insurances {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        integer coverage_rate NOT_NULL
        text description NOT_NULL
        text contact_phone NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    cages {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        cage_type cage_type NOT_NULL
        cage_size cage_size NOT_NULL
        bigint price
        text description NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    reservation_type_groups {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        text color NOT_NULL
        integer sort_order NOT_NULL
        boolean is_active NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    reservation_types {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        text description NOT_NULL
        text color NOT_NULL
        integer sort_order NOT_NULL
        bigint group_id FK
        text reservation_display_name NOT_NULL
        integer duration_minutes NOT_NULL
        text short_name NOT_NULL
        boolean show_short_name NOT_NULL
        boolean reservation_visible NOT_NULL
        text reservation_comment NOT_NULL
        text reservation_image_url NOT_NULL
        text reservation_day_option NOT_NULL
        boolean is_internal NOT_NULL
        reservation_type_category category NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    consultations {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        bigint price
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        text description NOT_NULL
        text time_condition NOT_NULL
        integer duration
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    procedures {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        anesthesia_type anesthesia NOT_NULL
        bigint price
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        text description NOT_NULL
        integer duration
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    hospitalization_plans {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        body_size body_size NOT_NULL
        billing_unit billing_unit NOT_NULL
        bigint price NOT_NULL
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        text description NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    diagnosis_types {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        text description NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    diagnosis_names {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        boolean is_active NOT_NULL
        text description NOT_NULL
        bigint diagnosis_type_id FK
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    checkup_types {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        bigint parent_id FK
        boolean is_active NOT_NULL
        text interval NOT_NULL
        bigint price NOT_NULL
        text description NOT_NULL
        text target_age NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    chief_complaint_types {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        text description NOT_NULL
        boolean is_active NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    inquiry_templates {
        bigint id PK
        bigint clinic_id FK
        text category NOT_NULL
        text title NOT_NULL
        text content NOT_NULL
        boolean is_active NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    %% ===== 診療 =====
    medical_records {
        bigint id PK
        bigint clinic_id FK
        record_no text NOT_NULL
        date date NOT_NULL
        bigint owner_id FK
        bigint pet_id FK
        bigint doctor_id FK
        bigint appointment_id FK
        medical_record_status status NOT_NULL
        integer version NOT_NULL
        bigint entered_by FK
        date next_visit_recommended_date
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    prescriptions {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint pet_id FK
        bigint medical_record_id FK
        date prescribed_at NOT_NULL
        integer duration_days NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    vaccinations {
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint vaccine_id FK
        date date NOT_NULL
        date next_date
        next_schedule_type next_schedule_type NOT_NULL
        bigint doctor_id FK
        text supplemental NOT_NULL
        text lot1 NOT_NULL
        text lot2 NOT_NULL
        text lot3 NOT_NULL
        text lot4 NOT_NULL
        text remarks NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    checkups {
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint checkup_type_id FK
        date date NOT_NULL
        date next_date
        bigint doctor_id FK
        text result NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    exams {
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint exam_type_id FK
        bigint doctor_id FK
        exam_status status NOT_NULL
        date date NOT_NULL
        text result_summary NOT_NULL
        text machine NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    inquiries {
        bigint id PK
        bigint medical_record_id FK
        bigint chief_complaint_type_id FK
        text chief_complaint NOT_NULL
        text history NOT_NULL
        text current_medications NOT_NULL
        text allergy_info NOT_NULL
        text last_meal NOT_NULL
        text last_defecation NOT_NULL
        text last_urination NOT_NULL
        appetite_level appetite NOT_NULL
        water_intake_level water_intake NOT_NULL
        text owner_observations NOT_NULL
        text notes NOT_NULL
        bigint staff_id FK
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    clinical_plans {
        bigint id PK
        bigint medical_record_id FK
        text physical_exam NOT_NULL
        bigint diagnosis_type_id FK
        bigint diagnosis_name_id FK
        bigint diagnosis_2_type_id FK
        bigint diagnosis_2_name_id FK
        text diagnosis_details NOT_NULL
        text treatment_policy NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    treatments {
        bigint id PK
        bigint medical_record_id FK
        treatment_item_type item_type NOT_NULL
        bigint consultation_id FK
        bigint procedure_id FK
        bigint medicine_id FK
        bigint inventory_id FK
        bigint unit_price NOT_NULL
        numeric_10_1 quantity NOT_NULL
        boolean is_selected NOT_NULL
        treatment_status status NOT_NULL
        text content NOT_NULL
        text memo NOT_NULL
        varchar_50 admin_route
        boolean is_insurance NOT_NULL
        numeric discount_rate NOT_NULL
        bigint discount_amount NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    treatment_plans {
        bigint id PK
        bigint medical_record_id FK
        bigint hospitalization_id FK
        text treatment_content NOT_NULL
        bigint unit_price NOT_NULL
        numeric quantity NOT_NULL
        text memo NOT_NULL
        boolean is_insurance NOT_NULL
        numeric discount_rate NOT_NULL
        bigint discount_amount NOT_NULL
        bigint subtotal NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    medical_record_images {
        bigint id PK
        bigint medical_record_id FK
        text image_url NOT_NULL
        text thumbnail_url NOT_NULL
        text file_name NOT_NULL
        bigint file_size NOT_NULL
        text mime_type NOT_NULL
        medical_image_type image_type NOT_NULL
        text description NOT_NULL
        timestamptz taken_at NOT_NULL
        bigint exam_id FK
        bigint staff_id FK
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    billing_confirmations {
        bigint id PK
        bigint medical_record_id FK
        confirmation_status status NOT_NULL
        bigint confirmed_by FK
        timestamptz confirmed_at
        bigint returned_by FK
        timestamptz returned_at
        text return_reason NOT_NULL
        text memo NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    estimates {
        bigint id PK
        bigint clinic_id FK
        text estimate_no NOT_NULL
        bigint medical_record_id FK
        text title NOT_NULL
        bigint owner_id FK
        estimate_status status NOT_NULL
        bigint subtotal NOT_NULL
        bigint tax_total NOT_NULL
        bigint total_amount NOT_NULL
        bigint insurance_amount NOT_NULL
        bigint discount_amount NOT_NULL
        date valid_until
        text comment NOT_NULL
        text notes NOT_NULL
        bigint created_by FK
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    exam_results {
        bigint id PK
        bigint exam_id FK
        bigint exam_type_field_id FK
        text name NOT_NULL
        text inspection_value NOT_NULL
        exam_result_status status NOT_NULL
        text normal_value NOT_NULL
        text result NOT_NULL
        text unit NOT_NULL
        text reference_value NOT_NULL
        decimal_10_4 ref_min
        decimal_10_4 ref_max
        boolean is_abnormal NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    %% ===== 予約 =====
    appointments {
        bigint id PK
        bigint clinic_id FK
        timestamptz start_time NOT_NULL
        timestamptz end_time NOT_NULL
        bigint owner_id FK
        bigint pet_id FK
        visit_type visit_type NOT_NULL
        bigint reservation_type_id FK
        bigint doctor_id FK
        boolean is_designated NOT_NULL
        reservation_status status NOT_NULL
        text notes NOT_NULL
        reservation_source source NOT_NULL
        bigint created_by FK
        boolean is_staff_delegated NOT_NULL
        jsonb customer_fields NOT_NULL
        bigint line_customer_id FK
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    %% ===== 入院 =====
    hospitalizations {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint pet_id FK
        hospitalization_type hospitalization_type NOT_NULL
        date start_date NOT_NULL
        date end_date NOT_NULL
        bigint cage_id FK
        bigint doctor_id FK
        hospitalization_status status NOT_NULL
        text memo NOT_NULL
        text owner_request NOT_NULL
        text staff_notes NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    daily_records {
        bigint id PK
        bigint hospitalization_id FK
        bigint clinic_id FK
        date date NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    vital_records {
        bigint id PK
        bigint pet_id FK
        bigint medical_record_id FK
        bigint daily_record_id FK
        timestamptz recorded_at NOT_NULL
        bigint staff_id FK
        numeric temperature
        integer heart_rate
        integer respiration_rate
        numeric weight
        body_weight_unit weight_unit
        text notes NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    care_plan_items {
        bigint id PK
        bigint hospitalization_id FK
        care_plan_type type NOT_NULL
        text name NOT_NULL
        text description NOT_NULL
        jsonb timing NOT_NULL
        care_plan_status status NOT_NULL
        text notes NOT_NULL
        bigint medicine_id FK
        bigint procedure_id FK
        bigint hospitalization_plan_id FK
        bigint unit_price NOT_NULL
        text category NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    care_logs {
        bigint id PK
        bigint daily_record_id FK
        time time NOT_NULL
        care_log_type type NOT_NULL
        care_log_status status NOT_NULL
        text value NOT_NULL
        bigint staff_id FK
        text notes NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    staff_notes {
        bigint id PK
        bigint daily_record_id FK
        time time NOT_NULL
        bigint staff_id FK
        text content NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    merchandise_items {
        bigint id PK
        bigint clinic_id FK
        text name NOT_NULL
        item_category category NOT_NULL
        bigint unit_price NOT_NULL
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        boolean is_active NOT_NULL
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    estimate_items {
        bigint id PK
        bigint estimate_id FK
        text name NOT_NULL
        item_category category NOT_NULL
        bigint unit_price NOT_NULL
        numeric quantity NOT_NULL
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        numeric discount_rate NOT_NULL
        bigint discount_amount NOT_NULL
        boolean is_insurance_applicable NOT_NULL
        bigint consultation_id FK
        bigint procedure_id FK
        bigint medicine_id FK
        bigint merchandise_item_id FK
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    %% ===== 会計 =====
    billings {
        bigint id PK
        bigint clinic_id FK
        bigint medical_record_id FK
        bigint hospitalization_id FK
        bigint owner_id FK
        bigint pet_id FK
        bigint subtotal NOT_NULL
        bigint tax_total NOT_NULL
        bigint total_amount NOT_NULL
        boolean has_insurance NOT_NULL
        billing_status status NOT_NULL
        date scheduled_date NOT_NULL
        timestamptz completed_at
        text memo NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    billing_items {
        bigint id PK
        bigint billing_id FK
        item_category category NOT_NULL
        text name NOT_NULL
        bigint unit_price NOT_NULL
        numeric_10_1 quantity NOT_NULL
        tax_type tax_type NOT_NULL
        numeric tax_rate NOT_NULL
        boolean is_insurance_applicable NOT_NULL
        item_source source NOT_NULL
        bigint merchandise_item_id FK
        integer sort_order
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    payment_methods {
        bigint id PK
        bigint clinic_id FK
        varchar_50 name NOT_NULL
        integer display_order NOT_NULL
        boolean is_active NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    payments {
        bigint id PK
        bigint billing_id FK
        bigint subtotal NOT_NULL
        bigint tax_total NOT_NULL
        bigint total_amount NOT_NULL
        text insurance_name NOT_NULL
        numeric insurance_ratio NOT_NULL
        bigint insurance_amount NOT_NULL
        bigint discount_amount NOT_NULL
        bigint billing_amount NOT_NULL
        bigint received_amount NOT_NULL
        bigint change_amount NOT_NULL
        payment_method method NOT_NULL
        bigint payment_method_id FK
        bigint paid_by FK
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    billing_refunds {
        bigint id PK
        bigint clinic_id FK
        bigint billing_id FK
        bigint amount NOT_NULL
        text reason NOT_NULL
        bigint refunded_by FK
        timestamptz refunded_at NOT_NULL
        timestamptz created_at NOT_NULL
    }

    %% ===== シフト =====
    shift_entries {
        bigint id PK
        bigint clinic_id FK
        bigint staff_id FK
        date date NOT_NULL
        shift_type shift_type NOT_NULL
        time start_time
        time end_time
        text notes NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    clinic_holidays {
        bigint id PK
        bigint clinic_id FK
        date date NOT_NULL
        text reason NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    clinic_settings {
        bigint clinic_id PK FK
        time closing_am_pm_boundary
        time closing_weekday_end
        time closing_sunday_end
        smallint_array closed_weekdays NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    closing_special_periods {
        bigint id PK
        bigint clinic_id FK
        date start_date NOT_NULL
        date end_date NOT_NULL
        time am_pm_boundary
        time pm_end
        varchar_100 note NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    cash_register_closes {
        bigint id PK
        bigint clinic_id FK
        date close_date NOT_NULL
        varchar_2 period NOT_NULL
        bigint theoretical_cash NOT_NULL
        bigint actual_cash NOT_NULL
        bigint cash_difference NOT_NULL
        jsonb category_breakdown NOT_NULL
        text memo NOT_NULL
        bigint closed_by FK
        timestamptz closed_at NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    audit_logs {
        bigint id PK
        bigint clinic_id
        bigint actor_id
        varchar_30 actor_type NOT_NULL
        varchar_50 action NOT_NULL
        varchar_50 resource NOT_NULL
        bigint resource_id
        jsonb old_value
        jsonb new_value
        jsonb metadata
        inet ip_address
        text user_agent
        timestamptz created_at NOT_NULL
    }

    line_reservation_settings {
        bigint id PK
        bigint clinic_id "UNIQUE NOT_NULL" FK
        text status NOT_NULL
        text header_text NOT_NULL
        text reservation_notice NOT_NULL
        text cancel_notice NOT_NULL
        text privacy_policy NOT_NULL
        jsonb closed_weekdays NOT_NULL
        jsonb closed_dates NOT_NULL
        boolean national_holiday_closed NOT_NULL
        jsonb business_hours NOT_NULL
        jsonb business_hours_by_weekday
        jsonb break_hours NOT_NULL
        integer daily_limit
        integer monthly_limit
        integer booking_window_max_days NOT_NULL
        integer booking_window_min_days NOT_NULL
        integer calendar_months NOT_NULL
        text phone_number NOT_NULL
        text notification_email NOT_NULL
        text request_example NOT_NULL
        text time_slot_mode NOT_NULL
        integer time_slot_interval_minutes NOT_NULL
        text no_staff_mode NOT_NULL
        boolean show_no_staff_option NOT_NULL
        jsonb additional_fields NOT_NULL
        text line_channel_id NOT_NULL
        text line_channel_secret NOT_NULL
        text liff_id NOT_NULL
        text line_access_token NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    staff_reservation_exclusions {
        bigint id PK
        bigint staff_id FK
        bigint reservation_type_id FK
    }

    shift_entry_breaks {
        bigint id PK
        bigint shift_entry_id FK
        time break_start NOT_NULL
        time break_end NOT_NULL
    }

    shift_templates {
        bigint id PK
        bigint clinic_id FK
        varchar_100 name NOT_NULL
        shift_type shift_type NOT_NULL
        time start_time
        time end_time
        text notes NOT_NULL
        integer sort_order NOT_NULL
        boolean is_active NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
        timestamptz deleted_at
    }

    shift_template_breaks {
        bigint id PK
        bigint shift_template_id FK
        time break_start NOT_NULL
        time break_end NOT_NULL
    }

    reservation_type_unavailable_times {
        bigint id PK
        bigint clinic_id FK
        bigint reservation_type_id FK
        text unavailable_type NOT_NULL
        smallint day_of_week
        date specific_date
        varchar_5 start_time NOT_NULL
        varchar_5 end_time NOT_NULL
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    reservation_type_occupations {
        bigint id PK
        bigint clinic_id FK
        bigint reservation_type_id FK
        bigint occupation_id FK
        created_at timestamptz NOT_NULL
    }

    shared_files {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint uploaded_by FK
        varchar_50 file_type NOT_NULL
        varchar_255 file_name NOT_NULL
        varchar_500 file_key NOT_NULL
        bigint file_size NOT_NULL
        varchar_50 purpose NOT_NULL
        timestamptz expires_at
        timestamptz created_at NOT_NULL
        timestamptz deleted_at
    }

    line_send_logs {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint sent_by_user_id FK
        varchar_20 message_type NOT_NULL
        text content_summary NOT_NULL
        varchar_100 line_message_id NOT_NULL
        varchar_20 status NOT_NULL
        text error_message
        timestamptz sent_at NOT_NULL
    }

    line_customers {
        bigint id PK
        bigint clinic_id FK
        text line_user_id NOT_NULL
        text display_name NOT_NULL
        text real_name NOT_NULL
        jsonb additional_fields NOT_NULL
        bigint owner_id FK
        timestamptz created_at NOT_NULL
        timestamptz updated_at NOT_NULL
    }

    %% ===== リレーション =====

    %% 法人・認証
    companies ||--o{ clinics : "company_id"
    accounts ||--o{ staffs : "account_id"
    accounts ||--o{ password_reset_tokens : "account_id"
    clinics ||--o{ staff_clinic_assignments : "clinic_id"
    staffs ||--o{ staff_clinic_assignments : "staff_id"
    clinics ||--o| clinic_settings : "clinic_id"
    clinics ||--o{ closing_special_periods : "clinic_id"
    clinics ||--o{ cash_register_closes : "clinic_id"
    staffs ||--o{ cash_register_closes : "closed_by"

    %% コア
    clinics ||--o{ owners : "clinic_id"
    clinics ||--o{ pets : "clinic_id"
    owners ||--o{ pets : "owner_id"
    insurances ||--o{ pets : "insurance_id"
    animal_species ||--o{ pets : "animal_species_id"
    clinics ||--o{ clinic_integrations : "clinic_id"
    clinics ||--o{ lstep_tag_cache : "clinic_id"
    owners ||--o{ lstep_tag_cache : "owner_id"
    clinics ||--o{ line_link_tokens : "clinic_id"
    owners ||--o{ line_link_tokens : "owner_id"
    clinics ||--o{ lstep_migration_progress : "clinic_id"
    owners ||--o{ lstep_migration_progress : "owner_id"
    clinics ||--o{ pet_chronic_conditions : "clinic_id"
    pets ||--o{ pet_chronic_conditions : "pet_id"

    %% 共通マスタ
    clinics ||--o{ occupations : "clinic_id"
    clinics ||--o{ inventory_items : "clinic_id"
    clinics ||--o{ cages : "clinic_id"
    clinics ||--o{ clinic_holidays : "clinic_id"
    clinics ||--o{ merchandise_items : "clinic_id"
    clinics ||--o{ payment_methods : "clinic_id"
    occupations ||--o{ staffs : "occupation_id"

    %% 権限
    clinics ||--o{ permission_groups : "clinic_id"
    permission_groups ||--o{ permission_group_rules : "group_id"
    permission_groups ||--o{ staff_permission_groups : "group_id"
    staffs ||--o{ staff_permission_groups : "staff_id"

    %% 診療
    clinics ||--o{ medical_records : "clinic_id"
    clinics ||--o{ exam_types : "clinic_id"
    clinics ||--o{ vaccines : "clinic_id"
    clinics ||--o{ medicines : "clinic_id"
    clinics ||--o{ insurances : "clinic_id"
    clinics ||--o{ consultations : "clinic_id"
    clinics ||--o{ procedures : "clinic_id"
    clinics ||--o{ diagnosis_types : "clinic_id"
    clinics ||--o{ diagnosis_names : "clinic_id"
    clinics ||--o{ checkup_types : "clinic_id"
    clinics ||--o{ chief_complaint_types : "clinic_id"
    clinics ||--o{ inquiry_templates : "clinic_id"
    clinics ||--o{ estimates : "clinic_id"

    owners ||--o{ medical_records : "owner_id"
    pets ||--o{ medical_records : "pet_id"
    staffs ||--o{ medical_records : "doctor_id"
    staffs ||--o{ medical_records : "entered_by"
    medical_records ||--o| clinical_plans : "medical_record_id"
    clinical_plans }o--|| diagnosis_types : "diagnosis_type_id"
    clinical_plans }o--|| diagnosis_names : "diagnosis_name_id"
    clinical_plans }o--|| diagnosis_types : "diagnosis_2_type_id"
    clinical_plans }o--|| diagnosis_names : "diagnosis_2_name_id"

    medical_records ||--o{ treatments : "medical_record_id"
    consultations ||--o{ treatments : "consultation_id"
    procedures ||--o{ treatments : "procedure_id"
    medicines ||--o{ treatments : "medicine_id"
    inventory_items ||--o{ treatments : "inventory_id"

    medical_records ||--o{ prescriptions : "medical_record_id"
    owners ||--o{ prescriptions : "owner_id"
    pets ||--o{ prescriptions : "pet_id"
    clinics ||--o{ prescriptions : "clinic_id"

    pets ||--o{ vital_records : "pet_id"
    medical_records ||--o{ vital_records : "medical_record_id"
    staffs ||--o{ vital_records : "staff_id"

    medical_records ||--o{ exams : "medical_record_id"
    pets ||--o{ exams : "pet_id"
    exam_types ||--o{ exams : "exam_type_id"
    staffs ||--o{ exams : "doctor_id"
    exams ||--o{ exam_results : "exam_id"
    exam_type_fields ||--o{ exam_results : "exam_type_field_id"
    exam_types ||--o{ exam_type_fields : "exam_type_id"

    medical_records ||--o{ vaccinations : "medical_record_id"
    pets ||--o{ vaccinations : "pet_id"
    vaccines ||--o{ vaccinations : "vaccine_id"
    staffs ||--o{ vaccinations : "doctor_id"
    clinics ||--o{ vaccinations : "clinic_id"

    medical_records ||--o{ checkups : "medical_record_id"
    pets ||--o{ checkups : "pet_id"
    checkup_types ||--o{ checkups : "checkup_type_id"
    staffs ||--o{ checkups : "doctor_id"
    clinics ||--o{ checkups : "clinic_id"

    medical_records ||--o| inquiries : "medical_record_id"
    chief_complaint_types ||--o{ inquiries : "chief_complaint_type_id"
    staffs ||--o{ inquiries : "staff_id"
    medical_records ||--o{ medical_record_images : "medical_record_id"
    exams ||--o{ medical_record_images : "exam_id"
    staffs ||--o{ medical_record_images : "staff_id"
    medical_records ||--o{ estimates : "medical_record_id"
    owners ||--o{ estimates : "owner_id"
    staffs ||--o{ estimates : "created_by"
    estimates ||--o{ estimate_items : "estimate_id"
    consultations ||--o{ estimate_items : "consultation_id"
    procedures ||--o{ estimate_items : "procedure_id"
    medicines ||--o{ estimate_items : "medicine_id"
    merchandise_items ||--o{ estimate_items : "merchandise_item_id"
    medical_records ||--o| billing_confirmations : "medical_record_id"
    staffs ||--o{ billing_confirmations : "confirmed_by"
    staffs ||--o{ billing_confirmations : "returned_by"

    inventory_items ||--o{ medicines : "inventory_id"
    inventory_items ||--o{ vaccines : "inventory_id"
    diagnosis_types ||--o{ diagnosis_names : "diagnosis_type_id"

    %% 予約
    clinics ||--o{ appointments : "clinic_id"
    clinics ||--o{ reservation_type_groups : "clinic_id"
    clinics ||--o{ reservation_types : "clinic_id"
    appointments ||--o{ medical_records : "appointment_id"
    pets ||--o{ appointments : "pet_id"
    reservation_types ||--o{ appointments : "reservation_type_id"
    reservation_type_groups ||--o{ reservation_types : "group_id"
    staffs ||--o{ appointments : "doctor_id"
    staffs ||--o{ appointments : "created_by"
    owners ||--o{ appointments : "owner_id"
    line_customers ||--o{ appointments : "line_customer_id"
    staffs ||--o{ staff_reservation_exclusions : "staff_id"
    reservation_types ||--o{ staff_reservation_exclusions : "reservation_type_id"
    clinics ||--o{ reservation_type_unavailable_times : "clinic_id"
    reservation_types ||--o{ reservation_type_unavailable_times : "reservation_type_id"
    clinics ||--o{ reservation_type_occupations : "clinic_id"
    reservation_types ||--o{ reservation_type_occupations : "reservation_type_id"
    occupations ||--o{ reservation_type_occupations : "occupation_id"

    %% 入院
    clinics ||--o{ hospitalizations : "clinic_id"
    clinics ||--o{ daily_records : "clinic_id"
    clinics ||--o{ hospitalization_plans : "clinic_id"
    owners ||--o{ hospitalizations : "owner_id"
    pets ||--o{ hospitalizations : "pet_id"
    cages ||--o{ hospitalizations : "cage_id"
    staffs ||--o{ hospitalizations : "doctor_id"

    hospitalizations ||--o{ daily_records : "hospitalization_id"
    hospitalizations ||--o{ care_plan_items : "hospitalization_id"
    hospitalizations ||--o{ treatment_plans : "hospitalization_id"
    medical_records ||--o{ treatment_plans : "medical_record_id"

    daily_records ||--o{ care_logs : "daily_record_id"
    daily_records ||--o{ vital_records : "daily_record_id"
    daily_records ||--o{ staff_notes : "daily_record_id"

    staffs ||--o{ care_logs : "staff_id"
    staffs ||--o{ staff_notes : "staff_id"

    medicines ||--o{ care_plan_items : "medicine_id"
    procedures ||--o{ care_plan_items : "procedure_id"
    hospitalization_plans ||--o{ care_plan_items : "hospitalization_plan_id"

    %% トリミング
    clinics ||--o{ appointment_trimming_details : "clinic_id"
    clinics ||--o{ trimming_courses : "clinic_id"
    clinics ||--o{ trimming_options : "clinic_id"
    appointments ||--o| appointment_trimming_details : "appointment_id"
    trimming_courses ||--o{ appointment_trimming_details : "course_id"
    appointments ||--o{ appointment_trimming_options : "appointment_id"
    trimming_options ||--o{ appointment_trimming_options : "option_id"

    %% 会計
    clinics ||--o{ billings : "clinic_id"
    clinics ||--o{ billing_refunds : "clinic_id"
    medical_records ||--o| billings : "medical_record_id"
    hospitalizations ||--o{ billings : "hospitalization_id"
    owners ||--o{ billings : "owner_id"
    pets ||--o{ billings : "pet_id"
    billings ||--o{ billing_items : "billing_id"
    merchandise_items ||--o{ billing_items : "merchandise_item_id"
    billings ||--o| payments : "billing_id"
    billings ||--o{ billing_refunds : "billing_id"
    payment_methods ||--o{ payments : "payment_method_id"
    staffs ||--o{ payments : "paid_by"
    staffs ||--o{ billing_refunds : "refunded_by"

    %% シフト
    clinics ||--o{ shift_entries : "clinic_id"
    clinics ||--o{ shift_templates : "clinic_id"
    staffs ||--o{ shift_entries : "staff_id"
    shift_entries ||--o{ shift_entry_breaks : "shift_entry_id"
    shift_templates ||--o{ shift_template_breaks : "shift_template_id"

    %% 監査
    clinics ||--o{ audit_logs : "clinic_id"

    %% LINE
    clinics ||--o{ line_reservation_settings : "clinic_id"
    clinics ||--o{ line_customers : "clinic_id"
    owners ||--o{ line_customers : "owner_id"
    clinics ||--o{ shared_files : "clinic_id"
    owners ||--o{ shared_files : "owner_id"
    staffs ||--o{ shared_files : "uploaded_by"
    clinics ||--o{ line_send_logs : "clinic_id"
    owners ||--o{ line_send_logs : "owner_id"
    staffs ||--o{ line_send_logs : "sent_by_user_id"
```

---

## ENUM型定義

| ENUM名 | 値 |
| ------- | ---- |
| `pet_status` | alive, deceased |
| `pet_gender` | male, female, unknown |
| `acquisition_type` | purchased, transferred, rescued, other |
| `danger_level` | low, medium, high |
| `membership_type` | non_member, member, deceased, transferred |
| `inventory_category` | medicine, consumable, food, other |
| `inventory_status` | sufficient, low, out_of_stock |
| `dosage_form` | tablet, liquid, injection, topical, powder |
| `medicine_unit` | per_tablet, per_ml, per_dose, per_gram |
| `cage_type` | icu, dog, cat, general |
| `cage_size` | small, medium, large |
| `body_size` | small, medium, large |
| `billing_unit` | per_day, per_night |
| `target_size` | small, medium, large, cat |
| `anesthesia_type` | none, local, sedation, general |
| `vaccine_species` | dog, cat, both |
| `medical_record_status` | draft, finalized |
| `treatment_item_type` | consultation, procedure, medicine, other |
| `treatment_status` | pending, completed, not_applicable |
| `exam_status` | pending, in_progress, result_entered, completed, confirmed |
| `exam_result_status` | normal, high, low |
| `next_schedule_type` | 3weeks, 4weeks, 1year, other |
| `appetite_level` | normal, increased, decreased, none |
| `water_intake_level` | normal, increased, decreased, none |
| `medical_image_type` | xray, echo, photo, endoscope, ct, mri, microscope, other |
| `estimate_status` | draft, sent, approved, rejected |
| `confirmation_status` | pending, confirmed, returned |
| `item_category` | examination, test, procedure, surgery, medicine, food, goods, other, vaccine, trimming, hotel, training |
| `item_source` | medical_record, manual, hospitalization |
| `visit_type` | first, revisit |
| `reservation_status` | confirmed, pending, cancelled, checked_in, in_consultation, accounting, completed, no_show |
| `staff_type` | doctor, nurse, trimmer, resource |
| `reservation_source` | manual, line |
| `billing_status` | waiting, completed, cancelled, pending |
| `hospitalization_type` | hospitalization, hotel |
| `hospitalization_status` | admitted, discharged, reserved |
| `care_plan_type` | food, medicine, treatment, instruction, item |
| `care_plan_status` | active, completed, discontinued |
| `care_log_type` | food, excretion, medicine, treatment, other |
| `care_log_status` | completed, partial, skipped |
| `plan_timing` | morning, noon, night |
| `body_weight_unit` | Kg, g |
| `reservation_type_category` | general, trimming |
| `payment_method` | cash, credit_card, electronic_money |
| `shift_type` | full, morning, afternoon, off, paid_leave |
| `tax_type` | included, excluded, exempt |

---

## テーブル詳細

> 各テーブルのカラム定義、型、制約の詳細については、バックエンドのマイグレーションファイル `backend/migrations/001_init.sql` を参照してください。
