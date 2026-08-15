package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// HospitalizationHandler serves the hospitalization HTTP boundary. Moved from internal/handler
// (BE9-2D ⑤ Batch C).
type HospitalizationHandler struct {
	service       HospitalizationService
	hasPermission PermissionChecker
}

// NewHospitalizationHandler initializes a HospitalizationHandler.
// hasPermission is required for nested treatment_plans discount:create guards (TASK-001-BE).
// Pass nil only in tests that never exercise nested discounted plans.
func NewHospitalizationHandler(service HospitalizationService, hasPermission PermissionChecker) *HospitalizationHandler {
	return &HospitalizationHandler{service: service, hasPermission: hasPermission}
}

// ListHospitalizations godoc
func (h *HospitalizationHandler) ListHospitalizations(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	q := newListHospitalizationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	hospitalizations, total, err := h.service.List(
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
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(hospitalizations, toHospitalizationResponse), total, page, limit))
}

// GetHospitalization godoc
func (h *HospitalizationHandler) GetHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	hospitalization, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationResponse(hospitalization))
}

// CreateHospitalization godoc
func (h *HospitalizationHandler) CreateHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var input createHospitalizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// Nested plans reuse the same discount:create guards as POST .../treatment-plans.
	for i := range input.TreatmentPlans {
		plan := input.TreatmentPlans[i]
		if err := requireDiscountCreateFloat(c, h.hasPermission, plan.DiscountRate); err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := requireDiscountCreateInt(c, h.hasPermission, plan.DiscountAmount); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	svcInput, err := input.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}
	ctx := c.Request.Context()
	hospitalization, err := h.service.Create(ctx, clinicID, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d", hospitalization.ID))
	c.JSON(http.StatusCreated, toHospitalizationResponse(hospitalization))
}

// UpdateHospitalization godoc
func (h *HospitalizationHandler) UpdateHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateHospitalizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()
	hosp, err := h.service.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toHospitalizationResponse(hosp))
}

// DischargeWithBilling godoc
// POST /hospitalizations/:id/discharge-with-billing
func (h *HospitalizationHandler) DischargeWithBilling(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if staffID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid user context"))
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req dischargeWithBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	result, err := h.service.DischargeWithBilling(c.Request.Context(), clinicID, id, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDischargeWithBillingResponse(result))
}

// DeleteHospitalization godoc
func (h *HospitalizationHandler) DeleteHospitalization(c *gin.Context) {
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
