package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

func (h *Handler) ListTreatmentPlansByMedicalRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req createTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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

func (h *Handler) UpdateTreatmentPlan(c *gin.Context) {
	planID, err := strconv.ParseUint(c.Param("planId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid planId"))
		return
	}
	var req updateTreatmentPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	input := &service.UpdateTreatmentPlanInput{
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
	plan, err := h.svc.TreatmentPlan.Update(c.Request.Context(), planID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))
}

func (h *Handler) DeleteTreatmentPlan(c *gin.Context) {
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
	rg.PATCH("/:id/treatment-plans/:planId", h.UpdateTreatmentPlan)
	rg.DELETE("/:id/treatment-plans/:planId", h.DeleteTreatmentPlan)
}

func (h *Handler) RegisterTreatmentPlanHospitalizationRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatment-plans", h.ListTreatmentPlansByHospitalization)
	rg.POST("/:id/treatment-plans", h.CreateTreatmentPlanForHospitalization)
}
