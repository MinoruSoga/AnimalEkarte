# TASK-195: closing_settings_service.go — DeleteSpecialPeriod 存在確認欠落・UpdateSpecialPeriod slog 欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/closing_settings_service.go`

## 問題概要

### 問題1: DeleteSpecialPeriod に FindByID 存在確認なし（行224-229）
Delete 前に `periodRepo.FindByID` による存在確認がない。
存在しない ID を指定した場合に 404 ではなく別のエラーが返る可能性がある。
他の service（clinic_holiday_service 等）は Delete 前に `IsNotFound` チェックを行っており不一致。

```go
// 現状（NG）
func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
    slog.InfoContext(ctx, "deleting closing special period", ...)
    return s.periodRepo.Delete(ctx, clinicID, id)
    // ← 存在確認なし
}
```

```go
// あるべき姿
func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.periodRepo.FindByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to get special period")
    }
    slog.InfoContext(ctx, "deleting closing special period", ...)
    return s.periodRepo.Delete(ctx, clinicID, id)
}
```

### 問題2: UpdateSpecialPeriod に成功 slog が欠落（行222付近）
`CreateSpecialPeriod` や `DeleteSpecialPeriod` は slog.InfoContext を出力しているが、
`UpdateSpecialPeriod` だけ欠落しており操作追跡の一貫性が取れていない。

```go
// 修正後: return の前に slog を追加
slog.InfoContext(ctx, "special period updated", slog.Uint64("id", id))
return result, nil
```

## 完了条件
- [ ] `DeleteSpecialPeriod` に `FindByID` 存在確認を追加
- [ ] `UpdateSpecialPeriod` に成功 `slog.InfoContext` を追加
- [ ] `go test ./backend/internal/...` がパス
