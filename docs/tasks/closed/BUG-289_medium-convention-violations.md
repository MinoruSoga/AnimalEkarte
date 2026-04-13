# BUG-289: MEDIUM 規約違反 6件

## 概要

第6回監査で検出された MEDIUM レベルの規約違反。

## 1. c.JSON 直接エラーレスポンス（2件）

### `handler/reservation_course_handler.go:162`
### `handler/reservation_staff_handler.go:180`

```go
c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
```

**規約違反**: `c.JSON` による直接エラーレスポンス禁止。`RespondError` に委譲すべき。

**修正方針**: `apperrors` に `WrapNotImplemented` を追加するか、`apperrors.WrapInternalServerError("not implemented")` に変更。

## 2. handler 層の slog 使用（1件）

### `handler/record_image_handler.go:234`

```go
slog.WarnContext(c.Request.Context(), "failed to clean up uploaded file", "key", key, "error", removeErr)
```

**規約違反**: slog はサービス層のみ。ただし auth_handler の監査ログ例外に近い「ベストエフォートのクリーンアップ失敗ログ」であり、実害度は低い。

## 3. 裸の return err（3件）

### `service/reservation_staff_service.go:108,134,147`

```go
if _, err := s.GetByID(ctx, clinicID, id); err != nil {
    return nil, err  // ← apperrors.Wrap なし
}
```

**規約違反**: Service 層での裸 return。`GetByID` が既に wrap 済みだが、コンテキスト（"update"/"delete"/"patch status"）が付与されない。

**修正方針**:
```go
if _, err := s.GetByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to verify reservation staff ownership")
}
```

## 4. slog 監査ログ欠落（1件）

### `service/reservation_staff_service.go:145-156` — `PatchStatus`

`Update` と `Delete` には `slog.InfoContext` があるが、`PatchStatus` にはない。

**修正方針**:
```go
slog.InfoContext(ctx, "reservation staff status patched",
    slog.Uint64("staff_id", id),
    slog.Uint64("clinic_id", clinicID),
    slog.Bool("is_active", isActive))
```

## 優先度

**Medium** — 動作上の問題はないが、規約一貫性とデバッグ追跡性に影響。

## 関連チケット

- BUG-287: 第6回監査 親チケット
