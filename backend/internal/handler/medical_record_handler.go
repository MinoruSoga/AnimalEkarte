package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListMedicalRecords godoc
func (h *Handler) ListMedicalRecords(c *gin.Context) {
	// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
	clinicIDs, ok := resolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	q := newListMedicalRecordQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	records, total, err := h.svc.MedicalRecord.List(
		c.Request.Context(),
		clinicIDs,
		filters,
		page,
		limit,
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(records, func(r *model.MedicalRecord) medicalRecordResponse {
		return toMedicalRecordResponseWithVisitCount(r, 0)
	}), total, page, limit))
}

// GetMedicalRecord godoc
func (h *Handler) GetMedicalRecord(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := resolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	record, err := h.svc.MedicalRecord.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		RespondError(c, err)
		return
	}

	// pet_id がある場合のみ来院回数を取得する（clinicIDs[0] = mainClinicID for cross-clinic scope）
	var visitCount int64
	if record.PetID != nil {
		visitCount, err = h.svc.MedicalRecord.CountByPetID(c.Request.Context(), record.ClinicID, *record.PetID)
		if err != nil {
			RespondError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, toMedicalRecordResponseWithVisitCount(record, visitCount))
}

// CreateMedicalRecord godoc
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	var input createMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	svcInput, err := input.toServiceInput(staffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	ctx := c.Request.Context()
	record, err := h.svc.MedicalRecord.Create(ctx, clinicID, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	h.svc.MedicalRecord.CreateSubRecords(ctx, clinicID, record.ID, input.toSubRecordsInput())
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d", record.ID))
	c.JSON(http.StatusCreated, toMedicalRecordResponseWithVisitCount(record, 0))
}

// UpdateMedicalRecord godoc
func (h *Handler) UpdateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	svcInput, err := input.toServiceInput(staffID)
	if err != nil {
		RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	record, err := h.svc.MedicalRecord.Update(ctx, clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicalRecordResponse(record))
}

// PatchMedicalRecordRecommendationReason godoc
// PATCH /medical-records/:id/recommendation-reason — 受診推奨理由を更新する（FEAT-381-2）。
func (h *Handler) PatchMedicalRecordRecommendationReason(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchMedicalRecordRecommendationReasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	record, err := h.svc.MedicalRecord.UpdateRecommendationReason(
		c.Request.Context(), clinicID, id,
		req.toServiceInput(),
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicalRecordResponse(record))
}

// DeleteMedicalRecord godoc
func (h *Handler) DeleteMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.MedicalRecord.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
