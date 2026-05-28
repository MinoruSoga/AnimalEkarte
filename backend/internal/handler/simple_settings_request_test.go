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

func TestUpsertLineReservationSettingRequest_ToServiceInput(t *testing.T) {
	dailyLimit := 0
	closedWeekdays := jsonRawOrEmpty(`[1,2]`)
	req := upsertLineReservationSettingRequest{
		Status:                  "active",
		ClosedWeekdays:          closedWeekdays,
		DailyLimit:              &dailyLimit,
		BookingWindowMaxDays:    30,
		TimeSlotIntervalMinutes: 15,
		ShowNoStaffOption:       true,
		LineChannelSecret:       "secret",
	}

	input := req.toServiceInput()

	if input.Status != req.Status {
		t.Errorf("Status = %q, want %q", input.Status, req.Status)
	}
	if string(input.ClosedWeekdays) != string(closedWeekdays) {
		t.Errorf("ClosedWeekdays = %s, want %s", input.ClosedWeekdays, closedWeekdays)
	}
	if input.DailyLimit == nil || *input.DailyLimit != dailyLimit {
		t.Errorf("DailyLimit = %v, want %d", input.DailyLimit, dailyLimit)
	}
	if !input.ShowNoStaffOption {
		t.Error("ShowNoStaffOption = false, want true")
	}
	if input.LineChannelSecret != req.LineChannelSecret {
		t.Errorf("LineChannelSecret = %q, want %q", input.LineChannelSecret, req.LineChannelSecret)
	}
}
