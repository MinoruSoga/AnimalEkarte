package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createExaminationInput struct {
	MedicalRecordID uint64    `json:"medical_record_id" binding:"required"`
	PetID           *uint64   `json:"pet_id"`
	ExamTypeID      uint64    `json:"exam_type_id"      binding:"required"`
	DoctorID        *uint64   `json:"doctor_id"`
	Date            time.Time `json:"date"              binding:"required"`
	ResultSummary   string    `json:"result_summary"`
	Machine         string    `json:"machine"`
	Status          string    `json:"status"`
}

type updateExaminationInput struct {
	MedicalRecordID uint64     `json:"medical_record_id"`
	PetID           *uint64    `json:"pet_id"`
	ExamTypeID      uint64     `json:"exam_type_id"`
	DoctorID        *uint64    `json:"doctor_id"`
	Date            *time.Time `json:"date"`
	ResultSummary   string     `json:"result_summary"`
	Machine         string     `json:"machine"`
	Status          string     `json:"status"`
}

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	exams, total, err := h.svc.Examination.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: exams, Total: total, Page: page, Limit: limit})
}

// GetExamination godoc
func (h *Handler) GetExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	exam, err := h.svc.Examination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, exam)
}

// CreateExamination godoc
func (h *Handler) CreateExamination(c *gin.Context) {
	var input createExaminationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam := &model.Exam{
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
	}
	if input.Status != "" {
		exam.Status = model.ExaminationStatus(input.Status)
	}

	if err := h.svc.Examination.Create(c.Request.Context(), exam); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, exam)
}

// UpdateExamination godoc
func (h *Handler) UpdateExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateExaminationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exam := &model.Exam{
		ID:              id,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
	}
	if input.Date != nil {
		exam.Date = *input.Date
	}
	if input.Status != "" {
		exam.Status = model.ExaminationStatus(input.Status)
	}

	if err := h.svc.Examination.Update(c.Request.Context(), clinicID, exam); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, exam)
}

// DeleteExamination godoc
func (h *Handler) DeleteExamination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Examination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
