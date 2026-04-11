package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetClinicalPlan godoc
// GET /medical-records/:id/clinical-plan
func (h *Handler) GetClinicalPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	plan, err := h.svc.ClinicalPlan.GetOrCreate(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// UpdateClinicalPlan godoc
// PATCH /medical-records/:id/clinical-plan
func (h *Handler) UpdateClinicalPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateClinicalPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := &service.UpdateClinicalPlanInput{
		PhysicalExam:         req.PhysicalExam,
		DiagnosisTypeID:  req.DiagnosisTypeID,
		DiagnosisNameID:      req.DiagnosisNameID,
		DiagnosisDetails:     req.DiagnosisDetails,
		TreatmentPolicy:      req.TreatmentPolicy,
		Diagnosis2CategoryID: req.Diagnosis2CategoryID,
		Diagnosis2NameID:     req.Diagnosis2NameID,
	}
	plan, err := h.svc.ClinicalPlan.Update(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// DeleteClinicalPlan godoc
// DELETE /medical-records/:id/clinical-plan
func (h *Handler) DeleteClinicalPlan(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.ClinicalPlan.Delete(c.Request.Context(), clinicID, medicalRecordID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicalPlanRoutes はClinicalPlan関連のルートを登録する
func (h *Handler) RegisterClinicalPlanRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/clinical-plan", h.GetClinicalPlan)
	rg.PATCH("/:id/clinical-plan", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateClinicalPlan)
	rg.DELETE("/:id/clinical-plan", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteClinicalPlan)
}
