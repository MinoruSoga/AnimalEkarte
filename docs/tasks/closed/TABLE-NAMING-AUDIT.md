# TABLE-NAMING-AUDIT: DB命名規則の全面監査

- **対象**: `backend/migrations/001_init.sql` 全64テーブル + カラム + ENUM + Go/API層
- **作成日**: 2026-04-11
- **ステータス**: Phase 1-3 完了 / Phase 4 (INFO) は対応任意

---

## 概要

全64テーブルの命名規則を以下の観点で監査した。

1. テーブル名（同義語重複、複数形、プレフィックス、接尾辞の一貫性）
2. カラム名（冗長性、boolean プレフィックス、テキスト列の概念混在、FK 命名、省略形）
3. ENUM型名（テーブル名との不整合）
4. Go モデル / API エンドポイントとの命名乖離

---

## Part 1: テーブル名

### CRITICAL（3件）

| 現在 | 提案 | 理由 |
|------|------|------|
| `company` | `companies` | 全テーブル複数形の規約に違反。シングルトンでもテーブル名は複数形が PostgreSQL 標準 |
| `reservation_appointments` | `appointments` | **同義語の重複**。reservation（予約）と appointment（予約）は同じ意味。source カラムで LINE/手動を区別済み |
| `staff_excluded_reservation_categories` | `staff_reservation_exclusions` | 38文字で最長。テーブル名に否定論理を埋め込みすぎ |

### WARNING（12件）

| 現在 | 提案 | 理由 |
|------|------|------|
| `reservation_settings` | `line_reservation_settings` | LINE専用テーブル（`line_channel_id`, `liff_id`, `line_access_token` 保持）に `line_` プレフィックス |
| `reservation_customers` | `line_customers` | LINE専用テーブル（`line_user_id NOT NULL`）に `line_` プレフィックス |
| `reservation_categories` | `reservation_types` | `_types`/`_categories` 統一（後述） |
| `reservation_category_groups` | `reservation_type_groups` | 上記に連動 |
| `diagnosis_categories` | `diagnosis_types` | `_types`/`_categories` 統一 |
| `chief_complaint_categories` | `chief_complaint_types` | `_types`/`_categories` 統一 |
| `exam_type_items` | `exam_type_fields` | `exam_items`（結果側）との混同回避。"fields"=定義、"results"=記録 |
| `exam_items` | `exam_results` | 上記に連動 |
| `care_log_records` | `care_logs` | "log" と "record" は同義で冗長 |
| `staff_note_records` | `staff_notes` | "_records" 冗長 |
| `record_images` | `medical_record_images` | "record" が何の record か曖昧 |
| `billing_reviews` | `billing_confirmations` | "review" が曖昧。実態は医師による会計確認（confirm/return） |

### INFO（3件）

| 現在 | 備考 |
|------|------|
| `staffs` | 英語文法上 "staff" は不可算名詞だが、日本語圏では慣例的に許容 |
| `vital_records` | "records" は冗長だが他テーブルとの一貫性あり |
| `daily_records` | "daily" だけでは入院文脈不明だが FK で十分 |

### 分類接尾辞の統一: `_types` vs `_categories`

| テーブル | 接尾辞 | 日本語 | 実態 |
|---------|--------|--------|------|
| `exam_types` | `_types` | 検査種別 | 種類定義 |
| `checkup_types` | `_types` | 健診種別 | 種類定義 |
| `diagnosis_categories` | `_categories` | 診断カテゴリ | 病名グルーピング |
| `chief_complaint_categories` | `_categories` | 主訴区分 | 主訴グルーピング |
| `reservation_categories` | `_categories` | 予約区分 | 予約種類定義 |

**方針**: `_types` に統一。SQL/Go で「このマスタの接尾辞は types? categories?」と迷う頻度を0にする。

---

## Part 2: カラム名

### CRITICAL（3件）

#### 2-1. テーブル名とカラム名の冗長な重複

