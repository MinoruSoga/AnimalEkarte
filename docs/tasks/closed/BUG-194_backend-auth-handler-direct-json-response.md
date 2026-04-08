# BUG-194: バックエンド auth_handler.go の RefreshToken で直接 c.JSON を使用（RespondError 未使用）

## 概要

`backend/internal/handler/auth_handler.go` の `RefreshToken` メソッドで、`c.JSON(http.StatusUnauthorized, gin.H{"error": ...})` が直接使用されており、プロジェクト標準の `RespondError(c, err)` パターンに従っていない。`.claude/CLAUDE.md` で「`c.JSON(http.StatusBadRequest, gin.H{"error": ...})` の直接使用は禁止」と明示されているにも関わらず、全 31 ハンドラ統一化から漏れている。

## 脆弱性分類
- **影響**: エラーレスポンスが構造化されず、エラーコード・タイムスタンプが返却されない。ログ出力が `RespondError` の統一ロギングをバイパスする。

## 再現手順

1. 無効なリフレッシュトークンで `POST /api/auth/refresh` を実行
   ```bash
   curl -X POST http://localhost:8080/api/auth/refresh \
     -H "Content-Type: application/json" \
     -d '{"refresh_token": "invalid-token"}'
   ```
2. **結果**: 
   ```json
   {"error": "invalid token"}
   ```
   ← 構造化エラー（`code`, `message`, `timestamp`）ではなく `gin.H{"error": ...}` の生レスポンス

## 期待する動作

```json
{
  "code": "UNAUTHORIZED",
  "message": "Token is invalid or expired",
  "timestamp": "2026-04-08T12:00:00Z"
}
```

## 現状コード

### `backend/internal/handler/auth_handler.go:316-350`
```go
// ❌ 直接 c.JSON で 401 を返す
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})  // ❌
        return
    }

    claims, err := h.authService.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})  // ❌
        return
    }

    // ...
}
```

### 比較: 正しい実装（参照実装）
```go
// ✅ 全ハンドラ統一パターン（例: owner_handler.go）
func (h *OwnerHandler) GetOwner(c *gin.Context) {
    // ...
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    owner, err := h.service.GetOwner(c.Request.Context(), id)
    if err != nil {
        RespondError(c, err)  // ✅ 統一 RespondError
        return
    }
}
```

## 影響範囲

| 対象ファイル | 行番号 | 問題 | 状態 |
|---|---|---|---|
| `backend/internal/handler/auth_handler.go` | 316-350 (RefreshToken) | c.JSON 直接使用（ShouldBindJSON エラー + ValidateRefreshToken エラー） | 未修正 |

## 修正方針

### `auth_handler.go:316-350` — `c.JSON` を `RespondError` に置換
```go
import "github.com/your-org/ekarte/backend/internal/apperrors"

func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // ✅ 統一パターン
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    claims, err := h.authService.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        // ✅ apperrors.ErrUnauthorized を使用（定義がなければ追加）
        RespondError(c, apperrors.WrapUnauthorized("invalid or expired token"))
        return
    }

    // ...
}
```

**注**: `apperrors.WrapUnauthorized` が未定義の場合、`errors/errors.go` に以下を追加する:
```go
var ErrUnauthorized = errors.New("unauthorized")

func WrapUnauthorized(msg string) error {
    return fmt.Errorf("%s: %w", msg, ErrUnauthorized)
}
```

そして `RespondError` の switch 文に 401 ケースを追加:
```go
case errors.Is(err, apperrors.ErrUnauthorized):
    code = http.StatusUnauthorized
    message = "Unauthorized"
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> **Handler**: `RespondError(c, err)` で統一レスポンス。
> `c.JSON(http.StatusBadRequest, gin.H{"error": ...})` の直接使用は禁止。

### `.claude/rules/error-handling.md` — HTTP ステータスマッピング
> Handler: `RespondError(c, err)` で統一レスポンス。

## 優先度
**High** — エラーログの統一性が破れており、監視・デバッグが困難になる。認証エンドポイントのエラーが構造化されないのはセキュリティログ観点でも問題。

## 関連チケット
- BUG-187 ではないが、バックエンドの構造的統一性の問題

## 関連ファイル
- `backend/internal/handler/auth_handler.go`
- `backend/internal/apperrors/errors.go`
- `backend/internal/handler/response.go`
