package billing

import (
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

	// held for EMR-506: service は常に固定エラー（422）で拒否する。
	respondRefundPending(c, h.svc.ListByBillingID(c.Request.Context(), clinicID, billingID))
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
	// held for EMR-506: body の形式検証の後、service が固定エラー（422）で拒否する。
	respondRefundPending(c, h.svc.Create(ctx, clinicID, billingID, CreateRefundInput{
		StaffID:       &staffID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		PaymentMethod: paymentMethod,
	}))
}

// respondRefundPending は quarantine 中の service の結果を応答する。
// held for EMR-506: service は常に固定エラー（422）を返す。nil が返っても成功応答は組み立てず 500 で止める（fail-closed）。
func respondRefundPending(c *gin.Context, err error) {
	if err == nil {
		err = apperrors.WrapInternalServerError("refund quarantine returned without refusal")
	}
	httpapi.RespondError(c, err)
}
