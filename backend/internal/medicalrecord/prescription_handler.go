package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// PrescriptionHandler serves the Prescription (per-medical-record 処方薬記録) HTTP boundary.
type PrescriptionHandler struct {
	service PrescriptionService
}

// NewPrescriptionHandler initializes a PrescriptionHandler.
func NewPrescriptionHandler(service PrescriptionService) *PrescriptionHandler {
	return &PrescriptionHandler{service: service}
}

// ListPrescriptions は指定カルテに紐づく処方薬記録の一覧を返す
func (h *PrescriptionHandler) ListPrescriptions(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMedicalRecords), "view") {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptions, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(prescriptions, toPrescriptionResponse))
}

// CreatePrescription は指定カルテに処方薬記録を作成する
func (h *PrescriptionHandler) CreatePrescription(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createPrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	p, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/prescriptions/%d", medicalRecordID, p.ID))
	c.JSON(http.StatusCreated, toPrescriptionResponse(p))
}

// UpdatePrescription は処方薬記録を部分更新する
func (h *PrescriptionHandler) UpdatePrescription(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptionID, ok := httpapi.ParseIDParam(c, "prescriptionId")
	if !ok {
		return
	}

	var req updatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	p, err := h.service.Update(c.Request.Context(), clinicID, medicalRecordID, prescriptionID, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPrescriptionResponse(p))
}

// DeletePrescription は処方薬記録を soft delete する
func (h *PrescriptionHandler) DeletePrescription(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	prescriptionID, ok := httpapi.ParseIDParam(c, "prescriptionId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, prescriptionID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
