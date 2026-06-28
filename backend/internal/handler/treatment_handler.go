package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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

// ListPetTreatmentHistory godoc
// GET /pets/:id/treatment-history?item_type=medicine|procedure|all&anesthesia_only=true&is_surgery=true (#158/#159 飼主レポート)
// ペット単位で治療履歴を medical_records.date 降順で返す。clinic_id 隔離は service/repository が medical_records JOIN で担保する。
func (h *Handler) ListPetTreatmentHistory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	itemType, ok := parseTreatmentItemTypeFilter(c)
	if !ok {
		return
	}

	filter := model.PetTreatmentHistoryFilter{
		ItemType:       itemType,
		AnesthesiaOnly: c.Query("anesthesia_only") == "true",
		IsSurgery:      c.Query("is_surgery") == "true",
	}

	treatments, total, err := h.svc.Treatment.ListPetHistory(c.Request.Context(), clinicID, petID, filter, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(treatments, toPetTreatmentHistoryResponse), total, page, limit))
}

// parseTreatmentItemTypeFilter は item_type クエリを optional な絞り込みにパースする。
// "" / "all" は nil（全 item_type）。不正値は 400 を書いて false を返す。
func parseTreatmentItemTypeFilter(c *gin.Context) (*model.TreatmentItemType, bool) {
	raw := c.Query("item_type")
	if raw == "" || raw == "all" {
		return nil, true
	}
	t := model.TreatmentItemType(raw)
	switch t {
	case model.TreatmentItemTypeConsultation,
		model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine,
		model.TreatmentItemTypeOther:
		return &t, true
	}
	RespondError(c, apperrors.WrapInvalidInput("item_type は medicine / procedure / consultation / other / all のいずれかを指定してください"))
	return nil, false
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

	createInput := req.toServiceInput()
	createInput.ActorID = optionalStaffID(c) // #201 B-2: 逸脱 audit の実施者
	treatment, err := h.svc.Treatment.Create(c.Request.Context(), clinicID, medicalRecordID, createInput)
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

	updateInput := req.toServiceInput()
	updateInput.ActorID = optionalStaffID(c) // #201 B-2: 逸脱 audit の実施者
	treatment, err := h.svc.Treatment.Update(c.Request.Context(), clinicID, medicalRecordID, treatmentID, updateInput)
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

	if err := h.svc.Treatment.BulkUpdateSortOrder(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput()); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterTreatmentRoutes は治療項目関連のルートをカルテサブリソースとして登録する
func (h *Handler) RegisterTreatmentRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/treatments", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListTreatments)
	rg.POST("/:id/treatments", h.RequirePermission(string(model.ResourceMedicalRecords), "create"), h.CreateTreatment)
	rg.PATCH("/:id/treatments/:treatmentId", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.UpdateTreatment)
	rg.DELETE("/:id/treatments/:treatmentId", h.RequirePermission(string(model.ResourceMedicalRecords), "delete"), h.DeleteTreatment)
	rg.PUT("/:id/treatments", h.RequirePermission(string(model.ResourceMedicalRecords), "edit"), h.BulkUpdateTreatments)
}
