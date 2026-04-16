# BUG-210: バックエンド Handler / Service 層の規約違反

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | エラーハンドリング / アーキテクチャ |

## 1. RefreshToken で直接 c.JSON（6箇所）

`backend/internal/handler/auth_handler.go`

RespondError を使わず `c.JSON(http.StatusUnauthorized, gin.H{"error": ...})` を直接使用。

| 行 | 内容 |
|-----|------|
| 316 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})` |
| 328 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})` |
| 334 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})` |
| 341 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})` |
| 346 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})` |
| 350 | `c.JSON(http.StatusUnauthorized, gin.H{"error": "このアカウントは無効です"})` |

修正: 全て `RespondError(c, apperrors.WrapUnauthorized("..."))` に統一。

## 2. trimming_handler で parseBindError 未使用（2箇所）

`backend/internal/handler/trimming_handler.go`

| 行 | 現状 |
|-----|------|
| 95 | `RespondError(c, apperrors.WrapInvalidInput(err.Error()))` |
| 145 | `RespondError(c, apperrors.WrapInvalidInput(err.Error()))` |

修正: `apperrors.WrapInvalidInput(parseBindError(err))` に統一。

## 3. permission_group_service 裸 return err（2箇所）

`backend/internal/service/permission_group_service.go`

| 行 | メソッド |
|-----|---------|
| 85 | Delete — `return err`（apperrors.Wrap なし） |
| 104 | Reorder — `return err`（apperrors.Wrap なし） |

修正:
```go
return apperrors.Wrap(err, "failed to delete permission group")
return apperrors.Wrap(err, "failed to reorder permission groups")
```

## 4. auth_handler に slog 11箇所（handler 層 slog 禁止違反）

`backend/internal/handler/auth_handler.go`

行: 127, 133-136, 140, 147, 155, 248, 259, 305, 428, 489, 583-585

規約:「slog は service 層のみ。handler・repository には記述しない」

## 5. cleanupLoop に context キャンセルなし

`backend/internal/middleware/rate_limit.go:30-41`

```go
go s.cleanupLoop()  // context なしで goroutine 起動

func (s *RateLimitStore) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {  // 永遠に止まらない
        s.evict(10 * time.Minute)
    }
}
```

## 6. WrapConflict が ErrAlreadyExists を使用

`backend/internal/errors/errors.go:88-94`

```go
func WrapConflict(message string) error {
    return &AppError{
        Code:    "CONFLICT",
        Message: message,
        Err:     ErrAlreadyExists,  // 独立した ErrConflict センチネルがない
    }
}
```

## 参照実装

- RespondError の正しい使用: `owner_handler.go` 全箇所
- parseBindError の正しい使用: `owner_handler.go:38`
- apperrors.Wrap の正しい使用: `owner_service.go` 全箇所
