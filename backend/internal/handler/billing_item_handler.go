package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// CreateBillingItem godoc
func (h *Handler) CreateBillingItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	item, err := h.svc.BillingItem.CreateItem(c.Request.Context(), &service.CreateBillingItemInput{
		ClinicID:              clinicID,
		BillingID:             req.BillingID,
		Category:              req.Category,
		Name:                  req.Name,
		UnitPrice:             req.UnitPrice,
		Quantity:              req.Quantity,
		TaxType:               req.TaxType,
		TaxRate:               req.TaxRate,
		IsInsuranceApplicable: req.IsInsuranceApplicable,
		Source:                req.Source,
		SortOrder:             req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/billing/%d/items/%d", item.BillingID, item.ID))
	c.JSON(http.StatusCreated, toBillingItemResponse(item))
}

// UpdateBillingItem godoc
func (h *Handler) UpdateBillingItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	var taxType *model.TaxType
	if req.TaxType != nil {
		if err := validateTaxType(*req.TaxType); err != nil {
			RespondError(c, err)
			return
		}
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

	item, err := h.svc.BillingItem.UpdateItem(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingItemResponse(item))
}

// DeleteBillingItem godoc
func (h *Handler) DeleteBillingItem(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.BillingItem.DeleteItem(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterBillingItemRoutes は明細関連のルートを登録する
func (h *Handler) RegisterBillingItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/billing-items")
	items.POST("", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateBillingItem)
	items.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateBillingItem)
	items.DELETE("/:id", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.DeleteBillingItem)
}
