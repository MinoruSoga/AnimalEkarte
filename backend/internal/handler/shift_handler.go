package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListShiftEntries godoc
func (h *Handler) ListShiftEntries(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	yearMonth := c.Query("date") // YYYY-MM

	var staffID *uint64
	if s := c.Query("staff_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid staff_id"))
			return
		}
		staffID = &id
	}

	shifts, err := h.svc.ShiftEntry.List(c.Request.Context(), clinicID, yearMonth, staffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toShiftResponseList(shifts))
}

// CreateShiftEntry godoc
func (h *Handler) CreateShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date: use YYYY-MM-DD"))
		return
	}

	var startTime, endTime *string
	if req.StartTime != "" {
		startTime = &req.StartTime
	}
	if req.EndTime != "" {
		endTime = &req.EndTime
	}
	breaks := make([]service.ShiftBreakInput, 0, len(req.Breaks))
	for _, b := range req.Breaks {
		breaks = append(breaks, service.ShiftBreakInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
	}
	shift, err := h.svc.ShiftEntry.Create(c.Request.Context(), clinicID, &service.CreateShiftEntryInput{
		StaffID:   req.StaffID,
		Date:      date,
		ShiftType: model.ShiftType(req.ShiftType),
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     req.Notes,
		Breaks:    breaks,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toShiftResponse(shift))
}

// UpdateShiftEntry godoc
func (h *Handler) UpdateShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateShiftEntryInput{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Notes:     req.Notes,
	}
	if req.ShiftType != nil {
		st := model.ShiftType(*req.ShiftType)
		input.ShiftType = &st
	}
	if req.Breaks != nil {
		breaks := make([]service.ShiftBreakInput, 0, len(*req.Breaks))
		for _, b := range *req.Breaks {
			breaks = append(breaks, service.ShiftBreakInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
		}
		input.Breaks = &breaks
	}

	shift, err := h.svc.ShiftEntry.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toShiftResponse(shift))
}

// DeleteShiftEntry godoc
func (h *Handler) DeleteShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	if err := h.svc.ShiftEntry.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterShiftRoutes はシフト関連のルートを登録する
func (h *Handler) RegisterShiftRoutes(rg *gin.RouterGroup) {
	shifts := rg.Group("/shifts")
	shifts.GET("", h.ListShiftEntries)
	shifts.POST("", h.RequirePermission(string(model.ResourceShifts), "create"), h.CreateShiftEntry)
	shifts.PATCH("/:id", h.RequirePermission(string(model.ResourceShifts), "edit"), h.UpdateShiftEntry)
	shifts.DELETE("/:id", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteShiftEntry)
}
