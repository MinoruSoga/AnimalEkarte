package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListPrescriptions は指定カルテに紐づく処方薬記録の一覧を返す
func (h *Handler) ListPrescriptions(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptions, err := h.svc.Prescription.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(prescriptions, toPrescriptionResponse))
}

// CreatePrescription は指定カルテに処方薬記録を作成する
func (h *Handler) CreatePrescription(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createPrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}

	p, err := h.svc.Prescription.Create(c.Request.Context(), clinicID, medicalRecordID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/prescriptions/%d", medicalRecordID, p.ID))
	c.JSON(http.StatusCreated, toPrescriptionResponse(p))
}

// UpdatePrescription は処方薬記録を部分更新する
func (h *Handler) UpdatePrescription(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptionID, ok := parseIDParam(c, "prescriptionId")
	if !ok {
		return
	}

	var req updatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}

	p, err := h.svc.Prescription.Update(c.Request.Context(), clinicID, medicalRecordID, prescriptionID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPrescriptionResponse(p))
}

// DeletePrescription は処方薬記録を soft delete する
func (h *Handler) DeletePrescription(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptionID, ok := parseIDParam(c, "prescriptionId")
	if !ok {
		return
	}

	if err := h.svc.Prescription.Delete(c.Request.Context(), clinicID, medicalRecordID, prescriptionID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterPrescriptionRoutes は処方薬記録関連のルートを登録する
func (h *Handler) RegisterPrescriptionRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/prescriptions", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListPrescriptions)
	rg.POST("/:id/prescriptions", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreatePrescription)
	rg.PATCH("/:id/prescriptions/:prescriptionId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdatePrescription)
	rg.DELETE("/:id/prescriptions/:prescriptionId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeletePrescription)
}
