package handler

import "testing"

func TestCreateClinicRequest_ToServiceInput(t *testing.T) {
	input := (&createClinicRequest{
		Name:               "Main Clinic",
		PostalCode:         "100-0001",
		Address:            "Tokyo",
		PhoneNumber:        "03-0000-0000",
		FaxNumber:          "03-0000-0001",
		RegistrationNumber: "REG-001",
		DirectorName:       "Dr. Test",
		Email:              "clinic@example.test",
		Website:            "https://example.test",
	}).toServiceInput()

	if input.Name != "Main Clinic" {
		t.Fatalf("Name = %q, want Main Clinic", input.Name)
	}
	if input.Email != "clinic@example.test" {
		t.Fatalf("Email = %q, want clinic@example.test", input.Email)
	}
	if input.Website != "https://example.test" {
		t.Fatalf("Website = %q, want https://example.test", input.Website)
	}
}

func TestUpdateClinicRequest_ToServiceInput(t *testing.T) {
	name := ""
	isActive := false
	standardTaxRate := 0.1
	reducedTaxRate := 0.08

	input := (&updateClinicRequest{
		Name:            &name,
		IsActive:        &isActive,
		StandardTaxRate: &standardTaxRate,
		ReducedTaxRate:  &reducedTaxRate,
	}).toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.StandardTaxRate != &standardTaxRate {
		t.Fatalf("StandardTaxRate pointer was not preserved")
	}
	if input.ReducedTaxRate != &reducedTaxRate {
		t.Fatalf("ReducedTaxRate pointer was not preserved")
	}
}

func TestUpdateClinicRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateClinicRequest{}).toServiceInput()

	if input.Name != nil || input.IsActive != nil || input.StandardTaxRate != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}
