package billing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func estimateItemDiscountKey(name string, sortOrder int) string {
	return fmt.Sprintf("%s\x1f%d", name, sortOrder)
}

// EstimateHandler は EstimateService の HTTP handler。
type EstimateHandler struct {
	svc           EstimateService
	hasPermission httpapi.PermissionChecker
}

// NewEstimateHandler は EstimateHandler を構築する。
func NewEstimateHandler(svc EstimateService, hasPermission httpapi.PermissionChecker) *EstimateHandler {
	return &EstimateHandler{svc: svc, hasPermission: hasPermission}
}

// ListEstimates godoc
func (h *EstimateHandler) ListEstimates(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceEstimates), "view") {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters, err := newListEstimateQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	estimates, total, err := h.svc.List(c.Request.Context(), clinicID, filters.OwnerID, filters.MedicalRecordID, filters.Status, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(estimates, toEstimateResponse), total, page, limit))
}

// GetEstimate godoc
func (h *EstimateHandler) GetEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceEstimates), "view") {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	estimate, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// CreateEstimate godoc
func (h *EstimateHandler) CreateEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req createEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// PR665: 作成時は既存行と一致させる相手が存在しないため、一致ヒントは全面的に拒否する。
	for i, item := range req.Items {
		if item.ExistingItemID != nil {
			httpapi.RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].existing_item_id is not allowed when creating an estimate", i)))
			return
		}
	}
	// BUG-372: 明細の discount_amount / discount_rate にゼロ以外を指定する場合は権限要（ヘッダー割引は 0 以外を service が 422 で拒否する）
	for _, item := range req.Items {
		if err := httpapi.RequireDiscountCreateInt(c, h.hasPermission, item.DiscountAmount); err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := httpapi.RequireDiscountCreateExact(c, h.hasPermission, item.DiscountRate.orZero()); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Create(ctx, clinicID, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/estimates/%d", estimate.ID))
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}

// UpdateEstimate godoc
func (h *EstimateHandler) UpdateEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount_amount を変更する場合は既存値と比較し権限チェック
	prepared, err := h.prepareUpdateEstimateRequest(c, clinicID, id, &req)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	estimate, err := h.svc.Update(ctx, clinicID, id, prepared.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toEstimateResponse(estimate))
}

