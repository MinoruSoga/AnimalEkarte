package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListAccountings godoc
func (h *Handler) ListAccountings(c *gin.Context) {
	clinicIDs, ok := resolveListClinicIDs(c)
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

	var accountings []model.Billing
	var total int64
	if len(clinicIDs) == 1 {
		accountings, total, err = h.svc.Accounting.List(
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
		accountings, total, err = h.svc.Accounting.ListForClinics(
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
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(accountings, toAccountingResponse), total, page, limit))
}

// GetAccounting godoc
func (h *Handler) GetAccounting(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := resolveAllClinicIDs(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	accounting, err := h.svc.Accounting.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
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
		c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(accountings, toAccountingResponse), total, page, limit))
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
// GET /v1/accountings/daily-summary?date=YYYY-MM-DD[&clinic_ids=1,2]
// clinic_ids が複数の場合は per_clinic 配列を追加で返す (#86 段階3 論点4=2)。
func (h *Handler) GetDailySummary(c *gin.Context) {
	clinicIDs, ok := resolveListClinicIDs(c)
	if !ok {
		return
	}

	query := newDailySummaryQuery(c.Request.URL.Query())

	if len(clinicIDs) == 1 {
		result, err := h.svc.Accounting.GetDailySummary(c.Request.Context(), clinicIDs[0], query.Date)
		if err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, toDailySummaryResponse(result))
		return
	}

	perClinic, err := h.svc.Accounting.GetDailySummaryForClinics(c.Request.Context(), clinicIDs, query.Date)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := make([]clinicDailySummaryResponse, 0, len(perClinic))
	for _, cs := range perClinic {
		items = append(items, clinicDailySummaryResponse{
			ClinicID: cs.ClinicID,
			Summary:  toDailySummaryResponse(cs.Summary),
		})
	}
	c.JSON(http.StatusOK, gin.H{"per_clinic": items})
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
