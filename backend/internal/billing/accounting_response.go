package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-368: レジ締め日次集計レスポンス

type paymentMethodTotalResponse struct {
	Method string `json:"method"`
	Total  int64  `json:"total"`
}

type categoryTotalResponse struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

type dailySummaryResponse struct {
	PaymentTotals  []paymentMethodTotalResponse `json:"payment_totals"`
	CategoryTotals []categoryTotalResponse      `json:"category_totals"`
	BillingCount   int64                        `json:"billing_count"`
	GrandTotal     int64                        `json:"grand_total"`
}

func toDailySummaryResponse(s *DailySummaryResult) dailySummaryResponse {
	paymentTotals := make([]paymentMethodTotalResponse, 0, len(s.PaymentTotals))
	for _, p := range s.PaymentTotals {
		paymentTotals = append(paymentTotals, paymentMethodTotalResponse(p))
	}
	categoryTotals := make([]categoryTotalResponse, 0, len(s.CategoryTotals))
	for _, ct := range s.CategoryTotals {
		categoryTotals = append(categoryTotals, categoryTotalResponse(ct))
	}
	return dailySummaryResponse{
		PaymentTotals:  paymentTotals,
		CategoryTotals: categoryTotals,
		BillingCount:   s.BillingCount,
		GrandTotal:     s.GrandTotal,
	}
}

// clinicDailySummaryResponse は拠点別日次集計レスポンス (#86 段階3 論点4=2)。
type clinicDailySummaryResponse struct {
	ClinicID uint64               `json:"clinic_id"`
	Summary  dailySummaryResponse `json:"summary"`
}

type dailySummaryForClinicsResponse struct {
	PerClinic []clinicDailySummaryResponse `json:"per_clinic"`
}

func toDailySummaryForClinicsResponse(items []ClinicDailySummary) dailySummaryForClinicsResponse {
	perClinic := make([]clinicDailySummaryResponse, 0, len(items))
	for _, item := range items {
		perClinic = append(perClinic, clinicDailySummaryResponse{
			ClinicID: item.ClinicID,
			Summary:  toDailySummaryResponse(item.Summary),
		})
	}
	return dailySummaryForClinicsResponse{PerClinic: perClinic}
}

// BUG-370: 月末未納者一覧レスポンス

type unpaidOwnerResponse struct {
	OwnerID         uint64 `json:"owner_id"`
	OwnerName       string `json:"owner_name"`
	Count           int64  `json:"count"`
	TotalAmount     int64  `json:"total_amount"`
	OldestScheduled string `json:"oldest_scheduled"`
	LatestScheduled string `json:"latest_scheduled"`
}

type unpaidSummaryResponse struct {
	TotalAmount  int64 `json:"total_amount"`
	BillingCount int64 `json:"billing_count"`
	OwnerCount   int64 `json:"owner_count"`
}

