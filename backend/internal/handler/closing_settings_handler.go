package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- request/response types ----

type updateClinicSettingsRequest struct {
	ClosingAmPmBoundary *string `json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   *string `json:"closing_weekday_end"`
	ClosingSundayEnd    *string `json:"closing_sunday_end"`
	ClosedWeekdays      []int64 `json:"closed_weekdays"`
}

type createSpecialPeriodRequest struct {
	StartDate    string `json:"start_date"    binding:"required"` // YYYY-MM-DD
	EndDate      string `json:"end_date"      binding:"required"` // YYYY-MM-DD
	AmPmBoundary string `json:"am_pm_boundary" binding:"required"`
	PmEnd        string `json:"pm_end"         binding:"required"`
	Note         string `json:"note"`
}

type updateSpecialPeriodRequest struct {
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	AmPmBoundary *string `json:"am_pm_boundary"`
	PmEnd        *string `json:"pm_end"`
	Note         *string `json:"note"`
}

// ---- handlers ----

// GetClosingSettings godoc
// GET /v1/closing-settings
func (h *Handler) GetClosingSettings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	resp, err := h.svc.ClosingSettings.Get(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateClosingSettings godoc
// PATCH /v1/closing-settings
func (h *Handler) UpdateClosingSettings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req updateClinicSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	result, err := h.svc.ClosingSettings.UpdateStandard(c.Request.Context(), clinicID, service.UpdateClinicSettingsInput{
		ClosingAmPmBoundary: req.ClosingAmPmBoundary,
		ClosingWeekdayEnd:   req.ClosingWeekdayEnd,
		ClosingSundayEnd:    req.ClosingSundayEnd,
		ClosedWeekdays:      req.ClosedWeekdays,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListSpecialPeriods godoc
// GET /v1/closing-settings/special-periods
func (h *Handler) ListSpecialPeriods(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	periods, err := h.svc.ClosingSettings.ListSpecialPeriods(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, periods)
}

// CreateSpecialPeriod godoc
// POST /v1/closing-settings/special-periods
func (h *Handler) CreateSpecialPeriod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createSpecialPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください"))
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください"))
		return
	}
	period, err := h.svc.ClosingSettings.CreateSpecialPeriod(c.Request.Context(), clinicID, service.CreateSpecialPeriodInput{
		StartDate:    startDate,
		EndDate:      endDate,
		AmPmBoundary: req.AmPmBoundary,
		PmEnd:        req.PmEnd,
		Note:         req.Note,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, period)
}

// UpdateSpecialPeriod godoc
// PATCH /v1/closing-settings/special-periods/:id
func (h *Handler) UpdateSpecialPeriod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateSpecialPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := service.UpdateSpecialPeriodInput{
		AmPmBoundary: req.AmPmBoundary,
		PmEnd:        req.PmEnd,
		Note:         req.Note,
	}
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください"))
			return
		}
		input.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください"))
			return
		}
		input.EndDate = &t
	}

	period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, period)
}

// DeleteSpecialPeriod godoc
// DELETE /v1/closing-settings/special-periods/:id
func (h *Handler) DeleteSpecialPeriod(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ClosingSettings.DeleteSpecialPeriod(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClosingSettingsRoutes は締め設定関連のルートを登録する
func (h *Handler) RegisterClosingSettingsRoutes(rg *gin.RouterGroup) {
	cs := rg.Group("/closing-settings")
	cs.GET("", h.RequirePermission(string(model.ResourceClosingSettings), "view"), h.GetClosingSettings)
	cs.PATCH("", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.UpdateClosingSettings)

	sp := cs.Group("/special-periods")
	sp.GET("", h.RequirePermission(string(model.ResourceClosingSettings), "view"), h.ListSpecialPeriods)
	sp.POST("", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.CreateSpecialPeriod)
	sp.PATCH("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.UpdateSpecialPeriod)
	sp.DELETE("/:id", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.DeleteSpecialPeriod)

	// 休診日は既存 ClinicHoliday ハンドラに委譲
	holidays := cs.Group("/holidays")
	holidays.GET("", h.ListClinicHolidays)
	holidays.POST("", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.SetClinicHoliday)
	holidays.DELETE("/:date", h.RequirePermission(string(model.ResourceClosingSettings), "edit"), h.DeleteClinicHoliday)
}
