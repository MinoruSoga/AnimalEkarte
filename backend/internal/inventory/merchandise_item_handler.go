package inventory

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListMerchandiseItems godoc
// マスタ API は全件取得（ページネーションなし）で配列を直接返す。
// 他のマスタ API（cages, staffs 等）と同じパターン。
func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query := newListMerchandiseItemsQuery(c.Request.URL.Query())
	items, err := h.merchandiseItem.List(c.Request.Context(), clinicID, query.Category)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toMerchandiseItemResponse))
}

// GetMerchandiseItem godoc
func (h *Handler) GetMerchandiseItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	item, err := h.merchandiseItem.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMerchandiseItemResponse(item))
}

// CreateMerchandiseItem godoc
func (h *Handler) CreateMerchandiseItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	item, err := h.merchandiseItem.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/merchandise-items/%d", item.ID))
	c.JSON(http.StatusCreated, toMerchandiseItemResponse(item))
}

// UpdateMerchandiseItem godoc
func (h *Handler) UpdateMerchandiseItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateMerchandiseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	item, err := h.merchandiseItem.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMerchandiseItemResponse(item))
}

// ReorderMerchandiseItems godoc
func (h *Handler) ReorderMerchandiseItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.merchandiseItem.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteMerchandiseItem godoc
func (h *Handler) DeleteMerchandiseItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.merchandiseItem.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
