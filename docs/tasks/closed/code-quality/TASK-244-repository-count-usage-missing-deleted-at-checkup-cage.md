# TASK-244: checkup_type_repository / cage_repository — CountUsage クエリに deleted_at IS NULL が欠落 [CRITICAL]

## 優先度
Critical

## 対象ファイル
- `backend/internal/repository/checkup_type_repository.go`（CountUsageByCheckupTypeID）
- `backend/internal/repository/cage_repository.go`（CountUsageByCageID）

## 問題概要
FK 依存チェック用 `CountUsage*` メソッドが論理削除済みレコードをカウントに含めている。
`model.Checkup` / `model.Hospitalization` は `gorm.DeletedAt` による論理削除を持つが、
WHERE 条件に `deleted_at IS NULL` が欠落しているため、
削除済みレコードが残っていると正常な削除がブロックされる。

既に同様の問題として TASK-240（medicine / procedure の CountUsage）が起票済みだが、
checkup_type と cage は別ファイルのため本タスクとして起票。

## 現状コード

### checkup_type_repository.go（行93〜103付近）
```go
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Checkup{}).
        Scopes(clinicScope(clinicID)).
        Where("checkup_type_id = ?", checkupTypeID).  // ❌ deleted_at IS NULL なし
        Count(&count).Error; err != nil {
```

### cage_repository.go（行88〜98付近）
```go
func (r *cageRepository) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Hospitalization{}).
        Scopes(clinicScope(clinicID)).
        Where("cage_id = ?", id).  // ❌ deleted_at IS NULL なし
        Count(&count).Error; err != nil {
```

## あるべき姿

```go
// checkup_type_repository.go
Where("checkup_type_id = ? AND deleted_at IS NULL", checkupTypeID)

// cage_repository.go
Where("cage_id = ? AND deleted_at IS NULL", id)
```

## 完了条件
- [ ] `CountUsageByCheckupTypeID` に `AND deleted_at IS NULL` を追加
- [ ] `CountUsageByCageID` に `AND deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
