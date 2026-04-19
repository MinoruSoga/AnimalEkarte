# TASK-016: SetPermissionGroupRules に clinic 所有権チェックなし — 横断クリニック権限書き換え可能

## 概要

`permission_group_handler.go` の `SetPermissionGroupRules` が `clinicID` による所有権検証なしに `groupID` のみでルールを書き換えている。他クリニックのグループ ID を指定すれば、そのクリニックの権限ルールを上書きできる。

## 優先度

CRITICAL（マルチテナント権限の横断アクセス）

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `backend/internal/handler/permission_group_handler.go` | L191 | `SetRules` 前に clinicID チェックなし |
| `backend/internal/handler/permission_group_handler.go` | L284 | `extractClinicID` の `bool` 戻り値を `_` で無視 |
| `backend/internal/service/permission_group_service.go` | `SetRules` | clinicID を引数に取っていない |
| `backend/internal/repository/permission_group_repository.go` | `SetRules` | clinicID を引数に取っていない |

## 修正方針

### handler 層

```go
// permission_group_handler.go（修正後）
func (h *PermissionGroupHandler) SetPermissionGroupRules(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    id := parseUintParam(c, "id")

    // SetRules の前に所有権確認
    _, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
    if err != nil {
        RespondError(c, err)
        return
    }

    // ... ルール更新処理
}
```

### service / repository 層

```go
// service: SetRules に clinicID を追加
func (s *permissionGroupService) SetRules(ctx context.Context, clinicID, groupID uint64, rules []RuleInput) error {
    // 所有権確認
    _, err := s.repo.GetByID(ctx, clinicID, groupID)
    if err != nil {
        return apperrors.Wrap(err, "permission group not found for clinic")
    }
    return s.repo.SetRules(ctx, groupID, rules)
}
```

## テスト

- 異なるクリニックの権限グループ ID を指定して `SetRules` をコールした場合、403/404 が返ることを確認
- `extractClinicID` が失敗した場合にハンドラが早期リターンすることを確認
