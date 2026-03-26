package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// createPermissionGroupRequest はグループ作成リクエスト
type createPermissionGroupRequest struct {
	Name        string `json:"name"        binding:"required,max=100"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// updatePermissionGroupRequest はグループ更新リクエスト（全フィールドオプション）
type updatePermissionGroupRequest struct {
	Name        *string `json:"name"        binding:"omitempty,max=100"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

// setPermissionGroupRulesRequest はルール一括更新リクエスト
type setPermissionGroupRulesRequest struct {
	Rules []ruleRequest `json:"rules" binding:"required"`
}

// ruleRequest は個別ルールのリクエスト
type ruleRequest struct {
	Resource  string `json:"resource"   binding:"required"`
	CanView   bool   `json:"can_view"`
	CanCreate bool   `json:"can_create"`
	CanEdit   bool   `json:"can_edit"`
	CanDelete bool   `json:"can_delete"`
}

// ListPermissionGroups godoc
// GET /api/v1/permission-groups — clinic_id（JWTクレーム）に紐づくグループ一覧を返す
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

// CreatePermissionGroup godoc
// POST /api/v1/permission-groups — 新規グループを作成する
func (h *Handler) CreatePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createPermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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

// UpdatePermissionGroup godoc
// PATCH /api/v1/permission-groups/:id — グループのname/description/colorを更新する
func (h *Handler) UpdatePermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updatePermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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

// DeletePermissionGroup godoc
// DELETE /api/v1/permission-groups/:id — グループを論理削除する
func (h *Handler) DeletePermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetPermissionGroupRules godoc
// PUT /api/v1/permission-groups/:id/rules — グループのルールを一括更新する（既存削除→新規作成）
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req setPermissionGroupRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
