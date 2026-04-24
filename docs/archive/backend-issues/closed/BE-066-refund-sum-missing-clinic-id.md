# BE-066: SumByBillingID に clinic_id フィルタなし（クロステナント集計リスク）

**Status**: Closed
**Priority**: High
**Affects**: `backend/internal/repository/refund_repository.go`
**Date Created**: 2026-03-26
**Related**: -

## Summary

`refund_repository.go` の `SumByBillingID` メソッドが `billing_id` のみを条件にしており、`clinic_id` フィルタが欠落している。`FindByBillingID` には `clinic_id = ?` が存在するのに `SumByBillingID` にはない不整合。マルチテナント設計の穴であり、異なるクリニックの返金額を合算してしまうリスクがある。

## 現状のコード

```go
// backend/internal/repository/refund_repository.go:47-57
func (r *refundRepository) SumByBillingID(ctx context.Context, billingID uint64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.BillingRefund{}).
		Where("billing_id = ?", billingID).  // ← clinic_id なし
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.Wrap(err, fmt.Sprintf("sum refunds for billing_id=%d", billingID))
	}
	return total, nil
}
```

インターフェース定義（`refund_repository.go:17`）も `clinicID` 引数がない:

```go
SumByBillingID(ctx context.Context, billingID uint64) (int64, error)
```

## 必要な変更

### 1. Repository インターフェース変更

```go
// backend/internal/repository/refund_repository.go:17
SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error)
```

### 2. Repository 実装変更

```go
func (r *refundRepository) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.BillingRefund{}).
		Where("clinic_id = ? AND billing_id = ?", clinicID, billingID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, apperrors.Wrap(err, fmt.Sprintf("sum refunds for billing_id=%d", billingID))
	}
	return total, nil
}
```

### 3. Service 呼び出し元の変更

`refund_service.go:42` の呼び出しを `clinicID` 渡しに変更:

```go
// 変更前
alreadyRefunded, err := s.repo.SumByBillingID(ctx, billingID)

// 変更後
alreadyRefunded, err := s.repo.SumByBillingID(ctx, clinicID, billingID)
```

## 完了条件

- [ ] `SumByBillingID` シグネチャに `clinicID uint64` 引数を追加
- [ ] WHERE 句に `clinic_id = ? AND billing_id = ?` を設定
- [ ] `refund_service.go` の呼び出し元を更新
- [ ] `docker compose exec backend go test ./... -v` がパス
