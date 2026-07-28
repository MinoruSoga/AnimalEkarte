package medicalrecord

import "testing"

// Moved from internal/handler/simple_settings_request_test.go in BE9-2D sub-batch④a
// alongside updateClinicalPlanRequest.
func TestUpdateClinicalPlanRequest_ToServiceInput(t *testing.T) {
	physicalExam := ""
	diagnosisTypeID := uint64(0)
	req := updateClinicalPlanRequest{
		PhysicalExam:    &physicalExam,
		DiagnosisTypeID: &diagnosisTypeID,
	}

	input := req.toServiceInput()

	if input.PhysicalExam == nil || *input.PhysicalExam != physicalExam {
		t.Errorf("PhysicalExam = %v, want empty string pointer", input.PhysicalExam)
	}
	if input.DiagnosisTypeID == nil || *input.DiagnosisTypeID != diagnosisTypeID {
		t.Errorf("DiagnosisTypeID = %v, want %d", input.DiagnosisTypeID, diagnosisTypeID)
	}
}
