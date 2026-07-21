package billing

import "testing"

func TestCreateInsuranceRequest_ToServiceInput(t *testing.T) {
	coverageRate := 70

	req := createInsuranceRequest{
		Name:         "Pet insurance",
		IsActive:     true,
		Description:  "Standard coverage",
		CoverageRate: &coverageRate,
		ContactPhone: "03-0000-0000",
		SortOrder:    4,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %v, want %v", input.IsActive, req.IsActive)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.CoverageRate != &coverageRate {
		t.Fatalf("CoverageRate pointer was not preserved")
	}
	if input.ContactPhone != req.ContactPhone {
		t.Fatalf("ContactPhone = %q, want %q", input.ContactPhone, req.ContactPhone)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateInsuranceRequest_ToServiceInput(t *testing.T) {
	name := ""
	isActive := false
	description := ""
	coverageRate := 0
	contactPhone := ""
	sortOrder := 0

	req := updateInsuranceRequest{
		Name:         &name,
		IsActive:     &isActive,
		Description:  &description,
		CoverageRate: &coverageRate,
		ContactPhone: &contactPhone,
		SortOrder:    &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.Description != &description {
		t.Fatalf("Description pointer was not preserved")
	}
	if input.CoverageRate != &coverageRate {
		t.Fatalf("CoverageRate pointer was not preserved")
	}
	if input.ContactPhone != &contactPhone {
		t.Fatalf("ContactPhone pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateInsuranceRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateInsuranceRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.CoverageRate != nil {
		t.Fatalf("CoverageRate = %v, want nil", input.CoverageRate)
	}
	if input.ContactPhone != nil {
		t.Fatalf("ContactPhone = %v, want nil", input.ContactPhone)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
