package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

	q := newListHospitalizationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	hospitalizations, total, err := h.svc.Hospitalization.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.Status,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(hospitalizations, toHospitalizationResponse), total, page, limit))
}

// GetHospitalization godoc
func (h *Handler) GetHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	hospitalization, err := h.svc.Hospitalization.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationResponse(hospitalization))
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

	svcInput, err := input.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}
	ctx := c.Request.Context()
	hospitalization, err := h.svc.Hospitalization.Create(ctx, clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d", hospitalization.ID))
	c.JSON(http.StatusCreated, toHospitalizationResponse(hospitalization))
}

// UpdateHospitalization godoc
func (h *Handler) UpdateHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateHospitalizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()
	hosp, err := h.svc.Hospitalization.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationResponse(hosp))
}

// DischargeWithBilling godoc
// POST /hospitalizations/:id/discharge-with-billing
func (h *Handler) DischargeWithBilling(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dischargeWithBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	result, err := h.svc.Hospitalization.DischargeWithBilling(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDischargeWithBillingResponse(result))
}

// DeleteHospitalization godoc
func (h *Handler) DeleteHospitalization(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Hospitalization.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
