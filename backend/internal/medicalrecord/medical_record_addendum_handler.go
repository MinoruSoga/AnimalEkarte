package medicalrecord

import (
	"fmt"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// MedicalRecordAddendumHandler serves the medical-record-addendum HTTP boundary. Moved from internal/handler
// (BE9-2D ⑦ Batch C).
type MedicalRecordAddendumHandler struct {
	service MedicalRecordAddendumService
}

// NewMedicalRecordAddendumHandler initializes a MedicalRecordAddendumHandler.
func NewMedicalRecordAddendumHandler(service MedicalRecordAddendumService) *MedicalRecordAddendumHandler {
	return &MedicalRecordAddendumHandler{service: service}
}

func (h *MedicalRecordAddendumHandler) ListMedicalRecordAddenda(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	addenda, svcErr := h.service.FindByMedicalRecordID(c.Request.Context(), clinicID, medicalRecordID)
	if svcErr != nil {
		httpapi.RespondError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, toMedicalRecordAddendumListResponse(addenda))
}

func (h *MedicalRecordAddendumHandler) CreateMedicalRecordAddendum(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req CreateMedicalRecordAddendumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	addendum, svcErr := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput(medicalRecordID, staffID))
	if svcErr != nil {
		httpapi.RespondError(c, svcErr)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/addenda/%d", medicalRecordID, addendum.ID))
	c.JSON(http.StatusCreated, toMedicalRecordAddendumResponse(addendum))
}
