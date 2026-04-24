# TASK-240: medicine_repository / procedure_repository — CountUsage クエリの care_plan_items に deleted_at IS NULL が欠落

## 優先度
High

## 対象ファイル
- `backend/internal/repository/medicine_repository.go`（行75付近）
- `backend/internal/repository/procedure_repository.go`（行97付近）

## 問題概要
FK 削除チェック用の `CountUsage*` メソッドで、`care_plan_items` テーブルへのカウントクエリに
`care_plan_items.deleted_at IS NULL` フィルタが欠落している。

論理削除済みの `care_plan_items` レコードがカウントに含まれ、
削除済みの参照によって削除が不当にブロックされる可能性がある。

## 現状コード

### medicine_repository.go（行75付近）
```go
if err := r.db.WithContext(ctx).
    Model(&model.CarePlanItem{}).
    Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id"+
        " AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
    Where("care_plan_items.medicine_id = ?", medicineID).  // ❌ deleted_at IS NULL なし
    Count(&carePlanCount).Error; err != nil {
```

### procedure_repository.go（行97付近）
```go
if err := r.db.WithContext(ctx).
    Model(&model.CarePlanItem{}).
    Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id"+
        " AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
    Where("care_plan_items.procedure_id = ?", procedureID).  // ❌ deleted_at IS NULL なし
    Count(&carePlanCount).Error; err != nil {
```

## 比較（正しい treatments カウント）
同ファイル内の treatments カウントクエリ:
```go
Where("treatments.medicine_id = ? AND treatments.deleted_at IS NULL", medicineID)  // ✅
```

## あるべき姿

```go
// medicine_repository.go
Where("care_plan_items.medicine_id = ? AND care_plan_items.deleted_at IS NULL", medicineID)

// procedure_repository.go
Where("care_plan_items.procedure_id = ? AND care_plan_items.deleted_at IS NULL", procedureID)
```

## 完了条件
- [ ] `medicine_repository.go` の care_plan_items カウントに `AND care_plan_items.deleted_at IS NULL` を追加
- [ ] `procedure_repository.go` の care_plan_items カウントに `AND care_plan_items.deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
