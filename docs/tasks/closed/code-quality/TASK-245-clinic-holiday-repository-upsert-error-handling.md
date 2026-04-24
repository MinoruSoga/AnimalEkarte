# TASK-245: clinic_holiday_repository.go — Upsert の ErrRecordNotFound 判定不備 + RowsAffected 未確認

## 優先度
High

## 対象ファイル
- `backend/internal/repository/clinic_holiday_repository.go`

## 問題概要

### 問題1: 全エラーを「レコードなし」として処理している
`First()` の結果を `if err != nil` で一括判定し、全てのエラーを
「レコードが存在しないので Create する」として扱っている。
DB 接続タイムアウト・制約違反など他のエラーでも Create が実行され、
意図しないレコード生成が起きる可能性がある。

### 問題2: Update 後の RowsAffected を確認していない
`Update(...)` 実行後、`RowsAffected == 0` のチェックがない。
`existing.ID` で特定したレコードが Update 直前に削除されていた場合、
0件更新が「成功」として返される。

## 現状コード（行47〜70付近）

```go
// ❌ 問題1: 全エラーを not found として扱う
err := r.db.WithContext(ctx).
    Scopes(clinicScope(holiday.ClinicID)).
    Where("date = ?", holiday.Date).
    First(&existing).Error

if err != nil {
    // レコードなし → 新規作成（接続エラーでもここに来る）
    if err := r.db.WithContext(ctx).Create(holiday).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinic_holiday", ...)
    }
    return holiday, nil
}

// ❌ 問題2: RowsAffected なし
if err := r.db.WithContext(ctx).
    Model(&model.ClinicHoliday{}).
    Where("id = ?", existing.ID).
    Update("reason", holiday.Reason).Error; err != nil {
    return nil, apperrors.FromGORM(err, "clinic_holiday", ...)
}
existing.Reason = holiday.Reason
return holiday, nil  // 0件更新でも success を返す
```

## あるべき姿

```go
// ✅ 問題1修正: ErrRecordNotFound のみ新規作成
err := r.db.WithContext(ctx).
    Scopes(clinicScope(holiday.ClinicID)).
    Where("date = ?", holiday.Date).
    First(&existing).Error

if err != nil {
    if !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
    }
    // レコードなし → 新規作成
    if err := r.db.WithContext(ctx).Create(holiday).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
    }
    return holiday, nil
}

// ✅ 問題2修正: RowsAffected チェック追加
result := r.db.WithContext(ctx).
    Scopes(clinicScope(holiday.ClinicID)).
    Model(&model.ClinicHoliday{}).
    Where("id = ?", existing.ID).
    Update("reason", holiday.Reason)
if result.Error != nil {
    return nil, apperrors.FromGORM(result.Error, "clinic_holiday", holiday.Date.Format("2006-01-02"))
}
if result.RowsAffected == 0 {
    return nil, apperrors.WrapNotFound("clinic_holiday", holiday.Date.Format("2006-01-02"))
}
existing.Reason = holiday.Reason
return &existing, nil
```

## 完了条件
- [ ] `First()` のエラー判定を `errors.Is(err, gorm.ErrRecordNotFound)` に変更
- [ ] `Update()` 後に `RowsAffected == 0` チェックを追加
- [ ] `go test ./backend/internal/...` がパス
