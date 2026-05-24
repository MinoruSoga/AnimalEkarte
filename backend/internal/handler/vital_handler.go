package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListVitals は指定カルテIDのバイタル一覧を返す
// GET /medical-records/:id/vitals
func (h *Handler) ListVitals(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitals, err := h.svc.Vital.List(c.Request.Context(), clinicID, medicalRecordID)
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
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	mr, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID)
	if !ok {
		return
	}

	var req createVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var petID uint64
	if mr.PetID != nil {
		petID = *mr.PetID
	}
	input := &service.CreateVitalInput{
		ClinicID:        clinicID,
		PetID:           petID,
		RecordedAt:      req.RecordedAt,
		StaffID:         req.StaffID,
		Temperature:     req.Temperature,
		HeartRate:       req.HeartRate,
		RespirationRate: req.RespirationRate,
		Weight:          req.Weight,
		WeightUnit:      toBodyWeightUnit(req.WeightUnit),
		Notes:           req.Notes,
	}

	vital, err := h.svc.Vital.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/vitals/%d", medicalRecordID, vital.ID))
	c.JSON(http.StatusCreated, toVitalResponse(vital))
}

// UpdateVital は指定バイタルを部分更新する
// PATCH /medical-records/:id/vitals/:vitalId
func (h *Handler) UpdateVital(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitalID, ok := parseIDParam(c, "vitalId")
	if !ok {
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	var req updateVitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateVitalInput{
		RecordedAt:      req.RecordedAt,
		StaffID:         req.StaffID,
		Temperature:     req.Temperature,
		HeartRate:       req.HeartRate,
		RespirationRate: req.RespirationRate,
		Weight:          req.Weight,
		WeightUnit:      toBodyWeightUnit(req.WeightUnit),
		Notes:           req.Notes,
		ActorID:         &staffID,
	}

	vital, err := h.svc.Vital.Update(c.Request.Context(), clinicID, medicalRecordID, vitalID, input)
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
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID); !ok {
		return
	}

	vitalID, ok := parseIDParam(c, "vitalId")
	if !ok {
		return
	}

	if err := h.svc.Vital.Delete(c.Request.Context(), clinicID, medicalRecordID, vitalID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// toBodyWeightUnit は文字列ポインタを *model.BodyWeightUnit に変換するヘルパー。
// nil の場合は nil を返し、サービス層でデフォルト値（Kg）が適用される。
func toBodyWeightUnit(s *string) *model.BodyWeightUnit {
	if s == nil {
		return nil
	}
	u := model.BodyWeightUnit(*s)
	return &u
}

// RegisterVitalRoutes はバイタル関連のルートをmedical-recordsグループに登録する
func (h *Handler) RegisterVitalRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/vitals", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListVitals)
	rg.POST("/:id/vitals", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.CreateVital)
	rg.PATCH("/:id/vitals/:vitalId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateVital)
	rg.DELETE("/:id/vitals/:vitalId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteVital)
}
