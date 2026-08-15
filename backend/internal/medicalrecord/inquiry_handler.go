package medicalrecord

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// InquiryHandler serves the Inquiry (per-medical-record 問診) HTTP boundary.
type InquiryHandler struct {
	service InquiryService
}

// NewInquiryHandler initializes an InquiryHandler.
func NewInquiryHandler(service InquiryService) *InquiryHandler {
	return &InquiryHandler{service: service}
}

// UpdateInquiry godoc
// PATCH /medical-records/:id/inquiries
func (h *InquiryHandler) UpdateInquiry(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	inquiry, err := h.service.Save(c.Request.Context(), req.toServiceInput(clinicID, medicalRecordID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInquiryResponse(inquiry))
}
