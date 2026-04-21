package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetShiftTemplate GET /api/v1/shift-templates/:id
func (h *Handler) GetShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
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
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	breaks := make([]service.ShiftBreakTemplateInput, 0, len(req.Breaks))
	for _, b := range req.Breaks {
		breaks = append(breaks, service.ShiftBreakTemplateInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
	}
	tpl, err := h.svc.ShiftTemplate.Create(c.Request.Context(), clinicID, &service.CreateShiftTemplateInput{
		Name:      req.Name,
		ShiftType: req.ShiftType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Notes:     req.Notes,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
		Breaks:    breaks,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/shift-templates/%d", tpl.ID))
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
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := &service.UpdateShiftTemplateInput{
		Name:      req.Name,
		Notes:     req.Notes,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	if req.ShiftType != nil {
		st := model.ShiftType(*req.ShiftType)
		input.ShiftType = &st
	}
	if req.Breaks != nil {
		breaks := make([]service.ShiftBreakTemplateInput, 0, len(*req.Breaks))
		for _, b := range *req.Breaks {
			breaks = append(breaks, service.ShiftBreakTemplateInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		input.Breaks = &breaks
	}
	tpl, err := h.svc.ShiftTemplate.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
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

// ReorderShiftTemplates PUT /api/v1/shift-templates/reorder
func (h *Handler) ReorderShiftTemplates(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
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
	g.GET("", h.ListShiftTemplates)
	g.POST("", h.RequirePermission(string(model.ResourceShifts), "create"), h.CreateShiftTemplate)
	g.PATCH("/reorder", h.RequirePermission(string(model.ResourceShifts), "edit"), h.ReorderShiftTemplates)
	g.GET("/:id", h.GetShiftTemplate)
	g.PATCH("/:id", h.RequirePermission(string(model.ResourceShifts), "edit"), h.UpdateShiftTemplate)
	g.DELETE("/:id", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteShiftTemplate)
}
