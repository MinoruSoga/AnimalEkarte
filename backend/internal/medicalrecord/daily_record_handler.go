package medicalrecord

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// DailyRecordHandler serves the daily-record HTTP boundary. Moved from internal/handler
// (BE9-2D ⑤ Batch C).
type DailyRecordHandler struct {
	service DailyRecordService
}

// NewDailyRecordHandler initializes a DailyRecordHandler.
func NewDailyRecordHandler(service DailyRecordService) *DailyRecordHandler {
	return &DailyRecordHandler{service: service}
}

// ListDailyRecords godoc
// GET /hospitalizations/:id/daily-records
func (h *DailyRecordHandler) ListDailyRecords(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceHospitalization), "view") {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	records, err := h.service.List(c.Request.Context(), clinicID, hospitalizationID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items := httpapi.MapSlice(records, toDailyRecordResponse)
	c.JSON(http.StatusOK, items)
}

// GetDailyRecord godoc
// GET /hospitalizations/:id/daily-records/:date
func (h *DailyRecordHandler) GetDailyRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceHospitalization), "view") {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	// view GET は読み取り専用。未存在日は NotFound（FirstOrCreate しない — AUD-003）。
	record, err := h.service.GetByDate(c.Request.Context(), clinicID, hospitalizationID, date)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDailyRecordResponse(record))
}

// CreateDailyRecord godoc
// POST /hospitalizations/:id/daily-records
func (h *DailyRecordHandler) CreateDailyRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Date string `json:"date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	date, err := time.ParseInLocation(time.DateOnly, req.Date, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	record, err := h.service.FindOrCreateByDate(c.Request.Context(), clinicID, hospitalizationID, date)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", record.HospitalizationID, record.Date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddVitalRecord godoc
// POST /hospitalizations/:id/daily-records/:date/vitals
func (h *DailyRecordHandler) AddVitalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addVitalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.service.AddVitalRecord(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddCareLog godoc
// POST /hospitalizations/:id/daily-records/:date/care-logs
func (h *DailyRecordHandler) AddCareLog(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addCareLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.service.AddCareLog(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}

// AddStaffNote godoc
// POST /hospitalizations/:id/daily-records/:date/staff-notes
func (h *DailyRecordHandler) AddStaffNote(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	dateStr := c.Param("date")
	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD"))
		return
	}

	var req addStaffNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.service.AddStaffNote(c.Request.Context(), clinicID, hospitalizationID, date, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", hospitalizationID, date.In(time.Local).Format(time.DateOnly)))
	c.JSON(http.StatusCreated, toDailyRecordResponse(record))
}
