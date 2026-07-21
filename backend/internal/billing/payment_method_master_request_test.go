package billing

import "testing"

func TestCreatePaymentMethodRequest_ToServiceInput(t *testing.T) {
	req := createPaymentMethodRequest{
		Name:         "Credit card",
		DisplayOrder: 5,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.DisplayOrder != req.DisplayOrder {
		t.Fatalf("DisplayOrder = %d, want %d", input.DisplayOrder, req.DisplayOrder)
	}
}

func TestUpdatePaymentMethodRequest_ToServiceInput(t *testing.T) {
	name := ""
	displayOrder := 0
	isActive := false

	req := updatePaymentMethodRequest{
		Name:         &name,
		DisplayOrder: &displayOrder,
		IsActive:     &isActive,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.DisplayOrder != &displayOrder {
		t.Fatalf("DisplayOrder pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
}

func TestUpdatePaymentMethodRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updatePaymentMethodRequest{}).toServiceInput()

	if input.Name != nil || input.DisplayOrder != nil || input.IsActive != nil {
		t.Fatalf("input = %+v, want all nil fields", input)
	}
}
