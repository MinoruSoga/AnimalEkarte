package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
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
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceHospitalization), "view") {
		return
	}
	q := newListHospitalizationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	if h.maybeListHospitalizationSnapshot(c, clinicID, filters) {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	hospitalizations, total, err := h.service.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.LifecycleState,
		filters.PlannedEntryFrom,
		filters.PlannedEntryTo,
		page,
		limit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	projections, err := h.service.StayProjections(c.Request.Context(), clinicID, hospitalizations)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(hospitalizationResponses(hospitalizations, projections), total, page, limit))
}

// singleHospitalizationResponse renders one row with its stay projection (EMR-499 U1).
func (h *HospitalizationHandler) singleHospitalizationResponse(c *gin.Context, clinicID uint64, row *model.Hospitalization) (HospitalizationResponse, bool) {
	projections, err := h.service.StayProjections(c.Request.Context(), clinicID, []model.Hospitalization{*row})
	if err != nil {
		httpapi.RespondError(c, err)
		return HospitalizationResponse{}, false
	}
	return hospitalizationResponses([]model.Hospitalization{*row}, projections)[0], true
}

// GetHospitalization godoc
func (h *HospitalizationHandler) GetHospitalization(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceHospitalization), "view") {
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
	resp, ok := h.singleHospitalizationResponse(c, clinicID, hospitalization)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, resp)
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
		respondToServiceInputError(c, err)
		return
	}
	ctx := c.Request.Context()
	hospitalization, err := h.service.Create(ctx, clinicID, svcInput)
	if err != nil {
		respondHospitalizationCageWriteError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d", hospitalization.ID))
	resp, ok := h.singleHospitalizationResponse(c, clinicID, hospitalization)
	if !ok {
		return
	}
	c.JSON(http.StatusCreated, resp)
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
		respondToServiceInputError(c, err)
		return
	}

	ctx := c.Request.Context()
	hosp, err := h.service.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		respondHospitalizationCageWriteError(c, err)
		return
	}
	resp, ok := h.singleHospitalizationResponse(c, clinicID, hosp)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, resp)
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

// SwapCages godoc
// POST /hospitalizations/cage-swap
// EMR-306 review: 2本の独立 PATCH による半スワップ（片側のみ commit）を防ぐため、
// 2件の入院のケージ交換を単一トランザクションで実行する。
// expected_cage_ids は hospitalization_ids と同じ位置の表示時ケージID。
func (h *HospitalizationHandler) SwapCages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req cageSwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	first, second, err := h.service.SwapCages(
		c.Request.Context(), clinicID, req.HospitalizationIDs[0], req.HospitalizationIDs[1],
		[2]uint64{req.ExpectedCageIDs[0], req.ExpectedCageIDs[1]})
	if err != nil {
		respondHospitalizationCageWriteError(c, err)
		return
	}
	pair := []model.Hospitalization{*first, *second}
	projections, err := h.service.StayProjections(c.Request.Context(), clinicID, pair)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, CageSwapResponse{Hospitalizations: hospitalizationResponses(pair, projections)})
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
