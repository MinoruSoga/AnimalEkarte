package clinic

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestCreateClinicRequest_ToServiceInput(t *testing.T) {
	input := (&CreateClinicRequest{
		Name:               "Main Clinic",
		PostalCode:         "100-0001",
		Address:            "Tokyo",
		PhoneNumber:        "03-0000-0000",
		FaxNumber:          "03-0000-0001",
		RegistrationNumber: "REG-001",
		DirectorName:       "Dr. Test",
		Email:              "clinic@example.test",
		Website:            "https://example.test",
	}).ToServiceInput()

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

	input := (&UpdateClinicRequest{
		Name:            &name,
		IsActive:        &isActive,
		StandardTaxRate: &standardTaxRate,
		ReducedTaxRate:  &reducedTaxRate,
	}).ToServiceInput()

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
	input := (&UpdateClinicRequest{}).ToServiceInput()

	if input.Name != nil || input.IsActive != nil || input.StandardTaxRate != nil {
		t.Fatalf("expected nil optional fields, got %#v", input)
	}
}

func TestClinicContactBinding_RejectsInvalidFormats(t *testing.T) {
	RegisterContactBindingValidators()
	// POC-17: email/phone/postal must match owner contact formats at bind time.
	badEmail := "not-an-email"
	badPhone := "12345"
	badPostal := "abc"
	okEmail := "clinic@example.com"
	okPhone := "03-1234-5678"
	okPostal := "123-4567"

	if err := binding.Validator.ValidateStruct(&CreateClinicRequest{
		Name: "x", Email: badEmail,
	}); err == nil {
		t.Fatal("expected invalid email rejection")
	}
	if err := binding.Validator.ValidateStruct(&CreateClinicRequest{
		Name: "x", PhoneNumber: badPhone,
	}); err == nil {
		t.Fatal("expected invalid phone rejection")
	}
	if err := binding.Validator.ValidateStruct(&CreateClinicRequest{
		Name: "x", PostalCode: badPostal,
	}); err == nil {
		t.Fatal("expected invalid postal rejection")
	}
	if err := binding.Validator.ValidateStruct(&CreateClinicRequest{
		Name: "x", Email: okEmail, PhoneNumber: okPhone, PostalCode: okPostal, FaxNumber: okPhone,
	}); err != nil {
		t.Fatalf("valid contact fields rejected: %v", err)
	}
	if err := binding.Validator.ValidateStruct(&UpdateClinicRequest{Email: &badEmail}); err == nil {
		t.Fatal("expected update email rejection")
	}
	if err := binding.Validator.ValidateStruct(&UpdateCompanyRequest{PhoneNumber: &badPhone}); err == nil {
		t.Fatal("expected company phone rejection")
	}
	if err := binding.Validator.ValidateStruct(&UpdateCompanyRequest{Email: &okEmail, PostalCode: &okPostal}); err != nil {
		t.Fatalf("valid company contact rejected: %v", err)
	}
}
