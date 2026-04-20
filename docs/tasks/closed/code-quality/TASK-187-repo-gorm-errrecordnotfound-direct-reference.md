# TASK-187: Repository 層 — gorm.ErrRecordNotFound 直接参照を apperrors パターンへ統一

## 優先度
High

## 対象ファイル（4箇所）
- `backend/internal/repository/clinic_settings_repository.go`（行31）
- `backend/internal/repository/closing_special_period_repository.go`（行63）
- `backend/internal/repository/reservation_staff_repository.go`（行137）
- `backend/internal/repository/reservation_type_liff_repository.go`（行106）

## 問題概要
プロジェクト規約「`gorm.ErrRecordNotFound` を Repository 層で直接 `errors.Is` でチェックしない
（`FromGORM` に委ねる）」に対して、上記4箇所が規約に違反している。

`gorm` パッケージへの直接依存が Repository 実装の制御フローに残ることで、
将来 ORM を交換する際の変更コストが増加する。

## 該当箇所と修正方針

### `clinic_settings_repository.go`（デフォルト値返却パターン）
```go
// 現状（NG）
if errors.Is(err, gorm.ErrRecordNotFound) {
    return &model.ClinicSettings{...}, nil  // デフォルト値
}
if err != nil {
    return nil, apperrors.FromGORM(...)
}

// あるべき姿
if err != nil {
    wrapped := apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
    if errors.Is(wrapped, apperrors.ErrNotFound) {
        return &model.ClinicSettings{...}, nil  // デフォルト値
    }
    return nil, wrapped
}
```

### その他3箇所
同様に `FromGORM` → `errors.Is(wrapped, apperrors.ErrNotFound)` パターンに置換する。

## 完了条件
- [ ] 4ファイルすべてで `gorm.ErrRecordNotFound` の直接参照を除去
- [ ] `apperrors.FromGORM` → `errors.Is(wrapped, apperrors.ErrNotFound)` パターンで統一
- [ ] `go test ./backend/internal/...` がパス
