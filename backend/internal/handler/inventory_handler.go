package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListInventory godoc
func (h *Handler) ListInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var category *string
	if cat := c.Query("category"); cat != "" {
		category = &cat
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	items, total, err := h.svc.Inventory.List(c.Request.Context(), clinicID, category, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(items, total, page, limit))
}

// GetInventory godoc
func (h *Handler) GetInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.Inventory.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateInventory godoc
func (h *Handler) CreateInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createInventoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// Parse dates with fallback: YYYY-MM-DD → RFC3339
	expiryDate, err := parseDate(input.ExpiryDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid expiry_date: %v", err)))
		return
	}
	lastRestocked, err := parseDate(input.LastRestocked)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid last_restocked: %v", err)))
		return
	}

	item := &model.InventoryItem{
		ClinicID:      clinicID,
		Name:          input.Name,
		Category:      model.InventoryCategory(input.Category),
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		MinStockLevel: input.MinStockLevel,
		Location:      input.Location,
		ExpiryDate:    expiryDate,
		Supplier:      input.Supplier,
		LastRestocked: lastRestocked,
	}
	if input.Status != "" {
		item.Status = model.InventoryStatus(input.Status)
	}

	if err := h.svc.Inventory.Create(c.Request.Context(), clinicID, item); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

// UpdateInventory godoc
func (h *Handler) UpdateInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateInventoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var category *model.InventoryCategory
	if input.Category != nil {
		cat := model.InventoryCategory(*input.Category)
		category = &cat
	}
	var status *model.InventoryStatus
	if input.Status != nil {
		s := model.InventoryStatus(*input.Status)
		status = &s
	}

	// Parse dates with fallback: YYYY-MM-DD → RFC3339
	expiryDate, err := parseDate(input.ExpiryDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid expiry_date: %v", err)))
		return
	}
	lastRestocked, err := parseDate(input.LastRestocked)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid last_restocked: %v", err)))
		return
	}

	svcInput := service.UpdateInventoryInput{
		Name:          input.Name,
		Category:      category,
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		MinStockLevel: input.MinStockLevel,
		Location:      input.Location,
		ExpiryDate:    expiryDate,
		Supplier:      input.Supplier,
		LastRestocked: lastRestocked,
		Status:        status,
	}

	item, err := h.svc.Inventory.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Inventory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
