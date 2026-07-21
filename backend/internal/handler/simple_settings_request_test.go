package handler

import "testing"

func TestUpdateCompanyRequest_ToServiceInput(t *testing.T) {
	name := ""
	email := "clinic@example.com"
	req := updateCompanyRequest{Name: &name, Email: &email}

	input := req.toServiceInput()

	if input.Name == nil || *input.Name != name {
		t.Errorf("Name = %v, want empty string pointer", input.Name)
	}
	if input.Email == nil || *input.Email != email {
		t.Errorf("Email = %v, want %q", input.Email, email)
	}
}

// TestUpdateClinicalPlanRequest_ToServiceInput moved to
// internal/medicalrecord/clinical_plan_request_test.go (BE9-2D sub-batch④a — updateClinicalPlanRequest moved there).

// TestUpdateInquiryRequest_ToServiceInput moved to
// internal/medicalrecord/inquiry_request_test.go (BE9-2D — updateInquiryRequest moved there).
