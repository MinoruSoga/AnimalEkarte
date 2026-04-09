package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListCarePlanItems godoc
// GET /hospitalizations/:id/care-plan-items
func (h *Handler) ListCarePlanItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	hospitalizationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req createCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.CreateCarePlanItemInput{
		Type:                  req.Type,
		Name:                  req.Name,
		Description:           req.Description,
		Timing:                req.Timing,
		Status:                req.Status,
		Notes:                 req.Notes,
		MedicineID:            req.MedicineID,
		ProcedureID:           req.ProcedureID,
		HospitalizationPlanID: req.HospitalizationPlanID,
		UnitPrice:             req.UnitPrice,
		Category:              req.Category,
		SortOrder:             req.SortOrder,
	}

	item, err := h.svc.CarePlanItem.Create(c.Request.Context(), clinicID, hospitalizationID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toCarePlanItemResponse(item))
}

// UpdateCarePlanItem godoc
// PATCH /hospitalizations/:id/care-plan-items/:itemId
func (h *Handler) UpdateCarePlanItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	hospitalizationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid itemId"))
		return
	}

	var req updateCarePlanItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// Timing: pass nil when the key is absent from request vs empty slice to clear.
	// Since Go JSON unmarshals missing array as nil and present empty array as []string{},
	// we can pass req.Timing directly: nil means not provided, non-nil means update.
	input := &service.UpdateCarePlanItemInput{
		Type:                  req.Type,
		Name:                  req.Name,
		Description:           req.Description,
		Timing:                req.Timing,
		Status:                req.Status,
		Notes:                 req.Notes,
		MedicineID:            req.MedicineID,
		ProcedureID:           req.ProcedureID,
		HospitalizationPlanID: req.HospitalizationPlanID,
		UnitPrice:             req.UnitPrice,
		Category:              req.Category,
		SortOrder:             req.SortOrder,
	}

	item, err := h.svc.CarePlanItem.Update(c.Request.Context(), clinicID, hospitalizationID, itemID, input)
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
	hospitalizationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid itemId"))
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
	rg.GET("/:id/care-plan-items", h.ListCarePlanItems)
	rg.POST("/:id/care-plan-items", h.RequirePermission(string(model.ResourceHospitalization), "create"), h.CreateCarePlanItem)
	rg.PATCH("/:id/care-plan-items/:itemId", h.RequirePermission(string(model.ResourceHospitalization), "edit"), h.UpdateCarePlanItem)
	rg.DELETE("/:id/care-plan-items/:itemId", h.RequirePermission(string(model.ResourceHospitalization), "delete"), h.DeleteCarePlanItem)
}
