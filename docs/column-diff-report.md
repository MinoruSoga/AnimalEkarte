# カラム差分レポート

生成日: 2026-03-14
調査対象: 55テーブル

## 凡例

| 記号 | 意味 |
|------|------|
| ✅ | カラムが存在する |
| ❌ | カラムが存在しない（差分あり） |
| N/A | スキーマ未定義（差分として扱わない） |

## 重大度

| 記号 | 意味 |
|------|------|
| 🔴 CRITICAL | ERD ↔ Goモデル間の不一致（要修正） |
| 🟡 WARNING | api.yaml 不整合（要確認） |

---

## サマリー

| 項目 | 数 |
|------|-----|
| 総テーブル数 | 55 |
| 差分ゼロのテーブル数 | 55 |
| 差分ありテーブル数 | 0 |
| 🔴 CRITICAL 差分カラム数 | 0 |
| 🟡 WARNING 差分カラム数 | 0 |

---

## 差分ゼロテーブル一覧

差分がないテーブルは一覧のみ。カラム順序の違いは差分として扱わない。

| # | テーブル名 | api.yaml スキーマ |
|---|-----------|-----------------|
| 1 | `animal_species` | AnimalSpecies |
| 2 | `billing_items` | BillingItem |
| 3 | `billing_reviews` | N/A |
| 4 | `billings` | Billing |
| 5 | `cages` | Cage |
| 6 | `care_log_records` | N/A |
| 7 | `checkup_types` | CheckupType |
| 8 | `checkups` | N/A |
| 9 | `chief_complaint_categories` | N/A |
| 10 | `clinical_plans` | N/A |
| 11 | `clinics` | Clinic |
| 12 | `company` | Company |
| 13 | `consultations` | Consultation |
| 14 | `daily_records` | N/A |
| 15 | `diagnosis_categories` | DiagnosisCategory |
| 16 | `diagnosis_names` | DiagnosisName |
| 17 | `estimate_items` | N/A |
| 18 | `estimates` | N/A |
| 19 | `exam_items` | ExaminationItem |
| 20 | `exam_type_items` | N/A |
| 21 | `exam_types` | ExaminationType |
| 22 | `exams` | Examination |
| 23 | `hospitalization_plans` | HospitalizationPlan |
| 24 | `hospitalizations` | Hospitalization |
| 25 | `inquiries` | Inquiry |
| 26 | `inquiry_templates` | N/A |
| 27 | `insurances` | Insurance |
| 28 | `inventory_items` | InventoryItem |
| 29 | `job_titles` | JobTitle |
| 30 | `medical_records` | MedicalRecord |
| 31 | `medicines` | Medicine |
| 32 | `owners` | Owner |
| 33 | `payments` | Payment |
| 34 | `pets` | Pet |
| 35 | `procedures` | Procedure |
| 36 | `record_images` | N/A |
| 37 | `reservation_appointments` | ReservationAppointment |
| 38 | `service_types` | ServiceType |
| 39 | `shift_entries` | N/A |
| 40 | `staff_note_records` | N/A |
| 41 | `staffs` | Staff |
| 42 | `treatment_plans` | N/A |
| 43 | `treatments` | N/A |
| 44 | `trimming_courses` | TrimmingCourse |
| 45 | `trimming_options` | TrimmingOption |
| 46 | `trimming_record_options` | N/A |
| 47 | `trimming_records` | TrimmingRecord |
| 48 | `user_accounts` | N/A |
| 49 | `user_clinic_memberships` | N/A |
| 50 | `user_permissions` | N/A |
| 51 | `vaccines` | Vaccine |
| 52 | `vaccinations` | Vaccination |
| 53 | `vital_records` | N/A |
| 54 | `vitals` | N/A |
| 55 | `care_plan_items` | N/A |

---

## ✅ 対応済み事項

### `company` — api.yaml スキーマ分離（対応済み）

`/company` エンドポイントが `Clinic` スキーマを流用していたため `is_active`, `company_id` が混入していた。
`api.yaml` に `Company` 専用スキーマを追加し、`/company` エンドポイントの全参照を `Company` スキーマに変更済み。

現在の `company` テーブルのカラム整合性：

| カラム名 | ERD | Goモデル | api.yaml | 備考 |
|---------|-----|----------|---------|------|
| id | ✅ | ✅ | ✅ | |
| name | ✅ | ✅ | ✅ | |
| postal_code | ✅ | ✅ | ✅ | |
| address | ✅ | ✅ | ✅ | |
| phone_number | ✅ | ✅ | ✅ | |
| fax_number | ✅ | ✅ | ✅ | |
| registration_number | ✅ | ✅ | ✅ | |
| director_name | ✅ | ✅ | ✅ | |
| email | ✅ | ✅ | ✅ | |
| website | ✅ | ✅ | ✅ | |
| logo_url | ✅ | ✅ | ✅ | |
| created_at | ✅ | ✅ | ✅ | |
| updated_at | ✅ | ✅ | ✅ | |
