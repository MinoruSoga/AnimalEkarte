package inventory

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListInventory godoc
func (h *Handler) ListInventory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters := newListInventoryQuery(c.Request.URL.Query()).toServiceFilters()

	items, total, err := h.inventory.List(c.Request.Context(), clinicID, filters.Category, filters.Status, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(items, toInventoryResponse), total, page, limit))
}

// GetInventory godoc
func (h *Handler) GetInventory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.inventory.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInventoryResponse(item))
}

// CreateInventory godoc
func (h *Handler) CreateInventory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	created, err := h.inventory.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/inventory/%d", created.ID))
	c.JSON(http.StatusCreated, toInventoryResponse(created))
}

// UpdateInventory godoc
func (h *Handler) UpdateInventory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	item, err := h.inventory.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toInventoryResponse(item))
}

func (h *Handler) DeleteInventory(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.inventory.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
