package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// errHandlerAlreadyResponded は ExtractStaffID 等が既にレスポンスを書いたことを示す。
// 呼び出し元は RespondError せず return する。
var errHandlerAlreadyResponded = errors.New("handler already responded")

// BillingItemHandler は BillingItemService の HTTP handler。
type BillingItemHandler struct {
	svc               BillingItemService
	cashRegister      cashRegisterCloseChecker
	hasPermission     httpapi.PermissionChecker
	requirePermission PermissionMiddleware
}

// NewBillingItemHandler は BillingItemHandler を構築する。
// cashRegister / hasPermission は締め後編集ゲート（#115 / BUG-463）用。テストでは nil 可
// （nil の場合は締め済み判定をスキップし IsPostClose=false として扱う）。
func NewBillingItemHandler(
	svc BillingItemService,
	cashRegister cashRegisterCloseChecker,
	hasPermission httpapi.PermissionChecker,
	requirePermission PermissionMiddleware,
) *BillingItemHandler {
	return &BillingItemHandler{
		svc:               svc,
		cashRegister:      cashRegister,
		hasPermission:     hasPermission,
		requirePermission: requirePermission,
	}
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
	input.StaffID = input.CreatedBy

	// #115 / BUG-463: レジ締め済み期間の明細登録は post-close 権限 + 理由必須
	if err := h.applyPostCloseFlags(c, clinicID, req.BillingID, 0, &input.IsPostClose, &input.PostCloseReason, &input.StaffID, req.PostCloseReason); err != nil {
		if !errors.Is(err, errHandlerAlreadyResponded) {
			httpapi.RespondError(c, err)
		}
		return
	}

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
	input.StaffID = httpapi.OptionalStaffID(c)

	// #115 / BUG-463: レジ締め済み期間の明細更新は post-close 権限 + 理由必須
	if err := h.applyPostCloseFlags(c, clinicID, 0, id, &input.IsPostClose, &input.PostCloseReason, &input.StaffID, req.PostCloseReason); err != nil {
		if !errors.Is(err, errHandlerAlreadyResponded) {
			httpapi.RespondError(c, err)
		}
		return
	}

	item, err := h.svc.UpdateItem(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToBillingItemResponse(item))
}

// applyPostCloseFlags は親請求の予定日がレジ締め済みかを判定し、権限チェックと IsPostClose 注入を行う。
// billingID > 0 のとき create 経路、itemID > 0 のとき update/delete 経路。
// cashRegister 未配線時は判定をスキップ（テスト互換）。
func (h *BillingItemHandler) applyPostCloseFlags(
	c *gin.Context,
	clinicID, billingID, itemID uint64,
	isPostClose *bool,
	postCloseReason **string,
	staffID **uint64,
	requestReason *string,
) error {
	*postCloseReason = requestReason
	if h.cashRegister == nil {
		return nil
	}

	ctx := c.Request.Context()
	var billing *model.Billing
	var err error
	switch {
	case billingID > 0:
		billing, err = h.svc.GetBilling(ctx, clinicID, billingID)
	case itemID > 0:
		billing, err = h.svc.GetBillingForItem(ctx, clinicID, itemID)
	default:
		return apperrors.WrapInternalServerError("post-close resolution requires billing or item id")
	}
	if err != nil {
		return err
	}
	if billing == nil {
		return apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}

	// BUG-009: 確定済み会計の明細更新は post-close-edit 権限 + 修正理由必須（締め有無に依存しない）。
	// create/delete 経路は従来どおり status guard（create）/ 締め後フラグ（delete）に委ねる。
	if itemID > 0 && billing.Status == model.BillingStatusCompleted {
		return h.applyCompletedItemPostCloseFlags(c, ctx, clinicID, billing, staffID, requestReason, postCloseReason, isPostClose)
	}

	closed, err := h.cashRegister.IsDateClosed(ctx, clinicID, billing.ScheduledDate)
	if err != nil {
		return err
	}
	if !closed {
		return nil
	}

	if h.hasPermission == nil || !h.hasPermission(c, string(model.ResourceAccountingPostCloseEdit), "edit") {
		return apperrors.WrapForbidden("レジ締め済み期間の会計編集には accounting-post-close-edit:edit 権限が必要です")
	}
	// 監査 actor 用に staff を必須化（Optional で取れない場合は Extract で応答済み）
	if staffID == nil || *staffID == nil {
		id, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return errHandlerAlreadyResponded
		}
		*staffID = &id
	}
	*isPostClose = true
	return nil
}

