package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListDailyRecords godoc
// GET /hospitalizations/:id/daily-records
func (h *Handler) ListDailyRecords(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	records, err := h.svc.DailyRecord.List(c.Request.Context(), clinicID, hospitalizationID)
	if err != nil {
		RespondError(c, err)
		return
	}

	items := mapSlice(records, toDailyRecordResponse)
	c.JSON(http.StatusOK, items)
}

// GetDailyRecord godoc
// GET /hospitalizations/:id/daily-records/:date
func (h *Handler) GetDailyRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	// view GET は読み取り専用。未存在日は NotFound（FirstOrCreate しない — AUD-003）。
	record, err := h.svc.DailyRecord.GetByDate(c.Request.Context(), clinicID, hospitalizationID, date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDailyRecordResponse(record))
}

// CreateDailyRecord godoc
// POST /hospitalizations/:id/daily-records
func (h *Handler) CreateDailyRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Date string `json:"date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	date, err := time.ParseInLocation(time.DateOnly, req.Date, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	record, err := h.svc.DailyRecord.FindOrCreateByDate(c.Request.Context(), clinicID, hospitalizationID, date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", record.HospitalizationID, record.Date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddVitalRecord godoc
// POST /hospitalizations/:id/daily-records/:date/vitals
func (h *Handler) AddVitalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addVitalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.svc.DailyRecord.AddVitalRecord(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddCareLog godoc
// POST /hospitalizations/:id/daily-records/:date/care-logs
func (h *Handler) AddCareLog(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addCareLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.svc.DailyRecord.AddCareLog(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddStaffNote godoc
// POST /hospitalizations/:id/daily-records/:date/staff-notes
func (h *Handler) AddStaffNote(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addStaffNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.svc.DailyRecord.AddStaffNote(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// RegisterDailyRecordRoutes は日次記録関連のルートを登録する
func (h *Handler) RegisterDailyRecordRoutes(rg *gin.RouterGroup) {
	permCreate := h.RequirePermission(string(model.ResourceHospitalization), "create")
	rg.GET("/:id/daily-records", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.ListDailyRecords)
	rg.POST("/:id/daily-records", permCreate, h.CreateDailyRecord)
	rg.GET("/:id/daily-records/:date", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.GetDailyRecord)
	rg.POST("/:id/daily-records/:date/vitals", permCreate, h.AddVitalRecord)
	rg.POST("/:id/daily-records/:date/care-logs", permCreate, h.AddCareLog)
	rg.POST("/:id/daily-records/:date/staff-notes", permCreate, h.AddStaffNote)
}
