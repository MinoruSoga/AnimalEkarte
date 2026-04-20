# TASK-209: reservation_staff_repository.go — staff_clinic_assignments JOIN に deleted_at IS NULL 欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/repository/reservation_staff_repository.go`

## 問題概要
`FindAllByClinicID`（行35-46付近）の JOIN 条件に `staff_clinic_assignments.deleted_at IS NULL` が含まれていない。
論理削除済みの割り当てレコードが JOIN にヒットし、退職済みスタッフが一覧に混入する可能性がある。

TASK-188（`staff_repository.go`）・TASK-201（`staff_clinic_assignment_repository.go`）と同様のパターン。

```go
// 現状（NG）— 行35-46付近
Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ?", clinicID)
// ↑ sca.deleted_at IS NULL が欠落
```

```go
// あるべき姿
Joins("JOIN staff_clinic_assignments sca ON sca.staff_id = staffs.id AND sca.clinic_id = ? AND sca.deleted_at IS NULL", clinicID)
```

## 完了条件
- [ ] `FindAllByClinicID` の JOIN 条件に `AND sca.deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
