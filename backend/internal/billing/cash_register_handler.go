package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CashRegisterHandler は CashRegisterService の HTTP handler。
type CashRegisterHandler struct {
	svc               CashRegisterService
	requirePermission PermissionMiddleware
}

// NewCashRegisterHandler は CashRegisterHandler を構築する。
func NewCashRegisterHandler(svc CashRegisterService, requirePermission PermissionMiddleware) *CashRegisterHandler {
	return &CashRegisterHandler{svc: svc, requirePermission: requirePermission}
}

// ---- handlers ----

// GetCashRegisterPreview godoc
// GET /v1/cash-register/preview?date=YYYY-MM-DD&period=am
func (h *CashRegisterHandler) GetCashRegisterPreview(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceCashRegisterClose), "view") {
		return
	}
	query := newCashRegisterPreviewQuery(c.Request.URL.Query())
	preview, err := h.svc.GetPreview(c.Request.Context(), clinicID, query.Date, query.Period)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCashRegisterPreviewResponse(preview))
}

// CloseCashRegister godoc
// POST /v1/cash-register/closes
func (h *CashRegisterHandler) CloseCashRegister(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req closeCashRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput(staffID)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.svc.Close(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/cash-register/closes/%d", record.ID))
	c.JSON(http.StatusCreated, toCashRegisterCloseResponse(record))
}

// ListCashRegisterCloses godoc
// GET /v1/cash-register/closes?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&page=1&limit=20
func (h *CashRegisterHandler) ListCashRegisterCloses(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceCashRegisterClose), "view") {
		return
	}

	filters, err := newListCashRegisterClosesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	closes, total, err := h.svc.List(c.Request.Context(), clinicID, filters.StartDate, filters.EndDate, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(closes, toCashRegisterCloseResponse), total, page, limit))
}

// GetCashRegisterClose godoc
// GET /v1/cash-register/closes/:id
func (h *CashRegisterHandler) GetCashRegisterClose(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceCashRegisterClose), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	record, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCashRegisterCloseResponse(record))
}
