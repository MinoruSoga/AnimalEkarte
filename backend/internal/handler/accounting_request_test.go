package handler

import (
	"testing"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestListAccountingQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listAccountingQuery{
		PetID:     "10",
		OwnerID:   "20",
		Status:    "completed",
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 20 {
		t.Fatalf("OwnerID = %v, want 20", filters.OwnerID)
	}
	if filters.Status == nil || *filters.Status != "completed" {
		t.Fatalf("Status = %v, want completed", filters.Status)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListAccountingQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listAccountingQuery
	}{
		{name: "pet_id", query: listAccountingQuery{PetID: "abc"}},
		{name: "owner_id", query: listAccountingQuery{OwnerID: "abc"}},
		{name: "start_date", query: listAccountingQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listAccountingQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listAccountingFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{
		BaseDate: "2026-05-28",
		GroupBy:  "billing",
	}).toServiceFilters("2026-05-01")
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.BaseDate != "2026-05-28" {
		t.Fatalf("BaseDate = %q, want 2026-05-28", filters.BaseDate)
	}
	if filters.GroupBy != "billing" {
		t.Fatalf("GroupBy = %q, want billing", filters.GroupBy)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_Defaults(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{}).toServiceFilters("2026-05-01")
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.BaseDate != "2026-05-01" {
		t.Fatalf("BaseDate = %q, want default", filters.BaseDate)
	}
	if filters.GroupBy != "owner" {
		t.Fatalf("GroupBy = %q, want owner", filters.GroupBy)
	}
}

func TestListUnpaidBillingsQuery_ToServiceFilters_InvalidBaseDate(t *testing.T) {
	filters, err := (&listUnpaidBillingsQuery{BaseDate: "2026/05/28"}).toServiceFilters("2026-05-01")
	if err == nil {
		t.Fatal("toServiceFilters returned nil error")
	}
	if filters != (listUnpaidBillingsFilters{}) {
		t.Fatalf("filters = %#v, want zero value", filters)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateAccountingRequest_ToServiceInput(t *testing.T) {
	scheduledDate := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	medicalRecordID := uint64(10)
	hospitalizationID := uint64(20)
	ownerID := uint64(30)
	petID := uint64(40)
	req := createAccountingRequest{
		MedicalRecordID:   &medicalRecordID,
		HospitalizationID: &hospitalizationID,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Subtotal:          1000,
		TaxTotal:          100,
		TotalAmount:       1100,
		HasInsurance:      true,
		Status:            string(model.BillingStatusCompleted),
		ScheduledDate:     scheduledDate,
		CompletedAt:       &completedAt,
		Memo:              "memo",
	}

	input := req.toServiceInput(1)

	if input.ClinicID != 1 {
		t.Fatalf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.MedicalRecordID == nil || *input.MedicalRecordID != medicalRecordID {
		t.Fatalf("MedicalRecordID = %v, want %d", input.MedicalRecordID, medicalRecordID)
	}
	if input.HospitalizationID == nil || *input.HospitalizationID != hospitalizationID {
		t.Fatalf("HospitalizationID = %v, want %d", input.HospitalizationID, hospitalizationID)
	}
	if input.OwnerID == nil || *input.OwnerID != ownerID {
		t.Fatalf("OwnerID = %v, want %d", input.OwnerID, ownerID)
	}
	if input.PetID == nil || *input.PetID != petID {
		t.Fatalf("PetID = %v, want %d", input.PetID, petID)
	}
	if input.Subtotal != req.Subtotal || input.TaxTotal != req.TaxTotal || input.TotalAmount != req.TotalAmount {
		t.Fatalf("amounts = (%d, %d, %d), want (%d, %d, %d)", input.Subtotal, input.TaxTotal, input.TotalAmount, req.Subtotal, req.TaxTotal, req.TotalAmount)
	}
	if !input.HasInsurance {
		t.Fatal("HasInsurance = false, want true")
	}
	if input.Status != model.BillingStatusCompleted {
		t.Fatalf("Status = %q, want %q", input.Status, model.BillingStatusCompleted)
	}
	if !input.ScheduledDate.Equal(scheduledDate) {
		t.Fatalf("ScheduledDate = %v, want %v", input.ScheduledDate, scheduledDate)
	}
	if input.CompletedAt == nil || !input.CompletedAt.Equal(completedAt) {
		t.Fatalf("CompletedAt = %v, want %v", input.CompletedAt, completedAt)
	}
	if input.Memo != req.Memo {
		t.Fatalf("Memo = %q, want %q", input.Memo, req.Memo)
	}
}

func TestUpdateAccountingRequest_ToServiceInput(t *testing.T) {
	status := string(model.BillingStatusCompleted)
	paymentMethod := string(model.PaymentMethodCash)
	memo := ""
	subtotal := int64(1000)
	taxTotal := int64(100)
	totalAmount := int64(1100)
	hasInsurance := false
	insuranceRatio := 50.0
	insuranceName := "insurance"
	insuranceAmount := int64(500)
	discountAmount := int64(0)
	billingAmount := int64(600)
	receivedAmount := int64(1000)
	changeAmount := int64(400)
	paymentMethodID := uint64(9)
	req := updateAccountingRequest{
		Subtotal:        &subtotal,
		TaxTotal:        &taxTotal,
		TotalAmount:     &totalAmount,
		HasInsurance:    &hasInsurance,
		Status:          &status,
		Memo:            &memo,
		PaymentMethod:   &paymentMethod,
		InsuranceRatio:  &insuranceRatio,
		InsuranceName:   &insuranceName,
		InsuranceAmount: &insuranceAmount,
		DiscountAmount:  &discountAmount,
		BillingAmount:   &billingAmount,
		ReceivedAmount:  &receivedAmount,
		ChangeAmount:    &changeAmount,
		PaymentSplits: []paymentSplitRequest{
			{
				Method:          string(model.PaymentMethodCash),
				PaymentMethodID: &paymentMethodID,
				Amount:          600,
				ReceivedAmount:  1000,
				ChangeAmount:    400,
			},
		},
	}

	input := req.toServiceInput(1, 2, 3)

	if input.ID != 1 || input.ClinicID != 2 {
		t.Fatalf("ID/ClinicID = %d/%d, want 1/2", input.ID, input.ClinicID)
	}
	if input.StaffID == nil || *input.StaffID != 3 {
		t.Fatalf("StaffID = %v, want 3", input.StaffID)
	}
	if input.Subtotal == nil || *input.Subtotal != subtotal {
		t.Fatalf("Subtotal = %v, want %d", input.Subtotal, subtotal)
	}
	if input.HasInsurance == nil || *input.HasInsurance != hasInsurance {
		t.Fatalf("HasInsurance = %v, want false", input.HasInsurance)
	}
	if input.Status == nil || *input.Status != model.BillingStatusCompleted {
		t.Fatalf("Status = %v, want %q", input.Status, model.BillingStatusCompleted)
	}
	if input.Memo == nil || *input.Memo != memo {
		t.Fatalf("Memo = %v, want explicit empty string", input.Memo)
	}
	if input.PaymentMethod == nil || *input.PaymentMethod != model.PaymentMethodCash {
		t.Fatalf("PaymentMethod = %v, want %q", input.PaymentMethod, model.PaymentMethodCash)
	}
	if input.DiscountAmount == nil || *input.DiscountAmount != 0 {
		t.Fatalf("DiscountAmount = %v, want explicit zero", input.DiscountAmount)
	}
	if len(input.PaymentSplits) != 1 {
		t.Fatalf("PaymentSplits length = %d, want 1", len(input.PaymentSplits))
	}
	split := input.PaymentSplits[0]
	if split.Method != model.PaymentMethodCash {
		t.Fatalf("split.Method = %q, want %q", split.Method, model.PaymentMethodCash)
	}
	if split.PaymentMethodID == nil || *split.PaymentMethodID != paymentMethodID {
		t.Fatalf("split.PaymentMethodID = %v, want %d", split.PaymentMethodID, paymentMethodID)
	}
	if split.Amount != 600 || split.ReceivedAmount != 1000 || split.ChangeAmount != 400 {
		t.Fatalf("split amounts = (%d, %d, %d), want (600, 1000, 400)", split.Amount, split.ReceivedAmount, split.ChangeAmount)
	}
}

func TestUpdateAccountingRequest_ToServiceInput_NilPaymentSplits(t *testing.T) {
	req := updateAccountingRequest{}

	input := req.toServiceInput(1, 2, 3)

	if input.PaymentSplits != nil {
		t.Fatalf("PaymentSplits = %#v, want nil", input.PaymentSplits)
	}
	if input.Status != nil {
		t.Fatalf("Status = %v, want nil", input.Status)
	}
	if input.PaymentMethod != nil {
		t.Fatalf("PaymentMethod = %v, want nil", input.PaymentMethod)
	}
}

func TestPaymentSplitRequest_ToServiceInput(t *testing.T) {
	paymentMethodID := uint64(8)
	req := paymentSplitRequest{
		Method:          string(model.PaymentMethodCreditCard),
		PaymentMethodID: &paymentMethodID,
		Amount:          1200,
		ReceivedAmount:  0,
		ChangeAmount:    0,
	}

	input := req.toServiceInput()

	if input.Method != model.PaymentMethodCreditCard {
		t.Fatalf("Method = %q, want %q", input.Method, model.PaymentMethodCreditCard)
	}
	if input.PaymentMethodID == nil || *input.PaymentMethodID != paymentMethodID {
		t.Fatalf("PaymentMethodID = %v, want %d", input.PaymentMethodID, paymentMethodID)
	}
	if input.Amount != req.Amount {
		t.Fatalf("Amount = %d, want %d", input.Amount, req.Amount)
	}
}
