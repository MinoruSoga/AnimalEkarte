package handler

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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

	item, err := h.svc.BillingItem.CreateItem(c.Request.Context(), req.toServiceInput(clinicID))
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

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, err)
		return
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

// GetUnbilledItems godoc
func (h *Handler) GetUnbilledItems(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	petID, err := newUnbilledItemsQuery(c.Request.URL.Query()).toPetID()
	if err != nil {
		RespondError(c, err)
		return
	}

	treatments, err := h.svc.BillingItem.GetUnbilledItems(c.Request.Context(), clinicID, petID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUnbilledItemsResponse(treatments))
}

func toUnbilledItemsResponse(treatments []model.Treatment) []billingItemResponse {
	const defaultTaxRate = 0.10
	items := make([]billingItemResponse, 0, len(treatments))
	for i := range treatments {
		t := &treatments[i]
		subtotal := int64(float64(t.UnitPrice) * t.Quantity)
		taxAmount := int64(math.Round(float64(subtotal) * defaultTaxRate))
		items = append(items, billingItemResponse{
			ID:                    t.ID,
			BillingID:             0,
			Category:              string(treatmentTypeToItemCategory(t.ItemType)),
			Name:                  t.Content,
			UnitPrice:             t.UnitPrice,
			Quantity:              t.Quantity,
			Subtotal:              subtotal,
			TaxType:               string(model.TaxTypeExcluded),
			TaxRate:               defaultTaxRate,
			TaxAmount:             taxAmount,
			IsInsuranceApplicable: t.IsInsurance,
			Source:                string(model.ItemSourceMedicalRecord),
			SortOrder:             t.SortOrder,
		})
	}
	return items
}

func treatmentTypeToItemCategory(t model.TreatmentItemType) model.ItemCategory {
	switch t {
	case model.TreatmentItemTypeConsultation:
		return model.ItemCategoryExamination
	case model.TreatmentItemTypeProcedure:
		return model.ItemCategoryProcedure
	case model.TreatmentItemTypeMedicine:
		return model.ItemCategoryMedicine
	default:
		return model.ItemCategoryOther
	}
}

// RegisterBillingItemRoutes は明細関連のルートを登録する
func (h *Handler) RegisterBillingItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/billing-items")
	items.GET("/unbilled", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetUnbilledItems)
	items.POST("", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateBillingItem)
	items.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateBillingItem)
	items.DELETE("/:id", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.DeleteBillingItem)
}
