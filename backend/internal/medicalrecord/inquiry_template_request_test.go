package medicalrecord

import "testing"

func TestCreateInquiryTemplateRequest_ToServiceInput(t *testing.T) {
	active := true
	req := createInquiryTemplateRequest{
		Category:  "first_visit",
		Title:     "Initial questionnaire",
		Content:   "Tell us about symptoms",
		IsActive:  &active,
		SortOrder: 2,
	}

	input := req.toServiceInput()

	if input.Category != req.Category {
		t.Fatalf("Category = %q, want %q", input.Category, req.Category)
	}
	if input.Title != req.Title {
		t.Fatalf("Title = %q, want %q", input.Title, req.Title)
	}
	if input.Content != req.Content {
		t.Fatalf("Content = %q, want %q", input.Content, req.Content)
	}
	if !input.IsActive {
		t.Fatalf("IsActive = %v, want true", input.IsActive)
	}
	if input.SortOrder != req.SortOrder {
		t.Fatalf("SortOrder = %d, want %d", input.SortOrder, req.SortOrder)
	}
}

func TestCreateInquiryTemplateRequest_IsActiveOmitted(t *testing.T) {
	req := createInquiryTemplateRequest{Category: "first_visit", Title: "title", Content: "c", SortOrder: 1}
	if req.IsActive != nil {
		t.Fatalf("omitted is_active must remain nil (presence absent), got %v", req.IsActive)
	}
	input := req.toServiceInput()
	if !input.IsActive {
		t.Fatalf("omitted is_active must resolve to true, got false")
	}
}

func TestCreateInquiryTemplateRequest_IsActiveFalse(t *testing.T) {
	active := false
	req := createInquiryTemplateRequest{Category: "first_visit", Title: "title", IsActive: &active}
	input := req.toServiceInput()
	if input.IsActive {
		t.Fatalf("explicit false must resolve to false")
	}
}

func TestUpdateInquiryTemplateRequest_ToServiceInput(t *testing.T) {
	category := ""
	title := ""
	content := ""
	isActive := false
	sortOrder := 0

	req := updateInquiryTemplateRequest{
		Category:  &category,
		Title:     &title,
		Content:   &content,
		IsActive:  &isActive,
		SortOrder: &sortOrder,
	}

	input := req.toServiceInput()

	if input.Category != &category {
		t.Fatalf("Category pointer was not preserved")
	}
	if input.Title != &title {
		t.Fatalf("Title pointer was not preserved")
	}
	if input.Content != &content {
		t.Fatalf("Content pointer was not preserved")
	}
	if input.IsActive != &isActive {
		t.Fatalf("IsActive pointer was not preserved")
	}
	if input.SortOrder != &sortOrder {
		t.Fatalf("SortOrder pointer was not preserved")
	}
}

func TestUpdateInquiryTemplateRequest_ToServiceInput_NilFields(t *testing.T) {
	input := (&updateInquiryTemplateRequest{}).toServiceInput()

	if input.Category != nil || input.Title != nil || input.Content != nil || input.IsActive != nil || input.SortOrder != nil {
		t.Fatalf("input = %+v, want all nil fields", input)
	}
}
