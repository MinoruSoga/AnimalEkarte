package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// billingStatusPtr は *string を *model.BillingStatus に変換する。nil の場合は nil を返す。
func billingStatusPtr(s *string) *model.BillingStatus {
	if s == nil {
		return nil
	}
	v := model.BillingStatus(*s)
	return &v
}

// paymentMethodPtr は *string を *model.PaymentMethod に変換する。nil の場合は nil を返す。
func paymentMethodPtr(s *string) *model.PaymentMethod {
	if s == nil {
		return nil
	}
	v := model.PaymentMethod(*s)
	return &v
}

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

	var petID *uint64
	if s := c.Query("pet_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
			return
		}
		petID = &id
	}
	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	startDate, err := parseDateQuery(c, "start_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	endDate, err := parseDateQuery(c, "end_date")
	if err != nil {
		RespondError(c, err)
		return
	}

	accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
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
	created, err := h.svc.Accounting.Create(ctx, &service.CreateAccountingInput{
		ClinicID:          clinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		Status:            model.BillingStatus(input.Status),
		ScheduledDate:     input.ScheduledDate,
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-011: CPM ステージタグ同期（completed 時のみ、best-effort）
	if model.BillingStatus(input.Status) == model.BillingStatusCompleted && created.OwnerID != nil {
		_ = h.svc.LstepTagSync.SyncCPMStageTag(ctx, clinicID, *created.OwnerID)
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
	updated, err := h.svc.Accounting.Update(ctx, &service.UpdateAccountingInput{
		ID:                id,
		ClinicID:          clinicID,
		StaffID:           &staffID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		ScheduledDate:     input.ScheduledDate,
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
		InsuranceRatio:    input.InsuranceRatio,
		InsuranceName:     input.InsuranceName,
		InsuranceAmount:   input.InsuranceAmount,
		DiscountAmount:    input.DiscountAmount,
		BillingAmount:     input.BillingAmount,
		ReceivedAmount:    input.ReceivedAmount,
		ChangeAmount:      input.ChangeAmount,
		Status:            billingStatusPtr(input.Status),
		PaymentMethod:     paymentMethodPtr(input.PaymentMethod),
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-011: CPM ステージタグ同期（completed 遷移時のみ、best-effort）
	if input.Status != nil && model.BillingStatus(*input.Status) == model.BillingStatusCompleted && updated.OwnerID != nil {
		_ = h.svc.LstepTagSync.SyncCPMStageTag(ctx, clinicID, *updated.OwnerID)
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

	baseDatePtr, err := parseDateQuery(c, "base_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	baseDate := time.Now().Format("2006-01-02")
	if baseDatePtr != nil {
		baseDate = *baseDatePtr
	}

	groupBy := c.DefaultQuery("group_by", "owner")
	ctx := c.Request.Context()

	switch groupBy {
	case "billing":
		// ListUnpaidByBilling は model.Billing スライスを返す（DBテーブル名 billings 由来）。
		// 会計ドメイン(accounting)と DB モデル(Billing)の命名差異はここで吸収する。
		accountings, total, err := h.svc.Accounting.ListUnpaidByBilling(ctx, clinicID, baseDate, page, limit)
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
		aggregates, total, summary, err := h.svc.Accounting.ListUnpaidByOwner(ctx, clinicID, baseDate, page, limit)
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

	dateStr := c.Query("date")
	result, err := h.svc.Accounting.GetDailySummary(c.Request.Context(), clinicID, dateStr)
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
