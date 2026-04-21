# TASK-236: clinic_holiday_repository.go — Upsert の Update 呼び出しに clinicScope が欠落 [CRITICAL]

## 優先度
Critical

## 対象ファイル
- `backend/internal/repository/clinic_holiday_repository.go`

## 問題概要
`Upsert` メソッド内の既存レコード更新パスで、`clinicScope` が適用されていない。
`Where("id = ?", existing.ID)` のみで更新しており、`existing` が正しく取得されていたとしても、
更新クエリ自体にテナント境界が設定されていない。

マルチテナント規約: **全更新・削除クエリに必ず `clinicScope(clinicID)` を適用する。**

## 現状コード（行63付近）

```go
func (r *clinicHolidayRepository) Upsert(ctx context.Context, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
    // ...
    if err := r.db.WithContext(ctx).
        Model(&model.ClinicHoliday{}).
        Where("id = ?", existing.ID).              // ❌ clinicScope なし
        Update("reason", holiday.Reason).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinic_holiday", fmt.Sprintf("%d", existing.ID))
    }
    // ...
}
```

## あるべき姿

```go
if err := r.db.WithContext(ctx).
    Scopes(clinicScope(holiday.ClinicID)).         // ✅ clinicScope を追加
    Model(&model.ClinicHoliday{}).
    Where("id = ?", existing.ID).
    Update("reason", holiday.Reason).Error; err != nil {
    return nil, apperrors.FromGORM(err, "clinic_holiday", fmt.Sprintf("%d", existing.ID))
}
```

## 完了条件
- [ ] Upsert の Update 呼び出しに `Scopes(clinicScope(holiday.ClinicID))` を追加
- [ ] `go test ./backend/internal/...` がパス
