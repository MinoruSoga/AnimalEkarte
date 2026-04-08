# BUG-140: 自分の権限グループルールを空にしてロックアウトできる

## 概要
`PUT /api/v1/masters/permission-groups/:id/rules` で自分が所属する権限グループのルールを
空配列 `{"rules": []}` に設定できる。設定後、自分自身を含む全メンバーが全 API で 403 になり、
**system_admin 以外では復旧不可能なロックアウト状態**になる。

## 脆弱性分類
- **CWE-285**: Improper Authorization
- **影響**: 管理者が自分の権限を誤って削除 → 復旧に system_admin 介入が必要。
  悪意のあるユーザーが意図的に全ユーザーの権限を破壊することも可能。

## 再現手順
1. `admin@example.com`（執行グループ ID=1、master-permission edit 権限あり）でログイン
2. `PUT /api/v1/masters/permission-groups/1/rules` で `{"rules": []}` を送信 → **200 OK**
3. 以降すべての API が **403 Forbidden**
4. 自分でルールを復元しようとしても **403**（ロックアウト）
5. **system_admin（admin@noavet.jp）でログインして復旧する以外に手段がない**

## ブラウザテスト結果
```
PUT /masters/permission-groups/1/rules {"rules": []} → 200 ⚠️
PUT /masters/permission-groups/1/rules {full rules} → 403 ❌ (ロックアウト)
→ system_admin で復旧
```

## 期待する動作

### 案A: 自分が所属するグループのルール変更時に警告/禁止
```go
// 自分のスタッフ ID から所属グループを取得
// 変更対象のグループ ID と一致する場合、ルールの完全削除を禁止
if isSelfGroup && len(rules) == 0 {
    return apperrors.WrapInvalidInput("自分が所属するグループのルールを全削除することはできません")
}
```

### 案B: master-permission の edit 権限は最低限残す
自分が所属するグループの `master-permission` リソースの `edit` 権限だけは削除できないようにする。

### 案C: 変更確認ダイアログ（フロントエンド側）
権限ルール変更時に「この変更であなたの権限が失われます。続行しますか？」という確認。
ただしバックエンドでもガードすべき。

## 修正方針

```go
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
    groupID := parseID(c.Param("id"))
    staffID, _ := extractStaffID(c)
    
    // 自分が所属するグループか確認
    myGroups, _ := h.repos.StaffPermissionGroup.FindByStaffID(ctx, staffID)
    isSelfGroup := false
    for _, g := range myGroups {
        if g.GroupID == groupID { isSelfGroup = true; break }
    }
    
    if isSelfGroup {
        // master-permission edit が残っているか確認
        hasMasterPermEdit := false
        for _, rule := range req.Rules {
            if rule.Resource == "master-permission" && rule.CanEdit {
                hasMasterPermEdit = true; break
            }
        }
        if !hasMasterPermEdit {
            RespondError(c, apperrors.WrapInvalidInput(
                "自分が所属するグループの権限管理権限を削除することはできません"))
            return
        }
    }
    // ... 通常処理
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md` — Authentication
> "Use secure session management"

自分自身の権限を破壊できる状態はセッション管理の欠陥。

### `.claude/rules/api.md` — Security
> "Validate all user input"

権限ルールの変更はシステムの根幹に関わる操作。入力バリデーションで自己ロックアウトを防止すべき。

## 優先度
**High** — 管理者の誤操作で全ユーザーがロックアウトされる。system_admin がいない環境では完全に復旧不能。

## 関連チケット
- BUG-134: 無効化ユーザーログイン（認証関連）

## 関連ファイル
- `backend/internal/handler/staff_handler.go` — SetPermissionGroupRules ハンドラ
- `backend/internal/repository/staff_permission_group_repository.go` — 所属グループ取得
