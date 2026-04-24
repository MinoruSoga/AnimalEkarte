# TASK-103: `buildUpdateFields` ヘルパー関数欠如（payment_method_master / closing_settings）

## 優先度

**Medium** — プロジェクト規約違反。他マスタとの統一実装パターンが崩れている。

---

## 概要

`payment_method_master_service.go` の `Update` メソッドと
`closing_settings_service.go` の `UpdateSpecialPeriod` メソッドが、
`map[string]any` の構築をメソッド本体にインラインで記述している。

プロジェクト規約（`.claude/rules/go-language.md`）では PATCH 系更新に
`buildXxxUpdateFields()` ヘルパー関数を使うパターンを標準としており、
`medicine_service.go`, `vaccine_service.go`, `cage_service.go`, `procedure_service.go`
はすべてこのパターンに準拠している。

---

## 問題箇所

### `service/payment_method_master_service.go:59-72`

```go
// ❌ ヘルパーなしでインライン構築
func (s *paymentMethodMasterService) Update(...) (*model.PaymentMethodMaster, error) {
    fields := map[string]any{}
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    if input.DisplayOrder != nil {
        fields["display_order"] = *input.DisplayOrder
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive
    }
    if len(fields) == 0 {
        return s.repo.FindByID(ctx, clinicID, id)
    }
    return s.repo.UpdateFields(ctx, clinicID, id, fields)
}
```

### `service/closing_settings_service.go:189-207`

```go
// ❌ UpdateSpecialPeriod にインライン構築
fields := map[string]any{}
if input.StartDate != nil {
    fields["start_date"] = *input.StartDate
}
if input.EndDate != nil {
    fields["end_date"] = *input.EndDate
}
...
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/cage_service.go:100,163
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    fields := buildCageUpdateFields(input)
    ...
}

func buildCageUpdateFields(input *UpdateCageInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    ...
    return fields
}
```

---

## 修正方針

### 1. `service/payment_method_master_service.go`

インラインの map 構築を `buildPaymentMethodUpdateFields` として抽出する。

```go
// ✅ 修正後
func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
    fields := buildPaymentMethodUpdateFields(input)
    if len(fields) == 0 {
        return s.repo.FindByID(ctx, clinicID, id)
    }
    return s.repo.UpdateFields(ctx, clinicID, id, fields)
}

func buildPaymentMethodUpdateFields(input UpdatePaymentMethodInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    if input.DisplayOrder != nil {
        fields["display_order"] = *input.DisplayOrder
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive
    }
    return fields
}
```

### 2. `service/closing_settings_service.go`

`UpdateSpecialPeriod` のインライン map 構築を `buildSpecialPeriodUpdateFields` として抽出する。

```go
// ✅ 修正後
func buildSpecialPeriodUpdateFields(input UpdateSpecialPeriodInput) map[string]any {
    fields := make(map[string]any)
    if input.StartDate != nil {
        fields["start_date"] = *input.StartDate
    }
    if input.EndDate != nil {
        fields["end_date"] = *input.EndDate
    }
    if input.AmPmBoundary != nil {
        fields["am_pm_boundary"] = *input.AmPmBoundary
    }
    if input.PmEnd != nil {
        fields["pm_end"] = *input.PmEnd
    }
    if input.Note != nil {
        fields["note"] = *input.Note
    }
    return fields
}
```

---

## 影響範囲

| ファイル | 対象メソッド | 状態 |
|---------|------------|------|
| `service/payment_method_master_service.go:59-72` | Update | ❌ インライン map |
| `service/closing_settings_service.go:189-207` | UpdateSpecialPeriod | ❌ インライン map |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — GORM PATCH（ポインタ型 + buildUpdateFields）

> ```go
> func buildOwnerUpdateFields(input UpdateOwnerInput) map[string]any {
>     fields := make(map[string]any)
>     if input.Name != nil { fields["name"] = *input.Name }
>     ...
>     return fields
> }
> ```

### プロジェクト内参照実装

- `service/cage_service.go:163` — `buildCageUpdateFields`
- `service/medicine_service.go:68` — `buildMedicineUpdateFields`
- `service/vaccine_service.go:141` — `buildVaccineUpdateFields`
- `service/procedure_service.go:200` — `buildProcedureUpdateFields`
