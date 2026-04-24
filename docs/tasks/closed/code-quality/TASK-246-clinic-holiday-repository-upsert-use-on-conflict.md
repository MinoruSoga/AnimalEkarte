# TASK-246: clinic_holiday_repository.go — Upsert を clause.OnConflict パターンに統一

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/clinic_holiday_repository.go`

## 問題概要
`Upsert` が手動の `First() → Create/Update` 分岐パターンを使っているが、
同プロジェクト内の `clinic_settings_repository.go` は PostgreSQL ネイティブの
`clause.OnConflict` を使ったアトミックな UPSERT を実装している。

手動パターンは First → Create の間に別プロセスが同じレコードを INSERT した場合、
UNIQUE 制約違反が発生するレースコンディションを持つ。

## 現状コード
```go
// First で検索 → 結果により Create または Update を分岐
First(&existing) → err != nil → Create(holiday)
                 → err == nil → Update("reason", holiday.Reason)
```

## 比較（正しい実装例: clinic_settings_repository.go）
```go
err := r.db.WithContext(ctx).
    Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "clinic_id"}},
        DoUpdates: clause.AssignmentColumns([]string{"open_time", "close_time", ..., "updated_at"}),
    }).
    Create(settings).Error
```

## あるべき姿

`clinic_holidays` テーブルの一意制約キー（`clinic_id` + `date`）を使って OnConflict で統一する。

```go
func (r *clinicHolidayRepository) Upsert(ctx context.Context, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
    err := r.db.WithContext(ctx).
        Clauses(clause.OnConflict{
            Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "date"}},
            DoUpdates: clause.AssignmentColumns([]string{"reason", "updated_at"}),
        }).
        Create(holiday).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
    }
    return holiday, nil
}
```

## 確認事項
- `clinic_holidays` テーブルに `(clinic_id, date)` のユニーク制約が存在することを確認してから実装
- OnConflict 実装後は TASK-245 の修正が不要になる（競合状態ごと解消）

## 完了条件
- [ ] Upsert を `clause.OnConflict` パターンに書き換え
- [ ] TASK-245 の手動パターン修正が不要になったことを確認
- [ ] `go test ./backend/internal/...` がパス
