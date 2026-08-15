package medicalrecord

import "testing"

func TestCreateTreatmentPlanRequest_ToServiceInput(t *testing.T) {
	req := createTreatmentPlanRequest{
		TreatmentContent: "Daily injection",
		Memo:             "memo",
		IsInsurance:      true,
		UnitPrice:        1500,
		Quantity:         2,
		DiscountRate:     5,
		DiscountAmount:   100,
		Subtotal:         2800,
		SortOrder:        4,
	}

	input := req.toServiceInput()

	if input.TreatmentContent != req.TreatmentContent {
		t.Errorf("TreatmentContent = %q, want %q", input.TreatmentContent, req.TreatmentContent)
	}
	if input.IsInsurance != req.IsInsurance {
		t.Errorf("IsInsurance = %v, want %v", input.IsInsurance, req.IsInsurance)
	}
	if input.UnitPrice != req.UnitPrice {
		t.Errorf("UnitPrice = %d, want %d", input.UnitPrice, req.UnitPrice)
	}
	if input.DiscountAmount != req.DiscountAmount {
		t.Errorf("DiscountAmount = %d, want %d", input.DiscountAmount, req.DiscountAmount)
	}
	if input.SortOrder != req.SortOrder {
		t.Errorf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateTreatmentPlanRequest_ToServiceInput(t *testing.T) {
	content := ""
	memo := "updated"
	isInsurance := false
	unitPrice := int64(0)
	quantity := 0.5
	discountRate := float64(0)
	discountAmount := int64(0)
	subtotal := int64(0)
	sortOrder := 0
	req := updateTreatmentPlanRequest{
		TreatmentContent: &content,
		Memo:             &memo,
		IsInsurance:      &isInsurance,
		UnitPrice:        &unitPrice,
		Quantity:         &quantity,
		DiscountRate:     &discountRate,
		DiscountAmount:   &discountAmount,
		Subtotal:         &subtotal,
		SortOrder:        &sortOrder,
	}

	input := req.toServiceInput()

	if input.TreatmentContent == nil || *input.TreatmentContent != content {
		t.Errorf("TreatmentContent = %v, want empty string pointer", input.TreatmentContent)
	}
	if input.IsInsurance == nil || *input.IsInsurance {
		t.Errorf("IsInsurance = %v, want false pointer", input.IsInsurance)
	}
	if input.UnitPrice == nil || *input.UnitPrice != unitPrice {
		t.Errorf("UnitPrice = %v, want %d", input.UnitPrice, unitPrice)
	}
	if input.DiscountRate == nil || *input.DiscountRate != discountRate {
		t.Errorf("DiscountRate = %v, want %v", input.DiscountRate, discountRate)
	}
	if input.SortOrder == nil || *input.SortOrder != sortOrder {
		t.Errorf("SortOrder = %v, want %d", input.SortOrder, sortOrder)
	}
}
