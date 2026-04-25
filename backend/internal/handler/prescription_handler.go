package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
		return
	}

	p, err := h.svc.Prescription.Create(c.Request.Context(), clinicID, medicalRecordID, &service.CreatePrescriptionInput{
		PrescribedAt: date,
		DurationDays: req.DurationDays,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-009: 処方タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncPrescriptionTag(c.Request.Context(), clinicID, p.OwnerID)
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

	var updateDate *time.Time
	if req.Date != nil && *req.Date != "" {
		d, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format"))
			return
		}
		updateDate = &d
	}

	p, err := h.svc.Prescription.Update(c.Request.Context(), clinicID, medicalRecordID, prescriptionID, &service.UpdatePrescriptionInput{
		PrescribedAt: updateDate,
		DurationDays: req.DurationDays,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-009: 処方タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncPrescriptionTag(c.Request.Context(), clinicID, p.OwnerID)
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

	// タグ同期用に削除前に ownerID を取得する
	existing, err := h.svc.Prescription.GetByID(c.Request.Context(), clinicID, prescriptionID)
	if err != nil {
		RespondError(c, err)
		return
	}
	ownerID := existing.OwnerID

	if err := h.svc.Prescription.Delete(c.Request.Context(), clinicID, medicalRecordID, prescriptionID); err != nil {
		RespondError(c, err)
		return
	}
	// BE-009: 処方タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncPrescriptionTag(c.Request.Context(), clinicID, ownerID)
	c.Status(http.StatusNoContent)
}

// RegisterPrescriptionRoutes は処方薬記録関連のルートを登録する
func (h *Handler) RegisterPrescriptionRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/prescriptions", h.ListPrescriptions)
	rg.POST("/:id/prescriptions", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreatePrescription)
	rg.PATCH("/:id/prescriptions/:prescriptionId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdatePrescription)
	rg.DELETE("/:id/prescriptions/:prescriptionId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeletePrescription)
}
