package medicalrecord

import (
	"fmt"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordHandler serves the medical-record HTTP boundary. Moved from internal/handler
// (BE9-2D ⑦ Batch C).
type MedicalRecordHandler struct {
	service MedicalRecordService
}

// NewMedicalRecordHandler initializes a MedicalRecordHandler.
func NewMedicalRecordHandler(service MedicalRecordService) *MedicalRecordHandler {
	return &MedicalRecordHandler{service: service}
}

// ListMedicalRecords godoc
func (h *MedicalRecordHandler) ListMedicalRecords(c *gin.Context) {
	// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	q := newListMedicalRecordQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	records, total, err := h.service.List(
		c.Request.Context(),
		clinicIDs,
		filters,
		page,
		limit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(records, func(r *model.MedicalRecord) MedicalRecordResponse {
		return toMedicalRecordResponseWithVisitCount(r, 0)
	}), total, page, limit))
}

// GetMedicalRecord godoc
func (h *MedicalRecordHandler) GetMedicalRecord(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := httpapi.ResolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	record, err := h.service.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// pet_id がある場合のみ来院回数を取得する（clinicIDs[0] = mainClinicID for cross-clinic scope）
	var visitCount int64
	if record.PetID != nil {
		visitCount, err = h.service.CountByPetID(c.Request.Context(), record.ClinicID, *record.PetID)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, toMedicalRecordResponseWithVisitCount(record, visitCount))
}

// CreateMedicalRecord godoc
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var input createMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	svcInput, err := input.toServiceInput(staffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	ctx := c.Request.Context()
	record, err := h.service.Create(ctx, clinicID, &svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// MRC-04 / DEC-32: do not return 201 when chief complaint / clinical plan subrecords failed.
	if err := h.service.CreateSubRecords(ctx, clinicID, record.ID, input.toSubRecordsInput()); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d", record.ID))
	c.JSON(http.StatusCreated, toMedicalRecordResponseWithVisitCount(record, 0))
}

// UpdateMedicalRecord godoc
func (h *MedicalRecordHandler) UpdateMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	svcInput, err := input.toServiceInput(staffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	record, err := h.service.Update(ctx, clinicID, id, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicalRecordResponse(record))
}

// UpdateMedicalRecordRecommendationReason godoc
// PATCH /medical-records/:id/recommendation-reason — 受診推奨理由を更新する（FEAT-381-2）。
func (h *MedicalRecordHandler) UpdateMedicalRecordRecommendationReason(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchMedicalRecordRecommendationReasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	record, err := h.service.UpdateRecommendationReason(
		c.Request.Context(), clinicID, id,
		req.toServiceInput(),
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicalRecordResponse(record))
}

// DeleteMedicalRecord godoc
func (h *MedicalRecordHandler) DeleteMedicalRecord(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
