package lstep

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// --- handlers: auto_managed_prefixes ---

// ListAutoManagedPrefixes godoc
// GET /api/v1/lstep-tag-config/auto-managed-prefixes
func (h *Handler) ListAutoManagedPrefixes(c *gin.Context) {
	items, err := h.tagConfig.ListAutoManagedPrefixes(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAutoManagedPrefixListResponse(items))
}

// CreateAutoManagedPrefix godoc
// POST /api/v1/lstep-tag-config/auto-managed-prefixes
func (h *Handler) CreateAutoManagedPrefix(c *gin.Context) {
	var req createAutoManagedPrefixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	item, err := h.tagConfig.CreateAutoManagedPrefix(c.Request.Context(), req.toServiceInput())
	if err != nil {
		// BUG-026: emit stable prefix-conflict code + params when present.
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/auto-managed-prefixes/%d", item.ID))
	c.JSON(http.StatusCreated, toAutoManagedPrefixResponse(item))
}

// DeleteAutoManagedPrefix godoc
// DELETE /api/v1/lstep-tag-config/auto-managed-prefixes/:id
func (h *Handler) DeleteAutoManagedPrefix(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.tagConfig.DeleteAutoManagedPrefix(c.Request.Context(), id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- handlers: condition_tag_mappings ---

// ListConditionTagMappings godoc
// GET /api/v1/lstep-tag-config/condition-tag-mappings
func (h *Handler) ListConditionTagMappings(c *gin.Context) {
	items, err := h.tagConfig.ListConditionTagMappings(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toConditionTagMappingListResponse(items))
}

// CreateConditionTagMapping godoc
// POST /api/v1/lstep-tag-config/condition-tag-mappings
func (h *Handler) CreateConditionTagMapping(c *gin.Context) {
	var req createConditionTagMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	item, err := h.tagConfig.CreateConditionTagMapping(c.Request.Context(), req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/condition-tag-mappings/%d", item.ID))
	c.JSON(http.StatusCreated, toConditionTagMappingResponse(item))
}

// DeleteConditionTagMapping godoc
// DELETE /api/v1/lstep-tag-config/condition-tag-mappings/:id
func (h *Handler) DeleteConditionTagMapping(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.tagConfig.DeleteConditionTagMapping(c.Request.Context(), id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- handlers: send_purpose_tag_prefixes ---

// ListSendPurposeTagPrefixes godoc
// GET /api/v1/lstep-tag-config/send-purpose-tag-prefixes
func (h *Handler) ListSendPurposeTagPrefixes(c *gin.Context) {
	items, err := h.tagConfig.ListSendPurposeTagPrefixes(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSendPurposeTagPrefixListResponse(items))
}

// CreateSendPurposeTagPrefix godoc
// POST /api/v1/lstep-tag-config/send-purpose-tag-prefixes
func (h *Handler) CreateSendPurposeTagPrefix(c *gin.Context) {
	var req createSendPurposeTagPrefixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	item, err := h.tagConfig.CreateSendPurposeTagPrefix(c.Request.Context(), req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-config/send-purpose-tag-prefixes/%d", item.ID))
	c.JSON(http.StatusCreated, toSendPurposeTagPrefixResponse(item))
}

// DeleteSendPurposeTagPrefix godoc
// DELETE /api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id
func (h *Handler) DeleteSendPurposeTagPrefix(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.tagConfig.DeleteSendPurposeTagPrefix(c.Request.Context(), id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
