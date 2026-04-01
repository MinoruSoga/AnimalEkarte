# BE: BUG-020 PATCH/PUT API の権限チェック未実装

## 概要

フロントエンドの create/edit ルートガードは修正済み（router.tsx）。
しかしバックエンドの PATCH/PUT/POST API に権限チェックが実装されていないため、
API 直接呼び出しで `can_edit=false` のユーザーが変更を保存できる。

## フロントエンド対応（修正済み）

- `/new` ルートに `RequirePermission action="create"` ガード追加
- `/edit` ルートに `RequirePermission action="edit"` ガード追加

## 残存問題（バックエンド）

直接 API を叩けば権限のないユーザーが変更可能：

```
PATCH /api/v1/owners/1
Authorization: Bearer <staff_token_with_can_edit=false>
→ HTTP 200 で保存成功（期待: 403 Forbidden）
```

## 実装場所

各ハンドラの PATCH/PUT/POST エンドポイントに権限チェックを追加：
- `internal/middleware/` に `RequireResourcePermission(resource, action)` ミドルウェアを作成
- 各ルートに `RequireResourcePermission("owners", "edit")` を適用

```go
// ミドルウェア例
func RequireResourcePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetAuthUser(c)
        if !user.HasPermission(resource, action) {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 優先度

High（セキュリティ）

## 関連

- `docs/tasks/open/security/BUG-020_edit_url_bypass.md`
- FUNCTIONAL_TEST_REPORT.md BUG-020
