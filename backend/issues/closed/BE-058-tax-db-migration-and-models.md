# BE-058: 消費税区分 DB migration + Go モデル変更 + codegen

**Status**: Closed
**Priority**: High
**Affects**: DB schema, 全モデル（clinic, consultation, procedure, medicine, merchandise_item, accounting）
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-059, BE-060, BE-061

## Summary

消費税区分（内税/外税/非課税）を保持する ENUM と関連カラムを DB に追加し、Go モデルを更新する。
`make codegen` で `models.ts` を再生成し、フロントエンドの型ベースを確立する。

## 現状のコード

```sql
-- backend/migrations/001_init.sql 現在の税関連
-- billing_items.tax_rate (default 0.10) -- 存在する
-- estimate_items.tax_rate (default 0.10) -- 存在する
-- merchandise_items.tax_rate (default 0.10) -- 存在する
-- consultations.price -- tax_rate/tax_type なし
-- procedures.price -- tax_rate/tax_type なし
-- medicines.price -- tax_rate/tax_type なし
-- clinics -- standard_tax_rate/reduced_tax_rate なし
-- TaxType ENUM -- 存在しない
```

```go
// backend/internal/model/merchandise_item.go:15-16
TaxRate float64 `gorm:"type:numeric;not null;default:0.10" json:"tax_rate"`
// TaxType フィールドなし

// backend/internal/model/accounting.go:83
TaxRate float64 `gorm:"type:numeric(3,2);default:0.10" json:"tax_rate"`
// BillingItem.TaxType フィールドなし

// backend/internal/model/clinic.go
// StandardTaxRate/ReducedTaxRate フィールドなし
```

## 必要な変更

### 1. DB マイグレーション（001_init.sql）

**新規 ENUM 追加**（clinics テーブル定義より前に挿入）:
```sql
CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt');
-- included = 内税, excluded = 外税, exempt = 非課税
```

**clinics テーブルへのカラム追加**:
```sql
-- clinics テーブルの既存カラム末尾（updated_at の前）に追加
standard_tax_rate  numeric  NOT NULL DEFAULT 0.10,
reduced_tax_rate   numeric  NOT NULL DEFAULT 0.08,
```

**consultations テーブルへのカラム追加**:
```sql
-- sort_order の前に追加
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
tax_rate  numeric   NOT NULL DEFAULT 0.10,
```

**procedures テーブルへのカラム追加**:
```sql
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
tax_rate  numeric   NOT NULL DEFAULT 0.10,
```

**medicines テーブルへのカラム追加**:
```sql
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
tax_rate  numeric   NOT NULL DEFAULT 0.10,
```

**hospitalization_plans テーブルへのカラム追加**:
```sql
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
tax_rate  numeric   NOT NULL DEFAULT 0.10,
```

**merchandise_items テーブルへのカラム追加**:
```sql
-- tax_rate は既存（DEFAULT 0.10）のため追加不要
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
```

**billing_items テーブルへのカラム追加**:
```sql
-- tax_rate は既存（DEFAULT 0.10）のため追加不要
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
```

**estimate_items テーブルへのカラム追加**:
```sql
-- tax_rate は既存（DEFAULT 0.10）のため追加不要
tax_type  tax_type  NOT NULL DEFAULT 'excluded',
```

### 2. Go モデル変更

**新規 TaxType 型の定義**（`backend/internal/model/accounting.go` または新規共通ファイルに追加）:

```go
// backend/internal/model/accounting.go の先頭部分に追加
type TaxType string

const (
    TaxTypeIncluded TaxType = "included" // 内税
    TaxTypeExcluded TaxType = "excluded" // 外税
    TaxTypeExempt   TaxType = "exempt"   // 非課税
)
```

**model/clinic.go の変更**:
```go
// Before: StandardTaxRate, ReducedTaxRate なし
// After: 以下を追加（既存フィールドの末尾、CreatedAt の前）
StandardTaxRate float64 `gorm:"type:numeric;not null;default:0.10" json:"standard_tax_rate"`
ReducedTaxRate  float64 `gorm:"type:numeric;not null;default:0.08" json:"reduced_tax_rate"`
```

**model/consultation.go の変更**:
```go
// Before: tax_rate/tax_type なし
// After: 以下を追加（SortOrder の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric;not null;default:0.10"     json:"tax_rate"`
```

**model/procedure.go の変更**:
```go
// 同上（SortOrder の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric;not null;default:0.10"     json:"tax_rate"`
```

**model/medicine.go の変更**:
```go
// 同上（SortOrder の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric;not null;default:0.10"     json:"tax_rate"`
```

**model/hospitalization_plan.go の変更**:
```go
// 同上（SortOrder の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric;not null;default:0.10"     json:"tax_rate"`
```

**model/merchandise_item.go の変更**:
```go
// Before:
TaxRate float64 `gorm:"type:numeric;not null;default:0.10" json:"tax_rate"`
// After: TaxType を追加（TaxRate の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric;not null;default:0.10"     json:"tax_rate"`
```

**model/accounting.go の BillingItem への変更**:
```go
// Before:
TaxRate float64 `gorm:"type:numeric(3,2);default:0.10" json:"tax_rate"`
// After: TaxType を追加（TaxRate の前）
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
TaxRate float64 `gorm:"type:numeric(3,2);default:0.10"         json:"tax_rate"`
```

**model/accounting.go の EstimateItem への変更**:
```go
// BillingItem と同様に TaxType を追加
TaxType TaxType `gorm:"type:tax_type;not null;default:excluded" json:"tax_type"`
```

### 3. codegen 実行

```bash
make codegen
# → frontend/src/types/generated/models.ts が自動更新される
# TaxType 型・StandardTaxRate/ReducedTaxRate・各モデルの TaxType フィールドが追加される
```

## フロントエンド影響

- `make codegen` 後に `models.ts` が更新される
- `TaxType = "included" | "excluded" | "exempt"` が生成される
- 型エラーが出る場合は FE-121〜FE-123 で対応

## 完了条件

- [x] `tax_type` ENUM が 001_init.sql に定義されている
- [x] 8 テーブル（clinics, consultations, procedures, medicines, hospitalization_plans, merchandise_items, billing_items, estimate_items）に正しいカラムが追加されている
- [x] Go モデル 7 ファイルが更新されている
- [x] `TaxType` 型が定義されている
- [x] `make codegen` が成功し `models.ts` に TaxType が含まれる
- [x] `docker compose exec backend go build ./...` が通る
- [x] `schema_drift_test.go` が通る（DBスキーマとモデルの整合性チェック）

## クローズ情報

- **Closed At**: 2026-03-25
- **変更ファイル**:
  - `backend/migrations/001_init.sql` — `tax_type` ENUM 追加 + 8 テーブルにカラム追加
  - `backend/internal/model/accounting.go` — `TaxType` 型定義 + `BillingItem.TaxType` 追加
  - `backend/internal/model/estimate.go` — `EstimateItem.TaxType` 追加
  - `backend/internal/model/clinic.go` — `StandardTaxRate`, `ReducedTaxRate` 追加
  - `backend/internal/model/consultation.go` — `TaxType`, `TaxRate` 追加
  - `backend/internal/model/procedure.go` — `TaxType`, `TaxRate` 追加
  - `backend/internal/model/medicine.go` — `TaxType`, `TaxRate` 追加
  - `backend/internal/model/hospitalization_plan.go` — `TaxType`, `TaxRate` 追加
  - `backend/internal/model/merchandise_item.go` — `TaxType` 追加
  - `frontend/src/types/generated/models.ts` — `make codegen` で自動更新
