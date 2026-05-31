package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListCarePlanItems godoc
// GET /hospitalizations/:id/care-plan-items
func (h *Handler) ListCarePlanItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	items, err := h.svc.CarePlanItem.List(c.Request.Context(), clinicID, hospitalizationID)
	if err != nil {
		RespondError(c, err)
		return
	}

	resp := make([]carePlanItemResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toCarePlanItemResponse(&items[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// CreateCarePlanItem godoc
// POST /hospitalizations/:id/care-plan-items
func (h *Handler) CreateCarePlanItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	item, err := h.svc.CarePlanItem.Create(c.Request.Context(), clinicID, hospitalizationID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/hospitalizations/%d/care-plan-items/%d", hospitalizationID, item.ID))
	c.JSON(http.StatusCreated, toCarePlanItemResponse(item))
}

// UpdateCarePlanItem godoc
// PATCH /hospitalizations/:id/care-plan-items/:itemId
func (h *Handler) UpdateCarePlanItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}

	var req updateCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	item, err := h.svc.CarePlanItem.Update(c.Request.Context(), clinicID, hospitalizationID, itemID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCarePlanItemResponse(item))
}

// DeleteCarePlanItem godoc
// DELETE /hospitalizations/:id/care-plan-items/:itemId
func (h *Handler) DeleteCarePlanItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}

	if err := h.svc.CarePlanItem.Delete(c.Request.Context(), clinicID, hospitalizationID, itemID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterCarePlanItemRoutes はケアプランアイテム関連のルートを登録する
func (h *Handler) RegisterCarePlanItemRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id/care-plan-items", h.RequirePermission(string(model.ResourceHospitalization), "view"), h.ListCarePlanItems)
	rg.POST("/:id/care-plan-items", h.RequirePermission(string(model.ResourceHospitalization), "create"), h.CreateCarePlanItem)
	rg.PATCH("/:id/care-plan-items/:itemId", h.RequirePermission(string(model.ResourceHospitalization), "edit"), h.UpdateCarePlanItem)
	rg.DELETE("/:id/care-plan-items/:itemId", h.RequirePermission(string(model.ResourceHospitalization), "delete"), h.DeleteCarePlanItem)
}
