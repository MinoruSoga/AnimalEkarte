# TASK-202: closing_settings_service.go — DeleteSpecialPeriod の裸エラーreturn・slog 前置き問題

## 優先度
High

## 対象ファイル
`backend/internal/service/closing_settings_service.go`

## 問題概要

### 問題1: DeleteSpecialPeriod が apperrors.Wrap なしで生エラーを return

Service 層のエラーハンドリング規約「Service は内部エラーを `apperrors.Wrap` でラッピングする」に違反している。
他ドメインの Delete メソッドは全て `apperrors.Wrap` でコンテキストを付加してから return している。

```go
// 現状（NG）
func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
    slog.InfoContext(ctx, "deleting closing special period", ...)
    return s.periodRepo.Delete(ctx, clinicID, id)  // 裸のエラー return
}
```

```go
// あるべき姿
func (s *closingSettingsService) DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error {
    if err := s.periodRepo.Delete(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to delete closing special period")
    }
    slog.InfoContext(ctx, "closing special period deleted",
        slog.Uint64("clinic_id", clinicID),
        slog.Uint64("id", id))
    return nil
}
```

### 問題2: slog.InfoContext が Delete 実行前に出力されている

`periodRepo.Delete` が失敗しても「削除した」というログが先に記録される。
他ドメインはすべて操作成功後に slog を出力している。

## 完了条件
- [ ] `DeleteSpecialPeriod` の `periodRepo.Delete` を `apperrors.Wrap` でラップ
- [ ] `slog.InfoContext` を Delete 成功後に移動
- [ ] `go test ./backend/internal/...` がパス
