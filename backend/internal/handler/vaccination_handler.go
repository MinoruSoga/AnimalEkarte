package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

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

	vaccinations, total, err := h.svc.Vaccination.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
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
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createVaccinationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	vaccination := &model.Vaccination{
		ClinicID:        clinicID,
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
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateVaccinationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	vaccination := &model.Vaccination{
		ID:              id,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		DoctorID:        input.DoctorID,
		NextDate:        input.NextDate,
	}
	if input.VaccineID != nil {
		vaccination.VaccineID = *input.VaccineID
	}
	if input.Date != nil {
		vaccination.Date = *input.Date
	}
	if input.NextScheduleType != nil {
		nst := model.NextScheduleType(*input.NextScheduleType)
		vaccination.NextScheduleType = &nst
	}
	if input.Supplemental != nil {
		vaccination.Supplemental = *input.Supplemental
	}
	if input.Lot1 != nil {
		vaccination.Lot1 = *input.Lot1
	}
	if input.Lot2 != nil {
		vaccination.Lot2 = *input.Lot2
	}
	if input.Lot3 != nil {
		vaccination.Lot3 = *input.Lot3
	}
	if input.Lot4 != nil {
		vaccination.Lot4 = *input.Lot4
	}
	if input.Remarks != nil {
		vaccination.Remarks = *input.Remarks
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
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
