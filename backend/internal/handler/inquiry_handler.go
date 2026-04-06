package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// UpdateInquiry godoc
// PATCH /medical-records/:id/inquiries
func (h *Handler) UpdateInquiry(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	inquiry, err := h.svc.Inquiry.Upsert(c.Request.Context(), service.UpsertInquiryInput{
		MedicalRecordID:          medicalRecordID,
		ChiefComplaintCategoryID: req.ChiefComplaintCategoryID,
		ChiefComplaint:           req.ChiefComplaint,
		Notes:                    req.Notes,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, inquiry)
}

// RegisterInquiryRoutes は問診関連のルートを登録する
func (h *Handler) RegisterInquiryRoutes(rg *gin.RouterGroup) {
	rg.PATCH("/:id/inquiries", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateInquiry)
}
