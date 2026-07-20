package medicalrecord

import "testing"

func TestCreateTreatmentRequest_ToServiceInput(t *testing.T) {
	consultationID := uint64(11)
	req := createTreatmentRequest{
		ItemType:       "consultation",
		ConsultationID: &consultationID,
		UnitPrice:      1200,
		Quantity:       2.5,
		IsSelected:     true,
		Status:         "completed",
		Content:        "General treatment",
		Memo:           "memo",
		AdminRoute:     "oral",
		IsInsurance:    true,
		DiscountRate:   10,
		DiscountAmount: 100,
		SortOrder:      3,
	}

	input := req.toServiceInput()

	if string(input.ItemType) != req.ItemType {
		t.Errorf("ItemType = %q, want %q", input.ItemType, req.ItemType)
	}
	if input.ConsultationID == nil || *input.ConsultationID != consultationID {
		t.Errorf("ConsultationID = %v, want %d", input.ConsultationID, consultationID)
	}
	if input.UnitPrice != req.UnitPrice {
		t.Errorf("UnitPrice = %d, want %d", input.UnitPrice, req.UnitPrice)
	}
	if input.Quantity != req.Quantity {
		t.Errorf("Quantity = %v, want %v", input.Quantity, req.Quantity)
	}
	if !input.IsSelected {
		t.Error("IsSelected = false, want true")
	}
	if input.DiscountAmount != req.DiscountAmount {
		t.Errorf("DiscountAmount = %d, want %d", input.DiscountAmount, req.DiscountAmount)
	}
}

func TestUpdateTreatmentRequest_ToServiceInput(t *testing.T) {
	itemType := "procedure"
	unitPrice := int64(800)
	quantity := 0.5
	isSelected := false
	status := "not_applicable"
	content := ""
	sortOrder := 0
	req := updateTreatmentRequest{
		ItemType:   &itemType,
		UnitPrice:  &unitPrice,
		Quantity:   &quantity,
		IsSelected: &isSelected,
		Status:     &status,
		Content:    &content,
		SortOrder:  &sortOrder,
	}

	input := req.toServiceInput()

	if input.ItemType == nil || string(*input.ItemType) != itemType {
		t.Errorf("ItemType = %v, want %q", input.ItemType, itemType)
	}
	if input.UnitPrice == nil || *input.UnitPrice != unitPrice {
		t.Errorf("UnitPrice = %v, want %d", input.UnitPrice, unitPrice)
	}
	if input.Quantity == nil || *input.Quantity != quantity {
		t.Errorf("Quantity = %v, want %v", input.Quantity, quantity)
	}
	if input.IsSelected == nil || *input.IsSelected {
		t.Errorf("IsSelected = %v, want false pointer", input.IsSelected)
	}
	if input.Content == nil || *input.Content != content {
		t.Errorf("Content = %v, want empty string pointer", input.Content)
	}
	if input.SortOrder == nil || *input.SortOrder != sortOrder {
		t.Errorf("SortOrder = %v, want %d", input.SortOrder, sortOrder)
	}
}

func TestBulkUpdateTreatmentsRequest_ToServiceInput(t *testing.T) {
	req := bulkUpdateTreatmentsRequest{
		Treatments: []bulkTreatmentItem{
			{ID: 1, SortOrder: 2},
			{ID: 3, SortOrder: 0},
		},
	}

	input := req.toServiceInput()

	if len(input.Treatments) != len(req.Treatments) {
		t.Fatalf("len(Treatments) = %d, want %d", len(input.Treatments), len(req.Treatments))
	}
	if input.Treatments[0].ID != 1 || input.Treatments[0].SortOrder != 2 {
		t.Errorf("Treatments[0] = %+v, want ID 1 SortOrder 2", input.Treatments[0])
	}
	if input.Treatments[1].ID != 3 || input.Treatments[1].SortOrder != 0 {
		t.Errorf("Treatments[1] = %+v, want ID 3 SortOrder 0", input.Treatments[1])
	}
}
