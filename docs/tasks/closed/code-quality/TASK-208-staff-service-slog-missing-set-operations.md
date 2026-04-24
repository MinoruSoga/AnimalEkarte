# TASK-208: staff_service.go — SetPermissionGroupIDs / SetExcludedReservationTypeIDs の slog 欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/staff_service.go`

## 問題概要
`SetPermissionGroupIDs`（行431付近）と `SetExcludedReservationTypeIDs`（行452付近）に
`slog.InfoContext` による操作ログが記録されていない。

同ファイルの他のメソッド（`Create`, `Update`, `Delete`, `SetClinicAssignments` 等）は
すべて操作ログを記録しており、この2メソッドのみ欠落している。

権限グループ設定・除外予約種別設定は重要な権限管理操作であり、
監査ログ（audit trail）として操作記録が必要。

※ TASK-196（Reorder slog 欠落）と同じパターンだが、別メソッドのため独立して起票。

## あるべき姿

```go
// SetPermissionGroupIDs 成功後
slog.InfoContext(ctx, "staff permission groups set",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("staff_id", staffID),
    slog.Any("group_ids", groupIDs))

// SetExcludedReservationTypeIDs 成功後
slog.InfoContext(ctx, "staff excluded reservation types set",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("staff_id", staffID),
    slog.Any("excluded_type_ids", typeIDs))
```

## 完了条件
- [ ] `SetPermissionGroupIDs` 成功後に `slog.InfoContext` を追加
- [ ] `SetExcludedReservationTypeIDs` 成功後に `slog.InfoContext` を追加
- [ ] `go test ./backend/internal/...` がパス