| テーブル | 現在 | 提案 | 理由 |
|---------|------|------|------|
| `owners` | `owner_name` | `name` | テーブルコンテキストで `owners.name` は自明。Go で `Owner.OwnerName` は冗長 |
| `owners` | `owner_name_kana` | `name_kana` | 同上 |
| `pets` | `pet_name_kana` | `name_kana` | 同上。`pets.name` は既に存在し `name_kana` との対称性が自然 |

#### 2-2. テキスト列の概念混在（memo / notes / remarks / comment）

同じ「フリーテキスト記述」に4種類のカラム名が混在:

| カラム名 | 使用テーブル |
|---------|------------|
| `memo` | hospitalizations, treatments, treatment_plans, billing_reviews, billings |
| `notes` | reservation_appointments, inquiries, vital_records, care_log_records, care_plan_items, estimates |
| `remarks` | owners, pets, vaccinations, trimming_records |
| `comment` | estimates（`comment` と `notes` が同じテーブルに共存） |

**統一方針**:
- `description`: マスタ定義の説明（変更不要、既に統一済み）
- `memo`: 運用時の内部メモ → **統一先**
- `notes`: 一般的な補足 → `memo` に統一
- `remarks`: 廃止 → `memo` に統一
- `comment`: 廃止 → `memo` に統一

#### 2-3. FK命名の不統一（doctor_id vs staff_id）

同じ `staffs(id)` への FK に2種類の名前が混在:

| カラム名 | 使用テーブル | 実態 |
|---------|------------|------|
| `doctor_id` | reservation_appointments, hospitalizations, medical_records, vaccinations, checkups, exams | `staffs(id)` への FK |
| `staff_id` | trimming_records, vital_records, care_log_records, staff_note_records, record_images, inquiries | `staffs(id)` への FK |
| `confirmed_by` | billing_reviews | `staffs(id)` への FK |
| `returned_by` | billing_reviews | `staffs(id)` への FK |
| `created_by` | estimates | `staffs(id)` への FK |

**問題**: 同じ参照先なのに `doctor_id` と `staff_id` が混在。`doctor_id` は `staff_type = 'doctor'` の暗黙的制約を示唆するが、DB レベルの制約はない。

**統一方針**:
- 担当医: `doctor_id`（現状維持。ドメイン上「担当医」が明確な場合）
- 操作者: `staff_id`（現状維持）
- 確認者/作成者: `confirmed_by_staff_id`, `returned_by_staff_id`, `created_by_staff_id`（現在の `confirmed_by` 等は FK 先が不明瞭）

### WARNING（5件）

#### 2-4. Boolean カラムの `is_` プレフィックス欠如

| テーブル | 現在 | 提案 | 理由 |
|---------|------|------|------|
| `trimming_options` | `combinable` | `is_combinable` | 他テーブルは `is_active`, `is_main` とプレフィックス付き |
| `treatments` | `selected` | `is_selected` | 同上 |
| `treatments` | `insurance` | `is_insurance` | boolean なのにプレフィックスなし。`is_insurance_applicable`（estimate_items）との不統一 |
| `reservation_appointments` | `is_designated` | — | ✅ OK（既にプレフィックス付き） |
| `reservation_appointments` | `is_staff_delegated` | — | ✅ OK |

#### 2-5. 省略形カラム名

| テーブル | 現在 | 提案 | 理由 |
|---------|------|------|------|
| `trimming_records` | `bw` | `body_weight` | 省略が不明瞭。`weight` でもよいが `vital_records.weight` と区別 |
| `trimming_records` | `bt` | `body_temperature` | 同上。`vital_records.temperature` と対応させるなら `temperature` |
| `exam_items` | `ref` | `reference_value` | "ref" は曖昧（reference? referral?） |

#### 2-6. 曖昧な `date` カラム

| テーブル | 現在 | 問題 |
|---------|------|------|
| `exams` | `date` | 検査日？記録日？ |
| `vaccinations` | `date` | 接種日？記録日？ |
| `checkups` | `date` | 健診日？ |
| `trimming_records` | `date` | 施術日？ |
| `medical_records` | `date` | 診療日（これは文脈で明確） |

**提案**: `exam_date`, `vaccination_date`, `checkup_date`, `trimming_date` に変更。ただし Go モデルでは `Exam.Date` で十分自明なので INFO 扱い。

