package handler

import (
	"net/http"
	"strconv"

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

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
		ExpiryDate:    input.ExpiryDate,
		Supplier:      input.Supplier,
		LastRestocked: input.LastRestocked,
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateInventoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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

	svcInput := service.UpdateInventoryInput{
		Name:          input.Name,
		Category:      category,
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		MinStockLevel: input.MinStockLevel,
		Location:      input.Location,
		ExpiryDate:    input.ExpiryDate,
		Supplier:      input.Supplier,
		LastRestocked: input.LastRestocked,
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Inventory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterInventoryRoutes は在庫関連のルートを登録する
func (h *Handler) RegisterInventoryRoutes(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.ListInventory)
	inventory.POST("", h.CreateInventory)
	inventory.GET("/:id", h.GetInventory)
	inventory.PATCH("/:id", h.UpdateInventory)
	inventory.DELETE("/:id", h.DeleteInventory)
}
