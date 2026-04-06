package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListTreatments godoc
// GET /medical-records/:id/treatments
func (h *Handler) ListTreatments(c *gin.Context) {
	_, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	treatments, err := h.svc.Treatment.List(c.Request.Context(), medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return
	}

	items := make([]treatmentResponse, 0, len(treatments))
	for i := range treatments {
		items = append(items, toTreatmentResponse(&treatments[i]))
	}
	c.JSON(http.StatusOK, items)
}

// CreateTreatment godoc
// POST /medical-records/:id/treatments
func (h *Handler) CreateTreatment(c *gin.Context) {
	_, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req createTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if err := validateTreatmentItemType(req.ItemType); err != nil {
		RespondError(c, err)
		return
	}

	input := &service.CreateTreatmentInput{
		ItemType:       model.TreatmentItemType(req.ItemType),
		ConsultationID: req.ConsultationID,
		ProcedureID:    req.ProcedureID,
		MedicineID:     req.MedicineID,
		InventoryID:    req.InventoryID,
		UnitPrice:      req.UnitPrice,
		Quantity:       req.Quantity,
		Selected:       req.Selected,
		Status:         req.Status,
		Content:        req.Content,
		Memo:           req.Memo,
		AdminRoute:     req.AdminRoute,
		Insurance:      req.Insurance,
		DiscountRate:   req.DiscountRate,
		DiscountAmount: req.DiscountAmount,
		SortOrder:      req.SortOrder,
	}

	treatment, err := h.svc.Treatment.Create(c.Request.Context(), medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTreatmentResponse(treatment))
}

// UpdateTreatment godoc
// PATCH /medical-records/:id/treatments/:treatmentId
func (h *Handler) UpdateTreatment(c *gin.Context) {
	_, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid treatmentId"))
		return
	}

	var req updateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateTreatmentInput{
		ConsultationID: req.ConsultationID,
		ProcedureID:    req.ProcedureID,
		MedicineID:     req.MedicineID,
		InventoryID:    req.InventoryID,
		UnitPrice:      req.UnitPrice,
		Quantity:       req.Quantity,
		Selected:       req.Selected,
		Status:         req.Status,
		Content:        req.Content,
		Memo:           req.Memo,
		AdminRoute:     req.AdminRoute,
		Insurance:      req.Insurance,
		DiscountRate:   req.DiscountRate,
		DiscountAmount: req.DiscountAmount,
		SortOrder:      req.SortOrder,
	}
	if req.ItemType != nil {
		itemType := model.TreatmentItemType(*req.ItemType)
		input.ItemType = &itemType
	}

	treatment, err := h.svc.Treatment.Update(c.Request.Context(), medicalRecordID, treatmentID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentResponse(treatment))
}

// DeleteTreatment godoc
// DELETE /medical-records/:id/treatments/:treatmentId
func (h *Handler) DeleteTreatment(c *gin.Context) {
	_, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	treatmentID, err := strconv.ParseUint(c.Param("treatmentId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid treatmentId"))
		return
	}

	if err := h.svc.Treatment.Delete(c.Request.Context(), medicalRecordID, treatmentID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkUpdateTreatments godoc
// PUT /medical-records/:id/treatments
func (h *Handler) BulkUpdateTreatments(c *gin.Context) {
	_, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req bulkUpdateTreatmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	items := make([]service.BulkTreatmentItem, 0, len(req.Treatments))
	for _, t := range req.Treatments {
		items = append(items, service.BulkTreatmentItem{
			ID:        t.ID,
			SortOrder: t.SortOrder,
		})
	}
	input := &service.BulkUpdateTreatmentsInput{Treatments: items}

	if err := h.svc.Treatment.BulkUpdateSortOrder(c.Request.Context(), medicalRecordID, input); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterTreatmentRoutes は治療項目関連のルートをカルテサブリソースとして登録する
func (h *Handler) RegisterTreatmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatments", h.ListTreatments)
	rg.POST("/:id/treatments", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateTreatment)
	rg.PATCH("/:id/treatments/:treatmentId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateTreatment)
	rg.DELETE("/:id/treatments/:treatmentId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteTreatment)
	rg.PUT("/:id/treatments", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.BulkUpdateTreatments)
}
