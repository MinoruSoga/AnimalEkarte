package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// --- handlers: auto_managed_prefixes ---

// ListAutoManagedPrefixes godoc
// GET /api/v1/lstep-tag-config/auto-managed-prefixes
func (h *Handler) ListAutoManagedPrefixes(c *gin.Context) {
	items, err := h.svc.LstepTagConfig.ListAutoManagedPrefixes(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAutoManagedPrefixListResponse(items))
}

// CreateAutoManagedPrefix godoc
// POST /api/v1/lstep-tag-config/auto-managed-prefixes
func (h *Handler) CreateAutoManagedPrefix(c *gin.Context) {
	var req createAutoManagedPrefixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	item, err := h.svc.LstepTagConfig.CreateAutoManagedPrefix(c.Request.Context(), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/auto-managed-prefixes/%d", item.ID))
	c.JSON(http.StatusCreated, toAutoManagedPrefixResponse(item))
}

// DeleteAutoManagedPrefix godoc
// DELETE /api/v1/lstep-tag-config/auto-managed-prefixes/:id
func (h *Handler) DeleteAutoManagedPrefix(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.LstepTagConfig.DeleteAutoManagedPrefix(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- handlers: condition_tag_mappings ---

// ListConditionTagMappings godoc
// GET /api/v1/lstep-tag-config/condition-tag-mappings
func (h *Handler) ListConditionTagMappings(c *gin.Context) {
	items, err := h.svc.LstepTagConfig.ListConditionTagMappings(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toConditionTagMappingListResponse(items))
}

// CreateConditionTagMapping godoc
// POST /api/v1/lstep-tag-config/condition-tag-mappings
func (h *Handler) CreateConditionTagMapping(c *gin.Context) {
	var req createConditionTagMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	item, err := h.svc.LstepTagConfig.CreateConditionTagMapping(c.Request.Context(), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/condition-tag-mappings/%d", item.ID))
	c.JSON(http.StatusCreated, toConditionTagMappingResponse(item))
}

// DeleteConditionTagMapping godoc
// DELETE /api/v1/lstep-tag-config/condition-tag-mappings/:id
func (h *Handler) DeleteConditionTagMapping(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.LstepTagConfig.DeleteConditionTagMapping(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- handlers: send_purpose_tag_prefixes ---

// ListSendPurposeTagPrefixes godoc
// GET /api/v1/lstep-tag-config/send-purpose-tag-prefixes
func (h *Handler) ListSendPurposeTagPrefixes(c *gin.Context) {
	items, err := h.svc.LstepTagConfig.ListSendPurposeTagPrefixes(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSendPurposeTagPrefixListResponse(items))
}

// CreateSendPurposeTagPrefix godoc
// POST /api/v1/lstep-tag-config/send-purpose-tag-prefixes
func (h *Handler) CreateSendPurposeTagPrefix(c *gin.Context) {
	var req createSendPurposeTagPrefixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	item, err := h.svc.LstepTagConfig.CreateSendPurposeTagPrefix(c.Request.Context(), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/send-purpose-tag-prefixes/%d", item.ID))
	c.JSON(http.StatusCreated, toSendPurposeTagPrefixResponse(item))
}

// DeleteSendPurposeTagPrefix godoc
// DELETE /api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id
func (h *Handler) DeleteSendPurposeTagPrefix(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.LstepTagConfig.DeleteSendPurposeTagPrefix(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterLstepTagConfigRoutes はLステップタグ設定のルートを登録する。
// hospital-settings リソースの権限で保護する（管理者専用）。
func (h *Handler) RegisterLstepTagConfigRoutes(rg *gin.RouterGroup) {
	perm := func(action string) gin.HandlerFunc {
		return h.RequirePermission(string(model.ResourceHospitalSettings), action)
	}

	prefixes := rg.Group("/lstep-tag-config/auto-managed-prefixes")
	prefixes.GET("", perm("view"), h.ListAutoManagedPrefixes)
	prefixes.POST("", perm("create"), h.CreateAutoManagedPrefix)
	prefixes.DELETE("/:id", perm("delete"), h.DeleteAutoManagedPrefix)

	conditions := rg.Group("/lstep-tag-config/condition-tag-mappings")
	conditions.GET("", perm("view"), h.ListConditionTagMappings)
	conditions.POST("", perm("create"), h.CreateConditionTagMapping)
	conditions.DELETE("/:id", perm("delete"), h.DeleteConditionTagMapping)

	purposes := rg.Group("/lstep-tag-config/send-purpose-tag-prefixes")
	purposes.GET("", perm("view"), h.ListSendPurposeTagPrefixes)
	purposes.POST("", perm("create"), h.CreateSendPurposeTagPrefix)
	purposes.DELETE("/:id", perm("delete"), h.DeleteSendPurposeTagPrefix)
}
