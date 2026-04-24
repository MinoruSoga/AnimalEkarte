package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListTreatments godoc
// GET /medical-records/:id/treatments
func (h *Handler) ListTreatments(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	treatments, err := h.svc.Treatment.List(c.Request.Context(), clinicID, medicalRecordID)
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createTreatmentRequest
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

	input := &service.CreateTreatmentInput{
		ItemType:       model.TreatmentItemType(req.ItemType),
		ConsultationID: req.ConsultationID,
		ProcedureID:    req.ProcedureID,
		MedicineID:     req.MedicineID,
		InventoryID:    req.InventoryID,
		UnitPrice:      req.UnitPrice,
		Quantity:       req.Quantity,
		IsSelected:     req.IsSelected,
		Status:         req.Status,
		Content:        req.Content,
		Memo:           req.Memo,
		AdminRoute:     req.AdminRoute,
		IsInsurance:    req.IsInsurance,
		DiscountRate:   req.DiscountRate,
		DiscountAmount: req.DiscountAmount,
		SortOrder:      req.SortOrder,
	}

	treatment, err := h.svc.Treatment.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/treatments/%d", medicalRecordID, treatment.ID))
	c.JSON(http.StatusCreated, toTreatmentResponse(treatment))
}

// UpdateTreatment godoc
// PATCH /medical-records/:id/treatments/:treatmentId
func (h *Handler) UpdateTreatment(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	treatmentID, ok := parseIDParam(c, "treatmentId")
	if !ok {
		return
	}

	var req updateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-372: discount フィールドを変更する場合は既存値と比較し権限チェック
	if req.DiscountRate != nil || req.DiscountAmount != nil {
		existing, err := h.svc.Treatment.GetByID(c.Request.Context(), clinicID, treatmentID)
		if err != nil {
			RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditFloat(c, req.DiscountRate, existing.DiscountRate); err != nil {
			RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditInt(c, req.DiscountAmount, existing.DiscountAmount); err != nil {
			RespondError(c, err)
			return
		}
	}

	input := &service.UpdateTreatmentInput{
		ConsultationID: req.ConsultationID,
		ProcedureID:    req.ProcedureID,
		MedicineID:     req.MedicineID,
		InventoryID:    req.InventoryID,
		UnitPrice:      req.UnitPrice,
		Quantity:       req.Quantity,
		IsSelected:     req.IsSelected,
		Status:         req.Status,
		Content:        req.Content,
		Memo:           req.Memo,
		AdminRoute:     req.AdminRoute,
		IsInsurance:    req.IsInsurance,
		DiscountRate:   req.DiscountRate,
		DiscountAmount: req.DiscountAmount,
		SortOrder:      req.SortOrder,
	}
	if req.ItemType != nil {
		itemType := model.TreatmentItemType(*req.ItemType)
		input.ItemType = &itemType
	}

	treatment, err := h.svc.Treatment.Update(c.Request.Context(), clinicID, medicalRecordID, treatmentID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentResponse(treatment))
}

// DeleteTreatment godoc
// DELETE /medical-records/:id/treatments/:treatmentId
func (h *Handler) DeleteTreatment(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	treatmentID, ok := parseIDParam(c, "treatmentId")
	if !ok {
		return
	}

	if err := h.svc.Treatment.Delete(c.Request.Context(), clinicID, medicalRecordID, treatmentID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkUpdateTreatments godoc
// PUT /medical-records/:id/treatments
func (h *Handler) BulkUpdateTreatments(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
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

	if err := h.svc.Treatment.BulkUpdateSortOrder(c.Request.Context(), clinicID, medicalRecordID, input); err != nil {
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
