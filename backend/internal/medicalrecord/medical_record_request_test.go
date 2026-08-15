package medicalrecord

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestListMedicalRecordQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{
		PetID:     "10",
		OwnerID:   "20",
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.PetID == nil || *filters.PetID != 10 {
		t.Fatalf("PetID = %v, want 10", filters.PetID)
	}
	if filters.OwnerID == nil || *filters.OwnerID != 20 {
		t.Fatalf("OwnerID = %v, want 20", filters.OwnerID)
	}
	if filters.StartDate == nil || *filters.StartDate != "2026-05-01" {
		t.Fatalf("StartDate = %v, want 2026-05-01", filters.StartDate)
	}
	if filters.EndDate == nil || *filters.EndDate != "2026-05-31" {
		t.Fatalf("EndDate = %v, want 2026-05-31", filters.EndDate)
	}
}

func TestListMedicalRecordQuery_ToServiceFilters_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		query listMedicalRecordQuery
	}{
		{name: "pet_id", query: listMedicalRecordQuery{PetID: "abc"}},
		{name: "owner_id", query: listMedicalRecordQuery{OwnerID: "abc"}},
		{name: "start_date", query: listMedicalRecordQuery{StartDate: "2026/05/01"}},
		{name: "end_date", query: listMedicalRecordQuery{EndDate: "2026/05/31"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := tt.query.toServiceFilters()
			if err == nil {
				t.Fatal("toServiceFilters returned nil error")
			}
			if filters != (listMedicalRecordFilters{}) {
				t.Fatalf("filters = %#v, want zero value", filters)
			}
			if !apperrors.IsInvalidInput(err) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestCreateMedicalRecordRequest_ToServiceInput(t *testing.T) {
	ownerID := "10"
	petID := "20"
	doctorID := "30"
	appointmentID := "40"
	visitDate := "2026-05-28"
	nextVisitDate := "2026-06-28"
	reason := "checkup"
	req := createMedicalRecordRequest{
		RecordNo:                 "MR-1",
		VisitDate:                &visitDate,
		VisitType:                string(model.VisitTypeFirst),
		OwnerID:                  &ownerID,
		PetID:                    &petID,
		DoctorID:                 &doctorID,
		AppointmentID:            &appointmentID,
		Status:                   string(model.MedicalRecordStatusFinalized),
		NextVisitRecommendedDate: &nextVisitDate,
		RecommendationReason:     &reason,
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}
	if got := input.Date.Format("2006-01-02"); got != visitDate {
		t.Fatalf("Date = %q, want %q", got, visitDate)
	}
	if input.OwnerID == nil || *input.OwnerID != 10 {
		t.Fatalf("OwnerID = %v, want 10", input.OwnerID)
	}
	if input.PetID == nil || *input.PetID != 20 {
		t.Fatalf("PetID = %v, want 20", input.PetID)
	}
	if input.DoctorID == nil || *input.DoctorID != 30 {
		t.Fatalf("DoctorID = %v, want 30", input.DoctorID)
	}
	if input.AppointmentID == nil || *input.AppointmentID != 40 {
		t.Fatalf("AppointmentID = %v, want 40", input.AppointmentID)
	}
	if input.VisitType != model.VisitTypeFirst {
		t.Fatalf("VisitType = %v, want %q", input.VisitType, model.VisitTypeFirst)
	}
	if input.Status == nil || *input.Status != model.MedicalRecordStatusFinalized {
		t.Fatalf("Status = %v, want %q", input.Status, model.MedicalRecordStatusFinalized)
	}
	if input.NextVisitRecommendedDate == nil || input.NextVisitRecommendedDate.Format("2006-01-02") != nextVisitDate {
		t.Fatalf("NextVisitRecommendedDate = %v, want %q", input.NextVisitRecommendedDate, nextVisitDate)
	}
	if input.RecommendationReason == nil || *input.RecommendationReason != reason {
		t.Fatalf("RecommendationReason = %v, want %q", input.RecommendationReason, reason)
	}
	if input.EnteredBy == nil || *input.EnteredBy != 9 {
		t.Fatalf("EnteredBy = %v, want 9", input.EnteredBy)
	}
}

func TestCreateMedicalRecordRequest_ToServiceInput_InvalidOwnerID(t *testing.T) {
	ownerID := "abc"
	petID := "20"
	req := createMedicalRecordRequest{
		OwnerID: &ownerID,
		PetID:   &petID,
	}

	input, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input.EnteredBy != nil {
		t.Fatalf("input = %#v, want zero value", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateMedicalRecordRequest_ToServiceInput_NextVisitDateValidation(t *testing.T) {
	ownerID := "10"
	petID := "20"
	visitDate := "2026-05-28"
	nextVisitDate := "2026-05-28"
	req := createMedicalRecordRequest{
		OwnerID:                  &ownerID,
		PetID:                    &petID,
		VisitDate:                &visitDate,
		NextVisitRecommendedDate: &nextVisitDate,
	}

	input, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input.EnteredBy != nil {
		t.Fatalf("input = %#v, want zero value", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

func TestCreateMedicalRecordRequest_ToSubRecordsInput(t *testing.T) {
	complaint := "cough"
	plan := "observe"
	diagnosis1CategoryID := uint64(11)
	req := createMedicalRecordRequest{
		ChiefComplaint:       &complaint,
		Plan:                 &plan,
		Diagnosis1CategoryID: &diagnosis1CategoryID,
	}

	input := req.toSubRecordsInput()

	if input.ChiefComplaint == nil || *input.ChiefComplaint != complaint {
		t.Fatalf("ChiefComplaint = %v, want %q", input.ChiefComplaint, complaint)
	}
	if input.Plan == nil || *input.Plan != plan {
		t.Fatalf("Plan = %v, want %q", input.Plan, plan)
	}
	if input.Diagnosis1CategoryID == nil || *input.Diagnosis1CategoryID != diagnosis1CategoryID {
		t.Fatalf("Diagnosis1CategoryID = %v, want %d", input.Diagnosis1CategoryID, diagnosis1CategoryID)
	}
}

func TestUpdateMedicalRecordRequest_ToServiceInput(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	status := string(model.MedicalRecordStatusFinalized)
	nextVisitDate := "2026-06-28"
	version := 3
	req := updateMedicalRecordRequest{
		Date:                     &date,
		Status:                   &status,
		Version:                  &version,
		NextVisitRecommendedDate: &nextVisitDate,
	}

	input, err := req.toServiceInput(9)
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Date == nil || !input.Date.Equal(date) {
		t.Fatalf("Date = %v, want %v", input.Date, date)
	}
	if input.Status == nil || *input.Status != model.MedicalRecordStatusFinalized {
		t.Fatalf("Status = %v, want %q", input.Status, model.MedicalRecordStatusFinalized)
	}
	if input.Version == nil || *input.Version != version {
		t.Fatalf("Version = %v, want %d", input.Version, version)
	}
	if input.NextVisitRecommendedDate == nil || input.NextVisitRecommendedDate.Format("2006-01-02") != nextVisitDate {
		t.Fatalf("NextVisitRecommendedDate = %v, want %q", input.NextVisitRecommendedDate, nextVisitDate)
	}
	if input.ActorID == nil || *input.ActorID != 9 {
		t.Fatalf("ActorID = %v, want 9", input.ActorID)
	}
}

func TestUpdateMedicalRecordRequest_ToServiceInput_InvalidNextVisitDate(t *testing.T) {
	nextVisitDate := "2026/06/28"
	req := updateMedicalRecordRequest{NextVisitRecommendedDate: &nextVisitDate}

	input, err := req.toServiceInput(9)
	if err == nil {
		t.Fatal("toServiceInput returned nil error")
	}
	if input.ActorID != nil {
		t.Fatalf("input = %#v, want zero value", input)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

// TestResolveMedicalRecordSort は B-1 follow-up の列ソート server 化における
// sort/order query の許可リスト検証を網羅する。
func TestResolveMedicalRecordSort(t *testing.T) {
	tests := []struct {
		name      string
		sort      string
		order     string
		wantSort  string
		wantOrder string
	}{
		{name: "allowed key date + asc", sort: "date", order: "asc", wantSort: "date", wantOrder: "asc"},
		{name: "allowed key owner_name + desc", sort: "owner_name", order: "desc", wantSort: "owner_name", wantOrder: "desc"},
		{name: "allowed key pet_name, no order defaults to desc", sort: "pet_name", order: "", wantSort: "pet_name", wantOrder: "desc"},
		{name: "allowed key status, invalid order defaults to desc", sort: "status", order: "invalid", wantSort: "status", wantOrder: "desc"},
		{name: "unknown sort key falls back to empty (default order)", sort: "unknown_column", order: "asc", wantSort: "", wantOrder: ""},
		{name: "empty sort falls back to empty (default order)", sort: "", order: "asc", wantSort: "", wantOrder: ""},
		{name: "SQL-injection-shaped sort key is rejected", sort: "date; DROP TABLE medical_records;--", order: "asc", wantSort: "", wantOrder: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSort, gotOrder := resolveMedicalRecordSort(tt.sort, tt.order)
			if gotSort != tt.wantSort {
				t.Fatalf("sort = %q, want %q", gotSort, tt.wantSort)
			}
			if gotOrder != tt.wantOrder {
				t.Fatalf("order = %q, want %q", gotOrder, tt.wantOrder)
			}
		})
	}
}

// TestListMedicalRecordQuery_ToServiceFilters_Sort は toServiceFilters が sort/order を
// listMedicalRecordFilters へ正しく伝播することを検証する。
func TestListMedicalRecordQuery_ToServiceFilters_Sort(t *testing.T) {
	filters, err := (&listMedicalRecordQuery{Sort: "owner_name", Order: "asc"}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}
	if filters.Sort != "owner_name" {
		t.Fatalf("Sort = %q, want owner_name", filters.Sort)
	}
	if filters.Order != "asc" {
		t.Fatalf("Order = %q, want asc", filters.Order)
	}
}

func TestPatchMedicalRecordRecommendationReasonRequest_ToServiceInput(t *testing.T) {
	req := patchMedicalRecordRecommendationReasonRequest{Reason: "exam"}

	input := req.toServiceInput()

	if input.Reason != req.Reason {
		t.Fatalf("Reason = %q, want %q", input.Reason, req.Reason)
	}
}
