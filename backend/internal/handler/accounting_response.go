package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

func toDailySummaryResponse(s *repository.DailySummaryResult) dailySummaryResponse {
	paymentTotals := make([]paymentMethodTotalResponse, 0, len(s.PaymentTotals))
	for _, p := range s.PaymentTotals {
		paymentTotals = append(paymentTotals, paymentMethodTotalResponse{
			Method: p.Method,
			Total:  p.Total,
		})
	}
	categoryTotals := make([]categoryTotalResponse, 0, len(s.CategoryTotals))
	for _, ct := range s.CategoryTotals {
		categoryTotals = append(categoryTotals, categoryTotalResponse{
			Category: ct.Category,
			Total:    ct.Total,
		})
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

func toUnpaidByOwnerResponse(items []repository.UnpaidOwnerAggregate, total int64, page, limit int, s repository.UnpaidSummary) unpaidByOwnerResponse {
	data := make([]unpaidOwnerResponse, 0, len(items))
	for _, it := range items {
		data = append(data, unpaidOwnerResponse{
			OwnerID:         it.OwnerID,
			OwnerName:       it.OwnerName,
			Count:           it.Count,
			TotalAmount:     it.TotalAmount,
			OldestScheduled: it.OldestScheduled,
			LatestScheduled: it.LatestScheduled,
		})
	}
	return unpaidByOwnerResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
		Summary: unpaidSummaryResponse{
			TotalAmount:  s.TotalAmount,
			BillingCount: s.BillingCount,
			OwnerCount:   s.OwnerCount,
		},
	}
}

type refundResponse struct {
	ID             uint64    `json:"id"`
	ClinicID       uint64    `json:"clinic_id"`
	BillingID      uint64    `json:"billing_id"`
	Amount         int64     `json:"amount"`
	Reason         string    `json:"reason"`
	RefundedBy     *uint64   `json:"refunded_by"`
	RefundedByName string    `json:"refunded_by_name"`
	PaymentMethod  *string   `json:"payment_method,omitempty"`
	RefundedAt     time.Time `json:"refunded_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func toRefundResponse(r *model.BillingRefund) refundResponse {
	var staffName string
	if r.RefundedByStaff != nil {
		staffName = r.RefundedByStaff.Name
	}
	var paymentMethod *string
	if r.PaymentMethod != nil {
		pm := string(*r.PaymentMethod)
		paymentMethod = &pm
	}
	return refundResponse{
		ID:             r.ID,
		ClinicID:       r.ClinicID,
		BillingID:      r.BillingID,
		Amount:         r.Amount,
		Reason:         r.Reason,
		RefundedBy:     r.RefundedBy,
		RefundedByName: staffName,
		PaymentMethod:  paymentMethod,
		RefundedAt:     r.RefundedAt,
		CreatedAt:      r.CreatedAt,
	}
}

type billingItemResponse struct {
	ID                    uint64    `json:"id"`
	BillingID             uint64    `json:"billing_id"`
	Category              string    `json:"category"`
	Name                  string    `json:"name"`
	UnitPrice             int64     `json:"unit_price"`
	Quantity              float64   `json:"quantity"`
	DiscountRate          float64   `json:"discount_rate"`
	DiscountAmount        int64     `json:"discount_amount"`
	Subtotal              int64     `json:"subtotal"`
	TaxType               string    `json:"tax_type"`
	TaxRate               float64   `json:"tax_rate"`
	TaxAmount             int64     `json:"tax_amount"`
	IsInsuranceApplicable bool      `json:"is_insurance_applicable"`
	Source                string    `json:"source"`
	TreatmentID           *uint64   `json:"treatment_id,omitempty"`
	AppointmentID         *uint64   `json:"appointment_id,omitempty"`
	TrimmingCourseID      *uint64   `json:"trimming_course_id,omitempty"`
	TrimmingOptionID      *uint64   `json:"trimming_option_id,omitempty"`
	SortOrder             int       `json:"sort_order"`
	CreatedAt             time.Time `json:"created_at"`
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
	HasInsurance        bool                    `json:"has_insurance"`
	Status              string                  `json:"status"`
	ScheduledDate       time.Time               `json:"scheduled_date"`
	CompletedAt         *time.Time              `json:"completed_at,omitempty"`
	Memo                string                  `json:"memo"`
	Items               []billingItemResponse   `json:"items,omitempty"`
	Payments            []paymentResponse       `json:"payments,omitempty"`
	PaymentSplits       []paymentSplitResponse  `json:"payment_splits,omitempty"`
	Refunds             []refundResponse        `json:"refunds,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

func toBillingItemResponse(item *model.BillingItem) billingItemResponse {
	// #85: subtotal は割引後（単価×数量 − 割引額）。負値は 0 クランプ。
	subtotal := max(int64(float64(item.UnitPrice)*item.Quantity)-item.DiscountAmount, 0)
	taxAmount := item.CalculateTaxAmount()
	return billingItemResponse{
		ID:                    item.ID,
		BillingID:             item.BillingID,
		Category:              string(item.Category),
		Name:                  item.Name,
		UnitPrice:             item.UnitPrice,
		Quantity:              item.Quantity,
		DiscountRate:          item.DiscountRate,
		DiscountAmount:        item.DiscountAmount,
		Subtotal:              subtotal,
		TaxType:               string(item.TaxType),
		TaxRate:               item.TaxRate,
		TaxAmount:             taxAmount,
		IsInsuranceApplicable: item.IsInsuranceApplicable,
		Source:                string(item.Source),
		TreatmentID:           item.TreatmentID,
		AppointmentID:         item.AppointmentID,
		TrimmingCourseID:      item.TrimmingCourseID,
		TrimmingOptionID:      item.TrimmingOptionID,
		SortOrder:             item.SortOrder,
		CreatedAt:             item.CreatedAt,
	}
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
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
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
		CreatedAt:       s.CreatedAt,
	}
}

func toAccountingResponse(b *model.Billing) accountingResponse {
	items := make([]billingItemResponse, 0, len(b.Items))
	for i := range b.Items {
		items = append(items, toBillingItemResponse(&b.Items[i]))
	}
	payments := make([]paymentResponse, 0, len(b.Payments))
	for i := range b.Payments {
		payments = append(payments, toPaymentResponse(&b.Payments[i]))
	}
	splits := make([]paymentSplitResponse, 0, len(b.PaymentSplits))
	for i := range b.PaymentSplits {
		splits = append(splits, toPaymentSplitResponse(&b.PaymentSplits[i]))
	}
	refunds := make([]refundResponse, 0, len(b.Refunds))
	for i := range b.Refunds {
		refunds = append(refunds, toRefundResponse(&b.Refunds[i]))
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
		HasInsurance:        b.HasInsurance,
		Status:              string(b.Status),
		ScheduledDate:       b.ScheduledDate,
		CompletedAt:         b.CompletedAt,
		Memo:                b.Memo,
		Items:               items,
		Payments:            payments,
		PaymentSplits:       splits,
		Refunds:             refunds,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}
