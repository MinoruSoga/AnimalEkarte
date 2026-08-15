package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// PaymentMethodMasterHandler は PaymentMethodMasterService の HTTP handler。
type PaymentMethodMasterHandler struct {
	svc PaymentMethodMasterService
}

// NewPaymentMethodMasterHandler は PaymentMethodMasterHandler を構築する。
func NewPaymentMethodMasterHandler(svc PaymentMethodMasterService) *PaymentMethodMasterHandler {
	return &PaymentMethodMasterHandler{svc: svc}
}

// ListPaymentMethods godoc
// GET /v1/payment-methods
func (h *PaymentMethodMasterHandler) ListPaymentMethods(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ms, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(ms, ToPaymentMethodResponse))
}

// GetPaymentMethod godoc
// GET /v1/payment-methods/:id
func (h *PaymentMethodMasterHandler) GetPaymentMethod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToPaymentMethodResponse(m))
}

// CreatePaymentMethod godoc
// POST /v1/payment-methods
func (h *PaymentMethodMasterHandler) CreatePaymentMethod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	m, err := h.svc.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/payment-methods/%d", m.ID))
	c.JSON(http.StatusCreated, ToPaymentMethodResponse(m))
}

// UpdatePaymentMethod godoc
// PATCH /v1/payment-methods/:id
func (h *PaymentMethodMasterHandler) UpdatePaymentMethod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	m, err := h.svc.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToPaymentMethodResponse(m))
}

// ReorderPaymentMethods godoc
// PATCH /v1/payment-methods/reorder
func (h *PaymentMethodMasterHandler) ReorderPaymentMethods(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeletePaymentMethod godoc
// DELETE /v1/payment-methods/:id
func (h *PaymentMethodMasterHandler) DeletePaymentMethod(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
