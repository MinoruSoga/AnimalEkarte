# BUG-149: パスワードリセット API が 500 Internal Server Error を返す

## 概要
`POST /api/v1/forgot-password` と `POST /api/v1/reset-password` が 500 エラーを返す。
フロントエンド（パスワードリセットページ）は実装済みだが、バックエンドの API がクラッシュする。

メール送信機能（SMTP）が未設定、または実装が不完全の可能性。

## 再現手順
```bash
# パスワードリセット依頼
curl -X POST /api/v1/forgot-password \
  -H 'Content-Type: application/json' \
  -d '{"email": "admin@example.com"}'
# → 500 Internal Server Error

# パスワードリセット実行
curl -X POST /api/v1/reset-password \
  -H 'Content-Type: application/json' \
  -d '{"token": "fake", "password": "newpass123"}'
# → 500 Internal Server Error
```

## 期待する動作
- `POST /forgot-password` → 200（メール送信成功。存在しないメールでも 200 を返す = ユーザー列挙防止）
- `POST /reset-password` + 無効トークン → 400（無効なトークン）

## 修正方針

### メール送信未設定の場合
500 ではなく適切なエラーハンドリングを追加:

```go
func (h *Handler) ForgotPassword(c *gin.Context) {
    // ... 
    if h.emailService == nil {
        // メール送信が未設定でもクラッシュしない
        slog.WarnContext(ctx, "email service not configured, skipping password reset email")
        c.JSON(http.StatusOK, gin.H{"message": "リセットリンクを送信しました"})
        return
    }
    // ...
}
```

### reset-password のトークン検証
```go
func (h *Handler) ResetPassword(c *gin.Context) {
    // トークン検証でエラーが発生しても 500 にしない
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無効またはの期限切れのリセットトークンです"})
        return
    }
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> Handler: `RespondError(c, err)` で統一レスポンス

500 ではなく適切なステータスコードで返すべき。

### `.claude/rules/api.md`
> "Return consistent error response format"

## 優先度
**Medium** — パスワードリセット機能が完全に動作しない。500 エラーはサーバー監視で誤報を生む。
ただし、自己パスワード変更（BUG-148）とは異なり、メール送信インフラが必要なため、
フル実装は別途対応でよい。最低限 500 を返さないようにすべき。

## 関連チケット
- BUG-148: 自己パスワード変更 API 未実装
- BUG-138: FK 違反が 500 を返す

## 関連ファイル
- `backend/internal/handler/auth_handler.go` — ForgotPassword, ResetPassword ハンドラ
- `frontend/src/features/auth/components/ForgotPasswordForm.tsx` — フロントエンド（実装済み）
