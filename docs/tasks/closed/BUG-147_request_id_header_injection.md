# BUG-147: X-Request-ID ヘッダーにユーザー入力がサニタイズなしで反映される

## 概要
リクエストの `X-Request-ID` ヘッダーに任意の値を設定すると、レスポンスの `X-Request-ID` に
そのまま反映される。`<script>alert(1)</script>` 等の XSS ペイロードがヘッダーに入る。

HTTP レスポンスヘッダーはブラウザが HTML として解釈しないため、直接的な XSS にはならないが、
ログビューアーやモニタリングダッシュボードで Stored XSS を引き起こす可能性がある。

## 脆弱性分類
- **CWE-113**: Improper Neutralization of CRLF Sequences in HTTP Headers
- **影響**: ログ汚染、モニタリングツールでの Stored XSS

## 再現手順
```bash
curl -v /api/v1/owners \
  -H 'X-Request-ID: <script>alert(1)</script>'

# レスポンスヘッダー:
# X-Request-ID: <script>alert(1)</script>
```

## ブラウザテスト結果
```
Request:  X-Request-ID: <script>alert(1)</script>
Response: X-Request-ID: <script>alert(1)</script>  ← そのまま反映
```

## 期待する動作
- ユーザー提供の `X-Request-ID` をサニタイズ（英数字とハイフンのみ許可）
- または無視してサーバー側で生成した ID のみを使用

## 修正方針

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.GetHeader("X-Request-ID")
        // サニタイズ: 英数字とハイフンのみ許可、最大36文字（UUID形式）
        if !isValidRequestID(id) {
            id = generateRequestID()
        }
        c.Header("X-Request-ID", id)
        c.Next()
    }
}

func isValidRequestID(id string) bool {
    if len(id) == 0 || len(id) > 36 { return false }
    for _, c := range id {
        if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' { return false }
    }
    return true
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md`
> "Sanitize all user input"

X-Request-ID もユーザー入力。

## 優先度
**Low** — 直接的な XSS にはならない。ログ/モニタリングツール経由の間接的リスク。

## 関連ファイル
- `backend/internal/middleware/` — Request ID ミドルウェア
