# BE-074: 権限グループ CRUD API

**Status**: Closed
**Priority**: High
**Affects**: permission_group_handler.go, permission_group_service.go, permission_group_repository.go（新規）
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-073（先に完了必要）, FE-130

## Summary

権限グループの作成・取得・更新・削除と、グループルール（ページ×CRUD）の一括更新APIを実装する。
handler → service → repository の3層構成。参照実装は `owner_handler.go` パターン。

## 現状のコード

新規実装。対象ファイルはすべて存在しない。

ルーティングの参照先:
```go
// backend/internal/handler/handler.go:61
protected.GET("/me", h.GetMe)
// 同ファイル内の RegisterUserRoutes 呼び出しパターンを踏襲する
```

## 必要な変更

### 1. Repository（新規: `backend/internal/repository/permission_group_repository.go`）

```go
package repository

import (
    "context"
    "gorm.io/gorm"
    "github.com/animal-ekarte/backend/internal/model"
)

type PermissionGroupRepository interface {
    FindByClinicID(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
    FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
    Create(ctx context.Context, group *model.PermissionGroup) error
    UpdateFields(ctx context.Context, id uint64, fields map[string]any) error
    Delete(ctx context.Context, id uint64) error
    SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
}

type permissionGroupRepository struct {
    db *gorm.DB
}

func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
    return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    var groups []model.PermissionGroup
    err := r.db.WithContext(ctx).
        Preload("Rules").
        Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
        Order("id ASC").
        Find(&groups).Error
    return groups, err
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
    var group model.PermissionGroup
    err := r.db.WithContext(ctx).
        Preload("Rules").
        Where("id = ? AND deleted_at IS NULL", id).
        First(&group).Error
    if err != nil {
        return nil, err
    }
    return &group, nil
}

func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
    return r.db.WithContext(ctx).Create(group).Error
}

func (r *permissionGroupRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
    return r.db.WithContext(ctx).
        Model(&model.PermissionGroup{}).
        Where("id = ? AND deleted_at IS NULL", id).
        Updates(fields).Error
}

func (r *permissionGroupRepository) Delete(ctx context.Context, id uint64) error {
    return r.db.WithContext(ctx).
        Where("id = ? AND deleted_at IS NULL", id).
        Delete(&model.PermissionGroup{}).Error
}

// SetRules はトランザクション内で既存ルールを全削除→新規一括挿入する
func (r *permissionGroupRepository) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("group_id = ?", groupID).Delete(&model.PermissionGroupRule{}).Error; err != nil {
            return err
        }
        if len(rules) == 0 {
            return nil
        }
        return tx.Create(&rules).Error
    })
}
```

### 2. Service（新規: `backend/internal/service/permission_group_service.go`）

```go
package service

import (
    "context"
    "fmt"
    "log/slog"
    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/animal-ekarte/backend/internal/repository"
)

type CreatePermissionGroupInput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Color       string `json:"color"`
}

type UpdatePermissionGroupInput struct {
    Name        *string `json:"name"`
    Description *string `json:"description"`
    Color       *string `json:"color"`
}

type SetPermissionGroupRulesInput struct {
    Rules []RuleInput `json:"rules"`
}

type RuleInput struct {
    Resource  string `json:"resource"`
    CanView   bool   `json:"can_view"`
    CanCreate bool   `json:"can_create"`
    CanEdit   bool   `json:"can_edit"`
    CanDelete bool   `json:"can_delete"`
}

type PermissionGroupService interface {
    List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
    GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
    Create(ctx context.Context, clinicID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error)
    Update(ctx context.Context, id uint64, input UpdatePermissionGroupInput) error
    Delete(ctx context.Context, id uint64) error
    SetRules(ctx context.Context, groupID uint64, input SetPermissionGroupRulesInput) error
}

type permissionGroupService struct {
    repo repository.PermissionGroupRepository
}

func NewPermissionGroupService(repo repository.PermissionGroupRepository) PermissionGroupService {
    return &permissionGroupService{repo: repo}
}

func (s *permissionGroupService) List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    return s.repo.FindByClinicID(ctx, clinicID)
}

func (s *permissionGroupService) GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
    g, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("permission group not found: %w", apperrors.ErrNotFound)
    }
    return g, nil
}

func (s *permissionGroupService) Create(ctx context.Context, clinicID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error) {
    color := input.Color
    if color == "" {
        color = "#6B7280"
    }
    group := &model.PermissionGroup{
        ClinicID:    clinicID,
        Name:        input.Name,
        Description: input.Description,
        Color:       color,
    }
    if err := s.repo.Create(ctx, group); err != nil {
        return nil, fmt.Errorf("failed to create permission group: %w", err)
    }
    slog.InfoContext(ctx, "permission group created", "id", group.ID, "clinic_id", clinicID, "name", group.Name)
    return group, nil
}

func (s *permissionGroupService) Update(ctx context.Context, id uint64, input UpdatePermissionGroupInput) error {
    fields := buildPermissionGroupUpdateFields(input)
    if len(fields) == 0 {
        return nil
    }
    return s.repo.UpdateFields(ctx, id, fields)
}

func (s *permissionGroupService) Delete(ctx context.Context, id uint64) error {
    slog.InfoContext(ctx, "deleting permission group", "id", id)
    return s.repo.Delete(ctx, id)
}

func (s *permissionGroupService) SetRules(ctx context.Context, groupID uint64, input SetPermissionGroupRulesInput) error {
    rules := make([]model.PermissionGroupRule, len(input.Rules))
    for i, r := range input.Rules {
        rules[i] = model.PermissionGroupRule{
            GroupID:   groupID,
            Resource:  r.Resource,
            CanView:   r.CanView,
            CanCreate: r.CanCreate,
            CanEdit:   r.CanEdit,
            CanDelete: r.CanDelete,
        }
    }
    slog.InfoContext(ctx, "setting permission group rules", "group_id", groupID, "rule_count", len(rules))
    return s.repo.SetRules(ctx, groupID, rules)
}

func buildPermissionGroupUpdateFields(input UpdatePermissionGroupInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name
    }
    if input.Description != nil {
        fields["description"] = *input.Description
    }
    if input.Color != nil {
        fields["color"] = *input.Color
    }
    return fields
}
```

