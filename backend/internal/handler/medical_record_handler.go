package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createMedicalRecordInput struct {
	RecordNo                 string    `json:"record_no"                   binding:"required"`
	Date                     time.Time `json:"date"                        binding:"required"`
	OwnerID                  *uint64   `json:"owner_id"`
	PetID                    *uint64   `json:"pet_id"`
	DoctorID                 *uint64   `json:"doctor_id"`
	ReservationAppointmentID *uint64   `json:"reservation_appointment_id"`
	Status                   string    `json:"status"`
}

type updateMedicalRecordInput struct {
	RecordNo                 string     `json:"record_no"`
	Date                     *time.Time `json:"date"`
	OwnerID                  *uint64    `json:"owner_id"`
	PetID                    *uint64    `json:"pet_id"`
	DoctorID                 *uint64    `json:"doctor_id"`
	ReservationAppointmentID *uint64    `json:"reservation_appointment_id"`
	Status                   string     `json:"status"`
}

// ListMedicalRecords godoc
func (h *Handler) ListMedicalRecords(c *gin.Context) {
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
	if petIDStr := c.Query("pet_id"); petIDStr != "" {
		id, err := strconv.ParseUint(petIDStr, 10, 64)
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

	records, total, err := h.svc.MedicalRecord.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(records, total, page, limit))
}

// GetMedicalRecord godoc
func (h *Handler) GetMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	record, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// CreateMedicalRecord godoc
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createMedicalRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := &model.MedicalRecord{
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		Date:                     input.Date,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		ReservationAppointmentID: input.ReservationAppointmentID,
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		record.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.MedicalRecord.Create(ctx, record); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "medical record created",
		slog.Uint64("record_id", record.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
	c.JSON(http.StatusCreated, record)
}

// UpdateMedicalRecord godoc
func (h *Handler) UpdateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateMedicalRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := &model.MedicalRecord{
		ID:                       id,
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		ReservationAppointmentID: input.ReservationAppointmentID,
	}
	if input.Date != nil {
		record.Date = *input.Date
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		record.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.MedicalRecord.Update(ctx, record); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "medical record updated",
		slog.Uint64("record_id", record.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
	c.JSON(http.StatusOK, record)
}

// DeleteMedicalRecord godoc
func (h *Handler) DeleteMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.MedicalRecord.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
