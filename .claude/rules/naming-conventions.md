---
description: DB/Go/API の命名規則（テーブル・カラム・ENUM・エンドポイント）
alwaysApply: false
globs: ["backend/migrations/**", "backend/internal/model/**", "backend/internal/repository/**", "backend/internal/handler/**", "backend/internal/service/**"]
---

# Naming Conventions — DB / Go / API

プロジェクト全体の命名規則。テーブル・カラム・ENUM・Go struct・API エンドポイントの一貫性を保証する。

---

## 1. テーブル名

### 必須ルール

| ルール | 例 | 禁止例 |
|--------|-----|--------|
| **常に複数形** | `companies`, `clinics`, `owners` | `company` |
| **snake_case** | `medical_records`, `billing_items` | `medicalRecords` |
| **同義語を重複させない** | `appointments`（予約） | `reservation_appointments`（reservation = appointment） |
| **テーブル名に否定論理を入れない** | `staff_reservation_exclusions` | `staff_excluded_reservation_categories` |
| **30文字以内を目安** | `staff_reservation_exclusions`（30） | `staff_excluded_reservation_categories`（38） |

### プレフィックス規則

| 条件 | プレフィックス | 例 |
|------|-------------|-----|
| LINE 専用テーブル | `line_` | `line_customers`, `line_reservation_settings` |
| 中間テーブル | `{parent}_{child}` | `staff_clinic_assignments`, `staff_permission_groups` |
| 子テーブル（明細行） | `{parent}_items` | `billing_items`, `estimate_items` |

### 分類マスタの接尾辞

**`_types` に統一**する。`_categories` は使わない。

```sql
-- ✅ 統一
exam_types, checkup_types, reservation_types, diagnosis_types, chief_complaint_types

-- ❌ 禁止: _categories との混在
reservation_categories, diagnosis_categories, chief_complaint_categories
```

例外: グルーピングテーブルは `_groups` を使用。
```sql
reservation_type_groups  -- 予約種別のグルーピング
```

### テーブル名 → Go struct 名 → API パスの対応

原則: struct 名はテーブル名の単数形 PascalCase。ただし既存コードでドメイン概念名を採用している場合はそのまま維持する。

| テーブル名 | Go struct（実装） | API パス | 備考 |
|-----------|------------------|---------|------|
| `appointments` | `ReservationAppointment` | `/reservations` | ドメイン概念名 |
| `medical_records` | `MedicalRecord` | `/medical-records` | — |
| `billing_items` | `BillingItem` | `/billings/{id}/items` | — |
| `line_customers` | `ReservationCustomer` | `/line-customers` | ドメイン概念名 |
| `exam_results` | `ExaminationItem` | （exams の子リソース） | ドメイン概念名 |
| `billing_confirmations` | `BillingReview` | （medical-records の子リソース） | ドメイン概念名 |

---

## 2. カラム名

### 必須ルール

| ルール | 正 | 誤 |
|--------|-----|-----|
| **テーブル名を繰り返さない** | `owners.name` | `owners.owner_name` |
| **Boolean は `is_`/`has_`/`can_` プレフィックス** | `is_active`, `is_selected`, `has_insurance` | `active`, `selected`, `insurance` |
| **省略形は禁止** | `body_weight`, `body_temperature`, `reference_value` | `bw`, `bt`, `ref` |
| **単複は複数形に統一**（テキスト列） | `notes` | `note` |

### テキスト列の用途別命名

| 用途 | カラム名 | 使用場面 |
|------|---------|---------|
| マスタの説明文 | `description` | `consultations.description`, `procedures.description` |
| 運用時のフリーテキスト | `memo` | `hospitalizations.memo`, `billings.memo` |
| 補足・備考 | `notes` | `appointments.notes`, `vital_records.notes`（既存で広く使用） |
| 飼主・ペットの備考 | `remarks` | `owners.remarks`, `pets.remarks`（既存で広く使用） |

新規テーブルでは `memo` を推奨。既存の `notes` / `remarks` は維持可（変更コスト対効果が低い）。

### FK カラム名

| 参照先 | 役割 | カラム名 |
|--------|------|---------|
| `staffs` | 担当医 | `doctor_id`（ドメイン上「医師」が明確な場合） |
| `staffs` | 操作者・記録者 | `staff_id` |
| `staffs` | 確認者 | `confirmed_by`（慣例として許容） |
| `staffs` | 作成者 | `created_by`（慣例として許容） |
| `staffs` | 差戻者 | `returned_by`（慣例として許容） |

`confirmed_by` 等は `staffs` テーブルへの FK であることが文脈で自明なため、`_staff_id` サフィックスは省略可。

### 日付・時刻カラム

| 用途 | 型 | 命名パターン | 例 |
|------|-----|------------|-----|
| システム管理 | `timestamptz` | `created_at`, `updated_at`, `deleted_at` | 全テーブル共通 |
| 操作タイムスタンプ | `timestamptz` | `{動詞}_at` | `confirmed_at`, `refunded_at`, `completed_at` |
| 業務日 | `date` | `date`（テーブル文脈で自明な場合） | `medical_records.date`, `vaccinations.date` |
| 期間 | `date` | `start_date`, `end_date` | `hospitalizations.start_date` |
| 時刻 | `time` | `start_time`, `end_time` | `appointments.start_time` |

### 金額カラム

| 用途 | カラム名 | テーブル |
|------|---------|---------|
| マスタ単価 | `price` | `exam_types`, `vaccines`, `medicines` 等 |
| 明細単価 | `unit_price` | `billing_items`, `estimate_items`, `treatments` |
| 小計 | `subtotal` | `billings`, `estimates`, `payments` |
| 合計 | `total_amount` | `billings`, `estimates` |
| 税額 | `tax_total` | `billings`, `estimates` |
| 割引額 | `discount_amount` | `billing_items`, `estimate_items` |
| 割引率 | `discount_rate` | `billing_items`, `estimate_items`, `owners` |

