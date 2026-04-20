# TASK-196: staff_service.go — SetClinicAssignments 二重 Wrap・Reorder slog 欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/staff_service.go`

## 問題概要

### 問題1: SetClinicAssignments でエラーが二重 Wrap されている（行308-328）
`WithTx` コールバック内で `apperrors.Wrap` したエラーが、
外側の `if err := s.tx.WithTx(...); err != nil` でさらに `apperrors.Wrap` される。
エラーメッセージが冗長に二重化される（例: `"failed to update: failed to create: ..."`）。

```go
// 現状（NG）
if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
    if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
        return apperrors.Wrap(err, "failed to create clinic assignment")  // 1層目
    }
    return nil
}); err != nil {
    return apperrors.Wrap(err, "failed to update clinic assignments")  // 2層目（二重）
}
```

```go
// あるべき姿
if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
    if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
        return apperrors.Wrap(err, "failed to create clinic assignment")
    }
    return nil
}); err != nil {
    return err  // コールバック内で既にラップ済み
}
```

他のトランザクション実装（`Create`, `CreateWithAccount`）は外側で `Wrap` していない。

### 問題2: Reorder メソッドに slog.InfoContext が欠落（行411付近）
同ファイル内の他のメソッドはすべて slog を出力しているが、
`Reorder` のみ欠落しており操作追跡の一貫性が取れていない。

```go
// 修正後: return の前に slog を追加
slog.InfoContext(ctx, "staff reordered", slog.Uint64("clinic_id", clinicID))
return nil
```

## 完了条件
- [ ] `SetClinicAssignments` の外側 `apperrors.Wrap` を除去（`return err` に変更）
- [ ] `Reorder` に `slog.InfoContext` を追加
- [ ] `go test ./backend/internal/...` がパス
