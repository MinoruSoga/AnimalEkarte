package medicalrecord

import (
	"reflect"
	"strings"
	"testing"
)

func TestCreateConsultationRequest_ToServiceInput(t *testing.T) {
	price := int64(3300)
	duration := 20
	parentID := uint64(8)
	taxRate := 0.1

	req := createConsultationRequest{
		Name:          "Initial consult",
		Price:         &price,
		IsActive:      true,
		Description:   "First visit",
		TimeCondition: "morning",
		Duration:      &duration,
		ParentID:      &parentID,
		SortOrder:     2,
		TaxType:       "included",
		TaxRate:       &taxRate,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if input.IsActive != req.IsActive {
		t.Fatalf("IsActive = %v, want %v", input.IsActive, req.IsActive)
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.TimeCondition != req.TimeCondition {
		t.Fatalf("TimeCondition = %q, want %q", input.TimeCondition, req.TimeCondition)
	}
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
	if input.TaxType == nil || *input.TaxType != req.TaxType {
		t.Fatalf("TaxType = %v, want %q", input.TaxType, req.TaxType)
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
}

func TestCreateConsultationRequest_ToServiceInput_EmptyTaxType(t *testing.T) {
	input := (&createConsultationRequest{}).toServiceInput()

	if input.TaxType != nil {
		t.Fatalf("TaxType = %v, want nil", input.TaxType)
	}
}

func TestUpdateConsultationRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	timeCondition := "evening"
	duration := 0
	parentID := uint64(0)
	sortOrder := 0
	taxType := "exempt"
	taxRate := 0.0

	req := updateConsultationRequest{
		Name:          &name,
		Price:         &price,
		IsActive:      &isActive,
		Description:   &description,
		TimeCondition: &timeCondition,
		Duration:      &duration,
		ParentID:      &parentID,
		ClearParentID: true,
		SortOrder:     &sortOrder,
		TaxType:       &taxType,
		TaxRate:       &taxRate,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.Description != &description {
		t.Fatalf("Description pointer was not preserved")
	}
	if input.TimeCondition != &timeCondition {
		t.Fatalf("TimeCondition pointer was not preserved")
	}
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.ParentID != &parentID {
		t.Fatalf("ParentID pointer was not preserved")
	}
	if !input.ClearParentID {
		t.Fatalf("ClearParentID = false, want true")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
	if input.TaxType != &taxType {
		t.Fatalf("TaxType pointer was not preserved")
	}
	if input.TaxRate != &taxRate {
		t.Fatalf("TaxRate pointer was not preserved")
	}
}

func TestConsultationRequest_TaxRateBindingTags_MRA03(t *testing.T) {
	// MRA-03: align with medicine/procedure masters (omitempty,min=0,max=1).
	createField, ok := reflect.TypeOf(createConsultationRequest{}).FieldByName("TaxRate")
	if !ok {
		t.Fatal("createConsultationRequest.TaxRate missing")
	}
	createBinding := createField.Tag.Get("binding")
	if !strings.Contains(createBinding, "min=0") || !strings.Contains(createBinding, "max=1") {
		t.Fatalf("create TaxRate binding = %q, want omitempty,min=0,max=1", createBinding)
	}
	updateField, ok := reflect.TypeOf(updateConsultationRequest{}).FieldByName("TaxRate")
	if !ok {
		t.Fatal("updateConsultationRequest.TaxRate missing")
	}
	updateBinding := updateField.Tag.Get("binding")
	if !strings.Contains(updateBinding, "min=0") || !strings.Contains(updateBinding, "max=1") {
		t.Fatalf("update TaxRate binding = %q, want omitempty,min=0,max=1", updateBinding)
	}
}

func TestUpdateConsultationRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateConsultationRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.Price != nil {
		t.Fatalf("Price = %v, want nil", input.Price)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
	if input.Description != nil {
		t.Fatalf("Description = %v, want nil", input.Description)
	}
	if input.TimeCondition != nil {
		t.Fatalf("TimeCondition = %v, want nil", input.TimeCondition)
	}
	if input.Duration != nil {
		t.Fatalf("Duration = %v, want nil", input.Duration)
	}
	if input.ParentID != nil {
		t.Fatalf("ParentID = %v, want nil", input.ParentID)
	}
	if input.ClearParentID {
		t.Fatalf("ClearParentID = true, want false")
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
	if input.TaxType != nil {
		t.Fatalf("TaxType = %v, want nil", input.TaxType)
	}
	if input.TaxRate != nil {
		t.Fatalf("TaxRate = %v, want nil", input.TaxRate)
	}
}
