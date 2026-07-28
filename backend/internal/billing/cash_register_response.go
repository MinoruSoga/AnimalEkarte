package billing

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type cashRegisterCloseResponse struct {
	ID                uint64          `json:"id"`
	ClinicID          uint64          `json:"clinic_id"`
	CloseDate         time.Time       `json:"close_date"`
	Period            string          `json:"period"`
	TheoreticalCash   int64           `json:"theoretical_cash"`
	ActualCash        int64           `json:"actual_cash"`
	CashDifference    int64           `json:"cash_difference"`
	CategoryBreakdown json.RawMessage `json:"category_breakdown"`
	Memo              string          `json:"memo"`
	ClosedBy          *uint64         `json:"closed_by,omitempty"`
	ClosedAt          time.Time       `json:"closed_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toCashRegisterCloseResponse(r *model.CashRegisterClose) cashRegisterCloseResponse {
	return cashRegisterCloseResponse{
		ID:                r.ID,
		ClinicID:          r.ClinicID,
		CloseDate:         httpapi.LocalTime(r.CloseDate),
		Period:            r.Period,
		TheoreticalCash:   r.TheoreticalCash,
		ActualCash:        r.ActualCash,
		CashDifference:    r.CashDifference,
		CategoryBreakdown: r.CategoryBreakdown,
		Memo:              r.Memo,
		ClosedBy:          r.ClosedBy,
		ClosedAt:          httpapi.LocalTime(r.ClosedAt),
		CreatedAt:         httpapi.LocalTime(r.CreatedAt),
		UpdatedAt:         httpapi.LocalTime(r.UpdatedAt),
	}
}

type cashRegisterTaxBreakdownEntryResponse struct {
	TaxableAmount int64 `json:"taxable_amount"`
	TaxAmount     int64 `json:"tax_amount"`
}

type cashRegisterTaxBreakdownResponse struct {
	Standard cashRegisterTaxBreakdownEntryResponse `json:"standard"`
	Reduced  cashRegisterTaxBreakdownEntryResponse `json:"reduced"`
}

type cashRegisterAggregateSummaryResponse struct {
	Categories      map[string]map[string]int64      `json:"categories"`
	PaymentMethods  []PaymentMethodResponse          `json:"payment_methods"`
	TheoreticalCash int64                            `json:"theoretical_cash"`
	TaxBreakdown    cashRegisterTaxBreakdownResponse `json:"tax_breakdown"`
}

type cashRegisterBillingDetailResponse struct {
	BillingID         uint64  `json:"billing_id"`
	PaidAt            string  `json:"paid_at"`
	OwnerName         string  `json:"owner_name"`
	PetName           string  `json:"pet_name"`
	IsHospitalization bool    `json:"is_hospitalization"`
	Category          string  `json:"category"`
	PaymentMethodID   *uint64 `json:"payment_method_id,omitempty"`
	PaymentMethodName string  `json:"payment_method_name"`
	BillingAmount     int64   `json:"billing_amount"`
	RefundAmount      int64   `json:"refund_amount"`
	NetAmount         int64   `json:"net_amount"`
}

type cashRegisterPreviewResponse struct {
	Date            string                               `json:"date"`
	Period          string                               `json:"period"`
	PeriodStart     string                               `json:"period_start"`
	PeriodEnd       string                               `json:"period_end"`
	IsAlreadyClosed bool                                 `json:"is_already_closed"`
	IsHoliday       bool                                 `json:"is_holiday"`
	Aggregate       cashRegisterAggregateSummaryResponse `json:"aggregate"`
	BillingDetails  []cashRegisterBillingDetailResponse  `json:"billing_details"`
}

func toCashRegisterPreviewResponse(p *CashRegisterPreview) cashRegisterPreviewResponse {
	pms := make([]PaymentMethodResponse, 0, len(p.Aggregate.PaymentMethods))
	for i := range p.Aggregate.PaymentMethods {
		pms = append(pms, ToPaymentMethodResponse(&p.Aggregate.PaymentMethods[i]))
	}

	details := make([]cashRegisterBillingDetailResponse, 0, len(p.BillingDetails))
	for i := range p.BillingDetails {
		d := &p.BillingDetails[i]
		details = append(details, cashRegisterBillingDetailResponse{
			BillingID:         d.BillingID,
			PaidAt:            d.PaidAt,
			OwnerName:         d.OwnerName,
			PetName:           d.PetName,
			IsHospitalization: d.IsHospitalization,
			Category:          d.Category,
			PaymentMethodID:   d.PaymentMethodID,
			PaymentMethodName: d.PaymentMethodName,
			BillingAmount:     d.BillingAmount,
			RefundAmount:      d.RefundAmount,
			NetAmount:         d.NetAmount,
		})
	}

	return cashRegisterPreviewResponse{
		Date:            p.Date,
		Period:          p.Period,
		PeriodStart:     p.PeriodStart,
		PeriodEnd:       p.PeriodEnd,
		IsAlreadyClosed: p.IsAlreadyClosed,
		IsHoliday:       p.IsHoliday,
		Aggregate: cashRegisterAggregateSummaryResponse{
			Categories:      p.Aggregate.Categories,
			PaymentMethods:  pms,
			TheoreticalCash: p.Aggregate.TheoreticalCash,
			TaxBreakdown: cashRegisterTaxBreakdownResponse{
				Standard: cashRegisterTaxBreakdownEntryResponse{
					TaxableAmount: p.Aggregate.TaxBreakdown.Standard.TaxableAmount,
					TaxAmount:     p.Aggregate.TaxBreakdown.Standard.TaxAmount,
				},
				Reduced: cashRegisterTaxBreakdownEntryResponse{
					TaxableAmount: p.Aggregate.TaxBreakdown.Reduced.TaxableAmount,
					TaxAmount:     p.Aggregate.TaxBreakdown.Reduced.TaxAmount,
				},
			},
		},
		BillingDetails: details,
	}
}
