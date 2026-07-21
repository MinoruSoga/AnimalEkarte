package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// BillingConfirmationHandler は BillingConfirmationService の HTTP handler。
type BillingConfirmationHandler struct {
	svc               BillingConfirmationService
	requirePermission PermissionMiddleware
}

// NewBillingConfirmationHandler は BillingConfirmationHandler を構築する。
func NewBillingConfirmationHandler(svc BillingConfirmationService, requirePermission PermissionMiddleware) *BillingConfirmationHandler {
	return &BillingConfirmationHandler{svc: svc, requirePermission: requirePermission}
}

// GetBillingConfirmation は指定カルテIDの会計医師確認を取得または初期化して返す
// GET /medical-records/:id/billing-confirmation
func (h *BillingConfirmationHandler) GetBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	review, err := h.svc.GetOrCreate(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ConfirmBillingConfirmation は会計を医師確認済みにする
// POST /medical-records/:id/billing-confirmation/confirm
func (h *BillingConfirmationHandler) ConfirmBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req confirmBillingConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	review, err := h.svc.Confirm(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ReturnBillingConfirmation は会計を差し戻す
// POST /medical-records/:id/billing-confirmation/return
func (h *BillingConfirmationHandler) ReturnBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req returnBillingConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	review, err := h.svc.Return(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}
