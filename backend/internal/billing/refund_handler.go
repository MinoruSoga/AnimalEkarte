package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// RefundHandler は RefundService の HTTP handler。
type RefundHandler struct {
	svc               RefundService
	requirePermission PermissionMiddleware
}

// NewRefundHandler は RefundHandler を構築する。
func NewRefundHandler(svc RefundService, requirePermission PermissionMiddleware) *RefundHandler {
	return &RefundHandler{svc: svc, requirePermission: requirePermission}
}

// ListRefunds godoc
// @Summary 返金一覧取得
// @Tags accounting
func (h *RefundHandler) ListRefunds(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
		return
	}

	billingID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	refunds, err := h.svc.ListByBillingID(c.Request.Context(), clinicID, billingID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(refunds, ToRefundResponse))
}

// CreateRefund godoc
// @Summary 返金登録
// @Tags accounting
func (h *RefundHandler) CreateRefund(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	billingID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	var paymentMethod *model.PaymentMethod
	if req.PaymentMethod != nil {
		pm := model.PaymentMethod(*req.PaymentMethod)
		paymentMethod = &pm
	}
	refund, err := h.svc.Create(ctx, clinicID, billingID, CreateRefundInput{
		StaffID:       &staffID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/billing/%d/refunds/%d", billingID, refund.ID))
	c.JSON(http.StatusCreated, ToRefundResponse(refund))
}
