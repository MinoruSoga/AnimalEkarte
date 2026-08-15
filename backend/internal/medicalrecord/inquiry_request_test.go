package medicalrecord

import "testing"

// TestUpdateInquiryRequest_ToServiceInput moved from
// internal/handler/simple_settings_request_test.go (BE9-2D — updateInquiryRequest /
// UpsertInquiryInput moved into this package).
func TestUpdateInquiryRequest_ToServiceInput(t *testing.T) {
	chiefComplaint := ""
	typeID := uint64(7)
	req := updateInquiryRequest{
		ChiefComplaint:       &chiefComplaint,
		ChiefComplaintTypeID: &typeID,
	}

	input := req.toServiceInput(1, 2)

	if input.ClinicID != 1 {
		t.Errorf("ClinicID = %d, want 1", input.ClinicID)
	}
	if input.MedicalRecordID != 2 {
		t.Errorf("MedicalRecordID = %d, want 2", input.MedicalRecordID)
	}
	if input.ChiefComplaint == nil || *input.ChiefComplaint != chiefComplaint {
		t.Errorf("ChiefComplaint = %v, want empty string pointer", input.ChiefComplaint)
	}
	if input.ChiefComplaintTypeID == nil || *input.ChiefComplaintTypeID != typeID {
		t.Errorf("ChiefComplaintTypeID = %v, want %d", input.ChiefComplaintTypeID, typeID)
	}
}
