package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

	filters, err := newListVaccinationQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	vaccinations, total, err := h.svc.Vaccination.List(
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
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(vaccinations, toVaccinationResponse), total, page, limit))
}

// GetVaccination godoc
func (h *Handler) GetVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	vaccination, err := h.svc.Vaccination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// CreateVaccination godoc
func (h *Handler) CreateVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}
	vaccination, err := h.svc.Vaccination.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/vaccinations/%d", vaccination.ID))
	c.JSON(http.StatusCreated, toVaccinationResponse(vaccination))
}

// UpdateVaccination godoc
func (h *Handler) UpdateVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
	}

	vaccination, err := h.svc.Vaccination.Update(c.Request.Context(), clinicID, id, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// DeleteVaccination godoc
func (h *Handler) DeleteVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Vaccination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
