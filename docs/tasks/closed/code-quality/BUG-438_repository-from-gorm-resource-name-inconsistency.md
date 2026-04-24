# BUG-438: repository の FromGORM resource_name がメソッド内で不統一（exam_type, checkup_type）

## 概要

`exam_type_repository.go` と `checkup_type_repository.go` の `CountUsageByXxxTypeID` メソッドで
`apperrors.FromGORM` に渡す resource_name が、同一ファイル内の他メソッドと不統一になっている。

## 問題箇所

### exam_type_repository.go:99

```go
// CountUsageByExamTypeID
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Examination{}).
        Where("exam_type_id = ? AND clinic_id = ?", examTypeID, clinicID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination_record", "")  // ← "examination_record"
    }
    return count, nil
}
```

同ファイル内の他メソッドでは `"exam_type"` を使用しているのに、
CountUsage のみ `"examination_record"` を使用しており不統一。

### checkup_type_repository.go:99

```go
// CountUsageByCheckupTypeID
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Checkup{}).
        Where("checkup_type_id = ? AND clinic_id = ?", checkupTypeID, clinicID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "checkup_record", "")  // ← "checkup_record"
    }
    return count, nil
}
```

同ファイル内では `"checkup_type"` を使用しているのに、
CountUsage のみ `"checkup_record"` を使用しており不統一。

## 修正方針

CountUsage メソッドの resource_name を同ファイル内の他メソッドと統一する。

```go
// exam_type_repository.go — 修正後
return 0, apperrors.FromGORM(err, "exam_type", "")

// checkup_type_repository.go — 修正後
return 0, apperrors.FromGORM(err, "checkup_type", "")
```

**補足**: `CountUsage` メソッドがカウント対象のテーブル（`examinations`, `checkups`）を
クエリしているため `"examination_record"` / `"checkup_record"` と命名したと思われるが、
`FromGORM` の resource_name は「操作対象となるリソース（エンティティのオーナー）」を
表すべきであり、マスタの名前 `"exam_type"` / `"checkup_type"` が正しい。

## 影響ファイル

- `backend/internal/repository/exam_type_repository.go` — 行 99
- `backend/internal/repository/checkup_type_repository.go` — 行 99

## 優先度

**Low** — 機能影響なし。エラーメッセージの resource_name 統一のみ。
