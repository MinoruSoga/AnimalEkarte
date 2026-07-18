package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (h *Handler) ListTreatmentPlansByMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, id); !ok {
		return
	}
	plans, err := h.svc.TreatmentPlan.ListByMedicalRecord(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := mapSlice(plans, toTreatmentPlanResponse)
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateTreatmentPlanForMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, id); !ok {
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	// BUG-372: discount フィールドにゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateFloat(c, req.DiscountRate); err != nil {
		RespondError(c, err)
		return
	}
	if err := h.requireDiscountCreateInt(c, req.DiscountAmount); err != nil {
		RespondError(c, err)
		return
	}
	plan, err := h.svc.TreatmentPlan.Create(c.Request.Context(), clinicID, &id, nil, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/treatment-plans/%d", id, plan.ID))
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

func (h *Handler) ListTreatmentPlansByHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	plans, err := h.svc.TreatmentPlan.ListByHospitalization(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := mapSlice(plans, toTreatmentPlanResponse)
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateTreatmentPlanForHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	// BUG-372: discount フィールドにゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateFloat(c, req.DiscountRate); err != nil {
		RespondError(c, err)
		return
	}
	if err := h.requireDiscountCreateInt(c, req.DiscountAmount); err != nil {
		RespondError(c, err)
		return
	}
	plan, err := h.svc.TreatmentPlan.Create(c.Request.Context(), clinicID, nil, &id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/treatment-plans/%d", id, plan.ID))
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

// BUG-372: 既存 TreatmentPlan を取得し discount フィールド変更時の権限チェックを行う。
// 権限 OK なら nil を返す。NG または取得エラーならその error を返す。
func (h *Handler) checkTreatmentPlanDiscountPermission(c *gin.Context, clinicID, planID uint64, req updateTreatmentPlanRequest) error {
	if req.DiscountRate == nil && req.DiscountAmount == nil {
		return nil
	}
	existing, err := h.svc.TreatmentPlan.GetByID(c.Request.Context(), clinicID, planID)
	if err != nil {
		return err
	}
	if err := h.requireDiscountEditFloat(c, req.DiscountRate, existing.DiscountRate); err != nil {
		return err
	}
	return h.requireDiscountEditInt(c, req.DiscountAmount, existing.DiscountAmount)
}

// UpdateTreatmentPlanInMedicalRecord は MedicalRecord コンテキストでのプラン更新
// PATCH /medical-records/:id/treatment-plans/:planId
func (h *Handler) UpdateTreatmentPlanInMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	mrID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, mrID); !ok {
		return
	}
	planID, ok := parseIDParam(c, "planId")
	if !ok {
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.checkTreatmentPlanDiscountPermission(c, clinicID, planID, req); err != nil {
		RespondError(c, err)
		return
	}
	plan, err := h.svc.TreatmentPlan.Update(c.Request.Context(), clinicID, planID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))
}

// DeleteTreatmentPlanInMedicalRecord は MedicalRecord コンテキストでのプラン削除
// DELETE /medical-records/:id/treatment-plans/:planId
func (h *Handler) DeleteTreatmentPlanInMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	mrID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.verifyMedicalRecordOwnership(c, clinicID, mrID); !ok {
		return
	}
	planID, ok := parseIDParam(c, "planId")
	if !ok {
		return
	}
	if err := h.svc.TreatmentPlan.Delete(c.Request.Context(), clinicID, planID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateTreatmentPlanInHospitalization は Hospitalization コンテキストでのプラン更新
// PATCH /hospitalizations/:id/treatment-plans/:planId
func (h *Handler) UpdateTreatmentPlanInHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		RespondError(c, err)
		return
	}
	planID, ok := parseIDParam(c, "planId")
	if !ok {
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.checkTreatmentPlanDiscountPermission(c, clinicID, planID, req); err != nil {
		RespondError(c, err)
		return
	}
	plan, err := h.svc.TreatmentPlan.Update(c.Request.Context(), clinicID, planID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))
}

// DeleteTreatmentPlanInHospitalization は Hospitalization コンテキストでのプラン削除
// DELETE /hospitalizations/:id/treatment-plans/:planId
func (h *Handler) DeleteTreatmentPlanInHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		RespondError(c, err)
		return
	}
	planID, ok := parseIDParam(c, "planId")
	if !ok {
		return
	}
	if err := h.svc.TreatmentPlan.Delete(c.Request.Context(), clinicID, planID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RegisterTreatmentPlanMedicalRecordRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatment-plans", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListTreatmentPlansByMedicalRecord)
	rg.POST("/:id/treatment-plans", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateTreatmentPlanForMedicalRecord)
	rg.PATCH("/:id/treatment-plans/:planId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateTreatmentPlanInMedicalRecord)
	rg.DELETE("/:id/treatment-plans/:planId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteTreatmentPlanInMedicalRecord)
}

func (h *Handler) RegisterTreatmentPlanHospitalizationRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatment-plans", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.ListTreatmentPlansByHospitalization)
	rg.POST("/:id/treatment-plans", h.RequirePermission(string(model.ResourceHospitalization), "create"), h.CreateTreatmentPlanForHospitalization)
	rg.PATCH("/:id/treatment-plans/:planId", h.RequirePermission(string(model.ResourceHospitalization), "edit"), h.UpdateTreatmentPlanInHospitalization)
	rg.DELETE("/:id/treatment-plans/:planId", h.RequirePermission(string(model.ResourceHospitalization), "delete"), h.DeleteTreatmentPlanInHospitalization)
}
