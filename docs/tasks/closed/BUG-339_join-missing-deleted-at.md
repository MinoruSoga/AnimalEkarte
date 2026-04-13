# BUG-339: JOIN 先テーブルの `deleted_at IS NULL` 条件欠落（7箇所）

## 概要

Repository の JOIN クエリで、JOIN 先テーブルの論理削除フィルタ（`deleted_at IS NULL`）が JOIN 条件に含まれていない。
GORM の global soft delete scope は JOIN 先には適用されないため、論理削除済みレコードを親に持つデータが一覧に混入するリスクがある。

## 影響範囲

| 対象ファイル | 行 | JOIN 先 | メソッド |
|------------|-----|--------|---------|
| `backend/internal/repository/billing_confirmation_repository.go` | 32 | `medical_records` | `FindByMedicalRecordID` |
| `backend/internal/repository/care_plan_item_repository.go` | 34 | `hospitalizations` | `ListByHospitalizationID` |
| `backend/internal/repository/care_plan_item_repository.go` | 49 | `hospitalizations` | `FindByID` |
| `backend/internal/repository/care_plan_item_repository.go` | 71 | `hospitalizations` | `Update` |
| `backend/internal/repository/care_plan_item_repository.go` | 85 | `hospitalizations` | `Delete` |
| `backend/internal/repository/hospitalization_repository.go` | 130 | `hospitalizations` | `CountCarePlanItemsByHospitalizationID` |
| `backend/internal/repository/clinical_plan_repository.go` | 55 | `medical_records` | `Update`（サブクエリ） |
| `backend/internal/repository/vaccination_repository.go` | 37 | `pets` | `FindAll`（ownerID フィルタ時） |
| `backend/internal/repository/examination_repository.go` | 38 | `pets` | `FindAll`（ownerID フィルタ時） |
| `backend/internal/repository/trimming_repository.go` | 40 | `pets` | `FindAll`（ownerID フィルタ時） |

## 現状コード

### `billing_confirmation_repository.go:32`
```go
Joins("JOIN medical_records ON medical_records.id = billing_confirmations.medical_record_id").
```

### `care_plan_item_repository.go:34,49,71,85`
```go
Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id").
```

### `hospitalization_repository.go:130`
```go
Joins("JOIN hospitalizations ON care_plan_items.hospitalization_id = hospitalizations.id").
```

### `clinical_plan_repository.go:55`（サブクエリ）
```go
Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
```

### `vaccination_repository.go:37`, `examination_repository.go:38`, `trimming_repository.go:40`
```go
q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id").Where("pets.owner_id = ?", *ownerID)
// examination_repository: exams.pet_id
// trimming_repository: trimming_records.pet_id
```

### 比較: 正しい実装（プロジェクト内参照実装）

`clinical_plan_repository.go:35`（FindByMedicalRecordID / Delete は正しく実装済み）
```go
Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id AND medical_records.deleted_at IS NULL").
```

## 修正方針

JOIN 条件の末尾に `AND {テーブル}.deleted_at IS NULL` を追加する。

### 1. `billing_confirmation_repository.go:32`
```go
Joins("JOIN medical_records ON medical_records.id = billing_confirmations.medical_record_id AND medical_records.deleted_at IS NULL").
```

### 2. `care_plan_item_repository.go:34,49,71,85`（4箇所同一パターン）
```go
Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
```

### 3. `hospitalization_repository.go:130`
```go
Joins("JOIN hospitalizations ON care_plan_items.hospitalization_id = hospitalizations.id AND hospitalizations.deleted_at IS NULL").
```

### 4. `clinical_plan_repository.go:55`（サブクエリ）
```go
Where("clinical_plans.id = ? AND clinical_plans.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
```

### 5. `vaccination_repository.go:37`
```go
q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
```

### 6. `examination_repository.go:38`
```go
q = q.Joins("JOIN pets ON pets.id = exams.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
```

### 7. `trimming_repository.go:40`
```go
q = q.Joins("JOIN pets ON pets.id = trimming_records.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
```

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — JOIN を含む repository メソッドのレビューチェックリスト
> **JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか**
> **JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれているか**

GORM の soft delete scope（`DeletedAt` フィールド）はプライマリテーブルにのみ自動適用される。JOIN 先には適用されないため、JOIN 条件に明示的に `AND {table}.deleted_at IS NULL` を追加しなければならない。

### プロジェクト内参照実装

正しい実装: `backend/internal/repository/clinical_plan_repository.go:35,68`
```go
Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id AND medical_records.deleted_at IS NULL")
```

## 優先度

**High** — 論理削除済みレコードを親に持つデータが API レスポンスに混入するリスク。`pets` の場合、論理削除済みペットを飼い主で検索すると意図せず削除済みペットのワクチン・検査・トリミング履歴が返りうる。

## 関連ファイル

- `backend/internal/repository/billing_confirmation_repository.go:32`
- `backend/internal/repository/care_plan_item_repository.go:34,49,71,85`
- `backend/internal/repository/hospitalization_repository.go:130`
- `backend/internal/repository/clinical_plan_repository.go:55`
- `backend/internal/repository/vaccination_repository.go:37`
- `backend/internal/repository/examination_repository.go:38`
- `backend/internal/repository/trimming_repository.go:40`
