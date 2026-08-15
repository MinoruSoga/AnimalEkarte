package medicalrecord

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ClinicalPlanHandler serves the clinical-plan (診療計画) HTTP boundary nested under a medical
// record. Unlike VitalHandler / MedicalRecordImageHandler it holds no medicalRecordGetter: the
// pre-move internal/handler clinical-plan handler never called verifyMedicalRecordOwnership — it
// delegates the clinic scope entirely to the service (GetOrCreate / Update / Delete all take
// clinicID).
type ClinicalPlanHandler struct {
	service ClinicalPlanService
}

// NewClinicalPlanHandler initializes a ClinicalPlanHandler.
func NewClinicalPlanHandler(service ClinicalPlanService) *ClinicalPlanHandler {
	return &ClinicalPlanHandler{service: service}
}

// GetClinicalPlan godoc
// GET /medical-records/:id/clinical-plan
func (h *ClinicalPlanHandler) GetClinicalPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	plan, err := h.service.GetOrCreate(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// UpdateClinicalPlan godoc
// PATCH /medical-records/:id/clinical-plan
func (h *ClinicalPlanHandler) UpdateClinicalPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var req updateClinicalPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input := req.toServiceInput()
	input.ActorID = &staffID
	plan, err := h.service.Update(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// DeleteClinicalPlan godoc
// DELETE /medical-records/:id/clinical-plan
func (h *ClinicalPlanHandler) DeleteClinicalPlan(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, &staffID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
