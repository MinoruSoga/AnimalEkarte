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

// ---- request types ----

type closeCashRegisterRequest struct {
	Date       string `json:"date"        binding:"required"` // YYYY-MM-DD
	Period     string `json:"period"      binding:"required"` // "am" or "pm"
	ActualCash int64  `json:"actual_cash"`
	Memo       string `json:"memo"`
}

// ---- handlers ----

// GetCashRegisterPreview godoc
// GET /v1/cash-register/preview?date=YYYY-MM-DD&period=am
func (h *Handler) GetCashRegisterPreview(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		RespondError(c, apperrors.WrapInvalidInput("date クエリパラメータは必須です"))
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください"))
		return
	}

	period := c.Query("period")
	if period == "" {
		RespondError(c, apperrors.WrapInvalidInput("period クエリパラメータは必須です"))
		return
	}

	preview, err := h.svc.CashRegister.GetPreview(c.Request.Context(), clinicID, date, period)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

// CloseCashRegister godoc
// POST /v1/cash-register/closes
func (h *Handler) CloseCashRegister(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	var req closeCashRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date は YYYY-MM-DD 形式で指定してください"))
		return
	}

	record, err := h.svc.CashRegister.Close(c.Request.Context(), clinicID, service.CloseRegisterInput{
		Date:       date,
		Period:     req.Period,
		ActualCash: req.ActualCash,
		Memo:       req.Memo,
		ClosedBy:   &staffID,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/cash-register/closes/%d", record.ID))
	c.JSON(http.StatusCreated, record)
}

// ListCashRegisterCloses godoc
// GET /v1/cash-register/closes?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&page=1&limit=20
func (h *Handler) ListCashRegisterCloses(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	startDateStr, err := parseDateQuery(c, "start_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	endDateStr, err := parseDateQuery(c, "end_date")
	if err != nil {
		RespondError(c, err)
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var startDate, endDate *time.Time
	if startDateStr != nil {
		t, _ := time.Parse("2006-01-02", *startDateStr)
		startDate = &t
	}
	if endDateStr != nil {
		t, _ := time.Parse("2006-01-02", *endDateStr)
		endDate = &t
	}

	closes, total, err := h.svc.CashRegister.List(c.Request.Context(), clinicID, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(closes, total, page, limit))
}

// GetCashRegisterClose godoc
// GET /v1/cash-register/closes/:id
func (h *Handler) GetCashRegisterClose(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	record, err := h.svc.CashRegister.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// RegisterCashRegisterRoutes はレジ締め関連のルートを登録する
func (h *Handler) RegisterCashRegisterRoutes(rg *gin.RouterGroup) {
	cr := rg.Group("/cash-register")
	cr.GET("/preview", h.RequirePermission(string(model.ResourceCashRegisterClose), "view"), h.GetCashRegisterPreview)
	cr.POST("/closes", h.RequirePermission(string(model.ResourceCashRegisterClose), "create"), h.CloseCashRegister)
	cr.GET("/closes", h.RequirePermission(string(model.ResourceCashRegisterClose), "view"), h.ListCashRegisterCloses)
	cr.GET("/closes/:id", h.RequirePermission(string(model.ResourceCashRegisterClose), "view"), h.GetCashRegisterClose)
}
