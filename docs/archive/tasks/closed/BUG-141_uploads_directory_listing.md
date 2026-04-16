# BUG-141: /uploads ディレクトリリスティングが有効

## 概要
バックエンド（port 8080）の `/uploads/` にアクセスすると、Gin の `r.Static()` により
ディレクトリリスティングが表示される。現在はファイルが存在しないため空だが、
画像アップロード機能が使用されるとファイル一覧が公開される。

## 脆弱性分類
- **CWE-548**: Exposure of Information Through Directory Listing
- **影響**: アップロードされたファイル名一覧が認証なしで取得可能

## 再現手順
```bash
curl http://localhost:8080/uploads/
# → 200
# <!doctype html>
# <meta name="viewport" content="width=device-width">
# <pre>
# </pre>
```

現在は空だが、ファイルが存在すれば `<pre>` 内にファイル一覧が表示される。

## 期待する動作
- ディレクトリリスティングを無効化
- 個別ファイルへのアクセスのみ許可
- または認証ミドルウェアを適用

## 現状コード

### `backend/internal/handler/handler.go`
```go
r.Static("/uploads", "/app/uploads")
// ❌ ディレクトリリスティングが有効
// ❌ 認証なしでアクセス可能
```

## 修正方針

### 案A: StaticFS でディレクトリリスティング無効化
```go
r.StaticFS("/uploads", http.Dir("/app/uploads"))
// Gin の StaticFS はディレクトリリスティングを無効化
```

### 案B: 認証ミドルウェアを適用
```go
uploads := r.Group("/uploads")
uploads.Use(middleware.Auth(h.cfg.JWTSecret))
uploads.Static("/", "/app/uploads")
```

## 優先度
**Low** — 現時点でファイルなし。画像アップロード機能が本番稼働する前に対応すべき。

## 関連ファイル
- `backend/internal/handler/handler.go` — Static ルート登録
