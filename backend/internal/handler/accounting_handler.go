package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListAccountings godoc
func (h *Handler) ListAccountings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	q := newListAccountingQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	accountings, total, err := h.svc.Accounting.List(
		c.Request.Context(),
		clinicID,
		filters.PetID,
		filters.OwnerID,
		filters.Status,
		filters.StartDate,
		filters.EndDate,
		page,
		limit,
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(accountings, toAccountingResponse), total, page, limit))
}

// GetAccounting godoc
func (h *Handler) GetAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	accounting, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(accounting))
}

// CreateAccounting godoc
func (h *Handler) CreateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	created, err := h.svc.Accounting.Create(ctx, input.toServiceInput(clinicID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/accountings/%d", created.ID))
	c.JSON(http.StatusCreated, toAccountingResponse(created))
}

// UpdateAccounting godoc
func (h *Handler) UpdateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateAccountingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-372: discount_amount にゼロ以外を指定する場合は discount:edit 権限要
	// Payment 既存値取得用リポジトリが未整備のため、ゼロ値は通過・非ゼロのみ権限チェック。
	// ゼロ値再送（通常操作）を阻害せず、非ゼロ（値引意図）のみ保護する。
	if input.DiscountAmount != nil && *input.DiscountAmount != 0 {
		zero := int64(0)
		if err := h.requireDiscountEditInt(c, input.DiscountAmount, zero); err != nil {
			RespondError(c, err)
			return
		}
	}

	ctx := c.Request.Context()
	updated, err := h.svc.Accounting.Update(ctx, input.toServiceInput(id, clinicID, staffID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountingResponse(updated))
}

// ListUnpaidBillings は月末未納者一覧を返す。BUG-370
// GET /v1/accountings/unpaid?base_date=YYYY-MM-DD&group_by=owner|billing&page=N&limit=N
func (h *Handler) ListUnpaidBillings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	filters, err := newListUnpaidBillingsQuery(c.Request.URL.Query()).toServiceFilters(time.Now().Format("2006-01-02"))
	if err != nil {
		RespondError(c, err)
		return
	}

	ctx := c.Request.Context()

	switch filters.GroupBy {
	case "billing":
		// ListUnpaidByBilling は model.Billing スライスを返す（DBテーブル名 billings 由来）。
		// 会計ドメイン(accounting)と DB モデル(Billing)の命名差異はここで吸収する。
		accountings, total, err := h.svc.Accounting.ListUnpaidByBilling(ctx, clinicID, filters.BaseDate, page, limit)
		if err != nil {
			RespondError(c, err)
			return
		}
		responses := make([]accountingResponse, 0, len(accountings))
		for i := range accountings {
			responses = append(responses, toAccountingResponse(&accountings[i]))
		}
		c.JSON(http.StatusOK, newPaginatedResponse(responses, total, page, limit))
	case "owner":
		aggregates, total, summary, err := h.svc.Accounting.ListUnpaidByOwner(ctx, clinicID, filters.BaseDate, page, limit)
		if err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, toUnpaidByOwnerResponse(aggregates, total, page, limit, summary))
	default:
		RespondError(c, apperrors.WrapInvalidInput("group_by must be owner or billing"))
	}
}

// GetDailySummary はレジ締め日次集計を返す。BUG-368
// GET /v1/accountings/daily-summary?date=YYYY-MM-DD
func (h *Handler) GetDailySummary(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	query := newDailySummaryQuery(c.Request.URL.Query())
	result, err := h.svc.Accounting.GetDailySummary(c.Request.Context(), clinicID, query.Date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toDailySummaryResponse(result))
}

// CancelAccounting は会計を論理削除（status=cancelled）する。
// BUG-371: 旧 DeleteAccounting (ハード削除) を本メソッドに置き換え。
// POST /accountings/:id/cancel
func (h *Handler) CancelAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Accounting.Cancel(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
