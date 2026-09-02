package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TreatmentPlanHandler serves the treatment-plan HTTP boundary (medical-record 配下 +
// hospitalization 配下の両サブリソース). Moved from internal/handler (BE9-2D ⑥ Batch C) —
// これにより ⑤ の Services.Hospitalization 残置 field と handler 側 carrier が解消。
type TreatmentPlanHandler struct {
	service         TreatmentPlanService
	hospitalization HospitalizationService
	medicalRecord   medicalRecordGetter
	hasPermission   PermissionChecker
}

// NewTreatmentPlanHandler initializes a TreatmentPlanHandler.
func NewTreatmentPlanHandler(service TreatmentPlanService, hospitalization HospitalizationService, medicalRecord medicalRecordGetter, hasPermission PermissionChecker) *TreatmentPlanHandler {
	return &TreatmentPlanHandler{service: service, hospitalization: hospitalization, medicalRecord: medicalRecord, hasPermission: hasPermission}
}

// verifyMedicalRecordOwnership は internal/handler/medical_record_ownership.go の同名の
// ローカル移植（vital handler の verifyOwnership 先例・medicalRecordGetter view 経由）。
func (h *TreatmentPlanHandler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) bool {
	_, err := h.medicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return false
	}
	return true
}

func (h *TreatmentPlanHandler) ListTreatmentPlansByMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, id) {
		return
	}
	plans, err := h.service.ListByMedicalRecord(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	items := httpapi.MapSlice(plans, toTreatmentPlanResponse)
	c.JSON(http.StatusOK, items)
}

func (h *TreatmentPlanHandler) CreateTreatmentPlanForMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, id) {
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	// BUG-372: discount フィールドにゼロ以外を指定する場合は権限要
	if err := requireDiscountCreateFloat(c, h.hasPermission, req.DiscountRate); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := requireDiscountCreateInt(c, h.hasPermission, req.DiscountAmount); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	plan, err := h.service.Create(c.Request.Context(), clinicID, &id, nil, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/treatment-plans/%d", id, plan.ID))
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

func (h *TreatmentPlanHandler) ListTreatmentPlansByHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.hospitalization.GetByID(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	plans, err := h.service.ListByHospitalization(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	items := httpapi.MapSlice(plans, toTreatmentPlanResponse)
	c.JSON(http.StatusOK, items)
}

func (h *TreatmentPlanHandler) CreateTreatmentPlanForHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.hospitalization.GetByID(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	// BUG-372: discount フィールドにゼロ以外を指定する場合は権限要
	if err := requireDiscountCreateFloat(c, h.hasPermission, req.DiscountRate); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := requireDiscountCreateInt(c, h.hasPermission, req.DiscountAmount); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	plan, err := h.service.Create(c.Request.Context(), clinicID, nil, &id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/treatment-plans/%d", id, plan.ID))
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

// BUG-372: 既存 TreatmentPlan を取得し discount フィールド変更時の権限チェックを行う。
// 権限 OK なら nil を返す。NG または取得エラーならその error を返す。
func (h *TreatmentPlanHandler) checkTreatmentPlanDiscountPermission(c *gin.Context, clinicID, planID uint64, req updateTreatmentPlanRequest) error {
	if req.DiscountRate == nil && req.DiscountAmount == nil {
		return nil
	}
	existing, err := h.service.GetByID(c.Request.Context(), clinicID, planID)
	if err != nil {
		return err
	}
	if err := requireDiscountEditFloat(c, h.hasPermission, req.DiscountRate, existing.DiscountRate); err != nil {
		return err
	}
	return requireDiscountEditInt(c, h.hasPermission, req.DiscountAmount, existing.DiscountAmount)
}

// UpdateTreatmentPlanInMedicalRecord は MedicalRecord コンテキストでのプラン更新
// PATCH /medical-records/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) UpdateTreatmentPlanInMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	mrID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, mrID) {
		return
	}
	planID, ok := httpapi.ParseIDParam(c, "planId")
	if !ok {
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.checkTreatmentPlanDiscountPermission(c, clinicID, planID, req); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	mrIDCopy := mrID
	input := req.toServiceInput()
	// SEC-CS-F10: pass discount:edit into the write TX for locked-row recheck.
	if h.hasPermission != nil {
		input.DiscountEditAllowed = h.hasPermission(c, string(model.ResourceDiscount), "edit")
	}
	plan, err := h.service.Update(c.Request.Context(), clinicID, planID, &mrIDCopy, nil, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))
}

// DeleteTreatmentPlanInMedicalRecord は MedicalRecord コンテキストでのプラン削除
// DELETE /medical-records/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) DeleteTreatmentPlanInMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	mrID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, mrID) {
		return
	}
	planID, ok := httpapi.ParseIDParam(c, "planId")
	if !ok {
		return
	}
	mrIDCopy := mrID
	if err := h.service.Delete(c.Request.Context(), clinicID, planID, &mrIDCopy, nil); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateTreatmentPlanInHospitalization は Hospitalization コンテキストでのプラン更新
// PATCH /hospitalizations/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) UpdateTreatmentPlanInHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	planID, ok := httpapi.ParseIDParam(c, "planId")
	if !ok {
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.checkTreatmentPlanDiscountPermission(c, clinicID, planID, req); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	hospIDCopy := hospID
	input := req.toServiceInput()
	// SEC-CS-F10: pass discount:edit into the write TX for locked-row recheck.
	if h.hasPermission != nil {
		input.DiscountEditAllowed = h.hasPermission(c, string(model.ResourceDiscount), "edit")
	}
	plan, err := h.service.Update(c.Request.Context(), clinicID, planID, nil, &hospIDCopy, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))
}

// DeleteTreatmentPlanInHospitalization は Hospitalization コンテキストでのプラン削除
// DELETE /hospitalizations/:id/treatment-plans/:planId
func (h *TreatmentPlanHandler) DeleteTreatmentPlanInHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	planID, ok := httpapi.ParseIDParam(c, "planId")
	if !ok {
		return
	}
	hospIDCopy := hospID
	if err := h.service.Delete(c.Request.Context(), clinicID, planID, nil, &hospIDCopy); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
