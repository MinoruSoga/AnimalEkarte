package medicalrecord

import "testing"

func TestCreateCarePlanItemRequest_ToServiceInput(t *testing.T) {
	medicineID := uint64(10)
	procedureID := uint64(20)
	hospitalizationPlanID := uint64(30)
	req := createCarePlanItemRequest{
		Type:                  "medicine",
		Name:                  "medication",
		Description:           "description",
		Timing:                []string{"morning", "night"},
		Status:                "active",
		Notes:                 "notes",
		MedicineID:            &medicineID,
		ProcedureID:           &procedureID,
		HospitalizationPlanID: &hospitalizationPlanID,
		UnitPrice:             1200,
		Category:              "medicine",
		SortOrder:             2,
	}

	input := req.toServiceInput()

	if input.Type != req.Type {
		t.Fatalf("Type = %q, want %q", input.Type, req.Type)
	}
	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if len(input.Timing) != 2 || input.Timing[0] != "morning" || input.Timing[1] != "night" {
		t.Fatalf("Timing = %#v, want [morning night]", input.Timing)
	}
	if input.MedicineID == nil || *input.MedicineID != medicineID {
		t.Fatalf("MedicineID = %v, want %d", input.MedicineID, medicineID)
	}
	if input.ProcedureID == nil || *input.ProcedureID != procedureID {
		t.Fatalf("ProcedureID = %v, want %d", input.ProcedureID, procedureID)
	}
	if input.HospitalizationPlanID == nil || *input.HospitalizationPlanID != hospitalizationPlanID {
		t.Fatalf("HospitalizationPlanID = %v, want %d", input.HospitalizationPlanID, hospitalizationPlanID)
	}
	if input.UnitPrice != req.UnitPrice || input.Category != req.Category || input.SortOrder != req.SortOrder {
		t.Fatalf("pricing/category/order = (%d, %q, %d), want (%d, %q, %d)", input.UnitPrice, input.Category, input.SortOrder, req.UnitPrice, req.Category, req.SortOrder)
	}
}

func TestUpdateCarePlanItemRequest_ToServiceInput(t *testing.T) {
	itemType := "instruction"
	name := "updated"
	status := "completed"
	notes := ""
	unitPrice := int64(0)
	sortOrder := 0
	req := updateCarePlanItemRequest{
		Type:      &itemType,
		Name:      &name,
		Timing:    []string{},
		Status:    &status,
		Notes:     &notes,
		UnitPrice: &unitPrice,
		SortOrder: &sortOrder,
	}

	input := req.toServiceInput()

	if input.Type == nil || *input.Type != itemType {
		t.Fatalf("Type = %v, want %q", input.Type, itemType)
	}
	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want %q", input.Name, name)
	}
	if input.Timing == nil {
		t.Fatal("Timing = nil, want explicit empty slice")
	}
	if len(input.Timing) != 0 {
		t.Fatalf("Timing length = %d, want 0", len(input.Timing))
	}
	if input.Notes == nil || *input.Notes != notes {
		t.Fatalf("Notes = %v, want explicit empty string", input.Notes)
	}
	if input.UnitPrice == nil || *input.UnitPrice != 0 {
		t.Fatalf("UnitPrice = %v, want explicit zero", input.UnitPrice)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
}

func TestUpdateCarePlanItemRequest_ToServiceInput_NilTiming(t *testing.T) {
	req := updateCarePlanItemRequest{}

	input := req.toServiceInput()

	if input.Timing != nil {
		t.Fatalf("Timing = %#v, want nil", input.Timing)
	}
}
