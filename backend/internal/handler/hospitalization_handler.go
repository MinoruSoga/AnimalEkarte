package handler

import (
	"log/slog"
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
	OwnerID             *uint64    `json:"owner_id"`
	PetID               *uint64    `json:"pet_id"`
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

	hospType, err := validateEnum(input.HospitalizationType,
		model.HospitalizationTypeInpatient,
		model.HospitalizationTypeHotel,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospitalization_type: " + err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		hospitalization.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.Hospitalization.Create(ctx, hospitalization); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
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
		CageID:       input.CageID,
		DoctorID:     input.DoctorID,
		Memo:         input.Memo,
		OwnerRequest: input.OwnerRequest,
		StaffNotes:   input.StaffNotes,
	}
	if input.OwnerID != nil {
		hospitalization.OwnerID = *input.OwnerID
	}
	if input.PetID != nil {
		hospitalization.PetID = *input.PetID
	}
	if input.HospitalizationType != "" {
		hospType, err := validateEnum(input.HospitalizationType,
			model.HospitalizationTypeInpatient,
			model.HospitalizationTypeHotel,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospitalization_type: " + err.Error()})
			return
		}
		hospitalization.HospitalizationType = hospType
	}
	if input.StartDate != nil {
		hospitalization.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		hospitalization.EndDate = *input.EndDate
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.HospitalizationStatusAdmitted,
			model.HospitalizationStatusDischarged,
			model.HospitalizationStatusReserved,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		hospitalization.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.Hospitalization.Update(ctx, hospitalization); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
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
