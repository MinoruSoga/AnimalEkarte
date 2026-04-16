# BUG-137: API レスポンスに Cache-Control ヘッダーがない

## 概要
`/api/v1/*` エンドポイントのレスポンスに `Cache-Control` ヘッダーが設定されていない。
認証済みユーザーの個人データ（飼主一覧、スタッフ情報等）がブラウザキャッシュや
CDN/プロキシにキャッシュされ、別ユーザーに返される可能性がある。

## 脆弱性分類
- **CWE-525**: Use of Web Browser Caching Containing Sensitive Information
- **OWASP A05:2021**: Security Misconfiguration
- **影響**: 共有端末やプロキシ経由でのデータ漏洩リスク

## 再現手順
```bash
curl -v /api/v1/owners -H 'Cookie: access_token=...'
# レスポンスヘッダー:
# Content-Type: application/json; charset=utf-8
# X-Content-Type-Options: nosniff
# ← Cache-Control ヘッダーがない
```

## ブラウザテスト結果
```
Cache-Control: null
Pragma: null
```

## 期待するヘッダー
```
Cache-Control: no-store, no-cache, must-revalidate, private
Pragma: no-cache
```

## 修正方針

### CORS ミドルウェアまたは専用ミドルウェアで追加

```go
func NoCacheAPI() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
        c.Next()
    }
}

// handler.go
protected.Use(NoCacheAPI())
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md`
> "Use secure session management"

キャッシュ制御はセッション管理の一環。認証済みデータはキャッシュ禁止が基本。

## 優先度
**Low** — 開発環境では問題にならない。本番環境（CDN/プロキシ経由）で対応が必要。

## 関連ファイル
- `backend/internal/middleware/cors.go` — CORS ミドルウェア（ヘッダー追加箇所）
- `backend/internal/handler/handler.go` — ミドルウェア登録