#### 2-7. `shift_entries.note` — 単複不統一

他テーブルは `notes`（複数形）だが、このテーブルだけ `note`（単数形）。

### INFO

#### 2-8. `cages.cage_type` / `cages.cage_size` — テーブル名との冗長

`cages.type`, `cages.size` で十分だが、ENUM型名 `cage_type`, `cage_size` との衝突を避けるために意図的に冗長にしている可能性がある。変更コストに対して効果が低い。

---

## Part 3: ENUM型名

### CRITICAL（2件）

#### 3-1. `examination_status` — テーブル名 `exams` と不整合

```sql
CREATE TYPE examination_status AS ENUM (...);
CREATE TABLE exams (status examination_status ...);  -- テーブル名は "exams"
```

ENUM型名が `examination_status` だが、テーブル名は `exams`。慣例では `exam_status` であるべき。

| 現在 | 提案 |
|------|------|
| `examination_status` | `exam_status` |
| `examination_result_status` | `exam_result_status` |

Go モデルでは `ExaminationStatus` 型として生成され、テーブル名 `exams` との乖離が顕著。

#### 3-2. `examination_result_status` — 同上

`exam_items` テーブルで使用されるが、ENUM名は `examination_result_status`。`exam_result_status` が適切。

### WARNING（1件）

#### 3-3. `billing_review_status` — 冗長

テーブル名 `billing_reviews` の status カラム専用 ENUM。`review_status` で十分。ただしテーブル名を `billing_confirmations` に変更する場合は `confirmation_status` に変更すべき。

### OK（44件）

残りの全 ENUM 型は問題なし。ENUM 値は全て snake_case で統一済み。

---

## Part 4: Go モデル / API エンドポイントとの乖離

### CRITICAL（1件）

#### 4-1. `reservation_categories` テーブルの二重 API 問題

**同一テーブルに対して 2つの API インターフェースが存在する:**

```
テーブル: reservation_categories
  ├─ Go Struct: ReservationCategory
  │
  ├─ API 1: /api/clinics/{id}/reservation-courses
  │  ├─ Handler: reservation_course_handler.go
  │  ├─ Service: ReservationCourseService
  │  └─ Repository: reservation_course_repository.go
  │
  └─ API 2: /api/clinics/{id}/reservation-categories
     ├─ Handler: reservation_category_handler.go
     ├─ Service: ReservationCategoryService
     └─ Repository: reservation_category_repository.go
```

**問題**: テーブル名は `reservation_categories` だが、API パスは `/reservation-courses` と `/reservation-categories` の両方が存在。同じデータに対して「courses」と「categories」という異なる名前でアクセスする設計は混乱の元。

**テーブルリネーム時の FK 影響**:

| 参照元テーブル | FK カラム | 影響度 |
|-------------|---------|--------|
| `reservation_appointments` | `reservation_category_id` | CRITICAL |
| `staff_excluded_reservation_categories` | `reservation_category_id` | CRITICAL |
| `reservation_categories` | `group_id` → `reservation_category_groups` | HIGH |

### WARNING（1件）

#### 4-2. `staffs` テーブルの予約 API での利用

`staffs` テーブルのデータが `/api/clinics/{id}/reservation-staffs` エンドポイントで公開されている。`ReservationStaffService` / `ReservationStaffHandler` という別レイヤーが存在するが、Go struct は `Staff` のまま。テーブルリネーム時の影響はないが、API 命名の一貫性に注意。

---

## 修正サマリー（全カテゴリ統合）

### CRITICAL（8件）

| カテゴリ | 現在 | 提案 | 理由 |
|---------|------|------|------|
| テーブル | `company` | `companies` | 複数形違反 |
| テーブル | `reservation_appointments` | `appointments` | 同義語重複 |
| テーブル | `staff_excluded_reservation_categories` | `staff_reservation_exclusions` | 長すぎ + 否定論理 |
| カラム | `owners.owner_name` | `owners.name` | テーブル名との冗長 |
| カラム | `owners.owner_name_kana` | `owners.name_kana` | 同上 |
| カラム | `pets.pet_name_kana` | `pets.name_kana` | 同上 |
| ENUM | `examination_status` | `exam_status` | テーブル名 `exams` と不整合 |
| ENUM | `examination_result_status` | `exam_result_status` | 同上 |

