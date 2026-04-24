# TASK-233: payment_method_master_repository.go — Delete フィルタパターンが他リポジトリと不統一

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/payment_method_master_repository.go`

## 問題概要
`Delete`（行78-79）が `clinicScope` を適用した後に `Delete(&model.PaymentMethodMaster{}, id)` と
GORM の位置引数パターンを使っている。他の全リポジトリは
`Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.X{})` の明示的パターンに統一されている。

GORM の `Delete(model, id)` は内部で PrimaryKey を WHERE に展開するが、
`clinicScope` による `clinic_id` 条件と組み合わせた場合、GORM バージョンによって
結合順が異なる可能性があり、レビューで意図が読みにくい。

## 現状コード（行76-79）

```go
func (r *paymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Delete(&model.PaymentMethodMaster{}, id)   // ❌ 位置引数パターン
```

## あるべき姿（他リポジトリとの統一）

```go
func (r *paymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        Delete(&model.PaymentMethodMaster{})       // ✅ 明示的 Where パターン
```

## 完了条件
- [ ] `Delete` を `Where("id = ?", id).Delete(&model.PaymentMethodMaster{})` パターンに変更
- [ ] `go test ./backend/internal/...` がパス
