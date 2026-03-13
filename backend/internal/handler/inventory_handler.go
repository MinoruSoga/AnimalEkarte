package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createInventoryInput struct {
	Name          string     `json:"name"            binding:"required"`
	Category      string     `json:"category"        binding:"required"`
	Quantity      int        `json:"quantity"`
	Unit          string     `json:"unit"            binding:"required"`
	MinStockLevel int        `json:"min_stock_level"`
	Location      string     `json:"location"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Supplier      string     `json:"supplier"`
	LastRestocked *time.Time `json:"last_restocked"`
	Status        string     `json:"status"`
}

type updateInventoryInput struct {
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Quantity      int        `json:"quantity"`
	Unit          string     `json:"unit"`
	MinStockLevel int        `json:"min_stock_level"`
	Location      string     `json:"location"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	Supplier      string     `json:"supplier"`
	LastRestocked *time.Time `json:"last_restocked"`
	Status        string     `json:"status"`
}

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
	c.JSON(http.StatusOK, PaginatedResponse{Data: items, Total: total, Page: page, Limit: limit})
}

// GetInventory godoc
func (h *Handler) GetInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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

	var input createInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &model.InventoryItem{
		ID:            id,
		ClinicID:      clinicID,
		Name:          input.Name,
		Quantity:      input.Quantity,
		Unit:          input.Unit,
		MinStockLevel: input.MinStockLevel,
		Location:      input.Location,
		ExpiryDate:    input.ExpiryDate,
		Supplier:      input.Supplier,
		LastRestocked: input.LastRestocked,
	}
	if input.Category != "" {
		item.Category = model.InventoryCategory(input.Category)
	}
	if input.Status != "" {
		item.Status = model.InventoryStatus(input.Status)
	}

	if err := h.svc.Inventory.Update(c.Request.Context(), clinicID, item); err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Inventory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
