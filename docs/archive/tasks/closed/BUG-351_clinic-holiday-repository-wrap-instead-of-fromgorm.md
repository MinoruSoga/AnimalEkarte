# BUG-351: clinic_holiday_repository が GORM エラーに apperrors.Wrap を誤用

## 概要

`clinic_holiday_repository.go` の `FindByYearMonth`（L40）と `Upsert`（L63）で、
GORM エラーを `apperrors.FromGORM` ではなく `apperrors.Wrap` で処理している。
規約「Repository: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換」に違反。

## 違反箇所

```go
// backend/internal/repository/clinic_holiday_repository.go:40（バグ）
if err := q.Find(&holidays).Error; err != nil {
    return nil, apperrors.Wrap(err, "failed to list clinic holidays") // → FromGORM に変更すべき
}

// 同:63（バグ）
return nil, apperrors.Wrap(result.Error, "failed to upsert clinic holiday") // → FromGORM に変更すべき
```

同ファイルの `FindByDate` は `apperrors.FromGORM` を正しく使っており、一貫性がない。

## 修正内容

```go
// FindByYearMonth
if err := q.Find(&holidays).Error; err != nil {
    return nil, apperrors.FromGORM(err, "clinic_holiday", fmt.Sprintf("clinic_id=%d,year=%d,month=%d", clinicID, year, month))
}

// Upsert
if result.Error != nil {
    return nil, apperrors.FromGORM(result.Error, "clinic_holiday", fmt.Sprintf("clinic_id=%d,date=%s", clinicID, holiday.Date))
}
```

## 影響

- 機能的影響は限定的（`Wrap` でも上位層の `RespondError` は動作する）
- `gorm.ErrRecordNotFound` を NotFound として正しく分類できない
- `pgconn.PgError` による詳細分類（UniqueViolation 等）が行われない

## 優先度

**MEDIUM** — 規約違反。機能バグではないが、エラー分類精度への影響がある。

## 関連ファイル

- `backend/internal/repository/clinic_holiday_repository.go:40,63`
