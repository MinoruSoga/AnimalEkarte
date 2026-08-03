package billing

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountingHandler は AccountingService の HTTP handler。
type AccountingHandler struct {
	svc           AccountingService
	cashRegister  cashRegisterCloseChecker
	hasPermission httpapi.PermissionChecker
}

// cashRegisterCloseChecker はレジ締め済み判定（cash_register=B⑤未移行 service）の最小view。
type cashRegisterCloseChecker interface {
	IsDateClosed(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

// NewAccountingHandler は AccountingHandler を構築する。
func NewAccountingHandler(svc AccountingService, cashRegister cashRegisterCloseChecker, hasPermission httpapi.PermissionChecker) *AccountingHandler {
	return &AccountingHandler{svc: svc, cashRegister: cashRegister, hasPermission: hasPermission}
}

// ListAccountings godoc
func (h *AccountingHandler) ListAccountings(c *gin.Context) {
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	q := newListAccountingQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var accountings []model.Billing
	var total int64
	if len(clinicIDs) == 1 {
		accountings, total, err = h.svc.List(
			c.Request.Context(),
			clinicIDs[0],
			filters.PetID,
			filters.OwnerID,
			filters.Status,
			filters.StartDate,
			filters.EndDate,
			page,
			limit,
		)
	} else {
		accountings, total, err = h.svc.ListForClinics(
			c.Request.Context(),
			clinicIDs,
			filters.PetID,
			filters.OwnerID,
			filters.Status,
			filters.StartDate,
			filters.EndDate,
			page,
			limit,
		)
	}
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(accountings, toAccountingResponse), total, page, limit))
}

// GetAccounting godoc
func (h *AccountingHandler) GetAccounting(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := httpapi.ResolveAllClinicIDs(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	accounting, err := h.svc.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(accounting))
}

// CreateAccounting godoc
func (h *AccountingHandler) CreateAccounting(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var input createAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	created, err := h.svc.Create(ctx, input.toServiceInput(clinicID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/accountings/%d", created.ID))
	c.JSON(http.StatusCreated, toAccountingResponse(created))
}

// CompleteAccounting は BUG-018: header/items/payments を単一 command で原子的に確定する。
// Idempotency-Key (UUID) 必須。初回 201、同一 digest retry 200、異 digest 409。
func (h *AccountingHandler) CompleteAccounting(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if err := ValidateIdempotencyKeyUUID(idempotencyKey); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var input completeAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: complete でも header/item 割引の非ゼロは discount:edit を要求（Update 経路と同等）。
	if input.DiscountAmount != nil && *input.DiscountAmount != 0 {
		zero := int64(0)
		if err := httpapi.RequireDiscountEditInt(c, h.hasPermission, input.DiscountAmount, zero); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}
	for i := range input.Items {
		if input.Items[i].DiscountAmount != 0 {
			amt := input.Items[i].DiscountAmount
			zero := int64(0)
			if err := httpapi.RequireDiscountEditInt(c, h.hasPermission, &amt, zero); err != nil {
				httpapi.RespondError(c, err)
				return
			}
		}
	}

	ctx := c.Request.Context()
	serviceInput := input.toServiceInput(clinicID, staffID, idempotencyKey)

	// Close 境界の候補 read（service が write 時に再評価）。
	isClosed, err := h.cashRegister.IsDateClosed(ctx, clinicID, serviceInput.ScheduledDate)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if isClosed {
		if !h.hasPermission(c, string(model.ResourceAccountingPostCloseEdit), "edit") {
			httpapi.RespondError(c, apperrors.WrapForbidden("レジ締め済み期間の会計編集には accounting-post-close-edit:edit 権限が必要です"))
			return
		}
		serviceInput.IsPostClose = true
	}

	result, err := h.svc.Complete(ctx, serviceInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if result == nil || result.Accounting == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("complete accounting returned empty result"))
		return
	}
	if result.Created {
		c.Header("Location", fmt.Sprintf("/v1/accountings/%d", result.Accounting.ID))
		c.JSON(http.StatusCreated, toAccountingResponse(result.Accounting))
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(result.Accounting))
}

// UpdateAccounting godoc
func (h *AccountingHandler) UpdateAccounting(c *gin.Context) {
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
	var input updateAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount_amount にゼロ以外を指定する場合は discount:edit 権限要
	// Payment 既存値取得用リポジトリが未整備のため、ゼロ値は通過・非ゼロのみ権限チェック。
	// ゼロ値再送（通常操作）を阻害せず、非ゼロ（値引意図）のみ保護する。
	if input.DiscountAmount != nil && *input.DiscountAmount != 0 {
		zero := int64(0)
		if err := httpapi.RequireDiscountEditInt(c, h.hasPermission, input.DiscountAmount, zero); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	ctx := c.Request.Context()

	// #115: 締め後編集チェック
	existing, err := h.svc.GetByID(ctx, clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	isClosed, err := h.cashRegister.IsDateClosed(ctx, clinicID, existing.ScheduledDate)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// 認可（締め後編集権限）はリクエストスコープの関心事のため handler に残す。
	// post_close_reason の必須検証は service 層（accountingService.Update 冒頭）へ移設し、
	// handler を迂回する呼び出し元にも不変条件を強制する（B4）。
	if isClosed {
		if !h.hasPermission(c, string(model.ResourceAccountingPostCloseEdit), "edit") {
			httpapi.RespondError(c, apperrors.WrapForbidden("レジ締め済み期間の会計編集には accounting-post-close-edit:edit 権限が必要です"))
			return
		}
	}

	serviceInput := input.toServiceInput(id, clinicID, staffID)
	serviceInput.IsPostClose = isClosed

	updated, err := h.svc.Update(ctx, serviceInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(updated))
}

// CorrectCreditPayment は確定済み会計のクレジット（カード）金額を確定後に訂正する（#189）。
// ルート側で ResourceAccountingPostCloseEdit:edit 権限を要求する（確定/締め済み訂正の既存ゲートを再利用）。
// 確定前編集の通常 PATCH 経路とは別の専用フローで、理由・監査を伴う。
func (h *AccountingHandler) CorrectCreditPayment(c *gin.Context) {
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

	var req correctCreditPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	ctx := c.Request.Context()

	// bug.md M-2（#189 follow-up）: 訂正対象会計の予定日が締め済み期間かを PATCH 経路（UpdateAccounting）と同じ判定で解決する。
	// 拒否はしない（ルートで post-close-edit 権限を既に要求）が、締め済み売上への訂正を service 側で
	// 監査ログ・WarnContext に可視化させるため、判定結果を IsPostClose として注入する。
	existing, err := h.svc.GetByID(ctx, clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	isClosed, err := h.cashRegister.IsDateClosed(ctx, clinicID, existing.ScheduledDate)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	serviceInput := req.toServiceInput(id, clinicID, staffID)
	serviceInput.IsPostClose = isClosed

	updated, err := h.svc.CorrectCreditPayment(ctx, serviceInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(updated))
}

// ListUnpaidBillings は未納者一覧を返す。#120
// GET /v1/accountings/unpaid?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&group_by=owner|billing&page=N&limit=N
// start_date / end_date は必須。欠けた場合は 400 を返す。
func (h *AccountingHandler) ListUnpaidBillings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	filters, err := newListUnpaidBillingsQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()

	switch filters.GroupBy {
	case "billing":
		accountings, total, err := h.svc.ListUnpaidByBilling(ctx, clinicID, filters.StartDate, filters.EndDate, page, limit)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(accountings, toAccountingResponse), total, page, limit))
	case "owner":
		aggregates, total, summary, err := h.svc.ListUnpaidByOwner(ctx, clinicID, filters.StartDate, filters.EndDate, page, limit)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, toUnpaidByOwnerResponse(aggregates, total, page, limit, summary))
	default:
		httpapi.RespondError(c, apperrors.WrapInvalidInput("group_by must be owner or billing"))
	}
}

