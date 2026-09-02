package lstep

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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
// Response body stays a JSON array for OpenAPI/FE compatibility.
// G2F-05: truncation is surfaced via headers (not silent 200-cap):
//
//	X-Total-Count, X-Limit, X-Truncated
func (h *LineCustomerHandler) ListLineCustomers(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	result, err := h.svc.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(result.Total, 10))
	c.Header("X-Limit", strconv.Itoa(result.Limit))
	if result.Truncated {
		c.Header("X-Truncated", "true")
	} else {
		c.Header("X-Truncated", "false")
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(result.Items, toLineCustomerResponse))
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
