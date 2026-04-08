# BUG-188: Handler 層で RespondError を使わず直接 c.JSON でエラー返却

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | エラーハンドリング統一 |

## 概要

`auth_handler.go` の RefreshToken 関数内で `RespondError(c, err)` を使わず
`c.JSON(http.StatusUnauthorized, gin.H{"error": "..."})` を直接呼び出している箇所が6つある。
また `trimming_handler.go` で `parseBindError` を経由せず `err.Error()` を直接渡している箇所が2つある。

## 現状コード

### `backend/internal/handler/auth_handler.go` — 直接 c.JSON（6箇所）
```go
// :316
c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
// :328
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
// :334
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
// :341
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// :346
c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
// :350
c.JSON(http.StatusUnauthorized, gin.H{"error": "このアカウントは無効です"})
```

### `backend/internal/handler/trimming_handler.go` — parseBindError 未使用（2箇所）
```go
// :95
RespondError(c, apperrors.WrapInvalidInput(err.Error()))
// :145
RespondError(c, apperrors.WrapInvalidInput(err.Error()))
```

### 参照実装（正しいパターン）: `backend/internal/handler/owner_handler.go`
```go
RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
```

## 修正方針

### auth_handler.go（6箇所）
```go
// Before
c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
// After
RespondError(c, apperrors.WrapUnauthorized("refresh token not found"))
```

### trimming_handler.go（2箇所）
```go
// Before
RespondError(c, apperrors.WrapInvalidInput(err.Error()))
// After
RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md`
> Handler: `RespondError(c, err)` で統一レスポンス。
> `c.JSON(http.StatusBadRequest, gin.H{"error": ...})` の直接使用は禁止。

## 優先度
**High** — エラーレスポンス形式の不統一。クライアント側のエラーパーサーが正しく動作しない可能性。
