package trimming

import "testing"

func TestCreateTrimmingCourseRequest_ToServiceInput(t *testing.T) {
	price := int64(5500)
	duration := 60
	active := true

	req := createTrimmingCourseRequest{
		Name:        "Basic groom",
		Price:       &price,
		IsActive:    &active,
		Description: "Bath and trim",
		TargetSize:  "medium",
		Duration:    &duration,
		SortOrder:   4,
	}

	input := req.toServiceInput()

	if input.Name != req.Name {
		t.Fatalf("Name = %q, want %q", input.Name, req.Name)
	}
	if input.Price != &price {
		t.Fatalf("Price pointer was not preserved")
	}
	if !input.IsActive {
		t.Fatalf("IsActive = false, want true")
	}
	if input.Description != req.Description {
		t.Fatalf("Description = %q, want %q", input.Description, req.Description)
	}
	if input.TargetSize != req.TargetSize {
		t.Fatalf("TargetSize = %q, want %q", input.TargetSize, req.TargetSize)
	}
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestUpdateTrimmingCourseRequest_ToServiceInput(t *testing.T) {
	name := ""
	price := int64(0)
	isActive := false
	description := ""
	targetSize := "cat"
	duration := 0
	sortOrder := 0

	req := updateTrimmingCourseRequest{
		Name:        &name,
		Price:       &price,
		IsActive:    &isActive,
		Description: &description,
		TargetSize:  &targetSize,
		Duration:    &duration,
		SortOrder:   &sortOrder,
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
	if input.TargetSize != &targetSize {
		t.Fatalf("TargetSize pointer was not preserved")
	}
	if input.Duration != &duration {
		t.Fatalf("Duration pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateTrimmingCourseRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateTrimmingCourseRequest{}).toServiceInput()

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
	if input.TargetSize != nil {
		t.Fatalf("TargetSize = %v, want nil", input.TargetSize)
	}
	if input.Duration != nil {
		t.Fatalf("Duration = %v, want nil", input.Duration)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
}
