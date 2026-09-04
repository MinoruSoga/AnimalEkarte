package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionChecker is medicalrecord's consumer-side view of the boolean permission check
// (transitional *handler.Handler.HasPermission method value; once auth migrates, an
// *auth.Handler method value instead) — TreatmentHandler の BUG-372 discount ガードは route
// middleware でなく inline の bool 判定を要するため、PermissionMiddleware とは別 view にする。
type PermissionChecker func(c *gin.Context, resource, action string) bool

// TreatmentHandler serves the treatment (治療項目) HTTP boundary nested under a medical record,
// plus the cross-record pet treatment history (#158/#159). Moved from internal/handler
// (BE9-2D sub-batch④b).
type TreatmentHandler struct {
	service       TreatmentService
	hasPermission PermissionChecker
}

// NewTreatmentHandler initializes a TreatmentHandler.
func NewTreatmentHandler(service TreatmentService, hasPermission PermissionChecker) *TreatmentHandler {
	return &TreatmentHandler{service: service, hasPermission: hasPermission}
}

// ── BUG-372 discount permission guards ──
// 実装は discount_permission.go の package-level 共有関数（⑥で treatment-plan と共用化）。
// 既存テスト互換のため thin method を維持する。

func (h *TreatmentHandler) requireDiscountEditFloat(c *gin.Context, newVal *float64, oldVal float64) error {
	return requireDiscountEditFloat(c, h.hasPermission, newVal, oldVal)
}

func (h *TreatmentHandler) requireDiscountEditInt(c *gin.Context, newVal *int64, oldVal int64) error {
	return requireDiscountEditInt(c, h.hasPermission, newVal, oldVal)
}

func (h *TreatmentHandler) requireDiscountCreateFloat(c *gin.Context, val float64) error {
	return requireDiscountCreateFloat(c, h.hasPermission, val)
}

func (h *TreatmentHandler) requireDiscountCreateInt(c *gin.Context, val int64) error {
	return requireDiscountCreateInt(c, h.hasPermission, val)
}

// ListTreatments godoc
// GET /medical-records/:id/treatments
func (h *TreatmentHandler) ListTreatments(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMedicalRecords), "view") {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	treatments, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items := httpapi.MapSlice(treatments, toTreatmentResponse)
	c.JSON(http.StatusOK, items)
}

// ListPetTreatmentHistory godoc
// GET /pets/:id/treatment-history?item_type=medicine|procedure|all&anesthesia_only=true&is_surgery=true (#158/#159 飼主レポート)
// ペット単位で治療履歴を medical_records.date 降順で返す。clinic_id 隔離は service/repository が medical_records JOIN で担保する。
func (h *TreatmentHandler) ListPetTreatmentHistory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMedicalRecords), "view") {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
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

	treatments, total, err := h.service.ListPetHistory(c.Request.Context(), clinicID, petID, filter, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(treatments, toPetTreatmentHistoryResponse), total, page, limit))
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
	httpapi.RespondError(c, apperrors.WrapInvalidInput("item_type は medicine / procedure / consultation / other / all のいずれかを指定してください"))
	return nil, false
}

// CreateTreatment godoc
// POST /medical-records/:id/treatments
func (h *TreatmentHandler) CreateTreatment(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount フィールドにゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateFloat(c, req.DiscountRate); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := h.requireDiscountCreateInt(c, req.DiscountAmount); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	createInput := req.toServiceInput()
	createInput.ActorID = httpapi.OptionalStaffID(c) // #201 B-2: 逸脱 audit の実施者
	treatment, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, createInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/treatments/%d", medicalRecordID, treatment.ID))
	c.JSON(http.StatusCreated, toTreatmentResponse(treatment))
}

// UpdateTreatment godoc
// PATCH /medical-records/:id/treatments/:treatmentId
func (h *TreatmentHandler) UpdateTreatment(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	treatmentID, ok := httpapi.ParseIDParam(c, "treatmentId")
	if !ok {
		return
	}

	var req updateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount フィールドを変更する場合は既存値と比較し権限チェック
	if req.DiscountRate != nil || req.DiscountAmount != nil {
		existing, err := h.service.GetByID(c.Request.Context(), clinicID, treatmentID)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditFloat(c, req.DiscountRate, existing.DiscountRate); err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditInt(c, req.DiscountAmount, existing.DiscountAmount); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	updateInput := req.toServiceInput()
	updateInput.ActorID = httpapi.OptionalStaffID(c) // #201 B-2: 逸脱 audit の実施者
	// SEC-CS-F09: pass discount:edit into the write TX so the service can recheck against the locked row.
	if h.hasPermission != nil {
		updateInput.DiscountEditAllowed = h.hasPermission(c, string(model.ResourceDiscount), "edit")
	}
	treatment, err := h.service.Update(c.Request.Context(), clinicID, medicalRecordID, treatmentID, updateInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentResponse(treatment))
}

// DeleteTreatment godoc
// DELETE /medical-records/:id/treatments/:treatmentId
func (h *TreatmentHandler) DeleteTreatment(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	treatmentID, ok := httpapi.ParseIDParam(c, "treatmentId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, treatmentID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkUpdateTreatments godoc
// PUT /medical-records/:id/treatments
func (h *TreatmentHandler) BulkUpdateTreatments(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req bulkUpdateTreatmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	if err := h.service.BulkUpdateSortOrder(c.Request.Context(), clinicID, medicalRecordID, req.toServiceInput()); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
