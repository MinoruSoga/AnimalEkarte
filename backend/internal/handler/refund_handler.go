package handler

import (
	"log/slog"
	"net/http"
	"strconv"

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

	billingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid billing id"))
		return
	}

	refunds, err := h.svc.Refund.ListByBillingID(c.Request.Context(), clinicID, billingID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toRefundResponseList(refunds))
}

// CreateRefund godoc
// @Summary 返金登録
// @Tags accounting
func (h *Handler) CreateRefund(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	billingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid billing id"))
		return
	}

	var req createRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	refund, err := h.svc.Refund.Create(ctx, clinicID, billingID, req.Amount, req.Reason)
	if err != nil {
		RespondError(c, err)
		return
	}

	slog.InfoContext(ctx, "refund created",
		slog.Uint64("billing_id", billingID),
		slog.Int64("amount", req.Amount))
	c.JSON(http.StatusCreated, toRefundResponse(refund))
}
