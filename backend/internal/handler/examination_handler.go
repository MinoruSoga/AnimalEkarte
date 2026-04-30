package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListExaminations godoc
func (h *Handler) ListExaminations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uint64
	if s := c.Query("pet_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
			return
		}
		petID = &id
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	startDate, err := parseDateQuery(c, "start_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	endDate, err := parseDateQuery(c, "end_date")
	if err != nil {
		RespondError(c, err)
		return
	}

	exams, total, err := h.svc.Examination.List(c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(exams, toExaminationResponse), total, page, limit))
}

// GetExamination godoc
func (h *Handler) GetExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	exam, err := h.svc.Examination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// CreateExamination godoc
func (h *Handler) CreateExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := &service.CreateExaminationInput{
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
		Status:          model.ExaminationStatus(input.Status),
	}
	exam, err := h.svc.Examination.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/examinations/%d", exam.ID))
	c.JSON(http.StatusCreated, toExaminationResponse(exam))
}

// UpdateExamination godoc
func (h *Handler) UpdateExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateExaminationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var status *model.ExaminationStatus
	if input.Status != nil {
		s := model.ExaminationStatus(*input.Status)
		status = &s
	}
	var examTypeID *uint64
	if input.ExamTypeID != 0 {
		v := input.ExamTypeID
		examTypeID = &v
	}

	svcInput := service.UpdateExaminationInput{
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      examTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
		Status:          status,
	}

	exam, err := h.svc.Examination.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationResponse(exam))
}

// ListExaminationItems godoc
//
//	@Summary	検査項目一覧を取得
//	@Tags		Examinations
//	@Produce	json
//	@Param		id	path	int	true	"Examination ID"
//	@Success	200	{object}	examItemsResponse
func (h *Handler) ListExaminationItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.Examination.ListItems(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamItemsResponse(items))
}

// ReplaceExaminationItems godoc
//
//	@Summary	検査項目を一括置換（PUT セマンティクス: 既存全削除→一括登録）
//	@Tags		Examinations
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int						true	"Examination ID"
//	@Param		body	body	replaceExamItemsRequest	true	"items"
//	@Success	200	{object}	examItemsResponse
func (h *Handler) ReplaceExaminationItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req replaceExamItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	inputs := make([]service.UpsertExamItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		inputs = append(inputs, service.UpsertExamItemInput{
			ExamTypeFieldID: it.ExamTypeFieldID,
			Name:            it.Name,
			InspectionValue: it.InspectionValue,
			NormalValue:     it.NormalValue,
			Unit:            it.Unit,
			ReferenceValue:  it.ReferenceValue,
			RefMin:          it.RefMin,
			RefMax:          it.RefMax,
			SortOrder:       it.SortOrder,
		})
	}

	saved, err := h.svc.Examination.ReplaceItems(c.Request.Context(), clinicID, id, inputs)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamItemsResponse(saved))
}

// DeleteExamination godoc
func (h *Handler) DeleteExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Examination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
