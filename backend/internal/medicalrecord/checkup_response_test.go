package medicalrecord

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToCheckupResponse(t *testing.T) {
	petID := uint64(3)
	doctorID := uint64(4)
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	nextDate := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		PetID:           &petID,
		Date:            date,
		NextDate:        &nextDate,
		DoctorID:        &doctorID,
		Result:          "normal",
		CreatedAt:       date,
		UpdatedAt:       date,
		CheckupType:     &model.CheckupType{ID: 5, Name: "血液検査"},
		Doctor:          &model.Staff{ID: 4, Name: "山田先生"},
	}

	resp := toCheckupResponse(checkup)

	if resp.ID != "1" {
		t.Errorf("ID = %q, want 1", resp.ID)
	}
	if resp.MedicalRecordID != "2" {
		t.Errorf("MedicalRecordID = %q, want 2", resp.MedicalRecordID)
	}
	if resp.CheckupTypeID != "5" {
		t.Errorf("CheckupTypeID = %q, want 5", resp.CheckupTypeID)
	}
	if resp.PetID == nil || *resp.PetID != "3" {
		t.Errorf("PetID = %v, want 3", resp.PetID)
	}
	if resp.Date != "2026-05-28" {
		t.Errorf("Date = %q, want 2026-05-28", resp.Date)
	}
	if resp.NextDate == nil || *resp.NextDate != "2026-11-28" {
		t.Errorf("NextDate = %v, want 2026-11-28", resp.NextDate)
	}
	if resp.DoctorID == nil || *resp.DoctorID != "4" {
		t.Errorf("DoctorID = %v, want 4", resp.DoctorID)
	}
	if resp.Result != "normal" {
		t.Errorf("Result = %q, want normal", resp.Result)
	}
	if resp.CheckupType == nil || resp.CheckupType.Name != "血液検査" {
		t.Errorf("CheckupType = %v, want 血液検査", resp.CheckupType)
	}
	if resp.Doctor == nil || resp.Doctor.Name != "山田先生" {
		t.Errorf("Doctor = %v, want 山田先生", resp.Doctor)
	}
}

func TestToCheckupResponse_NilOptionalFields(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		Date:            date,
		Result:          "",
		CreatedAt:       date,
		UpdatedAt:       date,
	}

	resp := toCheckupResponse(checkup)

	if resp.PetID != nil {
		t.Errorf("PetID = %v, want nil", resp.PetID)
	}
	if resp.NextDate != nil {
		t.Errorf("NextDate = %v, want nil", resp.NextDate)
	}
	if resp.DoctorID != nil {
		t.Errorf("DoctorID = %v, want nil", resp.DoctorID)
	}
	if resp.CheckupType != nil {
		t.Errorf("CheckupType = %v, want nil", resp.CheckupType)
	}
	if resp.Doctor != nil {
		t.Errorf("Doctor = %v, want nil", resp.Doctor)
	}
}

func TestToCheckupGlobalResponse(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	ownerID := uint64(9)
	petID := uint64(3)
	doctorID := uint64(4)
	nextDate := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		PetID:           &petID,
		Date:            date,
		NextDate:        &nextDate,
		DoctorID:        &doctorID,
		Result:          "normal",
		CheckupType:     &model.CheckupType{ID: 5, Name: "血液検査"},
		Doctor:          &model.Staff{ID: 4, Name: "山田先生"},
		MedicalRecord: &model.MedicalRecord{
			ID: 2,
			Pet: &model.Pet{
				ID:   3,
				Name: "ポチ",
				Owner: &model.Owner{
					ID:   ownerID,
					Name: "田中太郎",
				},
			},
		},
	}

	resp := toCheckupGlobalResponse(checkup)

	if resp.PetID == nil || *resp.PetID != "3" {
		t.Errorf("PetID = %v, want 3", resp.PetID)
	}
	if resp.NextDate == nil || *resp.NextDate != "2026-11-28" {
		t.Errorf("NextDate = %v, want 2026-11-28", resp.NextDate)
	}
	if resp.DoctorID == nil || *resp.DoctorID != "4" {
		t.Errorf("DoctorID = %v, want 4", resp.DoctorID)
	}
	if resp.PetName != "ポチ" {
		t.Errorf("PetName = %q, want ポチ", resp.PetName)
	}
	if resp.OwnerName != "田中太郎" {
		t.Errorf("OwnerName = %q, want 田中太郎", resp.OwnerName)
	}
	if resp.OwnerID == nil || *resp.OwnerID != "9" {
		t.Errorf("OwnerID = %v, want 9", resp.OwnerID)
	}
	if resp.CheckupType == nil || resp.CheckupType.Name != "血液検査" {
		t.Errorf("CheckupType = %v, want 血液検査", resp.CheckupType)
	}
	if resp.Doctor == nil || resp.Doctor.Name != "山田先生" {
		t.Errorf("Doctor = %v, want 山田先生", resp.Doctor)
	}
}

func TestToCheckupGlobalResponse_NilMedicalRecord(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		Date:            date,
		Result:          "normal",
	}

	resp := toCheckupGlobalResponse(checkup)

	if resp.PetName != "" {
		t.Errorf("PetName = %q, want empty", resp.PetName)
	}
	if resp.OwnerName != "" {
		t.Errorf("OwnerName = %q, want empty", resp.OwnerName)
	}
	if resp.OwnerID != nil {
		t.Errorf("OwnerID = %v, want nil", resp.OwnerID)
	}
}

func TestToCheckupGlobalResponse_NilPet(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		Date:            date,
		Result:          "normal",
		MedicalRecord:   &model.MedicalRecord{ID: 2},
	}

	resp := toCheckupGlobalResponse(checkup)

	if resp.PetName != "" {
		t.Errorf("PetName = %q, want empty", resp.PetName)
	}
	if resp.OwnerName != "" {
		t.Errorf("OwnerName = %q, want empty", resp.OwnerName)
	}
}

func TestToCheckupGlobalResponse_NilOwner(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	checkup := &model.Checkup{
		ID:              1,
		MedicalRecordID: 2,
		CheckupTypeID:   5,
		Date:            date,
		Result:          "normal",
		MedicalRecord: &model.MedicalRecord{
			ID:  2,
			Pet: &model.Pet{ID: 3, Name: "ポチ"},
		},
	}

	resp := toCheckupGlobalResponse(checkup)

	if resp.PetName != "ポチ" {
		t.Errorf("PetName = %q, want ポチ", resp.PetName)
	}
	if resp.OwnerName != "" {
		t.Errorf("OwnerName = %q, want empty", resp.OwnerName)
	}
	if resp.OwnerID != nil {
		t.Errorf("OwnerID = %v, want nil", resp.OwnerID)
	}
}
