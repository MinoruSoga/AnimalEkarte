package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// shiftTemplateBreakRequest は休憩時間テンプレートのリクエストDTO
type shiftTemplateBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required"`
	BreakEnd   string `json:"break_end"   binding:"required"`
}

// createShiftTemplateRequest はシフトテンプレート作成リクエスト
type createShiftTemplateRequest struct {
	Name      string                      `json:"name"       binding:"required"`
	ShiftType string                      `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	StartTime string                      `json:"start_time"`
	EndTime   string                      `json:"end_time"`
	Notes     string                      `json:"notes"`
	SortOrder int                         `json:"sort_order"`
	IsActive  *bool                       `json:"is_active"`
	Breaks    []shiftTemplateBreakRequest `json:"breaks"`
}

// updateShiftTemplateRequest はシフトテンプレート更新リクエスト（PATCH）
type updateShiftTemplateRequest struct {
	Name      *string                      `json:"name"`
	ShiftType *string                      `json:"shift_type" binding:"omitempty,oneof=full morning afternoon off paid_leave"`
	StartTime *string                      `json:"start_time"`
	EndTime   *string                      `json:"end_time"`
	Notes     *string                      `json:"notes"`
	SortOrder *int                         `json:"sort_order"`
	IsActive  *bool                        `json:"is_active"`
	Breaks    *[]shiftTemplateBreakRequest `json:"breaks"`
}

// reorderShiftTemplateRequest はシフトテンプレート並び替えリクエスト
type reorderShiftTemplateRequest struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

// shiftTemplateBreakResponse は休憩時間テンプレートのレスポンスDTO
type shiftTemplateBreakResponse struct {
	ID         string `json:"id"`
	BreakStart string `json:"break_start"`
	BreakEnd   string `json:"break_end"`
}

// shiftTemplateResponse はシフトテンプレートのレスポンスDTO
type shiftTemplateResponse struct {
	ID        string                       `json:"id"`
	ClinicID  string                       `json:"clinic_id"`
	Name      string                       `json:"name"`
	ShiftType string                       `json:"shift_type"`
	StartTime string                       `json:"start_time"`
	EndTime   string                       `json:"end_time"`
	Notes     string                       `json:"notes"`
	SortOrder int                          `json:"sort_order"`
	IsActive  bool                         `json:"is_active"`
	Breaks    []shiftTemplateBreakResponse `json:"breaks"`
}

func toShiftTemplateResponse(tpl *model.ShiftTemplate) shiftTemplateResponse {
	breaks := make([]shiftTemplateBreakResponse, 0, len(tpl.Breaks))
	for _, b := range tpl.Breaks {
		breaks = append(breaks, shiftTemplateBreakResponse{
			ID:         strconv.FormatUint(b.ID, 10),
			BreakStart: b.BreakStart,
			BreakEnd:   b.BreakEnd,
		})
	}
	startTime := ""
	if tpl.StartTime != nil {
		startTime = *tpl.StartTime
	}
	endTime := ""
	if tpl.EndTime != nil {
		endTime = *tpl.EndTime
	}
	return shiftTemplateResponse{
		ID:        strconv.FormatUint(tpl.ID, 10),
		ClinicID:  strconv.FormatUint(tpl.ClinicID, 10),
		Name:      tpl.Name,
		ShiftType: string(tpl.ShiftType),
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     tpl.Notes,
		SortOrder: tpl.SortOrder,
		IsActive:  tpl.IsActive,
		Breaks:    breaks,
	}
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
	result := make([]shiftTemplateResponse, 0, len(items))
	for i := range items {
		result = append(result, toShiftTemplateResponse(&items[i]))
	}
	c.JSON(http.StatusOK, result)
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
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	breaks := make([]service.ShiftBreakTemplateInput, 0, len(req.Breaks))
	for _, b := range req.Breaks {
		breaks = append(breaks, service.ShiftBreakTemplateInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
	}
	var startTime, endTime *string
	if req.StartTime != "" {
		startTime = &req.StartTime
	}
	if req.EndTime != "" {
		endTime = &req.EndTime
	}
	tpl, err := h.svc.ShiftTemplate.Create(c.Request.Context(), clinicID, &service.CreateShiftTemplateInput{
		Name:      req.Name,
		ShiftType: model.ShiftType(req.ShiftType),
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     req.Notes,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
		Breaks:    breaks,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toShiftTemplateResponse(tpl))
}

// UpdateShiftTemplate PATCH /api/v1/shift-templates/:id
func (h *Handler) UpdateShiftTemplate(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	var req reorderShiftTemplateRequest
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
	g.PUT("/reorder", h.RequirePermission(string(model.ResourceShifts), "edit"), h.ReorderShiftTemplates)
	g.PATCH("/:id", h.RequirePermission(string(model.ResourceShifts), "edit"), h.UpdateShiftTemplate)
	g.DELETE("/:id", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteShiftTemplate)
}
