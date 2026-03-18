// Package handler provides HTTP handler implementations for MerchandiseItem entity.
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// ListMerchandiseItems godoc
func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	category := c.Query("category") // optional category filter

	items, total, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, page, limit, category)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toMerchandiseItemResponseList(items), total, page, limit))
}

// GetMerchandiseItem godoc
func (h *Handler) GetMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.svc.MerchandiseItem.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMerchandiseItemResponse(item))
}

// CreateMerchandiseItem godoc
func (h *Handler) CreateMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := service.CreateMerchandiseItemInput{
		Name:      req.Name,
		Category:  req.Category,
		UnitPrice: req.UnitPrice,
		TaxRate:   req.TaxRate,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	item, err := h.svc.MerchandiseItem.Create(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/merchandise-items/%d", item.ID))
	c.JSON(http.StatusCreated, toMerchandiseItemResponse(item))
}

// UpdateMerchandiseItem godoc
func (h *Handler) UpdateMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := service.UpdateMerchandiseItemInput{
		Name:      req.Name,
		Category:  req.Category,
		UnitPrice: req.UnitPrice,
		TaxRate:   req.TaxRate,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	item, err := h.svc.MerchandiseItem.Update(c.Request.Context(), clinicID, id, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMerchandiseItemResponse(item))
}

// ReorderMerchandiseItems godoc
func (h *Handler) ReorderMerchandiseItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.MerchandiseItem.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// DeleteMerchandiseItem godoc
func (h *Handler) DeleteMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.MerchandiseItem.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
