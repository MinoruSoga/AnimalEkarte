package lstep

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LineCustomerHandler は LineCustomerService の HTTP handler。
type LineCustomerHandler struct {
	svc               LineCustomerService
	requirePermission PermissionMiddleware
}

// NewLineCustomerHandler は LineCustomerHandler を構築する。
func NewLineCustomerHandler(svc LineCustomerService, requirePermission PermissionMiddleware) *LineCustomerHandler {
	return &LineCustomerHandler{svc: svc, requirePermission: requirePermission}
}

// ListLineCustomers godoc
func (h *LineCustomerHandler) ListLineCustomers(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toLineCustomerResponse))
}

// LinkOwnerToLineCustomer godoc
func (h *LineCustomerHandler) LinkOwnerToLineCustomer(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "customerId")
	if !ok {
		return
	}
	var req linkOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	customer, err := h.svc.LinkOwner(c.Request.Context(), clinicID, id, req.OwnerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineCustomerResponse(customer))
}
