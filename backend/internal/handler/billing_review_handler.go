package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetBillingReview は指定カルテIDの会計医師確認を取得または初期化して返す
// GET /medical-records/:id/billing-review
func (h *Handler) GetBillingReview(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}

	review, err := h.svc.BillingReview.GetOrCreate(c.Request.Context(), medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingReviewResponse(review))
}

// ConfirmBillingReview は会計を医師確認済みにする
// POST /medical-records/:id/billing-review/confirm
func (h *Handler) ConfirmBillingReview(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}

	var req confirmBillingReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.ConfirmBillingReviewInput{
		ConfirmedBy: req.ConfirmedBy,
		Memo:        req.Memo,
	}

	review, err := h.svc.BillingReview.Confirm(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingReviewResponse(review))
}

// ReturnBillingReview は会計を差し戻す
// POST /medical-records/:id/billing-review/return
func (h *Handler) ReturnBillingReview(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}

	var req returnBillingReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.ReturnBillingReviewInput{
		ReturnedBy:   req.ReturnedBy,
		ReturnReason: req.ReturnReason,
		Memo:         req.Memo,
	}

	review, err := h.svc.BillingReview.Return(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingReviewResponse(review))
}

// RegisterBillingReviewRoutes は会計医師確認関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterBillingReviewRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/billing-review", h.GetBillingReview)
	rg.POST("/:id/billing-review/confirm", h.ConfirmBillingReview)
	rg.POST("/:id/billing-review/return", h.ReturnBillingReview)
}
