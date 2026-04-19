package handler

import (
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
	c.JSON(http.StatusOK, newPaginatedResponse(exams, total, page, limit))
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
