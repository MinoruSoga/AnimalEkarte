package handler

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

func TestUpdateLstepSettingsRequest_ToServiceInput(t *testing.T) {
	isSyncEnabled := false
	cpmVersion := "v2"
	noahLTV := int64(10000)
	lookbackDays := 30
	req := updateLstepSettingsRequest{
		LstepAPIKey:                  "key",
		IsSyncEnabled:                &isSyncEnabled,
		CPMVersion:                   &cpmVersion,
		CPMV1NoahLTV:                 &noahLTV,
		HealthPreventionLookbackDays: &lookbackDays,
	}

	input := req.toServiceInput()

	if input.LstepAPIKey != req.LstepAPIKey {
		t.Errorf("LstepAPIKey = %q, want %q", input.LstepAPIKey, req.LstepAPIKey)
	}
	if input.IsSyncEnabled == nil || *input.IsSyncEnabled {
		t.Errorf("IsSyncEnabled = %v, want false pointer", input.IsSyncEnabled)
	}
	if input.CPMVersion == nil || *input.CPMVersion != cpmVersion {
		t.Errorf("CPMVersion = %v, want %q", input.CPMVersion, cpmVersion)
	}
	if input.CPMV1NoahLTV == nil || *input.CPMV1NoahLTV != noahLTV {
		t.Errorf("CPMV1NoahLTV = %v, want %d", input.CPMV1NoahLTV, noahLTV)
	}
	if input.HealthPreventionLookbackDays == nil || *input.HealthPreventionLookbackDays != lookbackDays {
		t.Errorf("HealthPreventionLookbackDays = %v, want %d", input.HealthPreventionLookbackDays, lookbackDays)
	}
}
