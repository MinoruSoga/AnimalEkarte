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

// BUG-010: JSON 空文字は明示クリア（pointer to ""）、フィールド欠落は未送信（nil）として区別する。
func TestUpdateClinicalPlanRequest_EmptyStringVsOmitted(t *testing.T) {
	empty := ""
	withEmpty := updateClinicalPlanRequest{
		PhysicalExam:     &empty,
		DiagnosisDetails: &empty,
		TreatmentPolicy:  &empty,
	}
	inputEmpty := withEmpty.toServiceInput()
	if inputEmpty.PhysicalExam == nil || *inputEmpty.PhysicalExam != "" {
		t.Fatalf("empty PhysicalExam pointer = %v, want &\"\"", inputEmpty.PhysicalExam)
	}
	if inputEmpty.DiagnosisDetails == nil || *inputEmpty.DiagnosisDetails != "" {
		t.Fatalf("empty DiagnosisDetails pointer = %v, want &\"\"", inputEmpty.DiagnosisDetails)
	}
	if inputEmpty.TreatmentPolicy == nil || *inputEmpty.TreatmentPolicy != "" {
		t.Fatalf("empty TreatmentPolicy pointer = %v, want &\"\"", inputEmpty.TreatmentPolicy)
	}

	omitted := updateClinicalPlanRequest{}
	inputOmitted := omitted.toServiceInput()
	if inputOmitted.PhysicalExam != nil {
		t.Fatalf("omitted PhysicalExam = %v, want nil", inputOmitted.PhysicalExam)
	}
	if inputOmitted.DiagnosisDetails != nil {
		t.Fatalf("omitted DiagnosisDetails = %v, want nil", inputOmitted.DiagnosisDetails)
	}
	if inputOmitted.TreatmentPolicy != nil {
		t.Fatalf("omitted TreatmentPolicy = %v, want nil", inputOmitted.TreatmentPolicy)
	}

	fieldsEmpty := buildClinicalPlanUpdate(inputEmpty)
	if fieldsEmpty["physical_exam"] != "" || fieldsEmpty["diagnosis_details"] != "" || fieldsEmpty["treatment_policy"] != "" {
		t.Fatalf("empty clear fields = %#v, want empty strings for all three text columns", fieldsEmpty)
	}
	fieldsOmitted := buildClinicalPlanUpdate(inputOmitted)
	if len(fieldsOmitted) != 0 {
		t.Fatalf("omitted fields = %#v, want empty map (no-op update)", fieldsOmitted)
	}
}
