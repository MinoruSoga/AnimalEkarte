package medicalrecord

import "testing"

func TestCreateDiagnosisTypeRequest_ToServiceInput(t *testing.T) {
	req := createDiagnosisTypeRequest{
		Name:        "category",
		IsActive:    true,
		Description: "description",
		SortOrder:   2,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %t, want %t", input.IsActive, req.IsActive)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateDiagnosisTypeRequest_ToServiceInput(t *testing.T) {
	name := ""
	isActive := false
	description := ""
	sortOrder := 0
	req := updateDiagnosisTypeRequest{
		Name:        &name,
		IsActive:    &isActive,
		Description: &description,
		SortOrder:   &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want explicit empty string", input.Name)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", input.IsActive)
	}
	if input.Description == nil || *input.Description != description {
		t.Fatalf("Description = %v, want explicit empty string", input.Description)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
}

func TestUpdateDiagnosisTypeRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateDiagnosisTypeRequest{}

	input := req.toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}

func TestCreateDiagnosisNameRequest_ToServiceInput(t *testing.T) {
	req := createDiagnosisNameRequest{
		Name:            "diagnosis",
		DiagnosisTypeID: 10,
		IsActive:        true,
		Description:     "description",
		SortOrder:       3,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.DiagnosisTypeID != req.DiagnosisTypeID {
		t.Fatalf("DiagnosisTypeID = %d, want %d", input.DiagnosisTypeID, req.DiagnosisTypeID)
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %t, want %t", input.IsActive, req.IsActive)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateDiagnosisNameRequest_ToServiceInput(t *testing.T) {
	name := ""
	diagnosisTypeID := uint64(0)
	isActive := false
	description := ""
	sortOrder := 0
	req := updateDiagnosisNameRequest{
		Name:            &name,
		DiagnosisTypeID: &diagnosisTypeID,
		IsActive:        &isActive,
		Description:     &description,
		SortOrder:       &sortOrder,
	}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Fatalf("Name = %v, want explicit empty string", input.Name)
	}
	if input.DiagnosisTypeID == nil || *input.DiagnosisTypeID != 0 {
		t.Fatalf("DiagnosisTypeID = %v, want explicit zero", input.DiagnosisTypeID)
	}
	if input.IsActive == nil || *input.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", input.IsActive)
	}
	if input.Description == nil || *input.Description != description {
		t.Fatalf("Description = %v, want explicit empty string", input.Description)
	}
	if input.SortOrder == nil || *input.SortOrder != 0 {
		t.Fatalf("SortOrder = %v, want explicit zero", input.SortOrder)
	}
}

func TestUpdateDiagnosisNameRequest_ToServiceInput_NilFields(t *testing.T) {
	req := updateDiagnosisNameRequest{}

	input := req.toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.DiagnosisTypeID != nil {
		t.Fatalf("DiagnosisTypeID = %v, want nil", input.DiagnosisTypeID)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
