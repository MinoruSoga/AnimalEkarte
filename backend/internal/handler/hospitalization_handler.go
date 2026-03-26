package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

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

	hospitalizations, total, err := h.svc.Hospitalization.List(c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(hospitalizations, total, page, limit))
}

// GetHospitalization godoc
func (h *Handler) GetHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	var input createHospitalizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	hospType, err := validateEnum(input.HospitalizationType,
		model.HospitalizationTypeInpatient,
		model.HospitalizationTypeHotel,
	)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid hospitalization_type: "+err.Error()))
		return
	}

	hospitalization := &model.Hospitalization{
		ClinicID:            clinicID,
		OwnerID:             input.OwnerID,
		PetID:               input.PetID,
		HospitalizationType: hospType,
		StartDate:           input.StartDate,
		EndDate:             input.EndDate,
		CageID:              input.CageID,
		DoctorID:            input.DoctorID,
		Memo:                input.Memo,
		OwnerRequest:        input.OwnerRequest,
		StaffNotes:          input.StaffNotes,
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.HospitalizationStatusAdmitted,
			model.HospitalizationStatusDischarged,
			model.HospitalizationStatusReserved,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		hospitalization.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.Hospitalization.Create(ctx, hospitalization); err != nil {
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
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateHospitalizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateHospitalizationInput{
		OwnerID:      input.OwnerID,
		PetID:        input.PetID,
		StartDate:    input.StartDate,
		EndDate:      input.EndDate,
		CageID:       input.CageID,
		DoctorID:     input.DoctorID,
		Memo:         input.Memo,
		OwnerRequest: input.OwnerRequest,
		StaffNotes:   input.StaffNotes,
	}
	if input.HospitalizationType != nil {
		hospType, err := validateEnum(*input.HospitalizationType,
			model.HospitalizationTypeInpatient,
			model.HospitalizationTypeHotel,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid hospitalization_type: "+err.Error()))
			return
		}
		svcInput.HospitalizationType = &hospType
	}
	if input.Status != nil {
		status, err := validateEnum(*input.Status,
			model.HospitalizationStatusAdmitted,
			model.HospitalizationStatusDischarged,
			model.HospitalizationStatusReserved,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		svcInput.Status = &status
	}

	ctx := c.Request.Context()
	hosp, err := h.svc.Hospitalization.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, hosp)
}

// DeleteHospitalization godoc
func (h *Handler) DeleteHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Hospitalization.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterHospitalizationRoutes は入院関連のルートを登録する
func (h *Handler) RegisterHospitalizationRoutes(rg *gin.RouterGroup) {
	hospitalizations := rg.Group("/hospitalizations")
	hospitalizations.GET("", h.ListHospitalizations)
	hospitalizations.POST("", h.CreateHospitalization)
	hospitalizations.GET("/:id", h.GetHospitalization)
	hospitalizations.PATCH("/:id", h.UpdateHospitalization)
	hospitalizations.DELETE("/:id", h.DeleteHospitalization)

	h.RegisterDailyRecordRoutes(hospitalizations)
	h.RegisterCarePlanItemRoutes(hospitalizations)
	h.RegisterTreatmentPlanHospitalizationRoutes(hospitalizations)
}
