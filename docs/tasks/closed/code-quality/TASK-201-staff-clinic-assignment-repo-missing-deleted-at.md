# TASK-201: staff_clinic_assignment_repository.go — FindByStaffID に deleted_at IS NULL 欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/repository/staff_clinic_assignment_repository.go`

## 問題概要
`FindByStaffID`（行33）のクエリに `deleted_at IS NULL` 条件が欠落している。
`staff_clinic_assignments` テーブルに `deleted_at` カラムがある場合、
論理削除済みのレコードも返ることになる。

GORM の SoftDelete スコープ（`gorm.Model` 組み込み）が有効かどうかにもよるが、
他の repository では JOIN 条件に明示しているため不一致がある。

## 現状コード（行33）

```go
if err := dbOrTx(ctx, r.db).Where("staff_id = ?", staffID).Find(&assignments).Error; err != nil {
```

## あるべき姿

モデルに `gorm.DeletedAt` が含まれていない場合:
```go
if err := dbOrTx(ctx, r.db).
    Where("staff_id = ? AND deleted_at IS NULL", staffID).
    Find(&assignments).Error; err != nil {
```

GORM SoftDelete が有効な場合は自動フィルタが効くため追加不要。
まずモデル定義を確認すること。

## 完了条件
- [ ] `model.StaffClinicAssignment` の定義を確認（SoftDelete 有無）
- [ ] 必要な場合 `deleted_at IS NULL` 条件を追加
- [ ] `go test ./backend/internal/...` がパス
