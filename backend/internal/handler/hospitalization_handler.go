package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createHospitalizationInput struct {
	OwnerID             uint64    `json:"owner_id"              binding:"required"`
	PetID               uint64    `json:"pet_id"                binding:"required"`
	HospitalizationType string    `json:"hospitalization_type"  binding:"required"`
	StartDate           time.Time `json:"start_date"            binding:"required"`
	EndDate             time.Time `json:"end_date"              binding:"required"`
	Status              string    `json:"status"`
	CageID              *uint64   `json:"cage_id"`
	DoctorID            *uint64   `json:"doctor_id"`
	Memo                string    `json:"memo"`
	OwnerRequest        string    `json:"owner_request"`
	StaffNotes          string    `json:"staff_notes"`
}

type updateHospitalizationInput struct {
	OwnerID             uint64     `json:"owner_id"`
	PetID               uint64     `json:"pet_id"`
	HospitalizationType string     `json:"hospitalization_type"`
	StartDate           *time.Time `json:"start_date"`
	EndDate             *time.Time `json:"end_date"`
	Status              string     `json:"status"`
	CageID              *uint64    `json:"cage_id"`
	DoctorID            *uint64    `json:"doctor_id"`
	Memo                string     `json:"memo"`
	OwnerRequest        string     `json:"owner_request"`
	StaffNotes          string     `json:"staff_notes"`
}

// ListHospitalizations godoc
func (h *Handler) ListHospitalizations(c *gin.Context) {
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

	hospitalizations, total, err := h.svc.Hospitalization.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: hospitalizations, Total: total, Page: page, Limit: limit})
}

// GetHospitalization godoc
func (h *Handler) GetHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	hospitalization, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, hospitalization)
}

// CreateHospitalization godoc
func (h *Handler) CreateHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createHospitalizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hospitalization := &model.Hospitalization{
		ClinicID:            clinicID,
		OwnerID:             input.OwnerID,
		PetID:               input.PetID,
		HospitalizationType: model.HospitalizationType(input.HospitalizationType),
		StartDate:           input.StartDate,
		EndDate:             input.EndDate,
		CageID:              input.CageID,
		DoctorID:            input.DoctorID,
		Memo:                input.Memo,
		OwnerRequest:        input.OwnerRequest,
		StaffNotes:          input.StaffNotes,
	}
	if input.Status != "" {
		hospitalization.Status = model.HospitalizationStatus(input.Status)
	}

	if err := h.svc.Hospitalization.Create(c.Request.Context(), hospitalization); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, hospitalization)
}

// UpdateHospitalization godoc
func (h *Handler) UpdateHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateHospitalizationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hospitalization := &model.Hospitalization{
		ID:           id,
		ClinicID:     clinicID,
		OwnerID:      input.OwnerID,
		PetID:        input.PetID,
		CageID:       input.CageID,
		DoctorID:     input.DoctorID,
		Memo:         input.Memo,
		OwnerRequest: input.OwnerRequest,
		StaffNotes:   input.StaffNotes,
	}
	if input.HospitalizationType != "" {
		hospitalization.HospitalizationType = model.HospitalizationType(input.HospitalizationType)
	}
	if input.StartDate != nil {
		hospitalization.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		hospitalization.EndDate = *input.EndDate
	}
	if input.Status != "" {
		hospitalization.Status = model.HospitalizationStatus(input.Status)
	}

	if err := h.svc.Hospitalization.Update(c.Request.Context(), hospitalization); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, hospitalization)
}

// DeleteHospitalization godoc
func (h *Handler) DeleteHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Hospitalization.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
