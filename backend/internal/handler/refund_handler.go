package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListRefunds godoc
// @Summary 返金一覧取得
// @Tags accounting
func (h *Handler) ListRefunds(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	billingID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	refunds, err := h.svc.Refund.ListByBillingID(c.Request.Context(), clinicID, billingID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(refunds, toRefundResponse))
}

// CreateRefund godoc
// @Summary 返金登録
// @Tags accounting
func (h *Handler) CreateRefund(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	billingID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	refund, err := h.svc.Refund.Create(ctx, clinicID, billingID, &staffID, req.Amount, req.Reason)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/billing/%d/refunds/%d", billingID, refund.ID))
	c.JSON(http.StatusCreated, toRefundResponse(refund))
}
