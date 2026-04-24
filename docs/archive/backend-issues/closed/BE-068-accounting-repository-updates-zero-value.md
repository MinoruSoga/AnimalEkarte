# BE-068: accounting_repository の Update がモデル全体を渡しており GORM ゼロ値問題がある

**Status**: Closed
**Priority**: Medium
**Affects**: `backend/internal/repository/accounting_repository.go`
**Date Created**: 2026-03-26
**Related**: -

## Summary

`accounting_repository.go` の `Update` メソッドが `*model.Billing` をそのまま `Updates()` に渡している。GORM の `Updates` はゼロ値フィールド（`0`, `""`, `false`）をスキップするため、意図的にゼロ値へ更新したい場合に効かない。プロジェクト規約の `buildXxxUpdateFields() → map[string]any` パターンが未適用。

## 現状のコード

```go
// backend/internal/repository/accounting_repository.go:116-129
func (r *accountingRepository) Update(ctx context.Context, clinicID, billingID uint64, accounting *model.Billing) (*model.Billing, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Where("clinic_id = ? AND id = ?", clinicID, billingID).
		Updates(accounting).  // ← モデル全体を渡している（ゼロ値スキップ問題）
		First(accounting)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, fmt.Sprintf("update billing id=%d", billingID))
	}
	return accounting, nil
}
```

## 必要な変更

### 1. service 側で UpdateInput DTO を定義

```go
// backend/internal/service/accounting_service.go
type UpdateAccountingInput struct {
	Status      *model.BillingStatus
	Notes       *string
	DiscountAmount *int64
	// 更新対象フィールドのみポインタ型で定義
}
```

### 2. buildBillingUpdateFields を service に実装

```go
func buildBillingUpdateFields(input UpdateAccountingInput) map[string]any {
	fields := make(map[string]any)
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	return fields
}
```

### 3. repository のシグネチャを map[string]any に変更

```go
// backend/internal/repository/accounting_repository.go
UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)

func (r *accountingRepository) UpdateFields(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	var billing model.Billing
	if err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Where("clinic_id = ? AND id = ?", clinicID, billingID).
		Updates(fields).
		First(&billing, "clinic_id = ? AND id = ?", clinicID, billingID).Error; err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("update billing id=%d", billingID))
	}
	return &billing, nil
}
```

## 完了条件

- [ ] `UpdateAccountingInput` DTO を service 層に定義
- [ ] `buildBillingUpdateFields()` を service 層に実装
- [ ] repository の `Update` を `UpdateFields(fields map[string]any)` に変更
- [ ] インターフェース定義も更新
- [ ] `docker compose exec backend go test ./... -v` がパス