// GetOwnerUnpaidBalance は会計画面表示用に飼主の未納残高を返す。#182
// GET /v1/accountings/unpaid-balance?owner_id=N
func (h *AccountingHandler) GetOwnerUnpaidBalance(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, err := strconv.ParseUint(c.Query("owner_id"), 10, 64)
	if err != nil || ownerID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("owner_id is required"))
		return
	}
	result, err := h.svc.GetOwnerUnpaidBalance(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerUnpaidBalanceResponse(result))
}

// GetUnpaidMonthlySummary は月次未納繰越集計を返す。#114
// GET /v1/accountings/unpaid-monthly?year=YYYY&month=MM&page=N&limit=N
func (h *AccountingHandler) GetUnpaidMonthlySummary(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	year, month, err := newMonthlyUnpaidQuery(c.Request.URL.Query()).parse()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	ctx := c.Request.Context()
	items, total, summary, err := h.svc.GetMonthlyUnpaidCarryover(ctx, clinicID, year, month, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMonthlyUnpaidCarryoverResponse(items, total, page, limit, summary))
}

// GetDailySummary はレジ締め日次集計を返す。BUG-368
// GET /v1/accountings/daily-summary?date=YYYY-MM-DD[&clinic_ids=1,2]
// clinic_ids が複数の場合は per_clinic 配列を追加で返す (#86 段階3 論点4=2)。
func (h *AccountingHandler) GetDailySummary(c *gin.Context) {
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}

	query := newDailySummaryQuery(c.Request.URL.Query())

	if len(clinicIDs) == 1 {
		result, err := h.svc.GetDailySummary(c.Request.Context(), clinicIDs[0], query.Date)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, toDailySummaryResponse(result))
		return
	}

	perClinic, err := h.svc.GetDailySummaryForClinics(c.Request.Context(), clinicIDs, query.Date)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDailySummaryForClinicsResponse(perClinic))
}

// CancelAccounting は会計を論理削除（status=cancelled）する。
// BUG-371 / #118: 旧 DeleteAccounting (ハード削除) を本メソッドに置き換え。監査ログを記録する。
// POST /accountings/:id/cancel
func (h *AccountingHandler) CancelAccounting(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := optionalStaffID(c)
	if err := h.svc.Cancel(c.Request.Context(), clinicID, id, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// optionalStaffID は httpapi.OptionalStaffID への薄い委譲（エラーを書かない actor 取得・B④）。
func optionalStaffID(c *gin.Context) *uint64 {
	return httpapi.OptionalStaffID(c)
}
