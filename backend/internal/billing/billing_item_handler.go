package billing

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// BillingItemHandler は BillingItemService の HTTP handler。
type BillingItemHandler struct {
	svc               BillingItemService
	requirePermission PermissionMiddleware
}

// NewBillingItemHandler は BillingItemHandler を構築する。
func NewBillingItemHandler(svc BillingItemService, requirePermission PermissionMiddleware) *BillingItemHandler {
	return &BillingItemHandler{svc: svc, requirePermission: requirePermission}
}

// CreateBillingItem godoc
func (h *BillingItemHandler) CreateBillingItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput(clinicID)
	input.CreatedBy = httpapi.OptionalStaffID(c)
	item, err := h.svc.CreateItem(c.Request.Context(), input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/billing/%d/items/%d", item.BillingID, item.ID))
	c.JSON(http.StatusCreated, ToBillingItemResponse(item))
}

// UpdateBillingItem godoc
func (h *BillingItemHandler) UpdateBillingItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateBillingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	item, err := h.svc.UpdateItem(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToBillingItemResponse(item))
}

// DeleteBillingItem godoc
func (h *BillingItemHandler) DeleteBillingItem(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetUnbilledItems godoc
func (h *BillingItemHandler) GetUnbilledItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	petID, err := newUnbilledItemsQuery(c.Request.URL.Query()).toPetID()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	items, err := h.svc.GetUnbilledItems(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, ToBillingItemResponse))
}

// ungroupedSameDayResponse は #77 取り残し警告用レスポンス。
type ungroupedSameDayResponse struct {
	MedicalRecordCount int64 `json:"medical_record_count"`
	TrimmingCount      int64 `json:"trimming_count"`
	HasUngrouped       bool  `json:"has_ungrouped"`
}

func toUngroupedSameDayResponse(s UngroupedSameDaySummary) ungroupedSameDayResponse {
	return ungroupedSameDayResponse{
		MedicalRecordCount: s.MedicalRecordCount,
		TrimmingCount:      s.TrimmingCount,
		HasUngrouped:       s.MedicalRecordCount > 0 || s.TrimmingCount > 0,
	}
}

type discountSuggestionsResponse struct {
	Suggestions []DiscountSuggestion `json:"suggestions"`
}

func toDiscountSuggestionsResponse(suggestions []DiscountSuggestion) discountSuggestionsResponse {
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
func (h *BillingItemHandler) GetUngroupedSameDay(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, err := newUnbilledItemsQuery(c.Request.URL.Query()).toPetID()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	date := parseUngroupedDate(c.Query("date"))
	summary, err := h.svc.GetUngroupedSameDaySummary(c.Request.Context(), clinicID, petID, date)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUngroupedSameDayResponse(summary))
}

// GetBillingItemDiscountSuggestions は指定明細に適用可能な割引候補を返す (#81 Q-I スタッフ選択)。
func (h *BillingItemHandler) GetBillingItemDiscountSuggestions(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	itemID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	suggestions, err := h.svc.GetDiscountSuggestions(c.Request.Context(), clinicID, itemID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDiscountSuggestionsResponse(suggestions))
}
