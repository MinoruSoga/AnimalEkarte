package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// CreateBillingItem godoc
func (h *Handler) CreateBillingItem(c *gin.Context) {
	var req createBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	taxType := model.TaxTypeExcluded
	if req.TaxType != "" {
		taxType = model.TaxType(req.TaxType)
	}
	taxRate := 0.10
	if req.TaxRate > 0 {
		taxRate = req.TaxRate
	}
	source := model.ItemSourceManual
	if req.Source != "" {
		source = model.ItemSource(req.Source)
	}

	input := &service.CreateBillingItemInput{
		BillingID:             req.BillingID,
		Category:              model.ItemCategory(req.Category),
		Name:                  req.Name,
		UnitPrice:             req.UnitPrice,
		Quantity:              req.Quantity,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: req.IsInsuranceApplicable,
		Source:                source,
		SortOrder:             req.SortOrder,
	}

	item, err := h.svc.BillingItem.CreateItem(c.Request.Context(), input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toBillingItemResponse(item))
}

// UpdateBillingItem godoc
func (h *Handler) UpdateBillingItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req updateBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var taxType *model.TaxType
	if req.TaxType != nil {
		t := model.TaxType(*req.TaxType)
		taxType = &t
	}

	input := &service.UpdateBillingItemInput{
		UnitPrice:             req.UnitPrice,
		Quantity:              req.Quantity,
		TaxType:               taxType,
		TaxRate:               req.TaxRate,
		IsInsuranceApplicable: req.IsInsuranceApplicable,
	}

	item, err := h.svc.BillingItem.UpdateItem(c.Request.Context(), id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingItemResponse(item))
}

// DeleteBillingItem godoc
func (h *Handler) DeleteBillingItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	if err := h.svc.BillingItem.DeleteItem(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterBillingItemRoutes は明細関連のルートを登録する
func (h *Handler) RegisterBillingItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/billing-items")
	items.POST("", h.CreateBillingItem)
	items.PATCH("/:id", h.UpdateBillingItem)
	items.DELETE("/:id", h.DeleteBillingItem)
}
