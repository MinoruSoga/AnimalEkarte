package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListLineCustomers godoc
func (h *Handler) ListLineCustomers(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	items, err := h.svc.LineCustomer.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineCustomerResponseList(items))
}

// LinkOwnerToLineCustomer godoc
func (h *Handler) LinkOwnerToLineCustomer(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req linkOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	customer, err := h.svc.LineCustomer.LinkOwner(c.Request.Context(), clinicID, id, req.OwnerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineCustomerResponse(customer))
}
