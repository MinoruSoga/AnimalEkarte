package staff

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetShiftTemplate GET /api/v1/shift-templates/:id
func (h *Handler) GetShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceShifts), "view") {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	tpl, err := h.svc.ShiftTemplate.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toShiftTemplateResponse(tpl))
}

// ListShiftTemplates GET /api/v1/shift-templates
func (h *Handler) ListShiftTemplates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceShifts), "view") {
		return
	}
	items, err := h.svc.ShiftTemplate.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toShiftTemplateResponse))
}

// CreateShiftTemplate POST /api/v1/shift-templates
func (h *Handler) CreateShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createShiftTemplateRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	tpl, err := h.svc.ShiftTemplate.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		// BUG-026: emit stable name-conflict code + params when present.
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/shift-templates/%d", tpl.ID))
	c.JSON(http.StatusCreated, toShiftTemplateResponse(tpl))
}

// UpdateShiftTemplate PATCH /api/v1/shift-templates/:id
func (h *Handler) UpdateShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateShiftTemplateRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	tpl, err := h.svc.ShiftTemplate.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		// BUG-026: emit stable name-conflict code + params when present.
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toShiftTemplateResponse(tpl))
}

// DeleteShiftTemplate DELETE /api/v1/shift-templates/:id
func (h *Handler) DeleteShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ShiftTemplate.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderShiftTemplates PATCH /api/v1/shift-templates/reorder
func (h *Handler) ReorderShiftTemplates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if err := h.svc.ShiftTemplate.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterShiftTemplateRoutes はシフトテンプレートのルートを登録する
func (h *Handler) RegisterShiftTemplateRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/shift-templates")
	g.GET("", h.RequirePermission(string(model.ResourceShifts), "view"), h.ListShiftTemplates)
	g.POST("", h.RequirePermission(string(model.ResourceShifts), "create"), h.CreateShiftTemplate)
	g.PATCH("/reorder", h.RequirePermission(string(model.ResourceShifts), "edit"), h.ReorderShiftTemplates)
	g.GET("/:id", h.RequirePermission(string(model.ResourceShifts), "view"), h.GetShiftTemplate)
	g.PATCH("/:id", h.RequirePermission(string(model.ResourceShifts), "edit"), h.UpdateShiftTemplate)
	g.DELETE("/:id", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteShiftTemplate)
}
