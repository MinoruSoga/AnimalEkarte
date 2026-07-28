package inventory

import "testing"

func TestCreateMerchandiseItemRequest_ToServiceInput(t *testing.T) {
	taxRate := 0.1

	req := createMerchandiseItemRequest{
		Name:      "Prescription food",
		Category:  "food",
		UnitPrice: 1200,
		TaxType:   "included",
		TaxRate:   &taxRate,
		IsActive:  true,
		SortOrder: 3,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Category != req.Category {
		t.Fatalf("Category = %q, want %q", input.Category, req.Category)
	}
	if input.UnitPrice != req.UnitPrice {
		t.Fatalf("UnitPrice = %d, want %d", input.UnitPrice, req.UnitPrice)
	}
	if input.TaxType != req.TaxType {
		t.Fatalf("TaxType = %q, want %q", input.TaxType, req.TaxType)
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %v, want %v", input.IsActive, req.IsActive)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateMerchandiseItemRequest_ToServiceInput(t *testing.T) {
	name := ""
	category := "goods"
	unitPrice := int64(0)
	taxType := "exempt"
	taxRate := 0.0
	isActive := false
	sortOrder := 0

	req := updateMerchandiseItemRequest{
		Name:      &name,
		Category:  &category,
		UnitPrice: &unitPrice,
		TaxType:   &taxType,
		TaxRate:   &taxRate,
		IsActive:  &isActive,
		SortOrder: &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Category != &category {
		t.Fatalf("Category pointer was not preserved")
	}
	if input.UnitPrice != &unitPrice {
		t.Fatalf("UnitPrice pointer was not preserved")
	}
	if input.TaxType != &taxType {
		t.Fatalf("TaxType pointer was not preserved")
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateMerchandiseItemRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateMerchandiseItemRequest{}).toServiceInput()

	if input.Name != nil || input.Category != nil || input.UnitPrice != nil || input.TaxType != nil ||
		input.TaxRate != nil || input.IsActive != nil || input.SortOrder != nil {
		t.Fatalf("input = %+v, want all nil fields", input)
	}
}
