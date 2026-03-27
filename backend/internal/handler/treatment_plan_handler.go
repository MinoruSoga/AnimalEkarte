package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

func (h *Handler) ListTreatmentPlansByMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, id) {
		return
	}
	plans, err := h.svc.TreatmentPlan.ListByMedicalRecord(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := make([]treatmentPlanResponse, 0, len(plans))
	for i := range plans {
		items = append(items, toTreatmentPlanResponse(&plans[i]))
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateTreatmentPlanForMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, id) {
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := &service.CreateTreatmentPlanInput{
		TreatmentContent: req.TreatmentContent,
		Memo:             req.Memo,
		Insurance:        req.Insurance,
		UnitPrice:        req.UnitPrice,
		Quantity:         req.Quantity,
		DiscountRate:     req.DiscountRate,
		DiscountAmount:   req.DiscountAmount,
		Subtotal:         req.Subtotal,
		SortOrder:        req.SortOrder,
	}
	plan, err := h.svc.TreatmentPlan.Create(c.Request.Context(), &id, nil, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

func (h *Handler) ListTreatmentPlansByHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	plans, err := h.svc.TreatmentPlan.ListByHospitalization(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := make([]treatmentPlanResponse, 0, len(plans))
	for i := range plans {
		items = append(items, toTreatmentPlanResponse(&plans[i]))
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateTreatmentPlanForHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	input := &service.CreateTreatmentPlanInput{
		TreatmentContent: req.TreatmentContent,
		Memo:             req.Memo,
		Insurance:        req.Insurance,
		UnitPrice:        req.UnitPrice,
		Quantity:         req.Quantity,
		DiscountRate:     req.DiscountRate,
		DiscountAmount:   req.DiscountAmount,
		Subtotal:         req.Subtotal,
		SortOrder:        req.SortOrder,
	}
	plan, err := h.svc.TreatmentPlan.Create(c.Request.Context(), nil, &id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTreatmentPlanResponse(plan))
}

// buildUpdateTreatmentPlanInput は共通の更新入力を組み立てる
func buildUpdateTreatmentPlanInput(req updateTreatmentPlanRequest) *service.UpdateTreatmentPlanInput {
	return &service.UpdateTreatmentPlanInput{
		TreatmentContent: req.TreatmentContent,
		Memo:             req.Memo,
		Insurance:        req.Insurance,
		UnitPrice:        req.UnitPrice,
		Quantity:         req.Quantity,
		DiscountRate:     req.DiscountRate,
		DiscountAmount:   req.DiscountAmount,
		Subtotal:         req.Subtotal,
		SortOrder:        req.SortOrder,
	}
}

// UpdateTreatmentPlanInMedicalRecord は MedicalRecord コンテキストでのプラン更新
// PATCH /medical-records/:id/treatment-plans/:planId
func (h *Handler) UpdateTreatmentPlanInMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	mrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, mrID) {
		return
	}
	planID, err := strconv.ParseUint(c.Param("planId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid planId"))
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	plan, err := h.svc.TreatmentPlan.Update(c.Request.Context(), planID, buildUpdateTreatmentPlanInput(req))
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
	mrID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if !h.verifyMedicalRecordOwnership(c, clinicID, mrID) {
		return
	}
	planID, err := strconv.ParseUint(c.Param("planId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid planId"))
		return
	}
	if err := h.svc.TreatmentPlan.Delete(c.Request.Context(), planID); err != nil {
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
	hospID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		RespondError(c, err)
		return
	}
	planID, err := strconv.ParseUint(c.Param("planId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid planId"))
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	plan, err := h.svc.TreatmentPlan.Update(c.Request.Context(), planID, buildUpdateTreatmentPlanInput(req))
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
	hospID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if _, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, hospID); err != nil {
		RespondError(c, err)
		return
	}
	planID, err := strconv.ParseUint(c.Param("planId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid planId"))
		return
	}
	if err := h.svc.TreatmentPlan.Delete(c.Request.Context(), planID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RegisterTreatmentPlanMedicalRecordRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatment-plans", h.ListTreatmentPlansByMedicalRecord)
	rg.POST("/:id/treatment-plans", h.CreateTreatmentPlanForMedicalRecord)
	rg.PATCH("/:id/treatment-plans/:planId", h.UpdateTreatmentPlanInMedicalRecord)
	rg.DELETE("/:id/treatment-plans/:planId", h.DeleteTreatmentPlanInMedicalRecord)
}

func (h *Handler) RegisterTreatmentPlanHospitalizationRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatment-plans", h.ListTreatmentPlansByHospitalization)
	rg.POST("/:id/treatment-plans", h.CreateTreatmentPlanForHospitalization)
	rg.PATCH("/:id/treatment-plans/:planId", h.UpdateTreatmentPlanInHospitalization)
	rg.DELETE("/:id/treatment-plans/:planId", h.DeleteTreatmentPlanInHospitalization)
}
