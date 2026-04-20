// Package handler provides HTTP handler implementations for MerchandiseItem entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListMerchandiseItems godoc
// マスタ API は全件取得（ページネーションなし）で配列を直接返す。
// 他のマスタ API（cages, staffs 等）と同じパターン。
func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	category := c.Query("category") // optional category filter

	items, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, category)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toMerchandiseItemResponse))
}

// GetMerchandiseItem godoc
func (h *Handler) GetMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.CreateMerchandiseItemInput{
		Name:      req.Name,
		Category:  req.Category,
		UnitPrice: req.UnitPrice,
		TaxType:   req.TaxType,
		TaxRate:   req.TaxRate, // *float64: nil → service側でデフォルト 0.10 を適用
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	item, err := h.svc.MerchandiseItem.Create(c.Request.Context(), clinicID, input)
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
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateMerchandiseItemInput{
		Name:      req.Name,
		Category:  req.Category,
		UnitPrice: req.UnitPrice,
		TaxType:   req.TaxType,
		TaxRate:   req.TaxRate,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	item, err := h.svc.MerchandiseItem.Update(c.Request.Context(), clinicID, id, input)
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
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.MerchandiseItem.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteMerchandiseItem godoc
func (h *Handler) DeleteMerchandiseItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.MerchandiseItem.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
