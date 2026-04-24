# TASK-134: payment_method_master_repository — FindAll の GORM エラー処理が FromGORM でない

## 優先度
**Low**

## 対象ファイル
`backend/internal/repository/payment_method_master_repository.go`

---

## 問題: `FindAll` で `apperrors.Wrap` を使用しているが、規約は `apperrors.FromGORM`

### チェック項目
- **Repository のエラーハンドリング**: GORM エラーは必ず `apperrors.FromGORM(err, "resource_name", "")` で変換する。`apperrors.Wrap` は Service 層で使用するもの。

### 現状コード（repository.go L30–40）
```go
func (r *paymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
    var ms []model.PaymentMethodMaster
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("display_order ASC, id ASC").
        Find(&ms).Error
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to find payment methods")  // ❌ Wrap は Service 層のもの
    }
    return ms, nil
}
```

### 修正後コード
```go
func (r *paymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
    var ms []model.PaymentMethodMaster
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("display_order ASC, id ASC").
        Find(&ms).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "payment_method", "")  // ✅ Repository では FromGORM
    }
    return ms, nil
}
```

### 参照実装
```go
// cage_repository.go L35–38
if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
    return nil, apperrors.FromGORM(err, "cage", "")
}
```

---

## 備考
- `FindByID`, `Create`, `UpdateFields`, `Delete` はいずれも `apperrors.FromGORM` を正しく使用している。
- `CountUsageByID` は集計クエリのため `apperrors.Wrap` で適切。
- `FindAll` のみが例外的に `Wrap` を使っており統一が必要。