func (h *BillingItemHandler) applyCompletedItemPostCloseFlags(
	c *gin.Context,
	ctx context.Context,
	clinicID uint64,
	billing *model.Billing,
	staffID **uint64,
	requestReason *string,
	postCloseReason **string,
	isPostClose *bool,
) error {
	if h.hasPermission == nil || !h.hasPermission(c, string(model.ResourceAccountingPostCloseEdit), "edit") {
		return apperrors.WrapForbidden("確定済み会計の明細修正には accounting-post-close-edit:edit 権限が必要です")
	}
	if staffID == nil || *staffID == nil {
		id, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return errHandlerAlreadyResponded
		}
		*staffID = &id
	}
	if requestReason == nil || strings.TrimSpace(*requestReason) == "" {
		return apperrors.WrapInvalidInput("確定済み会計の明細修正には修正理由（post_close_reason）が必要です")
	}
	*postCloseReason = requestReason
	if h.cashRegister == nil {
		return nil
	}
	closed, err := h.cashRegister.IsDateClosed(ctx, clinicID, billing.ScheduledDate)
	if err != nil {
		return err
	}
	if closed {
		*isPostClose = true
	}
	return nil
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

	// Optional body: post_close_reason（締め後削除時）。空 body は非締め後として扱う。
	var req deleteBillingItemRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
			return
		}
	}

	// BUG-440: vaccination claim 解放監査の actor。認証経路では通常設定済み。
	// BUG-463 residual: レジ締め済みなら post-close 権限 + 理由必須を注入。
	input := &DeleteBillingItemInput{StaffID: httpapi.OptionalStaffID(c)}
	if err := h.applyPostCloseFlags(c, clinicID, 0, id, &input.IsPostClose, &input.PostCloseReason, &input.StaffID, req.PostCloseReason); err != nil {
		if !errors.Is(err, errHandlerAlreadyResponded) {
			httpapi.RespondError(c, err)
		}
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), clinicID, id, input); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetUnbilledItems godoc — legacy raw-array contract (BUG-013: kept for compatibility).
func (h *BillingItemHandler) GetUnbilledItems(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
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

// unbilledDetailsResponse は BUG-013 additive envelope。
// warnings は {source, code, count, blocking} のみ（内部情報を載せない）。
type unbilledDetailsResponse struct {
	Items    []BillingItemResponse `json:"items"`
	Warnings []UnbilledWarning     `json:"warnings"`
}

// GetUnbilledItemDetails godoc
// GET /billing-items/unbilled-details?pet_id=
func (h *BillingItemHandler) GetUnbilledItemDetails(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
		return
	}

	petID, err := newUnbilledItemsQuery(c.Request.URL.Query()).toPetID()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	details, err := h.svc.GetUnbilledItemDetails(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if details == nil {
		c.JSON(http.StatusOK, unbilledDetailsResponse{Items: []BillingItemResponse{}, Warnings: []UnbilledWarning{}})
		return
	}
	warnings := details.Warnings
	if warnings == nil {
		warnings = []UnbilledWarning{}
	}
	c.JSON(http.StatusOK, unbilledDetailsResponse{
		Items:    httpapi.MapSlice(details.Items, ToBillingItemResponse),
		Warnings: warnings,
	})
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
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
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
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
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
