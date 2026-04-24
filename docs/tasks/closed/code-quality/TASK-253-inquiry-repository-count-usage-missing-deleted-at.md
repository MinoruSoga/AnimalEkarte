# TASK-253: inquiry_repository.go — CountByChiefComplaintTypeID に deleted_at IS NULL が欠落

## 優先度
High

## 対象ファイル
- `backend/internal/repository/inquiry_repository.go`（CountByChiefComplaintTypeID、行96付近）

## 問題概要
TASK-240/244/252 と同カテゴリのパターン。chief_complaint_type の削除前FK依存チェックで
使用される `CountByChiefComplaintTypeID` が、`inquiries` テーブルを
`deleted_at IS NULL` フィルタなしでカウントしている。

論理削除済みの `inquiries` が参照カウントに含まれ、
既に削除済みの問診票を参照しているだけの chief_complaint_type が
削除できない状態が続く可能性がある。

## 現状コード（行96付近）

```go
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Inquiry{}).
        Scopes(clinicScope(clinicID)).
        Where("chief_complaint_type_id = ?", categoryID).  // ❌ deleted_at IS NULL なし
        Count(&count).Error; err != nil {
        return 0, apperrors.Wrap(err, "failed to count chief complaint type usage")
    }
    return count, nil
}
```

## あるべき姿

```go
Where("chief_complaint_type_id = ? AND deleted_at IS NULL", categoryID)
```

## 完了条件
- [ ] `WHERE` 条件に `AND deleted_at IS NULL` を追加
- [ ] `go test ./backend/internal/...` がパス
