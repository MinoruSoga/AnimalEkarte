# ノア動物病院 電子カルテシステム ER図 (Entity Relationship Diagram)

バージョン: v31.8（SQL マイグレーション 100% 同期）
更新日: 2026-04-17
状態: Production Ready

---

## 変更概要（v31.7 → v31.8）

| 変更内容 | 詳細 |
|---------|------|
| 定期監査 | 2026-04-17 時点の `001_init.sql` との完全同期を確認 |
| 整合性確認 | 69テーブルすべてのカラム、型、リレーション、ENUM定義の整合性を再検証 |
| 重複整理 | ドキュメント内の記載に重複や矛盾がないことを確認 |

---

## テーブル一覧（69テーブル）

> テーブル順序は `001_init.sql` の CREATE TABLE 順に準拠。

| # | テーブル名 | 区分 | 説明 |
|---|-----------|------|------|
| 1 | `companies` | 法人情報 | 法人（ノア動物病院）情報 |
| 2 | `clinics` | 医院情報 | 各医院（八王子・城東・敷島） |
| 3 | `animal_species` | マスタ | ペット種類マスタ（犬・猫・鳥等） |
| 4 | `occupations` | マスタ | 職種マスタ |
| 5 | `accounts` | 認証 | ログインアカウント |
| 6 | `staffs` | マスタ | スタッフ |
| 7 | `owners` | コア | 飼い主 |
| 8 | `inventory_items` | 在庫 | 在庫アイテム |
| 9 | `exam_types` | マスタ | 検査種別 |
| 10 | `exam_type_fields` | マスタ | 検査種別の検査項目定義 |
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
| 27 | `pets` | コア | ペット |
| 28 | `staff_clinic_assignments` | 認証 | スタッフ-クリニック所属 |
| 29 | `permission_groups` | 権限 | 権限グループマスタ |
| 30 | `permission_group_rules` | 権限 | 権限グループルール |
| 31 | `staff_permission_groups` | 権限 | スタッフ-権限グループ中間テーブル |
| 32 | `line_customers` | LINE予約 | LINE予約ユーザー管理 |
| 33 | `appointments` | 予約 | 予約 |
| 34 | `hospitalizations` | 入院 | 入院・ホテル |
| 35 | `appointment_trimming_details` | トリミング | 予約に紐づくトリミング詳細 |
| 36 | `appointment_trimming_options` | トリミング | 予約に紐づくトリミングオプション選択 |
| 37 | `medical_records` | 診療 | カルテ |
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
| 48 | `exam_results` | 診療 | 検査記録の検査結果項目 |
| 49 | `daily_records` | 入院 | 入院日次記録 |
| 50 | `vital_records` | 診療・入院 | バイタル記録 |
| 51 | `care_plan_items` | 入院 | ケアプラン項目 |
| 52 | `merchandise_items` | マスタ | 物販・フード・その他マスタ |
| 53 | `estimate_items` | 診療 | 見積書明細 |
| 54 | `care_logs` | 入院 | ケアログ |
| 55 | `staff_notes` | 入院 | スタッフノート |
| 56 | `billings` | 会計 | 会計 |
| 57 | `billing_items` | 会計 | 会計明細 |
| 58 | `payments` | 会計 | 支払い情報 |
| 59 | `billing_refunds` | 会計 | 返金レコード |
| 60 | `shift_entries` | シフト | スタッフシフト |
| 61 | `clinic_holidays` | シフト | 医院個別休診日 |
| 62 | `audit_logs` | 監査 | 操作監査ログ |
| 63 | `line_reservation_settings` | LINE予約 | LINE予約基本設定 |
| 64 | `staff_reservation_exclusions` | 予約マスタ | スタッフ × 非対応予約区分 |
| 65 | `shift_entry_breaks` | シフト | シフト中の休憩時間管理 |
| 66 | `shift_templates` | マスタ | シフトテンプレートマスタ |
| 67 | `shift_template_breaks` | マスタ | シフトテンプレートの休憩時間管理 |
| 68 | `reservation_type_unavailable_times` | 予約マスタ | 予約区分ごとの予約不可時間 |
| 69 | `reservation_type_occupations` | 予約マスタ | 予約区分と職種の対応 |

