# TASK-237: reservation_staff_repository.go — SwapSortOrder の Update 2箇所に clinicScope が欠落 [CRITICAL]

## 優先度
Critical

## 対象ファイル
- `backend/internal/repository/reservation_staff_repository.go`

## 問題概要
`SwapSortOrder` メソッド内のトランザクション Update 2箇所が `Where("id = ?", ...)` のみで
`clinicScope` を持たない。これにより、異なるクリニックの `sort_order` を書き換える可能性がある。

マルチテナント規約: **全更新クエリに必ず `clinicScope(clinicID)` を適用する。**

## 現状コード（行138付近）

```go
func (r *reservationStaffRepository) SwapSortOrder(ctx context.Context, clinicID, targetID, neighborID uint64) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // ...
        if err := tx.Model(&model.Staff{}).
            Where("id = ?", target.ID).            // ❌ clinicScope なし
            Update("sort_order", neighborOrder).Error; err != nil {
            return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", target.ID))
        }
        if err := tx.Model(&model.Staff{}).
            Where("id = ?", neighbor.ID).          // ❌ clinicScope なし
            Update("sort_order", targetOrder).Error; err != nil {
            return apperrors.FromGORM(err, "reservation_staff", fmt.Sprintf("%d", neighbor.ID))
        }
        return nil
    })
}
```

## あるべき姿

```go
if err := tx.Scopes(clinicScope(clinicID)).
    Model(&model.Staff{}).
    Where("id = ?", target.ID).
    Update("sort_order", neighborOrder).Error; err != nil {

if err := tx.Scopes(clinicScope(clinicID)).
    Model(&model.Staff{}).
    Where("id = ?", neighbor.ID).
    Update("sort_order", targetOrder).Error; err != nil {
```

## 完了条件
- [ ] SwapSortOrder 内の Update 2箇所に `Scopes(clinicScope(clinicID))` を追加
- [ ] `go test ./backend/internal/...` がパス
