# TASK-252: vaccine_repository.go — CountUsageByVaccineID に deleted_at IS NULL が欠落

## 優先度
High

## 対象ファイル
- `backend/internal/repository/vaccine_repository.go`（CountUsageByVaccineID、行98付近）

## 問題概要
TASK-240/244 と同カテゴリのパターン。`CountUsageByVaccineID` が `vaccinations` テーブルを
カウントする際に `deleted_at IS NULL` フィルタを設定していない。

`model.Vaccination` は `gorm.DeletedAt` による論理削除を持つが、
論理削除済みの `vaccinations` レコードがカウントに含まれるため、
ワクチンマスタの削除が不当にブロックされる可能性がある。

## 現状コード（行98付近）

```go
func (r *vaccineRepository) CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Vaccination{}).
        Scopes(clinicScope(clinicID)).
        Where("vaccine_id = ?", vaccineID).       // ❌ deleted_at IS NULL なし
        Count(&count).Error; err != nil {
        return 0, apperrors.Wrap(err, "failed to count vaccine usage")
    }
    return count, nil
}
```

## あるべき姿

```go
Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID)
```

## 完了条件
- [ ] `WHERE` 条件に `AND deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
