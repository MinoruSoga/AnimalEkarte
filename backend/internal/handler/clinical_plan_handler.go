package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetClinicalPlan godoc
// GET /medical-records/:id/clinical-plan
func (h *Handler) GetClinicalPlan(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	plan, err := h.svc.ClinicalPlan.GetOrCreate(c.Request.Context(), medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// UpdateClinicalPlan godoc
// PATCH /medical-records/:id/clinical-plan
func (h *Handler) UpdateClinicalPlan(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateClinicalPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := &service.UpdateClinicalPlanInput{
		PhysicalExam:         req.PhysicalExam,
		DiagnosisCategoryID:  req.DiagnosisCategoryID,
		DiagnosisNameID:      req.DiagnosisNameID,
		DiagnosisDetails:     req.DiagnosisDetails,
		TreatmentPolicy:      req.TreatmentPolicy,
		Diagnosis2CategoryID: req.Diagnosis2CategoryID,
		Diagnosis2NameID:     req.Diagnosis2NameID,
	}
	plan, err := h.svc.ClinicalPlan.Update(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicalPlanResponse(plan))
}

// DeleteClinicalPlan godoc
// DELETE /medical-records/:id/clinical-plan
func (h *Handler) DeleteClinicalPlan(c *gin.Context) {
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.ClinicalPlan.Delete(c.Request.Context(), medicalRecordID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicalPlanRoutes はClinicalPlan関連のルートを登録する
func (h *Handler) RegisterClinicalPlanRoutes(rg *gin.RouterGroup) {
	perm := h.RequirePermission(string(model.ResourceMedicalRecords), "edit")
	rg.GET("/:id/clinical-plan", h.GetClinicalPlan)
	rg.PATCH("/:id/clinical-plan", perm, h.UpdateClinicalPlan)
	rg.DELETE("/:id/clinical-plan", perm, h.DeleteClinicalPlan)
}
