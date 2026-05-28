package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpdateInquiry godoc
// PATCH /medical-records/:id/inquiries
func (h *Handler) UpdateInquiry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	inquiry, err := h.svc.Inquiry.Save(c.Request.Context(), req.toServiceInput(clinicID, medicalRecordID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryResponse(inquiry))
}

// RegisterInquiryRoutes は問診関連のルートを登録する
func (h *Handler) RegisterInquiryRoutes(rg *gin.RouterGroup) {
	rg.PATCH("/:id/inquiries", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateInquiry)
}
