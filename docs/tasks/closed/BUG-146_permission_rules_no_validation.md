# BUG-146: 権限グループルール設定に入力バリデーションがない

## 概要
`PUT /api/v1/masters/permission-groups/:id/rules` で以下の不正入力が許容される:
1. 空文字のリソース名 → 200（無意味なルールが作成される）
2. 存在しないリソース名 → 200（無効なルールが作成される）
3. 同一リソースの重複ルール → 500（DB UNIQUE 制約違反）

## 脆弱性分類
- **CWE-20**: Improper Input Validation
- **影響**: 権限管理データの汚染。重複ルールで 500 エラー。

## 再現手順

### 1. 空リソース名
```bash
curl -X PUT /api/v1/masters/permission-groups/7/rules \
  -H 'Content-Type: application/json' \
  -d '{"rules": [{"resource": "", "can_view": true, "can_create": true, "can_edit": true, "can_delete": true}]}'
# → 200 OK ❌
```

### 2. 存在しないリソース名
```bash
curl -X PUT /api/v1/masters/permission-groups/7/rules \
  -H 'Content-Type: application/json' \
  -d '{"rules": [{"resource": "nonexistent-resource", "can_view": true}]}'
# → 200 OK ❌
```

### 3. 重複リソース
```bash
curl -X PUT /api/v1/masters/permission-groups/7/rules \
  -H 'Content-Type: application/json' \
  -d '{"rules": [
    {"resource": "owners", "can_view": true, "can_create": true, "can_edit": true, "can_delete": true},
    {"resource": "owners", "can_view": false, "can_create": false, "can_edit": false, "can_delete": false}
  ]}'
# → 500 Internal Server Error ❌
```

## 期待する動作
- 空リソース名 → 400 `リソース名は必須です`
- 存在しないリソース名 → 400 `無効なリソース名: nonexistent-resource`
- 重複リソース → 400 `リソースが重複しています: owners`

## 修正方針

### Service 層でバリデーション

```go
func (s *PermissionGroupService) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
    seen := make(map[string]bool)
    for _, rule := range rules {
        // 空文字チェック
        if rule.Resource == "" {
            return apperrors.WrapInvalidInput("リソース名は必須です")
        }
        // 有効なリソース名か
        if !model.IsValidResource(rule.Resource) {
            return apperrors.WrapInvalidInput("無効なリソース名: " + rule.Resource)
        }
        // 重複チェック
        if seen[rule.Resource] {
            return apperrors.WrapInvalidInput("リソースが重複しています: " + rule.Resource)
        }
        seen[rule.Resource] = true
    }
    return s.repo.SetRules(ctx, groupID, rules)
}
```

### model/permission.go に IsValidResource 追加

```go
func IsValidResource(r string) bool {
    for _, res := range AllResources {
        if string(res) == r { return true }
    }
    return false
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> `c.JSON(http.StatusBadRequest, gin.H{"error": ...})` の直接使用は禁止

500 ではなく Service 層でバリデーションして 400 を返すべき。

### `.claude/rules/api.md`
> "Include input validation on all endpoints"

## 優先度
**Medium** — 権限管理データの汚染 + 500 エラー。重複ルールの 500 は BUG-138 と同根の問題。

## 関連チケット
- BUG-138: FK 違反が 500 を返す
- BUG-140: 権限グループ自己ロックアウト

## 関連ファイル
- `backend/internal/handler/staff_handler.go` — SetPermissionGroupRules
- `backend/internal/service/permission_group_service.go` — SetRules（バリデーション追加）
- `backend/internal/model/permission.go` — AllResources, IsValidResource 追加
