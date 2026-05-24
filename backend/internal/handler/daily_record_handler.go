package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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

	items := make([]dailyRecordResponse, 0, len(records))
	for i := range records {
		items = append(items, toDailyRecordResponse(&records[i]))
	}
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
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	record, err := h.svc.DailyRecord.FindOrCreateByDate(c.Request.Context(), clinicID, hospitalizationID, date)
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

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	record, err := h.svc.DailyRecord.FindOrCreateByDate(c.Request.Context(), clinicID, hospitalizationID, date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", record.HospitalizationID, record.Date.Format("2006-01-02")))
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
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addVitalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	parsedVitalTime, err := time.Parse("15:04:05", req.Time)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid time format, expected HH:MM:SS"))
		return
	}

	input := &service.CreateVitalRecordInput{
		Time:            parsedVitalTime,
		Temperature:     req.Temperature,
		HeartRate:       req.HeartRate,
		RespirationRate: req.RespirationRate,
		Weight:          req.Weight,
		Notes:           req.Notes,
		StaffID:         req.StaffID,
	}

	record, err := h.svc.DailyRecord.AddVitalRecord(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.Format("2006-01-02")))
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
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addCareLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	parsedCareLogTime, err := time.Parse("15:04:05", req.Time)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid time format, expected HH:MM:SS"))
		return
	}

	input := &service.CreateCareLogInput{
		Time:    parsedCareLogTime,
		Type:    req.Type,
		Status:  req.Status,
		Value:   req.Value,
		StaffID: req.StaffID,
		Notes:   req.Notes,
	}

	record, err := h.svc.DailyRecord.AddCareLog(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.Format("2006-01-02")))
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
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addStaffNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	parsedStaffNoteTime, err := time.Parse("15:04:05", req.Time)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid time format, expected HH:MM:SS"))
		return
	}

	input := &service.CreateStaffNoteInput{
		Time:    parsedStaffNoteTime,
		Content: req.Content,
		StaffID: req.StaffID,
	}

	record, err := h.svc.DailyRecord.AddStaffNote(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.Format("2006-01-02")))
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
