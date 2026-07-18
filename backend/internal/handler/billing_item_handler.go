package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

	items, err := h.svc.BillingItem.GetUnbilledItems(c.Request.Context(), clinicID, petID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(items, toBillingItemResponse))
}

// ungroupedSameDayResponse は #77 取り残し警告用レスポンス。
type ungroupedSameDayResponse struct {
	MedicalRecordCount int64 `json:"medical_record_count"`
	TrimmingCount      int64 `json:"trimming_count"`
	HasUngrouped       bool  `json:"has_ungrouped"`
}

func toUngroupedSameDayResponse(s service.UngroupedSameDaySummary) ungroupedSameDayResponse {
	return ungroupedSameDayResponse{
		MedicalRecordCount: s.MedicalRecordCount,
		TrimmingCount:      s.TrimmingCount,
		HasUngrouped:       s.MedicalRecordCount > 0 || s.TrimmingCount > 0,
	}
}

type discountSuggestionsResponse struct {
	Suggestions []service.DiscountSuggestion `json:"suggestions"`
}

func toDiscountSuggestionsResponse(suggestions []service.DiscountSuggestion) discountSuggestionsResponse {
	return discountSuggestionsResponse{Suggestions: suggestions}
}

func parseUngroupedDate(s string) time.Time {
	if s == "" {
		return time.Now().In(time.Local)
	}
	t, err := time.ParseInLocation(time.DateOnly, s, time.Local)
	if err != nil {
		return time.Now().In(time.Local)
	}
	return t
}

// GetUngroupedSameDay は同日同ペットの未会計対象化項目件数を返す(#77 取り残し警告)。
func (h *Handler) GetUngroupedSameDay(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := newUnbilledItemsQuery(c.Request.URL.Query()).toPetID()
	if err != nil {
		RespondError(c, err)
		return
	}
	date := parseUngroupedDate(c.Query("date"))
	summary, err := h.svc.BillingItem.GetUngroupedSameDaySummary(c.Request.Context(), clinicID, petID, date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUngroupedSameDayResponse(summary))
}

// GetBillingItemDiscountSuggestions は指定明細に適用可能な割引候補を返す (#81 Q-I スタッフ選択)。
func (h *Handler) GetBillingItemDiscountSuggestions(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	itemID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	suggestions, err := h.svc.BillingItem.GetDiscountSuggestions(c.Request.Context(), clinicID, itemID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiscountSuggestionsResponse(suggestions))
}

// RegisterBillingItemRoutes は明細関連のルートを登録する。
func (h *Handler) RegisterBillingItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/billing-items")
	items.GET("/unbilled", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetUnbilledItems)
	items.GET("/ungrouped-same-day", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetUngroupedSameDay)
	items.POST("", h.RequirePermission(string(model.ResourceAccounting), "create"), h.CreateBillingItem)
	items.PATCH("/:id", h.RequirePermission(string(model.ResourceAccounting), "edit"), h.UpdateBillingItem)
	items.DELETE("/:id", h.RequirePermission(string(model.ResourceAccounting), "delete"), h.DeleteBillingItem)
	items.GET("/:id/discount-suggestions", h.RequirePermission(string(model.ResourceAccounting), "view"), h.GetBillingItemDiscountSuggestions)
}
