# TASK-239: closing_settings_service.go — UpdateSpecialPeriod が UpdateFields エラーを raw return

## 優先度
High

## 対象ファイル
- `backend/internal/service/closing_settings_service.go`

## 問題概要
`UpdateSpecialPeriod` メソッドの `periodRepo.UpdateFields` 呼び出し後のエラーハンドリングが
`return nil, err` と生のエラーを返しており、`apperrors.Wrap` でのラッピングと `slog.ErrorContext` のログが欠落している。

規約: **Service 層は全エラーを `apperrors.Wrap(err, "message")` でラップし、`slog.ErrorContext` で記録する。**

## 現状コード（行222〜226付近）

```go
result, err := s.periodRepo.UpdateFields(ctx, clinicID, id, fields)
if err != nil {
    return nil, err  // ❌ raw return: Wrap なし、slog なし
}
slog.InfoContext(ctx, "special period updated", slog.Uint64("id", id))
```

## 比較（同ファイル内の正しい実装）

```go
// CreateSpecialPeriod パターン
if err := s.periodRepo.Create(ctx, period); err != nil {
    return nil, apperrors.Wrap(err, "failed to create special period")  // ✅
}
```

## あるべき姿

```go
result, err := s.periodRepo.UpdateFields(ctx, clinicID, id, fields)
if err != nil {
    slog.ErrorContext(ctx, "failed to update closing special period",
        slog.Uint64("clinic_id", clinicID),
        slog.Uint64("id", id),
        slog.Any("error", err))
    return nil, apperrors.Wrap(err, "failed to update special period")
}
slog.InfoContext(ctx, "closing special period updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("id", id))
```

## 完了条件
- [ ] `return nil, err` を `return nil, apperrors.Wrap(err, "failed to update special period")` に変更
- [ ] `slog.ErrorContext` を追加
- [ ] `go test ./backend/internal/...` がパス
