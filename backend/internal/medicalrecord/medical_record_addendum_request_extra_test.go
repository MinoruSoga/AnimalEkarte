package medicalrecord

import (
	"testing"
)

func TestCreateMedicalRecordAddendumRequest_ToServiceInput(t *testing.T) {
	req := CreateMedicalRecordAddendumRequest{AfterText: "after", Reason: "reason"}

	input := req.toServiceInput(1, 2)

	if input.MedicalRecordID != 1 {
		t.Errorf("MedicalRecordID = %d, want 1", input.MedicalRecordID)
	}
	if input.AuthorUserID != 2 {
		t.Errorf("AuthorUserID = %d, want 2", input.AuthorUserID)
	}
	if input.AfterText != req.AfterText {
		t.Errorf("AfterText = %q, want %q", input.AfterText, req.AfterText)
	}
}
