# TASK-194: permission_group_service.go — エラーのサイレント握りつぶしに slog 追加

## 優先度
Medium

## 対象ファイル
`backend/internal/service/permission_group_service.go`

## 問題概要
`canModifyPermissionGroup`（行169-172）にて、`GetGroupIDsByStaffID` のエラーを
`slog.WarnContext` なしで握りつぶし、空スライスに差し替えている。

ベストエフォート設計は理解できるが、エラーが監査ログに残らないため、
権限チェック失敗が「許可方向」でサイレントに通過する状況が観測不能になる。

## 現状コード（行169-172）

```go
staffGroupIDs, err := s.repo.GetGroupIDsByStaffID(ctx, actorStaffID)
if err != nil {
    // エラー時は空にして自己参照チェック不能なら許可方向（ベストエフォート）
    staffGroupIDs = []uint64{}
}
```

## あるべき姿

```go
staffGroupIDs, err := s.repo.GetGroupIDsByStaffID(ctx, actorStaffID)
if err != nil {
    slog.WarnContext(ctx, "failed to get staff group ids for self-reference check, allowing by best-effort",
        slog.Uint64("actor_staff_id", actorStaffID),
        slog.Any("error", err))
    staffGroupIDs = []uint64{}
}
```

## 完了条件
- [ ] エラー時に `slog.WarnContext` でエラーを記録するよう修正
- [ ] `go test ./backend/internal/...` がパス
