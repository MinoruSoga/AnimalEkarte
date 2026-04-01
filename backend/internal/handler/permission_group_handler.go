package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// extractCompanyID はJWT認証済みコンテキストのclinic_idからcompanyIDを取得するヘルパー
func (h *Handler) extractCompanyID(c *gin.Context) (uint64, bool) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return 0, false
	}
	clinic, err := h.svc.Clinic.GetClinicByID(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return 0, false
	}
	return clinic.CompanyID, true
}

// ListPermissionGroups godoc
// GET /api/v1/permission-groups — company単位の権限グループ一覧を返す
func (h *Handler) ListPermissionGroups(c *gin.Context) {
	companyID, ok := h.extractCompanyID(c)
	if !ok {
		return
	}
	groups, err := h.svc.PermissionGroup.List(c.Request.Context(), companyID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreatePermissionGroup godoc
// POST /api/v1/permission-groups — 新規グループを作成する
func (h *Handler) CreatePermissionGroup(c *gin.Context) {
	companyID, ok := h.extractCompanyID(c)
	if !ok {
		return
	}
	var req createPermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	group, err := h.svc.PermissionGroup.Create(c.Request.Context(), companyID, service.CreatePermissionGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	h.writeAuditLog(c, model.AuditActionPermissionGroupCreate, "permission_group", &group.ID,
		repository.MarshalAuditJSON(map[string]any{"name": req.Name, "description": req.Description}))
	c.JSON(http.StatusCreated, group)
}

// GetPermissionGroup godoc
// GET /api/v1/permission-groups/:id — 指定IDの権限グループを返す
func (h *Handler) GetPermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	group, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

// UpdatePermissionGroup godoc
// PATCH /api/v1/permission-groups/:id — グループのname/description/colorを更新する
func (h *Handler) UpdatePermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updatePermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
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
	h.writeAuditLog(c, model.AuditActionPermissionGroupUpdate, "permission_group", &id,
		repository.MarshalAuditJSON(map[string]any{"name": req.Name, "description": req.Description}))
	c.Status(http.StatusNoContent)
}

// DeletePermissionGroup godoc
// DELETE /api/v1/permission-groups/:id — グループを論理削除する
func (h *Handler) DeletePermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	h.writeAuditLog(c, model.AuditActionPermissionGroupDelete, "permission_group", &id, nil)
	c.Status(http.StatusNoContent)
}

// SetPermissionGroupRules godoc
// PUT /api/v1/permission-groups/:id/rules — グループのルールを一括更新する（既存削除→新規作成）
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req setPermissionGroupRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	rules := make([]service.RuleInput, len(req.Rules))
	for i, r := range req.Rules {
		if !isValidResource(r.Resource) {
			RespondError(c, apperrors.WrapInvalidInput("invalid resource: "+r.Resource))
			return
		}
		rules[i] = service.RuleInput{
			Resource:  model.Resource(r.Resource),
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
	h.writeAuditLog(c, model.AuditActionPermissionRulesUpdate, "permission_group", &id,
		repository.MarshalAuditJSON(req.Rules))
	c.Status(http.StatusNoContent)
}

// isValidResource は与えられた文字列が有効なリソース識別子かどうかを検証する
func isValidResource(r string) bool {
	for _, valid := range model.AllResources {
		if string(valid) == r {
			return true
		}
	}
	return false
}

// RegisterPermissionGroupRoutes はルーティングを登録する
func (h *Handler) RegisterPermissionGroupRoutes(rg *gin.RouterGroup) {
	pg := rg.Group("/permission-groups")
	pg.GET("", h.ListPermissionGroups)
	pg.POST("", h.CreatePermissionGroup)
	pg.GET("/:id", h.GetPermissionGroup)
	pg.PATCH("/:id", h.UpdatePermissionGroup)
	pg.DELETE("/:id", h.DeletePermissionGroup)
	pg.PUT("/:id/rules", h.SetPermissionGroupRules)
}
