# TASK-045: BE セキュリティヘッダミドルウェア追加

**作成日**: 2026-03-27
**ステータス**: Open
**優先度**: High
**領域**: Backend / Security

---

## 概要

以下のセキュリティヘッダが全レスポンスに設定されていない。

| ヘッダ | 役割 |
|--------|------|
| `X-Content-Type-Options: nosniff` | MIME タイプスニッフィング防止 |
| `X-Frame-Options: DENY` | クリックジャッキング防止 |
| `Referrer-Policy: strict-origin-when-cross-origin` | リファラ漏洩防止 |
| `Content-Security-Policy` | XSS 緩和（API サーバーは最小限で可） |
| `Permissions-Policy` | 不要ブラウザ機能の無効化 |

---

## 修正方針

```go
// middleware/security_headers.go（新規作成）
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'none'")
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        c.Next()
    }
}
```

`cmd/api/main.go` のグローバルミドルウェアに追加する。

HSTS (`Strict-Transport-Security`) は本番環境（HTTPS）のみ有効にする（`GIN_MODE=release` または環境変数で制御）。

---

## 受入条件

- [ ] `middleware/security_headers.go` が作成されている
- [ ] 全レスポンスに上記ヘッダが含まれている（curl で確認）
- [ ] `docker compose exec backend go build ./...` 成功