### 3. Handler（新規: `backend/internal/handler/permission_group_handler.go`）

```go
package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/animal-ekarte/backend/internal/service"
)

// -- Request/Response 型 --

type createPermissionGroupRequest struct {
    Name        string `json:"name"        binding:"required,max=100"`
    Description string `json:"description"`
    Color       string `json:"color"`
}

type updatePermissionGroupRequest struct {
    Name        *string `json:"name"        binding:"omitempty,max=100"`
    Description *string `json:"description"`
    Color       *string `json:"color"`
}

type setPermissionGroupRulesRequest struct {
    Rules []ruleRequest `json:"rules" binding:"required"`
}

type ruleRequest struct {
    Resource  string `json:"resource"   binding:"required"`
    CanView   bool   `json:"can_view"`
    CanCreate bool   `json:"can_create"`
    CanEdit   bool   `json:"can_edit"`
    CanDelete bool   `json:"can_delete"`
}

// -- Handler --

func (h *Handler) ListPermissionGroups(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    groups, err := h.svc.PermissionGroup.List(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, groups)
}

func (h *Handler) CreatePermissionGroup(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    var req createPermissionGroupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    group, err := h.svc.PermissionGroup.Create(c.Request.Context(), clinicID, service.CreatePermissionGroupInput{
        Name:        req.Name,
        Description: req.Description,
        Color:       req.Color,
    })
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, group)
}

func (h *Handler) UpdatePermissionGroup(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    var req updatePermissionGroupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    if err := h.svc.PermissionGroup.Update(c.Request.Context(), id, service.UpdatePermissionGroupInput{
        Name:        req.Name,
        Description: req.Description,
        Color:       req.Color,
    }); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}

func (h *Handler) DeletePermissionGroup(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    if err := h.svc.PermissionGroup.Delete(c.Request.Context(), id); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}

func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    var req setPermissionGroupRulesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, parseBindError(err))
        return
    }
    rules := make([]service.RuleInput, len(req.Rules))
    for i, r := range req.Rules {
        rules[i] = service.RuleInput{
            Resource:  r.Resource,
            CanView:   r.CanView,
            CanCreate: r.CanCreate,
            CanEdit:   r.CanEdit,
            CanDelete: r.CanDelete,
        }
    }
    if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, service.SetPermissionGroupRulesInput{Rules: rules}); err != nil {
        RespondError(c, err)
        return
    }
    c.Status(http.StatusNoContent)
}

// RegisterPermissionGroupRoutes はルーティングを登録する
func (h *Handler) RegisterPermissionGroupRoutes(rg *gin.RouterGroup) {
    pg := rg.Group("/permission-groups")
    pg.GET("", h.ListPermissionGroups)
    pg.POST("", h.CreatePermissionGroup)
    pg.PATCH("/:id", h.UpdatePermissionGroup)
    pg.DELETE("/:id", h.DeletePermissionGroup)
    pg.PUT("/:id/rules", h.SetPermissionGroupRules)
}
```

### 4. Services 構造体へ追加（`backend/internal/handler/handler.go` または service の DI）

`PermissionGroup PermissionGroupService` フィールドを Services 構造体に追加し、`main.go` でDI配線する。

## API レスポンス形式

**GET /api/permission-groups:**
```json
[
  {
    "id": 1,
    "clinic_id": 1,
    "name": "受付スタッフ",
    "description": "受付業務担当",
    "color": "#3B82F6",
    "rules": [
      { "id": 1, "group_id": 1, "resource": "accounting", "can_view": true, "can_create": true, "can_edit": false, "can_delete": false },
      { "id": 2, "group_id": 1, "resource": "reservations", "can_view": true, "can_create": true, "can_edit": true, "can_delete": false }
    ]
  }
]
```

## 完了条件

- [ ] `GET /api/permission-groups?clinic_id=1` でグループ一覧（rules込み）が返る
- [ ] `POST /api/permission-groups` でグループ作成できる
- [ ] `PATCH /api/permission-groups/1` でname/description/colorを更新できる
- [ ] `DELETE /api/permission-groups/1` でグループが論理削除される
- [ ] `PUT /api/permission-groups/1/rules` でルールが一括更新（既存削除→新規作成）される
- [ ] 全エンドポイントで clinic_id スコープが守られている（他クリニックのグループを操作できない）

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `backend/internal/repository/permission_group_repository.go` — 新規作成（PermissionGroupRepository）
  - `backend/internal/service/permission_group_service.go` — 新規作成（PermissionGroupService）
  - `backend/internal/handler/permission_group_handler.go` — 新規作成（CRUD + SetRules endpoints）
  - `backend/internal/repository/repositories.go` — PermissionGroup フィールド追加
  - `backend/internal/service/service.go` — PermissionGroup フィールド追加
  - `backend/internal/handler/handler.go` — RegisterPermissionGroupRoutes 追加
  - `backend/internal/handler/user_account_request.go` — gofmt 修正
  - `backend/internal/handler/user_account_response.go` — gofmt 修正
