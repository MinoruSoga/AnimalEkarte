package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// VitalHandler serves the vital-record (バイタル) HTTP boundary nested under a medical record. It
// holds the VitalService plus a medicalRecordGetter for the tenant-ownership guard the pre-move
// internal/handler vital handler performed via verifyMedicalRecordOwnership.
type VitalHandler struct {
	service       VitalService
	medicalRecord medicalRecordGetter
}

// NewVitalHandler initializes a VitalHandler.
func NewVitalHandler(service VitalService, medicalRecord medicalRecordGetter) *VitalHandler {
	return &VitalHandler{service: service, medicalRecord: medicalRecord}
}

// verifyOwnership は clinicID + medicalRecordID を検証しテナント分離を保証する（pre-move
// internal/handler.Handler.verifyMedicalRecordOwnership のローカル移植）。検証済みの
// MedicalRecord と成否を返す。
func (h *VitalHandler) verifyOwnership(c *gin.Context, clinicID, medicalRecordID uint64) (*model.MedicalRecord, bool) {
	mr, err := h.medicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return nil, false
	}
	return mr, true
}

// ListVitals は指定カルテIDのバイタル一覧を返す
// GET /medical-records/:id/vitals
func (h *VitalHandler) ListVitals(c *gin.Context) {
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
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitals, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items := httpapi.MapSlice(vitals, toVitalResponse)
	c.JSON(http.StatusOK, items)
}

// CreateVital は指定カルテIDにバイタルを追加する
// POST /medical-records/:id/vitals
func (h *VitalHandler) CreateVital(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	mr, ok := h.verifyOwnership(c, clinicID, medicalRecordID)
	if !ok {
		return
	}

	var req createVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := req.validate(); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var petID uint64
	if mr.PetID != nil {
		petID = *mr.PetID
	}
	vital, err := h.service.Create(c.Request.Context(), medicalRecordID, req.toServiceInput(clinicID, petID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/vitals/%d", medicalRecordID, vital.ID))
	c.JSON(http.StatusCreated, toVitalResponse(vital))
}

// UpdateVital は指定バイタルを部分更新する
// PATCH /medical-records/:id/vitals/:vitalId
func (h *VitalHandler) UpdateVital(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitalID, ok := httpapi.ParseIDParam(c, "vitalId")
	if !ok {
		return
	}

	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req updateVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := req.validate(); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	vital, err := h.service.Update(c.Request.Context(), clinicID, medicalRecordID, vitalID, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVitalResponse(vital))
}

// DeleteVital は指定バイタルを削除する
// DELETE /medical-records/:id/vitals/:vitalId
func (h *VitalHandler) DeleteVital(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitalID, ok := httpapi.ParseIDParam(c, "vitalId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, vitalID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