// DeleteEstimate godoc
func (h *EstimateHandler) DeleteEstimate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := staffID
	if err := h.svc.Delete(c.Request.Context(), clinicID, id, &actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateEstimateSuccessor godoc
// POST /api/v1/estimates/:id/successors — 確定見積の後継ドラフト作成（TASK-012 FINAL B）。
// 権限: estimates:create。原見積は不変。unlock は存在しない。
func (h *EstimateHandler) CreateEstimateSuccessor(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createEstimateSuccessorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	estimate, err := h.svc.CreateSuccessor(c.Request.Context(), clinicID, id, req.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/estimates/%d", estimate.ID))
	c.JSON(http.StatusCreated, toEstimateResponse(estimate))
}

func (h *EstimateHandler) prepareUpdateEstimateRequest(
	c *gin.Context,
	clinicID, id uint64,
	req *updateEstimateRequest,
) (*updateEstimateRequest, error) {
	prepared := *req
	if req.Items == nil {
		return &prepared, nil
	}
	existing, err := h.svc.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		return nil, err
	}
	resolved, err := h.assertEstimateItemDiscountPermissions(c, existing.Items, *req.Items)
	if err != nil {
		return nil, err
	}
	// 解決済みの明細は新しいリクエストへ載せ、元のリクエストを保持する。
	prepared.Items = &resolved
	return &prepared, nil
}

func (h *EstimateHandler) assertEstimateItemDiscountPermissions(
	c *gin.Context,
	existingItems []model.EstimateItem,
	items []createEstimateItemRequest,
) ([]createEstimateItemRequest, error) {
	matched, err := resolveEstimateItemMatches(existingItems, items)
	if err != nil {
		return nil, err
	}
	for i, item := range items {
		amt := item.DiscountAmount
		rate := item.DiscountRate.orZero()
		if old, ok := matched[i]; ok {
			if err := h.requireItemDiscountAmountEdit(c, amt, old.DiscountAmount); err != nil {
				return nil, err
			}
			// 既存の割引率が NULL（不明）なら ExactNumeric{} のままにする。Equal は NULL と値を不等とするため権限を要求する。
			oldRate := model.ExactNumeric{}
			if old.DiscountRate != nil {
				oldRate = *old.DiscountRate
			}
			if err := httpapi.RequireDiscountEditExact(c, h.hasPermission, &rate, oldRate); err != nil {
				return nil, err
			}
			continue
		}
		if err := httpapi.RequireDiscountCreateInt(c, h.hasPermission, amt); err != nil {
			return nil, err
		}
		if err := httpapi.RequireDiscountCreateExact(c, h.hasPermission, rate); err != nil {
			return nil, err
		}
	}
	resolved := make([]createEstimateItemRequest, len(items))
	copy(resolved, items)
	for i := range resolved {
		if old, ok := matched[i]; ok && resolved[i].ExistingItemID == nil {
			id := old.ID
			resolved[i].ExistingItemID = &id
		}
	}
	return resolved, nil
}

// requireItemDiscountAmountEdit は既存明細の割引額の変更に編集権限を要求する。
// 既存の割引額が NULL（不明）のときは 0 とみなさず、値にかかわらず編集権限を要求する（fail-closed、EMR-508 / Q12）。
func (h *EstimateHandler) requireItemDiscountAmountEdit(c *gin.Context, amount int64, old *int64) error {
	if old != nil {
		return httpapi.RequireDiscountEditInt(c, h.hasPermission, &amount, *old)
	}
	if h.hasPermission(c, string(model.ResourceDiscount), "edit") {
		return nil
	}
	return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
}

// resolveEstimateItemMatches は任意の existing_item_id ヒントを「この見積自身の行」と
// 1:1 で解決する純粋関数（入力を書き換えない）。解決結果はリクエスト行インデックス → 既存行のマップ。
// 未知 ID・重複 ID は fail-closed の 400。ヒントの無い行は従来どおり (name, sort_order)
// のレガシーキーで、未解決の既存行のみを対象に 1:1 でフォールバックする。
// 同一キーが未解決の複数既存行に重複する場合は曖昧として fail-closed の 400。
// 1 つの既存割引行が複数行の割引を正当化しないよう、どちらの経路でも一致は 1:1 に限定する。
func resolveEstimateItemMatches(
	existingItems []model.EstimateItem,
	items []createEstimateItemRequest,
) (map[int]model.EstimateItem, error) {
	byID := make(map[uint64]int, len(existingItems))
	for i := range existingItems {
		byID[existingItems[i].ID] = i
	}
	matched := make(map[int]model.EstimateItem, len(items))
	usedExisting := make(map[int]bool, len(existingItems))
	seenHint := make(map[uint64]int, len(items))
	for i, item := range items {
		if item.ExistingItemID == nil {
			continue
		}
		if prev, dup := seenHint[*item.ExistingItemID]; dup {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].existing_item_id %d duplicates items[%d]", i, *item.ExistingItemID, prev))
		}
		seenHint[*item.ExistingItemID] = i
		idx, ok := byID[*item.ExistingItemID]
		if !ok || usedExisting[idx] {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d].existing_item_id %d does not match an item of this estimate", i, *item.ExistingItemID))
		}
		usedExisting[idx] = true
		matched[i] = existingItems[idx]
	}
	claimedLegacy := make(map[int]bool)
	for i := range items {
		if items[i].ExistingItemID != nil {
			continue
		}
		key := estimateItemDiscountKey(items[i].Name, items[i].SortOrder)
		match := -1
		ambiguous := false
		for j := range existingItems {
			if usedExisting[j] || claimedLegacy[j] {
				continue
			}
			if existingItems[j].SortOrder == nil {
				// 並び順が NULL（不明）の既存行は legacy key で突き合わせない（新規行として扱い、権限を要求する）。
				continue
			}
			if estimateItemDiscountKey(existingItems[j].Name, *existingItems[j].SortOrder) == key {
				if match == -1 {
					match = j
				} else {
					ambiguous = true
				}
			}
		}
		if ambiguous {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("items[%d] matches multiple existing rows by (name, sort_order); specify existing_item_id", i))
		}
		if match != -1 {
			claimedLegacy[match] = true
			matched[i] = existingItems[match]
		}
	}
	return matched, nil
}
