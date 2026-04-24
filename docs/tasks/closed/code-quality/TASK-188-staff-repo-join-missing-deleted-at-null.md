# TASK-188: staff_repository.go — staff_clinic_assignments JOIN 条件に deleted_at IS NULL 欠落

## 優先度
High

## 対象ファイル
`backend/internal/repository/staff_repository.go`

## 問題概要
`FindAll`・`Update`・`Delete` の JOIN 条件で `staff_clinic_assignments.deleted_at IS NULL`
チェックが欠落しており、論理削除済みの割り当てレコードが JOIN にヒットしてしまう。

### `FindAll`（行38-41）
退職済みスタッフ（`staff_clinic_assignments.deleted_at` が非NULL）が
一覧に混入する可能性がある。

### `Update`・`Delete`（行93・111）
EXISTS サブクエリで削除済み割り当てを使って操作許可判定をしてしまう。

## 現状コード

```go
// FindAll（行38-41）
q := dbOrTx(ctx, r.db).Model(&model.Staff{}).
    Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id").
    Where("staff_clinic_assignments.clinic_id = ?", clinicID).
    Where("staffs.deleted_at IS NULL")
```

## あるべき姿

```go
// FindAll
q := dbOrTx(ctx, r.db).Model(&model.Staff{}).
    Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id"+
        " AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
    Where("staffs.deleted_at IS NULL")
```

Update・Delete の EXISTS サブクエリも同様に `staff_clinic_assignments.deleted_at IS NULL` を追加する。

## 完了条件
- [ ] `FindAll` の JOIN 条件に `AND staff_clinic_assignments.deleted_at IS NULL` を追加
- [ ] `Update` の EXISTS サブクエリに `deleted_at IS NULL` を追加
- [ ] `Delete` の EXISTS サブクエリに `deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
