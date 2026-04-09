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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// テナント分離: billing が同一クリニックに属することを確認
	if _, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, req.BillingID); err != nil {
		RespondError(c, err)
		return
	}

	taxType := model.TaxTypeExcluded
	if req.TaxType != "" {
		if err := validateTaxType(req.TaxType); err != nil {
			RespondError(c, err)
			return
		}
		taxType = model.TaxType(req.TaxType)
	}
	taxRate := 0.10
	if req.TaxRate > 0 {
		taxRate = req.TaxRate
	}
	source := model.ItemSourceManual
	if req.Source != "" {
		if err := validateItemSource(req.Source); err != nil {
			RespondError(c, err)
			return
		}
		source = model.ItemSource(req.Source)
	}

	if err := validateItemCategory(req.Category); err != nil {
		RespondError(c, err)
		return
	}

	input := &service.CreateBillingItemInput{
		ClinicID:              clinicID,
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

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

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
