package trimming

import "testing"

func TestCreateTrimmingCourseTypeRequest_ToServiceInput(t *testing.T) {
	req := createTrimmingCourseTypeRequest{Name: "フルコース", SortOrder: 3}

	input := req.toServiceInput()

	if input.Name != "フルコース" {
		t.Fatalf("Name = %q, want フルコース", input.Name)
	}
	if input.SortOrder != 3 {
		t.Fatalf("SortOrder = %d, want 3", input.SortOrder)
	}
}

func TestCreateTrimmingCourseTypeRequest_ToServiceInput_ZeroValue(t *testing.T) {
	input := (createTrimmingCourseTypeRequest{}).toServiceInput()

	if input.Name != "" {
		t.Fatalf("Name = %q, want empty", input.Name)
	}
	if input.SortOrder != 0 {
		t.Fatalf("SortOrder = %d, want 0", input.SortOrder)
	}
}

func TestUpdateTrimmingCourseTypeRequest_ToServiceInput(t *testing.T) {
	name := "更新後コース"
	sortOrder := 5
	isActive := false

	req := updateTrimmingCourseTypeRequest{
		Name:      &name,
		SortOrder: &sortOrder,
		IsActive:  &isActive,
	}

	input := req.toServiceInput()

	if input.Name != &name {
		t.Fatalf("Name pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
}

func TestUpdateTrimmingCourseTypeRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (updateTrimmingCourseTypeRequest{}).toServiceInput()

	if input.Name != nil {
		t.Fatalf("Name = %v, want nil", input.Name)
	}
	if input.SortOrder != nil {
		t.Fatalf("SortOrder = %v, want nil", input.SortOrder)
	}
	if input.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil", input.IsActive)
	}
}
