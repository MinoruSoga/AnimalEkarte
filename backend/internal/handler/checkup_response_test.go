package handler

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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

func TestToCheckupAlertItemResponse(t *testing.T) {
	today := time.Date(2026, 5, 28, 0, 0, 0, 0, time.Local)
	lastDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)

	t.Run("overdue: next_date before today", func(t *testing.T) {
		nextDate := today.AddDate(0, 0, -5)
		checkup := &model.Checkup{
			ID:          1,
			Date:        lastDate,
			NextDate:    &nextDate,
			CheckupType: &model.CheckupType{Name: "血液検査"},
			MedicalRecord: &model.MedicalRecord{
				Pet: &model.Pet{
					ID:   3,
					Name: "ポチ",
					Owner: &model.Owner{
						Name: "田中太郎",
					},
				},
			},
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.CheckupID != 1 {
			t.Errorf("CheckupID = %d, want 1", item.CheckupID)
		}
		if item.PetID != 3 {
			t.Errorf("PetID = %d, want 3", item.PetID)
		}
		if item.PetName != "ポチ" {
			t.Errorf("PetName = %q, want ポチ", item.PetName)
		}
		if item.OwnerName != "田中太郎" {
			t.Errorf("OwnerName = %q, want 田中太郎", item.OwnerName)
		}
		if item.CheckupTypeName != "血液検査" {
			t.Errorf("CheckupTypeName = %q, want 血液検査", item.CheckupTypeName)
		}
		if item.Days != 5 {
			t.Errorf("Days = %d, want 5 (overdue)", item.Days)
		}
	})

	t.Run("upcoming: next_date after today", func(t *testing.T) {
		nextDate := today.AddDate(0, 0, 7)
		checkup := &model.Checkup{
			ID:       2,
			Date:     lastDate,
			NextDate: &nextDate,
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.Days != 7 {
			t.Errorf("Days = %d, want 7 (upcoming)", item.Days)
		}
	})

	t.Run("nil NextDate leaves Days and NextDate zero-value", func(t *testing.T) {
		checkup := &model.Checkup{
			ID:   3,
			Date: lastDate,
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.NextDate != "" {
			t.Errorf("NextDate = %q, want empty", item.NextDate)
		}
		if item.Days != 0 {
			t.Errorf("Days = %d, want 0", item.Days)
		}
	})

	t.Run("nil MedicalRecord leaves pet/owner fields empty", func(t *testing.T) {
		checkup := &model.Checkup{
			ID:   4,
			Date: lastDate,
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.PetID != 0 {
			t.Errorf("PetID = %d, want 0", item.PetID)
		}
		if item.PetName != "" {
			t.Errorf("PetName = %q, want empty", item.PetName)
		}
		if item.OwnerName != "" {
			t.Errorf("OwnerName = %q, want empty", item.OwnerName)
		}
	})

	t.Run("nil Pet leaves pet/owner fields empty", func(t *testing.T) {
		checkup := &model.Checkup{
			ID:            5,
			Date:          lastDate,
			MedicalRecord: &model.MedicalRecord{},
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.PetID != 0 {
			t.Errorf("PetID = %d, want 0", item.PetID)
		}
		if item.PetName != "" {
			t.Errorf("PetName = %q, want empty", item.PetName)
		}
	})

	t.Run("nil Owner leaves owner name empty", func(t *testing.T) {
		checkup := &model.Checkup{
			ID:   6,
			Date: lastDate,
			MedicalRecord: &model.MedicalRecord{
				Pet: &model.Pet{ID: 3, Name: "ポチ"},
			},
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.PetName != "ポチ" {
			t.Errorf("PetName = %q, want ポチ", item.PetName)
		}
		if item.OwnerName != "" {
			t.Errorf("OwnerName = %q, want empty", item.OwnerName)
		}
	})

	t.Run("nil CheckupType leaves checkup type name empty", func(t *testing.T) {
		checkup := &model.Checkup{
			ID:   7,
			Date: lastDate,
		}

		item := toCheckupAlertItemResponse(checkup, today)

		if item.CheckupTypeName != "" {
			t.Errorf("CheckupTypeName = %q, want empty", item.CheckupTypeName)
		}
	})
}

func TestToCheckupAlertsResponse(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	result := &service.CheckupAlertsResult{
		Overdue:  []model.Checkup{{ID: 1, NextDate: &yesterday}},
		Upcoming: []model.Checkup{{ID: 2, NextDate: &tomorrow}},
	}

	resp := toCheckupAlertsResponse(result)

	if resp.Overdue.Count != 1 {
		t.Errorf("Overdue.Count = %d, want 1", resp.Overdue.Count)
	}
	if resp.Upcoming.Count != 1 {
		t.Errorf("Upcoming.Count = %d, want 1", resp.Upcoming.Count)
	}
	if len(resp.Overdue.Items) != 1 || resp.Overdue.Items[0].CheckupID != 1 {
		t.Errorf("Overdue.Items = %v, want CheckupID 1", resp.Overdue.Items)
	}
	if len(resp.Upcoming.Items) != 1 || resp.Upcoming.Items[0].CheckupID != 2 {
		t.Errorf("Upcoming.Items = %v, want CheckupID 2", resp.Upcoming.Items)
	}
}

func TestToCheckupAlertsResponse_Empty(t *testing.T) {
	result := &service.CheckupAlertsResult{}

	resp := toCheckupAlertsResponse(result)

	if resp.Overdue.Count != 0 {
		t.Errorf("Overdue.Count = %d, want 0", resp.Overdue.Count)
	}
	if resp.Upcoming.Count != 0 {
		t.Errorf("Upcoming.Count = %d, want 0", resp.Upcoming.Count)
	}
	if resp.Overdue.Items == nil {
		t.Error("Overdue.Items should be an empty slice, not nil")
	}
	if resp.Upcoming.Items == nil {
		t.Error("Upcoming.Items should be an empty slice, not nil")
	}
}
