package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// CarePlanItemHandler serves the care-plan-item HTTP boundary. Moved from internal/handler
// (BE9-2D ⑤ Batch C).
type CarePlanItemHandler struct {
	service CarePlanItemService
}

// NewCarePlanItemHandler initializes a CarePlanItemHandler.
func NewCarePlanItemHandler(service CarePlanItemService) *CarePlanItemHandler {
	return &CarePlanItemHandler{service: service}
}

// ListCarePlanItems godoc
// GET /hospitalizations/:id/care-plan-items
func (h *CarePlanItemHandler) ListCarePlanItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	items, err := h.service.List(c.Request.Context(), clinicID, hospitalizationID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	resp := httpapi.MapSlice(items, toCarePlanItemResponse)
	c.JSON(http.StatusOK, resp)
}

// CreateCarePlanItem godoc
// POST /hospitalizations/:id/care-plan-items
func (h *CarePlanItemHandler) CreateCarePlanItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	item, err := h.service.Create(c.Request.Context(), clinicID, hospitalizationID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/care-plan-items/%d", hospitalizationID, item.ID))
	c.JSON(http.StatusCreated, toCarePlanItemResponse(item))
}

// UpdateCarePlanItem godoc
// PATCH /hospitalizations/:id/care-plan-items/:itemId
func (h *CarePlanItemHandler) UpdateCarePlanItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	itemID, ok := httpapi.ParseIDParam(c, "itemId")
	if !ok {
		return
	}

	var req updateCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	item, err := h.service.Update(c.Request.Context(), clinicID, hospitalizationID, itemID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCarePlanItemResponse(item))
}

// DeleteCarePlanItem godoc
// DELETE /hospitalizations/:id/care-plan-items/:itemId
func (h *CarePlanItemHandler) DeleteCarePlanItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	itemID, ok := httpapi.ParseIDParam(c, "itemId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, hospitalizationID, itemID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