### WARNING（19件）

| カテゴリ | 現在 | 提案 |
|---------|------|------|
| テーブル | `reservation_settings` | `line_reservation_settings` |
| テーブル | `reservation_customers` | `line_customers` |
| テーブル | `reservation_categories` | `reservation_types` |
| テーブル | `reservation_category_groups` | `reservation_type_groups` |
| テーブル | `diagnosis_categories` | `diagnosis_types` |
| テーブル | `chief_complaint_categories` | `chief_complaint_types` |
| テーブル | `exam_type_items` | `exam_type_fields` |
| テーブル | `exam_items` | `exam_results` |
| テーブル | `care_log_records` | `care_logs` |
| テーブル | `staff_note_records` | `staff_notes` |
| テーブル | `record_images` | `medical_record_images` |
| テーブル | `billing_reviews` | `billing_confirmations` |
| カラム | `trimming_options.combinable` | `is_combinable` |
| カラム | `treatments.selected` | `is_selected` |
| カラム | `treatments.insurance` | `is_insurance` |
| カラム | `trimming_records.bw` | `body_weight` |
| カラム | `trimming_records.bt` | `body_temperature` |
| カラム | `exam_items.ref` | `reference_value` |
| ENUM | `billing_review_status` | `confirmation_status`（テーブル名変更に連動） |

### INFO

- `staffs`（英語文法上は不可算名詞）
- `memo`/`notes`/`remarks`/`comment` の統一（影響範囲が広く段階的対応が望ましい）
- `doctor_id` vs `staff_id` の FK 命名混在（ドメイン意味があるため維持可）
- `date` カラムの曖昧さ（Go struct 上では自明）
- `shift_entries.note` の単複不統一
- `cages.cage_type` / `cages.cage_size` の冗長

---

## 影響範囲（リネーム実施時）

| レイヤー | 影響ファイル |
|---------|------------|
| Migration | `backend/migrations/001_init.sql`（DDL + INDEX + COMMENT + ENUM） |
| Seed | `backend/migrations/002_seed_master.sql` |
| Go Model | `backend/internal/model/*.go`（`TableName()` + struct タグ） |
| Go Repository | `backend/internal/repository/*_repository.go` |
| Go Service | `backend/internal/service/*_service.go` |
| Go Handler | `backend/internal/handler/*_handler.go` + `*_routes.go` |
| TypeScript 型 | `frontend/src/types/generated/models.ts`（`make codegen` で自動更新） |
| TypeScript API | `frontend/src/features/*/api/*.ts`（エンドポイントパス） |
| TypeScript Transform | `frontend/src/features/*/api/transforms.ts`（フィールド名） |

**注意**: リリース前の DB リセット運用中のため、`001_init.sql` 直接編集で対応可能。incremental migration 不要。

---

## 実施方針

### Phase 1: CRITICAL 8件
1. テーブル名 3件（`companies`, `appointments`, `staff_reservation_exclusions`）
2. カラム名 3件（`owners.name`, `owners.name_kana`, `pets.name_kana`）
3. ENUM 2件（`exam_status`, `exam_result_status`）

### Phase 2: WARNING（LINE プレフィックス + 分類統一）
4. LINE テーブル 2件（`line_reservation_settings`, `line_customers`）
5. `_categories` → `_types` 統一 4件
6. `reservation-courses` / `reservation-categories` の API 二重命名解消

### Phase 3: WARNING（残り）
7. テーブル冗長接尾辞 4件（`care_logs`, `staff_notes`, `exam_results`, `exam_type_fields`）
8. テーブル曖昧名 2件（`medical_record_images`, `billing_confirmations`）
9. カラム boolean/省略 6件
10. ENUM 連動 1件

### Phase 4: INFO（対応任意）
- `memo`/`notes`/`remarks` 統一（影響範囲が大きいため別チケット推奨）