type unpaidByOwnerResponse struct {
	Data    []unpaidOwnerResponse `json:"data"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
	Summary unpaidSummaryResponse `json:"summary"`
}

func toUnpaidByOwnerResponse(items []UnpaidOwnerAggregate, total int64, page, limit int, s UnpaidSummary) unpaidByOwnerResponse {
	data := make([]unpaidOwnerResponse, 0, len(items))
	for _, it := range items {
		data = append(data, unpaidOwnerResponse(it))
	}
	return unpaidByOwnerResponse{
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Summary: unpaidSummaryResponse(s),
	}
}

// #182: 会計画面用 飼主未納残高レスポンス

type ownerUnpaidBalanceResponse struct {
	UnpaidTotal int64 `json:"unpaid_total"`
	UnpaidCount int64 `json:"unpaid_count"`
}

func toOwnerUnpaidBalanceResponse(b OwnerUnpaidBalance) ownerUnpaidBalanceResponse {
	return ownerUnpaidBalanceResponse{
		UnpaidTotal: b.TotalAmount,
		UnpaidCount: b.Count,
	}
}

// #114: 月次未納繰越集計レスポンス

type monthlyUnpaidOwnerPetResponse struct {
	OwnerID            uint64  `json:"owner_id"`
	OwnerName          string  `json:"owner_name"`
	PetID              *uint64 `json:"pet_id,omitempty"`
	PetName            string  `json:"pet_name"`
	PrevMonthCarryover int64   `json:"prev_month_carryover"`
	CurrentMonthUnpaid int64   `json:"current_month_unpaid"`
	NextMonthCarryover int64   `json:"next_month_carryover"`
}

type monthlyUnpaidSummaryResponse struct {
	PrevMonthCarryover int64 `json:"prev_month_carryover"`
	CurrentMonthUnpaid int64 `json:"current_month_unpaid"`
	NextMonthCarryover int64 `json:"next_month_carryover"`
}

type monthlyUnpaidCarryoverResponse struct {
	Data    []monthlyUnpaidOwnerPetResponse `json:"data"`
	Total   int64                           `json:"total"`
	Page    int                             `json:"page"`
	Limit   int                             `json:"limit"`
	Summary monthlyUnpaidSummaryResponse    `json:"summary"`
}

func toMonthlyUnpaidCarryoverResponse(items []MonthlyUnpaidOwnerPet, total int64, page, limit int, s MonthlyUnpaidSummary) monthlyUnpaidCarryoverResponse {
	data := make([]monthlyUnpaidOwnerPetResponse, 0, len(items))
	for _, it := range items {
		data = append(data, monthlyUnpaidOwnerPetResponse(it))
	}
	return monthlyUnpaidCarryoverResponse{
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Summary: monthlyUnpaidSummaryResponse(s),
	}
}

type paymentSplitResponse struct {
	ID              uint64    `json:"id"`
	ClinicID        uint64    `json:"clinic_id"`
	BillingID       uint64    `json:"billing_id"`
	Method          string    `json:"method"`
	PaymentMethodID *uint64   `json:"payment_method_id,omitempty"`
	Amount          int64     `json:"amount"`
	ReceivedAmount  int64     `json:"received_amount"`
	ChangeAmount    int64     `json:"change_amount"`
	PaidBy          *uint64   `json:"paid_by,omitempty"`
	PaidByName      string    `json:"paid_by_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type paymentResponse struct {
	ID              uint64    `json:"id"`
	BillingID       uint64    `json:"billing_id"`
	Subtotal        int64     `json:"subtotal"`
	TaxTotal        int64     `json:"tax_total"`
	TotalAmount     int64     `json:"total_amount"`
	InsuranceName   string    `json:"insurance_name"`
	InsuranceRatio  float64   `json:"insurance_ratio"`
	InsuranceAmount int64     `json:"insurance_amount"`
	DiscountAmount  int64     `json:"discount_amount"`
	BillingAmount   int64     `json:"billing_amount"`
	ReceivedAmount  int64     `json:"received_amount"`
	ChangeAmount    int64     `json:"change_amount"`
	Method          string    `json:"method"`
	PaidBy          *uint64   `json:"paid_by"`
	PaidByName      string    `json:"paid_by_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// accountingOwnerSummary は accountingResponse.owner に埋め込む軽量サマリ。
// BUG-374 TC-367-03: フロントの BackendAccounting は generated/models.Owner.name を期待するため、
// JSON タグは "name" に統一する（"owner_name" だと帳票で「様」のみ表示される）。
type accountingOwnerSummary struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

type accountingPetSummary struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type accountingResponse struct {
	ID                  uint64                  `json:"id"`
	ClinicID            uint64                  `json:"clinic_id"`
	MedicalRecordID     *uint64                 `json:"medical_record_id,omitempty"`
	HospitalizationID   *uint64                 `json:"hospitalization_id,omitempty"`
	OwnerID             *uint64                 `json:"owner_id,omitempty"`
	PetID               *uint64                 `json:"pet_id,omitempty"`
	Owner               *accountingOwnerSummary `json:"owner,omitempty"`
	Pet                 *accountingPetSummary   `json:"pet,omitempty"`
	Subtotal            int64                   `json:"subtotal"`
	TaxTotal            int64                   `json:"tax_total"`
	TotalAmount         int64                   `json:"total_amount"`
	TotalRefundedAmount int64                   `json:"total_refunded_amount"`
	// OutstandingAmount は未収残高（円）。waiting 全額 or クレジット訂正後の差額（BUG-007）。
	OutstandingAmount int64                  `json:"outstanding_amount"`
	HasInsurance      bool                   `json:"has_insurance"`
	Status            string                 `json:"status"`
	ScheduledDate     time.Time              `json:"scheduled_date"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	Memo              string                 `json:"memo"`
	Items             []BillingItemResponse  `json:"items,omitempty"`
	Payments          []paymentResponse      `json:"payments,omitempty"`
	PaymentSplits     []paymentSplitResponse `json:"payment_splits,omitempty"`
	Refunds           []RefundResponse       `json:"refunds,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

func toPaymentResponse(p *model.Payment) paymentResponse {
	var staffName string
	if p.PaidByStaff != nil {
		staffName = p.PaidByStaff.Name
	}
	return paymentResponse{
		ID:              p.ID,
		BillingID:       p.BillingID,
		Subtotal:        p.Subtotal,
		TaxTotal:        p.TaxTotal,
		TotalAmount:     p.TotalAmount,
		InsuranceName:   p.InsuranceName,
		InsuranceRatio:  p.InsuranceRatio,
		InsuranceAmount: p.InsuranceAmount,
		DiscountAmount:  p.DiscountAmount,
		BillingAmount:   p.BillingAmount,
		ReceivedAmount:  p.ReceivedAmount,
		ChangeAmount:    p.ChangeAmount,
		Method:          string(p.Method),
		PaidBy:          p.PaidBy,
		PaidByName:      staffName,
		CreatedAt:       httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(p.UpdatedAt),
	}
}

func toPaymentSplitResponse(s *model.PaymentSplit) paymentSplitResponse {
	var staffName string
	if s.PaidByStaff != nil {
		staffName = s.PaidByStaff.Name
	}
	return paymentSplitResponse{
		ID:              s.ID,
		ClinicID:        s.ClinicID,
		BillingID:       s.BillingID,
		Method:          string(s.Method),
		PaymentMethodID: s.PaymentMethodID,
		Amount:          s.Amount,
		ReceivedAmount:  s.ReceivedAmount,
		ChangeAmount:    s.ChangeAmount,
		PaidBy:          s.PaidBy,
		PaidByName:      staffName,
		CreatedAt:       httpapi.LocalTime(s.CreatedAt),
	}
}

func toAccountingResponse(b *model.Billing) accountingResponse {
	items := make([]BillingItemResponse, 0, len(b.Items))
	for i := range b.Items {
		items = append(items, ToBillingItemResponse(&b.Items[i]))
	}
	payments := make([]paymentResponse, 0, len(b.Payments))
	for i := range b.Payments {
		payments = append(payments, toPaymentResponse(&b.Payments[i]))
	}
	splits := make([]paymentSplitResponse, 0, len(b.PaymentSplits))
	for i := range b.PaymentSplits {
		splits = append(splits, toPaymentSplitResponse(&b.PaymentSplits[i]))
	}
	refunds := make([]RefundResponse, 0, len(b.Refunds))
	for i := range b.Refunds {
		refunds = append(refunds, ToRefundResponse(&b.Refunds[i]))
	}

	var owner *accountingOwnerSummary
	if b.Owner != nil {
		owner = &accountingOwnerSummary{
			ID:        b.Owner.ID,
			OwnerName: b.Owner.Name,
		}
	}

	var pet *accountingPetSummary
	if b.Pet != nil {
		pet = &accountingPetSummary{
			ID:   b.Pet.ID,
			Name: b.Pet.Name,
		}
	}

	return accountingResponse{
		ID:                  b.ID,
		ClinicID:            b.ClinicID,
		MedicalRecordID:     b.MedicalRecordID,
		HospitalizationID:   b.HospitalizationID,
		OwnerID:             b.OwnerID,
		PetID:               b.PetID,
		Owner:               owner,
		Pet:                 pet,
		Subtotal:            b.Subtotal,
		TaxTotal:            b.TaxTotal,
		TotalAmount:         b.TotalAmount,
		TotalRefundedAmount: b.TotalRefundedAmount,
		OutstandingAmount:   OutstandingAmount(b),
		HasInsurance:        b.HasInsurance,
		Status:              string(b.Status),
		ScheduledDate:       httpapi.LocalTime(b.ScheduledDate),
		CompletedAt:         httpapi.LocalTimePtr(b.CompletedAt),
		Memo:                b.Memo,
		Items:               items,
		Payments:            payments,
		PaymentSplits:       splits,
		Refunds:             refunds,
		CreatedAt:           httpapi.LocalTime(b.CreatedAt),
		UpdatedAt:           httpapi.LocalTime(b.UpdatedAt),
	}
}
