package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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

	filters := newListInventoryQuery(c.Request.URL.Query()).toServiceFilters()

	items, total, err := h.svc.Inventory.List(c.Request.Context(), clinicID, filters.Category, filters.Status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(items, toInventoryResponse), total, page, limit))
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
	c.JSON(http.StatusOK, toInventoryResponse(item))
}

// CreateInventory godoc
func (h *Handler) CreateInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	created, err := h.svc.Inventory.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/inventory/%d", created.ID))
	c.JSON(http.StatusCreated, toInventoryResponse(created))
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
	var req updateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	item, err := h.svc.Inventory.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInventoryResponse(item))
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