---

## システム全体 ER図

```mermaid
erDiagram
    %% ===== 法人・医院 =====
    companies {
        bigint id PK
        text name
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
        timestamptz created_at
        timestamptz updated_at
    }

    clinics {
        bigint id PK
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
        numeric standard_tax_rate
        numeric reduced_tax_rate
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 認証 =====
    accounts {
        bigint id PK
        text email "UNIQUE"
        text password_hash
        boolean is_active
        boolean is_system_admin
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    staff_clinic_assignments {
        bigint id PK
        bigint staff_id FK
        bigint clinic_id FK
        boolean is_main
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== コア =====
    owners {
        bigint id PK
        bigint clinic_id FK
        text name
        text phone
        text email
        date birth_date
        membership_type membership_type
        text name_kana
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
        text remarks
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    pets {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint animal_species_id FK
        text name
        pet_gender gender
        pet_status status
        bigint insurance_id FK
        text pet_number
        text name_kana
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
        bigint id PK
        text name
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    occupations {
        bigint id PK
        bigint clinic_id FK
        text name
        text description
        integer sort_order
        boolean is_active
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    staffs {
        bigint id PK
        bigint account_id FK
        text name
        boolean is_active
        text license_number
        bigint occupation_id FK
        integer sort_order
        staff_type staff_type
        text reservation_display_name
        boolean reservation_visible
        text reservation_comment
        text reservation_image_url
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    permission_groups {
        bigint id PK
        bigint clinic_id FK
        varchar_100 name
        text description
        varchar_7 color
        boolean is_active
        int sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    permission_group_rules {
        bigint id PK
        bigint group_id FK
        varchar_50 resource
        boolean can_view
        boolean can_create
        boolean can_edit
        boolean can_delete
        timestamptz created_at
        timestamptz updated_at
    }

    staff_permission_groups {
        bigint staff_id PK
        bigint group_id PK
        timestamptz created_at
    }

    inventory_items {
        bigint id PK
        bigint clinic_id FK
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
        timestamptz deleted_at
    }

    exam_types {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        bigint price
        text description
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    exam_type_fields {
        bigint id PK
        bigint exam_type_id FK
        text name
        integer sort_order
        text inspection_value
        text normal_value
        text unit
        timestamptz created_at
        timestamptz updated_at
    }

    vaccines {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        vaccine_species species
        bigint inventory_id FK
        bigint price
        text description
        text interval
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    medicines {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        dosage_form dosage_form
        bigint inventory_id FK
        bigint price
        tax_type tax_type
        numeric tax_rate
        text description
        medicine_unit medicine_unit
        numeric default_quantity
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    insurances {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        integer coverage_rate
        text description
        text contact_phone
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    cages {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        cage_type cage_type
        cage_size cage_size
        bigint price
        text description
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    reservation_type_groups {
        bigint id PK
        bigint clinic_id FK
        text name
        text color
        integer sort_order
        boolean is_active
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    reservation_types {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        text description
        text color
        integer sort_order
        bigint group_id FK
        text reservation_display_name
        integer duration_minutes
        text short_name
        boolean show_short_name
        boolean reservation_visible
        text reservation_comment
        text reservation_image_url
        text reservation_day_option
        boolean is_internal
        reservation_type_category category
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    consultations {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        bigint price
        tax_type tax_type
        numeric tax_rate
        text description
        text time_condition
        integer duration
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    procedures {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        anesthesia_type anesthesia
        bigint price
        tax_type tax_type
        numeric tax_rate
        text description
        integer duration
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    hospitalization_plans {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        body_size body_size
        billing_unit billing_unit
        bigint price
        tax_type tax_type
        numeric tax_rate
        text description
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    merchandise_items {
        bigint id PK
        bigint clinic_id FK
        text name
        item_category category
        bigint unit_price
        tax_type tax_type
        numeric tax_rate
        boolean is_active
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    trimming_courses {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        target_size target_size
        bigint price
        text description
        integer duration
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    trimming_options {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        boolean is_combinable
        bigint price
        text description
        integer duration
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    diagnosis_types {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        text description
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    diagnosis_names {
        bigint id PK
        bigint clinic_id FK
        text name
        boolean is_active
        text description
        bigint diagnosis_type_id FK
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    checkup_types {
        bigint id PK
        bigint clinic_id FK
        text name
        bigint parent_id FK
        boolean is_active
        text interval
        bigint price
        text description
        text target_age
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    chief_complaint_types {
        bigint id PK
        bigint clinic_id FK
        text name
        text description
        boolean is_active
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    inquiry_templates {
        bigint id PK
        bigint clinic_id FK
        text category
        text title
        text content
        boolean is_active
        integer sort_order
        timestamptz deleted_at
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 診療 =====
    medical_records {
        bigint id PK
        bigint clinic_id FK
        text record_no
        date date
        bigint owner_id FK
        bigint pet_id FK
        bigint doctor_id FK
        bigint appointment_id FK
        medical_record_status status
        integer version
        bigint entered_by FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    clinical_plans {
        bigint id PK
        bigint medical_record_id FK
        text physical_exam
        bigint diagnosis_type_id FK
        bigint diagnosis_name_id FK
        bigint diagnosis_2_type_id FK
        bigint diagnosis_2_name_id FK
        text diagnosis_details
        text treatment_policy
        timestamptz created_at
        timestamptz updated_at
    }

    treatments {
        bigint id PK
        bigint medical_record_id FK
        treatment_item_type item_type
        bigint consultation_id FK
        bigint procedure_id FK
        bigint medicine_id FK
        bigint inventory_id FK
        bigint unit_price
        numeric_10_1 quantity
        boolean is_selected
        treatment_status status
        text content
        text memo
        varchar_50 admin_route
        boolean is_insurance
        numeric discount_rate
        bigint discount_amount
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    vital_records {
        bigint id PK
        bigint pet_id FK
        bigint medical_record_id FK
        bigint daily_record_id FK
        timestamptz recorded_at
        bigint staff_id FK
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
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint exam_type_id FK
        bigint doctor_id FK
        exam_status status
        date date
        text result_summary
        text machine
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    exam_results {
        bigint id PK
        bigint exam_id FK
        bigint exam_type_field_id FK
        text name
        text inspection_value
        exam_result_status status
        text normal_value
        text result
        text unit
        text reference_value
        decimal ref_min
        decimal ref_max
        boolean is_abnormal
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    vaccinations {
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint vaccine_id FK
        date date
        date next_date
        next_schedule_type next_schedule_type
        bigint doctor_id FK
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
        bigint id PK
        bigint medical_record_id FK
        bigint clinic_id FK
        bigint pet_id FK
        bigint checkup_type_id FK
        date date
        date next_date
        bigint doctor_id FK
        text result
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    inquiries {
        bigint id PK
        bigint medical_record_id FK
        bigint chief_complaint_type_id FK
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
        bigint staff_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    medical_record_images {
        bigint id PK
        bigint medical_record_id FK
        text image_url
        text thumbnail_url
        text file_name
        bigint file_size
        text mime_type
        medical_image_type image_type
        text description
        timestamptz taken_at
        bigint exam_id FK
        bigint staff_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    estimates {
        bigint id PK
        bigint clinic_id FK
        text estimate_no
        bigint medical_record_id FK
        text title
        bigint owner_id FK
        estimate_status status
        bigint subtotal
        bigint tax_total
        bigint total_amount
        bigint insurance_amount
        bigint discount_amount
        date valid_until
        text comment
        text notes
        bigint created_by FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    estimate_items {
        bigint id PK
        bigint estimate_id FK
        text name
        item_category category
        bigint unit_price
        numeric quantity
        tax_type tax_type
        numeric tax_rate
        numeric discount_rate
        bigint discount_amount
        boolean is_insurance_applicable
        bigint consultation_id FK
        bigint procedure_id FK
        bigint medicine_id FK
        bigint merchandise_item_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    billing_confirmations {
        bigint id PK
        bigint medical_record_id FK
        confirmation_status status
        bigint confirmed_by FK
        timestamptz confirmed_at
        bigint returned_by FK
        timestamptz returned_at
        text return_reason
        text memo
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== 予約 =====
    appointments {
        bigint id PK
        bigint clinic_id FK
        timestamptz start_time
        timestamptz end_time
        bigint owner_id FK
        bigint pet_id FK
        visit_type visit_type
        bigint reservation_type_id FK
        bigint doctor_id FK
        boolean is_designated
        reservation_status status
        text notes
        reservation_source source
        bigint created_by FK
        boolean is_staff_delegated
        jsonb customer_fields
        bigint line_customer_id FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    %% ===== 入院 =====
    hospitalizations {
        bigint id PK
        bigint clinic_id FK
        bigint owner_id FK
        bigint pet_id FK
        hospitalization_type hospitalization_type
        date start_date
        date end_date
        bigint cage_id FK
        bigint doctor_id FK
        hospitalization_status status
        text memo
        text owner_request
        text staff_notes
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    daily_records {
        bigint id PK
        bigint hospitalization_id FK
        bigint clinic_id FK
        date date
        timestamptz created_at
        timestamptz updated_at
    }

    care_plan_items {
        bigint id PK
        bigint hospitalization_id FK
        care_plan_type type
        plan_timing[] timing
        bigint medicine_id FK
        bigint procedure_id FK
        bigint hospitalization_plan_id FK
        care_plan_status status
        text name
        text description
        text notes
        bigint unit_price
        text category
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    care_logs {
        bigint id PK
        bigint daily_record_id FK
        care_log_type type
        bigint staff_id FK
        time time
        care_log_status status
        text value
        text notes
        timestamptz created_at
        timestamptz updated_at
    }

    staff_notes {
        bigint id PK
        bigint daily_record_id FK
        time time
        bigint staff_id FK
        text content
        timestamptz created_at
        timestamptz updated_at
    }

    treatment_plans {
        bigint id PK
        bigint medical_record_id FK
        bigint hospitalization_id FK
        text treatment_content
        bigint unit_price
        numeric quantity
        text memo
        boolean is_insurance
        numeric discount_rate
        bigint discount_amount
        bigint subtotal
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    %% ===== トリミング =====
    appointment_trimming_details {
        bigint id PK
        bigint clinic_id FK
        bigint appointment_id FK
        bigint course_id FK
        text style_request
        numeric body_weight
        body_weight_unit bw_unit
        numeric body_temperature
        text used_shampoo
        text used_ribbon
        text remarks
        text style_image
        text completed_image
        timestamptz created_at
        timestamptz updated_at
    }

    appointment_trimming_options {
        bigint id PK
        bigint appointment_id FK
        bigint option_id FK
        integer sort_order
        timestamptz created_at
    }

    %% ===== 会計 =====
    billings {
        bigint id PK
        bigint clinic_id FK
        bigint medical_record_id FK
        bigint hospitalization_id FK
        bigint owner_id FK
        bigint pet_id FK
        bigint subtotal
        bigint tax_total
        bigint total_amount
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
        bigint id PK
        bigint billing_id FK
        item_category category
        text name
        bigint unit_price
        numeric_10_1 quantity
        tax_type tax_type
        numeric_3_2 tax_rate
        boolean is_insurance_applicable
        item_source source
        bigint merchandise_item_id FK
        integer sort_order
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    payments {
        bigint id PK
        bigint billing_id FK
        bigint subtotal
        bigint tax_total
        bigint total_amount
        text insurance_name
        numeric_3_2 insurance_ratio
        bigint insurance_amount
        bigint discount_amount
        bigint billing_amount
        bigint received_amount
        bigint change_amount
        payment_method method
        bigint paid_by FK
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    billing_refunds {
        bigint id PK
        bigint clinic_id FK
        bigint billing_id FK
        bigint amount
        text reason
        bigint refunded_by FK
        timestamptz refunded_at
        timestamptz created_at
    }

    %% ===== シフト =====
    shift_entries {
        bigint id PK
        bigint clinic_id FK
        bigint staff_id FK
        date date
        shift_type shift_type
        time start_time
        time end_time
        text notes
        timestamptz created_at
        timestamptz updated_at
    }

    shift_entry_breaks {
        bigint id PK
        bigint shift_entry_id FK
        time break_start
        time break_end
    }

    shift_templates {
        bigint id PK
        bigint clinic_id FK
        varchar_100 name
        shift_type shift_type
        time start_time
        time end_time
        text notes
        integer sort_order
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    shift_template_breaks {
        bigint id PK
        bigint shift_template_id FK
        time break_start
        time break_end
    }

    audit_logs {
        bigint id PK
        bigint clinic_id FK
        bigint actor_id
        varchar_30 actor_type
        varchar_50 action
        varchar_50 resource
        bigint resource_id
        jsonb old_value
        jsonb new_value
        inet ip_address
        text user_agent
        timestamptz created_at
    }

    %% ===== LINE予約 =====
    line_reservation_settings {
        bigint id PK
        bigint clinic_id FK
        text status
        text header_text
        text reservation_notice
        text cancel_notice
        text privacy_policy
        jsonb closed_weekdays
        jsonb closed_dates
        boolean national_holiday_closed
        jsonb business_hours
        jsonb business_hours_by_weekday
        jsonb break_hours
        integer daily_limit
        integer monthly_limit
        integer booking_window_max_days
        integer booking_window_min_days
        integer calendar_months
        text phone_number
        text notification_email
        text request_example
        text time_slot_mode
        integer time_slot_interval_minutes
        text no_staff_mode
        boolean show_no_staff_option
        jsonb additional_fields
        text line_channel_id
        text line_channel_secret
        text liff_id
        text line_access_token
        timestamptz created_at
        timestamptz updated_at
    }

    line_customers {
        bigint id PK
        bigint clinic_id FK
        text line_user_id
        text display_name
        text real_name
        jsonb additional_fields
        bigint owner_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    staff_reservation_exclusions {
        bigint id PK
        bigint staff_id FK
        bigint reservation_type_id FK
    }

    reservation_type_unavailable_times {
        bigint id PK
        bigint clinic_id FK
        bigint reservation_type_id FK
        text unavailable_type
        smallint day_of_week
        date specific_date
        varchar_5 start_time
        varchar_5 end_time
        timestamptz created_at
        timestamptz updated_at
    }

    reservation_type_occupations {
        bigint id PK
        bigint clinic_id FK
        bigint reservation_type_id FK
        bigint occupation_id FK
        timestamptz created_at
    }

    clinic_holidays {
        bigint id PK
        bigint clinic_id FK
        date date
        text reason
        timestamptz created_at
        timestamptz updated_at
    }

    %% ===== リレーション =====

    %% 法人・認証
    companies ||--o{ clinics : "company_id"
    accounts ||--o{ staffs : "account_id"
    clinics ||--o{ staff_clinic_assignments : "clinic_id"
    staffs ||--o{ staff_clinic_assignments : "staff_id"

    %% コア
    clinics ||--o{ owners : "clinic_id"
    clinics ||--o{ pets : "clinic_id"
    owners ||--o{ pets : "owner_id"
    insurances ||--o{ pets : "insurance_id"
    animal_species ||--o{ pets : "animal_species_id"

    %% 共通マスタ
    clinics ||--o{ occupations : "clinic_id"
    clinics ||--o{ inventory_items : "clinic_id"
    clinics ||--o{ cages : "clinic_id"
    clinics ||--o{ clinic_holidays : "clinic_id"
    clinics ||--o{ merchandise_items : "clinic_id"
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

    %% LINE予約
    clinics ||--o{ line_reservation_settings : "clinic_id"
    clinics ||--o{ line_customers : "clinic_id"
    owners ||--o{ line_customers : "owner_id"
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
| `item_category` | examination, test, procedure, surgery, medicine, food, goods, other |
| `item_source` | medical_record, manual, hospitalization |
| `visit_type` | first, revisit |
| `reservation_status` | confirmed, pending, cancelled, checked_in, in_consultation, accounting, completed |
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
