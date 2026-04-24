# BE-076: PUT /users/:id/permission-groups ユーザーへのグループ割当API

**Status**: Closed
**Priority**: High
**Affects**: user_account_handler.go, user_account_service.go, user_account_repository.go
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-073（先に完了必要）, BE-074（先に完了推奨）, FE-131

## Summary

ユーザーに権限グループを一括割当・解除するエンドポイントを実装する。
既存の `PUT /users/:id/permissions`（旧PermissionType）を削除し、`PUT /users/:id/permission-groups` に置き換える。
また `GET /users/:id` のレスポンスに割当済みグループIDリストを含める。

## 現状のコード

**廃止対象: `backend/internal/handler/user_account_handler.go`**
```go
// 廃止する（SetUserPermissions）
func (h *Handler) SetUserPermissions(c *gin.Context) { ... }
func (h *Handler) GetUserPermissions(c *gin.Context) { ... }

// ルーティング（廃止）
users.GET(":id/permissions", h.GetUserPermissions)
users.PUT(":id/permissions", h.SetUserPermissions)
```

**廃止対象: `backend/internal/service/user_account_service.go`**
```go
// 廃止
type SetPermissionsInput struct { Permissions []string }
func (s *userAccountService) GetPermissions(...) ([]model.UserPermission, error) { ... }
func (s *userAccountService) SetPermissions(...) error { ... }
```

**廃止対象: `backend/internal/repository/user_account_repository.go`**
```go
// 廃止
func (r *userAccountRepository) FindPermissions(...) ([]model.UserPermission, error) { ... }
func (r *userAccountRepository) SetPermissions(...) error { ... }
```

## 必要な変更

### 1. Repository に新メソッドを追加

**`backend/internal/repository/user_account_repository.go`:**

```go
// SetPermissionGroups はユーザーのグループ割当を全置換する（トランザクション）
func (r *userAccountRepository) SetPermissionGroups(ctx context.Context, userID uint64, groupIDs []uint64) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 既存割当を全削除
        if err := tx.Where("user_id = ?", userID).Delete(&model.UserPermissionGroup{}).Error; err != nil {
            return err
        }
        if len(groupIDs) == 0 {
            return nil
        }
        // 新規割当を一括挿入
        rows := make([]model.UserPermissionGroup, len(groupIDs))
        for i, gid := range groupIDs {
            rows[i] = model.UserPermissionGroup{UserID: userID, GroupID: gid}
        }
        return tx.Create(&rows).Error
    })
}

// FindPermissionGroupIDs はユーザーに割り当てられているグループIDリストを返す
func (r *userAccountRepository) FindPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error) {
    var rows []model.UserPermissionGroup
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Find(&rows).Error
    if err != nil {
        return nil, err
    }
    ids := make([]uint64, len(rows))
    for i, r := range rows {
        ids[i] = r.GroupID
    }
    return ids, nil
}
```

### 2. Service に新メソッドを追加

**`backend/internal/service/user_account_service.go`:**

```go
type SetPermissionGroupsInput struct {
    GroupIDs []uint64 `json:"group_ids"`
}

// SetPermissionGroups はユーザーのグループ割当を全置換する
func (s *userAccountService) SetPermissionGroups(ctx context.Context, userID uint64, input SetPermissionGroupsInput) error {
    slog.InfoContext(ctx, "setting user permission groups", "user_id", userID, "group_ids", input.GroupIDs)
    return s.repo.SetPermissionGroups(ctx, userID, input.GroupIDs)
}

// GetPermissionGroupIDs はユーザーの割当グループIDリストを返す
func (s *userAccountService) GetPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error) {
    return s.repo.FindPermissionGroupIDs(ctx, userID)
}
```

### 3. Handler を追加・置換

**`backend/internal/handler/user_account_handler.go`（追加・変更）:**

```go
// Request型
type setPermissionGroupsRequest struct {
    GroupIDs []uint64 `json:"group_ids" binding:"required"`
}

// SetUserPermissionGroups: PUT /users/:id/permission-groups
func (h *Handler) SetUserPermissionGroups(c *gin.Context) {
    userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    var req setPermissionGroupsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    if err := h.svc.UserAccount.SetPermissionGroups(c.Request.Context(), userID, service.SetPermissionGroupsInput{
        GroupIDs: req.GroupIDs,
    }); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}
```

### 4. ルーティング変更

**`backend/internal/handler/user_account_handler.go` の RegisterUserRoutes:**

```go
// Before
users.GET(":id/permissions", h.GetUserPermissions)  // 削除
users.PUT(":id/permissions", h.SetUserPermissions)  // 削除

// After
users.PUT(":id/permission-groups", h.SetUserPermissionGroups)  // 追加
```

### 5. GET /users/:id のレスポンスにグループIDリストを追加

**`backend/internal/handler/user_account_handler.go`（GetUser ハンドラ拡張）:**

```go
// userDetailResponse にフィールドを追加
type userDetailResponse struct {
    userResponse
    Memberships      []model.UserClinicMembership `json:"memberships"`
    PermissionGroupIDs []uint64                   `json:"permission_group_ids"` // 追加
}

func (h *Handler) GetUser(c *gin.Context) {
    // 既存コード...
    groupIDs, err := h.svc.UserAccount.GetPermissionGroupIDs(c.Request.Context(), id)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, userDetailResponse{
        userResponse:       toUserResponse(&data.UserAccount),
        Memberships:        data.Memberships,
        PermissionGroupIDs: groupIDs,
    })
}
```

## API レスポンス形式

**PUT /api/users/123/permission-groups（リクエスト）:**
```json
{ "group_ids": [1, 3] }
```

**GET /api/users/123:**
```json
{
  "id": "123",
  "display_name": "田中太郎",
  "permission_group_ids": [1, 3],
  "memberships": [...]
}
```

## 完了条件

- [ ] `PUT /api/users/123/permission-groups` でグループを一括割当できる（全置換）
- [ ] `PUT` に `{"group_ids": []}` を送るとグループ割当が全解除される
- [ ] `GET /api/users/123` のレスポンスに `permission_group_ids` が含まれる
- [ ] 旧 `GET/PUT /users/:id/permissions` エンドポイントが削除されている
- [ ] 旧 `GetPermissions` / `SetPermissions` のservice・repository コードが削除されている

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `backend/internal/repository/user_account_repository.go` — SetPermissionGroups/FindPermissionGroupIDs メソッド追加
  - `backend/internal/service/user_account_service.go` — SetPermissionGroupsInput/SetPermissionGroups/GetPermissionGroupIDs 追加
  - `backend/internal/handler/user_account_handler.go` — SetUserPermissionGroups ハンドラ追加、GetUser 拡張（groupIDs）、ルーティング追加
  - `backend/internal/handler/user_account_response.go` — userDetailResponse に PermissionGroupIDs フィールド追加
  - `backend/internal/service/user_account_service_test.go` — モックに SetPermissionGroups/FindPermissionGroupIDs 追加
