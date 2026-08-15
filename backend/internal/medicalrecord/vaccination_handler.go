package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// VaccinationHandler serves the Vaccination (接種記録) HTTP boundary.
type VaccinationHandler struct {
	service VaccinationService
}

// NewVaccinationHandler initializes a VaccinationHandler.
func NewVaccinationHandler(service VaccinationService) *VaccinationHandler {
	return &VaccinationHandler{service: service}
}

// ListVaccinations godoc
func (h *VaccinationHandler) ListVaccinations(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters, err := newListVaccinationQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	vaccinations, total, err := h.service.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(vaccinations, toVaccinationResponse), total, page, limit))
}

// GetVaccination godoc
func (h *VaccinationHandler) GetVaccination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	vaccination, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// CreateVaccination godoc
func (h *VaccinationHandler) CreateVaccination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	vaccination, err := h.service.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/vaccinations/%d", vaccination.ID))
	c.JSON(http.StatusCreated, toVaccinationResponse(vaccination))
}

// UpdateVaccination godoc
func (h *VaccinationHandler) UpdateVaccination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	vaccination, err := h.service.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// DeleteVaccination godoc
func (h *VaccinationHandler) DeleteVaccination(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
