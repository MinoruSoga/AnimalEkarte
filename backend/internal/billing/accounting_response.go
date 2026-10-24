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

// EMR-188: 月末未納者一覧（期間検索）レスポンス

type periodUnpaidOwnerPetResponse struct {
	OwnerID             uint64  `json:"owner_id"`
	OwnerName           string  `json:"owner_name"`
	PetID               *uint64 `json:"pet_id,omitempty"`
	PetName             string  `json:"pet_name"`
	PrevPeriodCarryover int64   `json:"prev_period_carryover"`
	CurrentPeriodUnpaid int64   `json:"current_period_unpaid"`
	PeriodEndCarryover  int64   `json:"period_end_carryover"`
	LatestScheduled     string  `json:"latest_scheduled"`
}

type periodUnpaidSummaryResponse struct {
	PrevPeriodCarryover int64 `json:"prev_period_carryover"`
	CurrentPeriodUnpaid int64 `json:"current_period_unpaid"`
	PeriodEndCarryover  int64 `json:"period_end_carryover"`
}

type periodUnpaidCarryoverResponse struct {
	Data    []periodUnpaidOwnerPetResponse `json:"data"`
	Total   int64                          `json:"total"`
	Page    int                            `json:"page"`
	Limit   int                            `json:"limit"`
	Summary periodUnpaidSummaryResponse    `json:"summary"`
}

func toPeriodUnpaidCarryoverResponse(items []PeriodUnpaidOwnerPet, total int64, page, limit int, s PeriodUnpaidSummary) periodUnpaidCarryoverResponse {
	data := make([]periodUnpaidOwnerPetResponse, 0, len(items))
	for _, it := range items {
		data = append(data, periodUnpaidOwnerPetResponse(it))
	}
	return periodUnpaidCarryoverResponse{
		Data:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Summary: periodUnpaidSummaryResponse(s),
	}
}

// accountingResponse は会計ヘッダと明細の応答。
// Q4(c): 合計・税額・状態・確定日時・保険有無・支払い・返金・未収残高は、保存先（算定固定・受領配賦・変更イベント）の
// reader が未接続のため導出せず、項目ごと省略する（absent = unknown。0 や空で埋めない）。
// owner / pet の入れ子 summary も、氏名の reader（EMR-491）が接続されるまで省略する（Q8）。held for EMR-506。
type accountingResponse struct {
	ID                uint64                `json:"id"`
	ClinicID          uint64                `json:"clinic_id"`
	MedicalRecordID   *uint64               `json:"medical_record_id,omitempty"`
	HospitalizationID *uint64               `json:"hospitalization_id,omitempty"`
	OwnerID           *uint64               `json:"owner_id,omitempty"`
	PetID             *uint64               `json:"pet_id,omitempty"`
	ScheduledDate     time.Time             `json:"scheduled_date"`
	Memo              string                `json:"memo"`
	Items             []BillingItemResponse `json:"items,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

func toAccountingResponse(b *model.Billing) accountingResponse {
	items := make([]BillingItemResponse, 0, len(b.Items))
	for i := range b.Items {
		items = append(items, ToBillingItemResponse(&b.Items[i]))
	}

	return accountingResponse{
		ID:                b.ID,
		ClinicID:          b.ClinicID,
		MedicalRecordID:   b.MedicalRecordID,
		HospitalizationID: b.HospitalizationID,
		OwnerID:           b.OwnerID,
		PetID:             b.PetID,
		ScheduledDate:     httpapi.LocalTime(b.ScheduledDate),
		Memo:              b.Memo,
		Items:             items,
		CreatedAt:         httpapi.LocalTime(b.CreatedAt),
		UpdatedAt:         httpapi.LocalTime(b.UpdatedAt),
	}
}
