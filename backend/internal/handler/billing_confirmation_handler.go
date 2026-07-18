package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetBillingConfirmation は指定カルテIDの会計医師確認を取得または初期化して返す
// GET /medical-records/:id/billing-confirmation
func (h *Handler) GetBillingConfirmation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	review, err := h.svc.BillingConfirmation.GetOrCreate(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ConfirmBillingConfirmation は会計を医師確認済みにする
// POST /medical-records/:id/billing-confirmation/confirm
func (h *Handler) ConfirmBillingConfirmation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req confirmBillingConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	review, err := h.svc.BillingConfirmation.Confirm(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ReturnBillingConfirmation は会計を差し戻す
// POST /medical-records/:id/billing-confirmation/return
func (h *Handler) ReturnBillingConfirmation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req returnBillingConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	review, err := h.svc.BillingConfirmation.Return(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// RegisterBillingConfirmationRoutes は会計医師確認関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterBillingConfirmationRoutes(rg *gin.RouterGroup) {
	permEdit := h.RequirePermission(string(model.ResourceAccounting), "edit")
	rg.GET("/:id/billing-confirmation", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetBillingConfirmation)
	rg.POST("/:id/billing-confirmation/confirm", permEdit, h.ConfirmBillingConfirmation)
	rg.POST("/:id/billing-confirmation/return", permEdit, h.ReturnBillingConfirmation)
}
