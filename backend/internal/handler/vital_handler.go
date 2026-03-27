package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListVitals は指定カルテIDのバイタル一覧を返す
// GET /medical-records/:id/vitals
func (h *Handler) ListVitals(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	vitals, err := h.svc.Vital.List(c.Request.Context(), medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}

	items := make([]vitalResponse, 0, len(vitals))
	for i := range vitals {
		items = append(items, toVitalResponse(&vitals[i]))
	}
	c.JSON(http.StatusOK, items)
}

// CreateVital は指定カルテIDにバイタルを追加する
// POST /medical-records/:id/vitals
func (h *Handler) CreateVital(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	var req createVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.CreateVitalInput{
		RecordedAt:      req.RecordedAt,
		StaffID:         req.StaffID,
		Temperature:     req.Temperature,
		HeartRate:       req.HeartRate,
		RespirationRate: req.RespirationRate,
		Weight:          req.Weight,
		Notes:           req.Notes,
	}

	vital, err := h.svc.Vital.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toVitalResponse(vital))
}

// UpdateVital は指定バイタルを部分更新する
// PATCH /medical-records/:id/vitals/:vitalId
func (h *Handler) UpdateVital(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	vitalID, err := strconv.ParseUint(c.Param("vitalId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid vital id"))
		return
	}

	var req updateVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.UpdateVitalInput{
		RecordedAt:      req.RecordedAt,
		StaffID:         req.StaffID,
		Temperature:     req.Temperature,
		HeartRate:       req.HeartRate,
		RespirationRate: req.RespirationRate,
		Weight:          req.Weight,
		Notes:           req.Notes,
	}

	vital, err := h.svc.Vital.Update(c.Request.Context(), medicalRecordID, vitalID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVitalResponse(vital))
}

// DeleteVital は指定バイタルを削除する
// DELETE /medical-records/:id/vitals/:vitalId
func (h *Handler) DeleteVital(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID) {
		return
	}

	vitalID, err := strconv.ParseUint(c.Param("vitalId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid vital id"))
		return
	}

	if err := h.svc.Vital.Delete(c.Request.Context(), medicalRecordID, vitalID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterVitalRoutes はバイタル関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterVitalRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/vitals", h.ListVitals)
	rg.POST("/:id/vitals", h.CreateVital)
	rg.PATCH("/:id/vitals/:vitalId", h.UpdateVital)
	rg.DELETE("/:id/vitals/:vitalId", h.DeleteVital)
}
