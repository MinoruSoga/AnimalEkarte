package billing

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToCashRegisterCloseResponse(t *testing.T) {
	closedBy := uint64(9)
	closeDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	closedAt := time.Date(2026, 5, 28, 12, 30, 0, 0, time.UTC)
	record := &model.CashRegisterClose{
		ID:                1,
		ClinicID:          2,
		CloseDate:         closeDate,
		Period:            "am",
		TheoreticalCash:   10000,
		ActualCash:        9900,
		CashDifference:    -100,
		CategoryBreakdown: json.RawMessage(`{"examination":{"cash":1000}}`),
		Memo:              "memo",
		ClosedBy:          &closedBy,
		ClosedAt:          closedAt,
		CreatedAt:         closedAt,
		UpdatedAt:         closedAt,
	}

	resp := toCashRegisterCloseResponse(record)

	if resp.ID != record.ID {
		t.Errorf("ID = %d, want %d", resp.ID, record.ID)
	}
	if resp.ClinicID != record.ClinicID {
		t.Errorf("ClinicID = %d, want %d", resp.ClinicID, record.ClinicID)
	}
	if resp.Period != record.Period {
		t.Errorf("Period = %q, want %q", resp.Period, record.Period)
	}
	if resp.TheoreticalCash != record.TheoreticalCash {
		t.Errorf("TheoreticalCash = %d, want %d", resp.TheoreticalCash, record.TheoreticalCash)
	}
	if resp.ActualCash != record.ActualCash {
		t.Errorf("ActualCash = %d, want %d", resp.ActualCash, record.ActualCash)
	}
	if resp.CashDifference != record.CashDifference {
		t.Errorf("CashDifference = %d, want %d", resp.CashDifference, record.CashDifference)
	}
	if string(resp.CategoryBreakdown) != string(record.CategoryBreakdown) {
		t.Errorf("CategoryBreakdown = %s, want %s", resp.CategoryBreakdown, record.CategoryBreakdown)
	}
	if resp.Memo != record.Memo {
		t.Errorf("Memo = %q, want %q", resp.Memo, record.Memo)
	}
	if resp.ClosedBy == nil || *resp.ClosedBy != closedBy {
		t.Errorf("ClosedBy = %v, want %d", resp.ClosedBy, closedBy)
	}
	if !resp.CloseDate.Equal(closeDate) {
		t.Errorf("CloseDate = %v, want %v", resp.CloseDate, closeDate)
	}
	if !resp.ClosedAt.Equal(closedAt) {
		t.Errorf("ClosedAt = %v, want %v", resp.ClosedAt, closedAt)
	}
}

func TestToCashRegisterCloseResponse_NilClosedBy(t *testing.T) {
	record := &model.CashRegisterClose{
		ID:                1,
		ClinicID:          2,
		Period:            "pm",
		CategoryBreakdown: json.RawMessage(`{}`),
	}

	resp := toCashRegisterCloseResponse(record)

	if resp.ClosedBy != nil {
		t.Errorf("ClosedBy = %v, want nil", resp.ClosedBy)
	}
}