---

## 3. ENUM 型名

### 命名規則

```
{テーブル名の単数形}_{カラム名}
```

| テーブル | カラム | ENUM 型名（実装） | 備考 |
|---------|--------|----------|------|
| `appointments` | `status` | `reservation_status` | ドメイン概念名を採用 |
| `exams` | `status` | `exam_status` | — |
| `exam_results` | `status` | `exam_result_status` | — |
| `billings` | `status` | `billing_status` | — |
| `billing_confirmations` | `status` | `confirmation_status` | — |
| `pets` | `gender` | `pet_gender` | — |
| `medicines` | `dosage_form` | `dosage_form` | 共通概念のため接頭辞不要 |

**禁止**: テーブル名と ENUM 型名を完全同一にしない。

```sql
-- ❌ カラム名と ENUM 型名が同じで SQL が読みにくい
cages.cage_type cage_type

-- ✅ テーブル名の単数形を接頭辞に
cages.cage_type cage_type  -- これは許容（テーブル名 cages ≠ ENUM 名 cage_type）
```

### ENUM 値

- 全て **snake_case**
- 肯定形を使う（`active` であり `not_inactive` ではない）
- 時系列を意識した並び順

```sql
CREATE TYPE exam_status AS ENUM (
    'pending',        -- 未実施
    'in_progress',    -- 実施中
    'result_entered', -- 結果入力済み
    'completed',      -- 完了
    'confirmed'       -- 確定
);
```

---

## 4. Go 層の命名

### Model (struct)

原則: テーブル名の**単数形 PascalCase**。既存コードでドメイン概念名を採用している場合はそのまま維持。

| DB テーブル | Go struct（実装） | Go ファイル | 備考 |
|-----------|------------------|------------|------|
| `appointments` | `ReservationAppointment` | `reservation.go` | ドメイン概念名 |
| `medical_records` | `MedicalRecord` | `medical_record.go` | — |
| `line_customers` | `ReservationCustomer` | `reservation_customer.go` | ドメイン概念名 |
| `exam_results` | `ExaminationItem` | `examination_record.go` | ドメイン概念名 |
| `billing_confirmations` | `BillingReview` | `billing_review.go` | ドメイン概念名 |

### ENUM 定数

```go
type ExaminationStatus string  // SQL: exam_status

const (
    ExaminationStatusPending       ExaminationStatus = "pending"
    ExaminationStatusInProgress    ExaminationStatus = "in_progress"
    ExaminationStatusResultEntered ExaminationStatus = "result_entered"
    ExaminationStatusCompleted     ExaminationStatus = "completed"
    ExaminationStatusConfirmed     ExaminationStatus = "confirmed"
)
```

パターン: `{GoStructName}{PascalCaseValue}`。Go 型名と SQL ENUM 名は一致しなくてもよい（GORM タグで橋渡し）。

### Repository / Service / Handler

| レイヤー | 命名パターン | 例 |
|---------|------------|-----|
| Repository interface | `{Entity}Repository` | `AppointmentRepository` |
| Repository impl | `appointment_repository.go` | — |
| Service interface | `{Entity}Service` | `AppointmentService` |
| Service impl | `appointment_service.go` | — |
| Handler | `{Entity}Handler` | `AppointmentHandler` |
| Handler impl | `appointment_handler.go` | — |
| Routes | `appointment_routes.go` | — |

**原則 1エンティティ = 1ファイルセット**。ただし、管理 API（`/v1/masters/`）と LIFF 公開 API（`/api/clinics/`）で同一テーブルを異なるレイヤーから提供する場合は、目的別に分離してよい。その場合も **API パス名はテーブル名と一致** させること。

```
# 同じ reservation_types テーブルに対して:
/v1/masters/reservation-types          ← 管理用（reservation_category_handler.go）
/api/clinics/{id}/reservation-types    ← LINE予約用（reservation_course_handler.go）
```

---

## 5. API エンドポイント

### パス命名

```
/api/clinics/{clinicId}/{resource}           -- 一覧・作成
/api/clinics/{clinicId}/{resource}/{id}      -- 取得・更新・削除
/api/{parent-resource}/{parentId}/{child}    -- 子リソース
```

| ルール | 正 | 誤 |
|--------|-----|-----|
| **kebab-case** | `/medical-records` | `/medicalRecords` |
| **複数形** | `/appointments` | `/appointment` |
| **テーブル名と1:1対応** | `/appointments`（← `appointments` テーブル） | `/reservation-appointments` |
| **LINE リソースは `line-` プレフィックス** | `/line-customers` | `/reservation-customers` |

---

## チェックリスト（新規テーブル追加時）

- [ ] テーブル名: 複数形 snake_case、30文字以内
- [ ] テーブル名: 同義語の重複なし
- [ ] LINE 専用なら `line_` プレフィックス付き
- [ ] 分類マスタなら `_types` 接尾辞
- [ ] カラム名: テーブル名を繰り返していない
- [ ] Boolean: `is_`/`has_`/`can_` プレフィックス付き
- [ ] テキスト列: `description`（マスタ説明）/ `memo`（運用メモ）のいずれか
- [ ] FK: 参照先と役割が名前から判別できる
- [ ] ENUM 型名: `{テーブル単数形}_{カラム名}` パターン
- [ ] ENUM 値: snake_case、時系列順
- [ ] Go struct: テーブル名の単数形 PascalCase
- [ ] API パス: テーブル名と 1:1、kebab-case
