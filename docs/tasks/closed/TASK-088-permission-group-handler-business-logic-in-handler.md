# TASK-088: permission_group_handler — BUG-140 バリデーション + 重複チェックがハンドラ層に実装されている

## 優先度

HIGH

---

## 概要

`permission_group_handler.go`（L212-257）に、BUG-140 対応として追加された
バリデーションロジックおよびリソース重複チェックが handler 層に実装されている。

ビジネスロジックは service 層に配置すべきであり、handler 層は
リクエストの解析・レスポンス生成・エラーマッピングのみを担う。

---

## 問題箇所

### permission_group_handler.go:212-257（推定）

```go
// ❌ handler 層にビジネスバリデーションが存在
func (h *permissionGroupHandler) Update(c *gin.Context) {
    // ...

    // BUG-140: 自グループの検証（ビジネスロジック）
    myGroupIDs := make(map[uint64]bool)
    for _, r := range req.Rules {
        // ... 自グループのID検証ループ
    }

    // BUG-140: リソース重複チェック（ビジネスロジック）
    seen := make(map[string]bool)
    for _, r := range req.Rules {
        key := fmt.Sprintf("%s-%d", r.Resource, r.Action)
        if seen[key] {
            // 重複エラー返却
        }
        seen[key] = true
    }
    // ...
}
```

---

## 影響

- handler がビジネスルールを知っている → 責任分離違反
- service 層をスキップした場合、バリデーションが実行されない
- service の単体テストで同じバリデーションをテストできない
- BUG 修正のたびに handler コードを変更しなければならない

---

## 修正方針

バリデーションロジックを **service 層に移動する**。

```go
// ✅ 修正後: handler は薄くする
func (h *permissionGroupHandler) Update(c *gin.Context) {
    var req updatePermissionGroupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    // ...
    pg, err := h.service.Update(ctx, clinicID, id, req.toInput())
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, toPermissionGroupResponse(pg))
}

// ✅ service 層でビジネスルールを実装
func (s *permissionGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
    // BUG-140: 自グループ検証
    if err := s.validateNotSelfReference(input.Rules, id); err != nil {
        return nil, err
    }

    // BUG-140: リソース重複チェック
    if err := s.validateNoDuplicateRules(input.Rules); err != nil {
        return nil, err
    }

    // ...
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/permission_group_handler.go` | BUG-140 バリデーションロジック（L212-257 付近）を削除し、service 呼び出しに集約 |
| `service/permission_group_service.go` | `validateNotSelfReference()` と `validateNoDuplicateRules()` を private メソッドとして実装 |

---

## 関連

- TASK-087: 同ハンドラの repository 直接 import 問題
- BUG-140: 本 TASK の起源となったバグ修正
