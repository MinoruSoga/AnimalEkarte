# カラム差分レポート

生成日: 2026-03-14

## 調査概要

以下の3ソース間でテーブル毎のカラム名差分を調査した。

| ソース | 場所 |
|--------|------|
| ERD.md | `docs/ERD.md` v19.0（55テーブル・3116行） |
| Goモデル | `backend/internal/model/*.go`（37ファイル） |
| api.yaml | `backend/docs/api.yaml` v2.0.0（5738行） |

### 調査の注意点

- **soft-delete (`deleted_at`)**: Goモデルでは `gorm.DeletedAt` 型で定義し `json:"-"` を付与。DBカラムとしては存在するがJSONレスポンスには含まれない。本レポートでは**DBカラムとして存在する**と判定した。
- **`password_hash`**: Goモデルでは `json:"-"` で定義済み（セキュリティ上の意図的除外）。api.yaml に `user_accounts` スキーマが存在しないため N/A 扱い。
- **api.yaml の N/A**: api.yaml にスキーマ定義がないテーブルは「N/A（スキーマ未定義）」と表記。DBカラムが全てAPIに公開される必要はないため、N/A は差分として扱わない。
- **リレーションフィールド**: GoモデルのJSON名がリレーション（例: `json:"category,omitempty"` で型が `*DiagnosisCategory`）の場合はDBカラムではないため除外した。

---

## サマリー

| 項目 | 数 |
|------|-----|
| 総テーブル数 | 55 |
| 完全一致テーブル数（差分なし） | 54 |
| 差分ありテーブル数 | 1 |
| 総差分カラム数 | 1 |
| DB重要差分（ERD ↔ Goモデル） | 0 |
| api.yaml のみ差分 | 1 |

**結論: ERD とGoモデルは全55テーブルで完全一致。api.yaml との差分は `company.is_active` の1件のみで、原因は api.yaml が Clinic スキーマを company エンドポイントに流用していることによる設計上の問題。**

---

## 差分なしテーブル一覧

以下の54テーブルは ERD・Goモデル・api.yaml（定義がある場合）すべてで完全一致。

| # | テーブル名 | api.yaml スキーマ |
|---|-----------|-----------------|
| 1 | `animal_species` | AnimalSpecies |
| 2 | `billing_items` | BillingItem |
| 3 | `billing_reviews` | N/A（未定義） |
| 4 | `billings` | Billing |
| 5 | `cages` | Cage |
| 6 | `care_log_records` | N/A |
| 7 | `care_plan_items` | N/A |
| 8 | `checkup_types` | CheckupType |
| 9 | `checkups` | N/A |
| 10 | `chief_complaint_categories` | N/A |
| 11 | `clinical_plans` | N/A |
| 12 | `clinics` | Clinic |
| 13 | `consultations` | Consultation |
| 14 | `daily_records` | N/A |
| 15 | `diagnosis_categories` | DiagnosisCategory |
| 16 | `diagnosis_names` | DiagnosisName |
| 17 | `estimate_items` | N/A |
| 18 | `estimates` | N/A |
| 19 | `exam_items` | ExaminationItem |
| 20 | `exam_type_items` | N/A（ExaminationType の items 配列に含む） |
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
| 48 | `user_accounts` | N/A（MeResponse は合成オブジェクト） |
| 49 | `user_clinic_memberships` | N/A |
| 50 | `user_permissions` | N/A |
| 51 | `vaccinations` | Vaccination |
| 52 | `vaccines` | Vaccine |
| 53 | `vital_records` | N/A |
| 54 | `vitals` | N/A |

---

## テーブル別差分詳細

差分があるテーブルのみ詳細を記載する。

### テーブル名: `company`

**概要**: api.yaml が `/company` エンドポイントのレスポンス定義として `Clinic` スキーマを流用しているため、`Clinic` スキーマが持つ `is_active` フィールドが混入している。`company` テーブル自体はシングルトンのため `is_active` は設計上不要。

| カラム名 | ERD.md | Goモデル | api.yaml | 備考 |
|---------|--------|----------|----------|------|
| `is_active` | ❌ | ❌ | ✅ | api.yaml が Clinic スキーマを company エンドポイントに流用。company テーブルには is_active は存在しない。api.yaml に Company 専用スキーマを作成するか、Clinic スキーマを流用する場合に is_active を除外すべき |

**影響**: api.yaml の `company` エンドポイント（`GET /company`, `PATCH /company`）のレスポンス定義に `is_active` フィールドが存在するが、実際のDBには当該カラムがない。フロントエンド側で `is_active` を参照している場合は常に `undefined` / ゼロ値になる。

---

## 補足情報

### soft-delete カラムの扱い

`deleted_at` を持つテーブルは以下の17テーブル。全てのテーブルで ERD・Goモデル・api.yaml（定義がある場合）が一致している。

| テーブル | ERD | Goモデル | api.yaml |
|---------|-----|----------|---------|
| `billing_items` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `billings` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `checkup_types` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `checkups` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | N/A |
| `estimates` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | N/A |
| `exams` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `hospitalizations` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `medical_records` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `owners` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `pets` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `reservation_appointments` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `staffs` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `treatment_plans` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | N/A |
| `treatments` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | N/A |
| `trimming_records` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |
| `user_accounts` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | N/A |
| `vaccinations` | ✅ | ✅（`gorm.DeletedAt`, `json:"-"`） | ✅ |

### `password_hash` の扱い

`user_accounts.password_hash` は ERD に存在し、Goモデルにも `json:"-"` タグで定義済み（セキュリティ上の意図的除外）。api.yaml には `user_accounts` 専用スキーマが存在しない（`MeResponse` は合成オブジェクト）。差分なし・意図的設計。

### api.yaml の N/A テーブルについて

以下のテーブルは api.yaml にスキーマ定義が存在しない。内部処理のみで使用されるか、親テーブル経由で参照されるため、個別のエンドポイント定義が不要な設計となっている。

- `billing_reviews`, `care_log_records`, `care_plan_items`, `checkups`, `chief_complaint_categories`, `clinical_plans`, `daily_records`, `estimate_items`, `estimates`, `inquiry_templates`, `record_images`, `shift_entries`, `staff_note_records`, `treatment_plans`, `treatments`, `trimming_record_options`, `user_accounts`（MeResponse経由）, `user_clinic_memberships`, `user_permissions`, `vital_records`, `vitals`

---

## 修正推奨事項

### 優先度: 低（api.yaml設計の改善）

**`company` テーブルの api.yaml スキーマ分離**

現在 `/company` エンドポイントが `Clinic` スキーマを流用しているため `is_active` フィールドが混入している。

修正案：
1. `api.yaml` に `Company` 専用スキーマを追加（`is_active` なし）
2. または既存の `Clinic` スキーマを流用したまま、フロントエンドで `company` レスポンスの `is_active` を無視する（現状維持）

```yaml
# 修正案: Company スキーマを追加
Company:
  type: object
  properties:
    id:
      type: integer
      format: int64
    name:
      type: string
    postal_code:
      type: string
    address:
      type: string
    phone_number:
      type: string
    fax_number:
      type: string
    registration_number:
      type: string
    director_name:
      type: string
    email:
      type: string
    website:
      type: string
    logo_url:
      type: string
    created_at:
      type: string
      format: date-time
    updated_at:
      type: string
      format: date-time
    # is_active は company テーブルに存在しないため除外
```