func TestToCashRegisterPreviewResponse(t *testing.T) {
	paymentMethodID := uint64(3)
	preview := &CashRegisterPreview{
		Date:            "2026-05-28",
		Period:          "am",
		PeriodStart:     "2026-05-28T09:00:00+09:00",
		PeriodEnd:       "2026-05-28T13:00:00+09:00",
		IsAlreadyClosed: true,
		IsHoliday:       false,
		Aggregate: CloseAggregateSummary{
			Categories: map[string]map[string]int64{
				string(model.ItemCategoryExamination): {"cash": 1000},
			},
			PaymentMethods: []model.PaymentMethodMaster{
				{ID: 1, ClinicID: 2, Name: "現金", DisplayOrder: 1, IsActive: true},
			},
			TheoreticalCash: 5000,
			TaxBreakdown: TaxBreakdownSummary{
				Standard: TaxBreakdownEntry{TaxableAmount: 1000, TaxAmount: 100},
				Reduced:  TaxBreakdownEntry{TaxableAmount: 500, TaxAmount: 40},
			},
		},
		BillingDetails: []CloseBillingDetail{
			{
				BillingID:         10,
				PaidAt:            "2026-05-28T10:00:00+09:00",
				OwnerName:         "田中太郎",
				PetName:           "ポチ",
				IsHospitalization: false,
				Category:          string(model.ItemCategoryExamination),
				PaymentMethodID:   &paymentMethodID,
				PaymentMethodName: "現金",
				BillingAmount:     1000,
				RefundAmount:      0,
				NetAmount:         1000,
			},
		},
	}

	resp := toCashRegisterPreviewResponse(preview)

	if resp.Date != preview.Date {
		t.Errorf("Date = %q, want %q", resp.Date, preview.Date)
	}
	if resp.Period != preview.Period {
		t.Errorf("Period = %q, want %q", resp.Period, preview.Period)
	}
	if resp.IsAlreadyClosed != preview.IsAlreadyClosed {
		t.Errorf("IsAlreadyClosed = %v, want %v", resp.IsAlreadyClosed, preview.IsAlreadyClosed)
	}
	if len(resp.Aggregate.PaymentMethods) != 1 {
		t.Fatalf("len(PaymentMethods) = %d, want 1", len(resp.Aggregate.PaymentMethods))
	}
	if resp.Aggregate.PaymentMethods[0].Name != "現金" {
		t.Errorf("PaymentMethods[0].Name = %q, want 現金", resp.Aggregate.PaymentMethods[0].Name)
	}
	if resp.Aggregate.TheoreticalCash != preview.Aggregate.TheoreticalCash {
		t.Errorf("TheoreticalCash = %d, want %d", resp.Aggregate.TheoreticalCash, preview.Aggregate.TheoreticalCash)
	}
	if resp.Aggregate.TaxBreakdown.Standard.TaxAmount != 100 {
		t.Errorf("Standard.TaxAmount = %d, want 100", resp.Aggregate.TaxBreakdown.Standard.TaxAmount)
	}
	if resp.Aggregate.TaxBreakdown.Reduced.TaxAmount != 40 {
		t.Errorf("Reduced.TaxAmount = %d, want 40", resp.Aggregate.TaxBreakdown.Reduced.TaxAmount)
	}
	if len(resp.BillingDetails) != 1 {
		t.Fatalf("len(BillingDetails) = %d, want 1", len(resp.BillingDetails))
	}
	detail := resp.BillingDetails[0]
	if detail.BillingID != 10 {
		t.Errorf("BillingID = %d, want 10", detail.BillingID)
	}
	if detail.OwnerName != "田中太郎" {
		t.Errorf("OwnerName = %q, want 田中太郎", detail.OwnerName)
	}
	if detail.PaymentMethodID == nil || *detail.PaymentMethodID != paymentMethodID {
		t.Errorf("PaymentMethodID = %v, want %d", detail.PaymentMethodID, paymentMethodID)
	}
	if detail.NetAmount != 1000 {
		t.Errorf("NetAmount = %d, want 1000", detail.NetAmount)
	}
}

func TestToCashRegisterPreviewResponse_EmptySlices(t *testing.T) {
	preview := &CashRegisterPreview{
		Date:   "2026-05-28",
		Period: "pm",
	}

	resp := toCashRegisterPreviewResponse(preview)

	if resp.Aggregate.PaymentMethods == nil {
		t.Error("PaymentMethods should be an empty slice, not nil")
	}
	if len(resp.Aggregate.PaymentMethods) != 0 {
		t.Errorf("len(PaymentMethods) = %d, want 0", len(resp.Aggregate.PaymentMethods))
	}
	if resp.BillingDetails == nil {
		t.Error("BillingDetails should be an empty slice, not nil")
	}
	if len(resp.BillingDetails) != 0 {
		t.Errorf("len(BillingDetails) = %d, want 0", len(resp.BillingDetails))
	}
}
