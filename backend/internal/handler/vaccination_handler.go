package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createVaccinationInput struct {
	MedicalRecordID  *uint64    `json:"medical_record_id"`
	PetID            *uint64    `json:"pet_id"`
	VaccineID        uint64     `json:"vaccine_id"          binding:"required"`
	Date             time.Time  `json:"date"                binding:"required"`
	DoctorID         *uint64    `json:"doctor_id"`
	NextDate         *time.Time `json:"next_date"`
	NextScheduleType string     `json:"next_schedule_type"`
	Supplemental     string     `json:"supplemental"`
	Lot1             string     `json:"lot1"`
	Lot2             string     `json:"lot2"`
	Lot3             string     `json:"lot3"`
	Lot4             string     `json:"lot4"`
	Remarks          string     `json:"remarks"`
}

type updateVaccinationInput struct {
	MedicalRecordID  *uint64    `json:"medical_record_id"`
	PetID            *uint64    `json:"pet_id"`
	VaccineID        uint64     `json:"vaccine_id"`
	Date             *time.Time `json:"date"`
	DoctorID         *uint64    `json:"doctor_id"`
	NextDate         *time.Time `json:"next_date"`
	NextScheduleType string     `json:"next_schedule_type"`
	Supplemental     string     `json:"supplemental"`
	Lot1             string     `json:"lot1"`
	Lot2             string     `json:"lot2"`
	Lot3             string     `json:"lot3"`
	Lot4             string     `json:"lot4"`
	Remarks          string     `json:"remarks"`
}

// ListVaccinations godoc
func (h *Handler) ListVaccinations(c *gin.Context) {
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

	vaccinations, total, err := h.svc.Vaccination.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(vaccinations, total, page, limit))
}

// GetVaccination godoc
func (h *Handler) GetVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	vaccination, err := h.svc.Vaccination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccination)
}

// CreateVaccination godoc
func (h *Handler) CreateVaccination(c *gin.Context) {
	var input createVaccinationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vaccination := &model.Vaccination{
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		VaccineID:       input.VaccineID,
		Date:            input.Date,
		DoctorID:        input.DoctorID,
		NextDate:        input.NextDate,
		Supplemental:    input.Supplemental,
		Lot1:            input.Lot1,
		Lot2:            input.Lot2,
		Lot3:            input.Lot3,
		Lot4:            input.Lot4,
		Remarks:         input.Remarks,
	}
	if input.NextScheduleType != "" {
		nst := model.NextScheduleType(input.NextScheduleType)
		vaccination.NextScheduleType = &nst
	}

	if err := h.svc.Vaccination.Create(c.Request.Context(), vaccination); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, vaccination)
}

// UpdateVaccination godoc
func (h *Handler) UpdateVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateVaccinationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vaccination := &model.Vaccination{
		ID:              id,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		VaccineID:       input.VaccineID,
		DoctorID:        input.DoctorID,
		NextDate:        input.NextDate,
		Supplemental:    input.Supplemental,
		Lot1:            input.Lot1,
		Lot2:            input.Lot2,
		Lot3:            input.Lot3,
		Lot4:            input.Lot4,
		Remarks:         input.Remarks,
	}
	if input.Date != nil {
		vaccination.Date = *input.Date
	}
	if input.NextScheduleType != "" {
		nst := model.NextScheduleType(input.NextScheduleType)
		vaccination.NextScheduleType = &nst
	}

	if err := h.svc.Vaccination.Update(c.Request.Context(), clinicID, vaccination); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccination)
}

// DeleteVaccination godoc
func (h *Handler) DeleteVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Vaccination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterVaccinationRoutes は予防接種関連のルートを登録する
func (h *Handler) RegisterVaccinationRoutes(rg *gin.RouterGroup) {
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", h.ListVaccinations)
	vaccinations.POST("", h.CreateVaccination)
	vaccinations.GET("/:id", h.GetVaccination)
	vaccinations.PATCH("/:id", h.UpdateVaccination)
	vaccinations.DELETE("/:id", h.DeleteVaccination)
}
