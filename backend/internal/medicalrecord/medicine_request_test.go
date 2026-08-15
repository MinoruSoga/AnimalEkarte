package medicalrecord

import "testing"

func TestCreateMedicineRequest_ToServiceInput(t *testing.T) {
	parentID := uint64(10)
	price := int64(1200)
	dosageForm := "tablet"
	medicineUnit := "tablet"
	inventoryID := uint64(20)
	taxType := "excluded"
	taxRate := 0.1
	req := createMedicineRequest{
		Name:            "medicine",
		ParentID:        &parentID,
		Price:           &price,
		IsActive:        true,
		Description:     "description",
		DosageForm:      &dosageForm,
		MedicineUnit:    &medicineUnit,
		InventoryID:     &inventoryID,
		DefaultQuantity: 1.5,
		SortOrder:       2,
		TaxType:         &taxType,
		TaxRate:         &taxRate,
		IsNonInsurance:  true,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.ParentID == nil || *input.ParentID != parentID {
		t.Fatalf("ParentID = %v, want %d", input.ParentID, parentID)
	}
	if input.Price == nil || *input.Price != price {
		t.Fatalf("Price = %v, want %d", input.Price, price)
	}
	if !input.IsActive {
		t.Fatal("IsActive = false, want true")
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.DosageForm == nil || *input.DosageForm != dosageForm {
		t.Fatalf("DosageForm = %v, want %q", input.DosageForm, dosageForm)
	}
	if input.MedicineUnit == nil || *input.MedicineUnit != medicineUnit {
		t.Fatalf("MedicineUnit = %v, want %q", input.MedicineUnit, medicineUnit)
	}
	if input.InventoryID == nil || *input.InventoryID != inventoryID {
		t.Fatalf("InventoryID = %v, want %d", input.InventoryID, inventoryID)
	}
	if input.DefaultQuantity != req.DefaultQuantity || input.SortOrder != req.SortOrder {
		t.Fatalf("quantity/order = (%f, %d), want (%f, %d)", input.DefaultQuantity, input.SortOrder, req.DefaultQuantity, req.SortOrder)
	}
	if input.TaxType == nil || *input.TaxType != taxType {
		t.Fatalf("TaxType = %v, want %q", input.TaxType, taxType)
	}
	if input.TaxRate == nil || *input.TaxRate != taxRate {
		t.Fatalf("TaxRate = %v, want %f", input.TaxRate, taxRate)
	}
	if !input.IsNonInsurance {
		t.Fatal("IsNonInsurance = false, want true")
	}
}

func TestUpdateMedicineRequest_ToServiceInput(t *testing.T) {
	name := ""
	parentID := uint64(10)
	price := int64(0)
	isActive := false
	description := ""
	dosageForm := ""
	medicineUnit := ""
	inventoryID := uint64(20)
	defaultQuantity := 0.0
	sortOrder := 0
	taxType := "exempt"
	taxRate := 0.0
	isNonInsurance := false
	req := updateMedicineRequest{
		Name:            &name,
		ParentID:        &parentID,
		ClearParentID:   true,
		Price:           &price,
		IsActive:        &isActive,
		Description:     &description,
		DosageForm:      &dosageForm,
		MedicineUnit:    &medicineUnit,
		InventoryID:     &inventoryID,
		DefaultQuantity: &defaultQuantity,
		SortOrder:       &sortOrder,
		TaxType:         &taxType,
		TaxRate:         &taxRate,
		IsNonInsurance:  &isNonInsurance,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want explicit empty string", input.Name)
	}
	if input.ParentID == nil || *input.ParentID != parentID {
		t.Fatalf("ParentID = %v, want %d", input.ParentID, parentID)
	}
	if !input.ClearParentID {
		t.Fatal("ClearParentID = false, want true")
	}
	if input.Price == nil || *input.Price != 0 {
		t.Fatalf("Price = %v, want explicit zero", input.Price)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", input.IsActive)
	}
	if input.Description == nil || *input.Description != description {
		t.Fatalf("Description = %v, want explicit empty string", input.Description)
	}
	if input.DosageForm == nil || *input.DosageForm != dosageForm {
		t.Fatalf("DosageForm = %v, want explicit empty string", input.DosageForm)
	}
	if input.MedicineUnit == nil || *input.MedicineUnit != medicineUnit {
		t.Fatalf("MedicineUnit = %v, want explicit empty string", input.MedicineUnit)
	}
	if input.DefaultQuantity == nil || *input.DefaultQuantity != 0 {
		t.Fatalf("DefaultQuantity = %v, want explicit zero", input.DefaultQuantity)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
	if input.TaxRate == nil || *input.TaxRate != 0 {
		t.Fatalf("TaxRate = %v, want explicit zero", input.TaxRate)
	}
	if input.IsNonInsurance == nil || *input.IsNonInsurance {
		t.Fatalf("IsNonInsurance = %v, want explicit false", input.IsNonInsurance)
	}
}

func TestUpdateMedicineRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateMedicineRequest{}

	input := req.toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.ParentID != nil {
		t.Fatalf("ParentID = %v, want nil", input.ParentID)
	}
	if input.ClearParentID {
		t.Fatal("ClearParentID = true, want false")
	}
	if input.Price != nil {
		t.Fatalf("Price = %v, want nil", input.Price)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.TaxType != nil {
		t.Fatalf("TaxType = %v, want nil", input.TaxType)
	}
}
